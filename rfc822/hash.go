package rfc822

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"mime/quotedprintable"
	"slices"
	"strconv"
	"strings"

	pkgutils "github.com/ProtonMail/gluon/pkg/utils"
	"github.com/ProtonMail/gluon/rfc5322"
	"github.com/sirupsen/logrus"
)

// gluonInternalHeaderKey mirrors internal/ids.InternalIDKey.
const gluonInternalHeaderKey = "X-Pm-Gluon-Id"

// GetMessageHash returns the hash of the given message.
// This takes into account:
// - the Subject header,
// - the From/To/Cc headers,
// - the Content-Type header of each (leaf) part,
// - the Content-Disposition header of each (leaf) part,
// - the (decoded) body of each part.
func GetMessageHash(b []byte) (string, error) {
	section := Parse(b)

	header, err := section.ParseHeader()
	if err != nil {
		return "", err
	}

	h := sha256.New()

	if _, err := h.Write([]byte(header.Get("Subject"))); err != nil {
		return "", err
	}

	if _, err := h.Write([]byte(getAddresses(header.Get("From")))); err != nil {
		return "", err
	}

	if _, err := h.Write([]byte(getAddresses(header.Get("To")))); err != nil {
		return "", err
	}

	if _, err := h.Write([]byte(getAddresses(header.Get("Cc")))); err != nil {
		return "", err
	}

	if _, err := h.Write([]byte(getAddresses(header.Get("Reply-To")))); err != nil {
		return "", err
	}

	if _, err := h.Write([]byte(getAddresses(header.Get("In-Reply-To")))); err != nil {
		return "", err
	}

	if err := section.Walk(func(section *Section) error {
		children, err := section.Children()
		if err != nil {
			return err
		} else if len(children) > 0 {
			return nil
		}

		header, err := section.ParseHeader()
		if err != nil {
			return err
		}

		contentType := header.Get("Content-Type")
		mimeType, values, err := ParseMIMEType(contentType)
		if err != nil {
			logrus.Warnf("Message contains invalid mime type: %v", contentType)
		} else if err := writeContentTypeParams(h, mimeType, values); err != nil {
			return err
		}

		if _, err := h.Write([]byte(header.Get("Content-Disposition"))); err != nil {
			return err
		}

		body := section.Body()
		if err := hashBody(h, body, mimeType, header.Get("Content-Transfer-Encoding")); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(h.Sum(nil)), nil
}

// LiteralsStructurallyEqual returns whether two messages literals are structurally equivalent.
// It's based on `GetMessageHash` which skips the Content-Type boundary parameter.
func LiteralsStructurallyEqual(a, b []byte) (bool, error) {
	ha, err := GetMessageHash(a)
	if err != nil {
		return false, err
	}

	hb, err := GetMessageHash(b)
	if err != nil {
		return false, err
	}

	return ha == hb, nil
}

// writeContentTypeParams writes the MIME type and its parameters (excluding the boundary parameter) to the writer.
// It is shared by the content hash and the header fingerprint so both ignore MIME boundary drift identically.
func writeContentTypeParams(w io.Writer, mimeType MIMEType, values map[string]string) error {
	if _, err := io.WriteString(w, string(mimeType)); err != nil {
		return err
	}

	keys := pkgutils.Keys(values)
	slices.Sort(keys)

	for _, k := range keys {
		if strings.EqualFold(k, "boundary") {
			continue
		}

		if _, err := io.WriteString(w, k); err != nil {
			return err
		}

		if _, err := io.WriteString(w, values[k]); err != nil {
			return err
		}
	}

	return nil
}

// GetMessageHeaderFingerprint returns a hash of the message's headers across the entire
// MIME tree. Unlike GetMessageHash, this captures every header - including Date, Message-Id and
// X-Pm-*.
//
// It is boundary-neutral: the Content-Type boundary parameter is excluded so pure MIME
// boundary drift does not change the fingerprint. The Gluon-internal X-Pm-Gluon-Id
// header is skipped because it is injected by the mailserver. Known address headers are
// normalized the same way as the content hash so display/folding differences do not
// cause false mismatches. Header ordering is ignored and bodies are not included.
//
// Note: Content-Transfer-Encoding is included per part, so a CTE-only rebuild (same
// decoded body, different encoding) will differ here even though GetMessageHash ignores
// it. This is intentional for wire-metadata comparison.
func GetMessageHeaderFingerprint(b []byte) (string, error) {
	section := Parse(b)

	h := sha256.New()

	if err := section.Walk(func(section *Section) error {
		header, err := section.ParseHeader()
		if err != nil {
			return err
		}

		// Section separator + MIME path so identical headers on different parts do not merge.
		if _, err := io.WriteString(h, "\x01"+sectionPath(section)+"\x02"); err != nil {
			return err
		}

		var entries []string

		header.Entries(func(key, val string) {
			if strings.EqualFold(key, gluonInternalHeaderKey) {
				return
			}

			entries = append(entries, canonicalHeaderEntry(key, val))
		})

		// Sort so header ordering does not affect the fingerprint.
		slices.Sort(entries)

		for _, e := range entries {
			if _, err := io.WriteString(h, e); err != nil {
				return err
			}

			if _, err := h.Write([]byte{0}); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(h.Sum(nil)), nil
}

// canonicalHeaderEntry renders a single header entry into a deterministic, comparable
// form: Content-Type has its boundary parameter stripped, address headers are normalized
// to their addresses only, and all other headers keep their merged value verbatim.
func canonicalHeaderEntry(key, val string) string {
	lkey := strings.ToLower(key)

	switch lkey {
	case "content-type":
		if mimeType, values, err := ParseMIMEType(val); err == nil {
			var sb strings.Builder

			sb.WriteString(lkey)
			sb.WriteByte(':')

			// writeContentTypeParams only errors on writer failure; strings.Builder never fails.
			_ = writeContentTypeParams(&sb, mimeType, values)

			return sb.String()
		}

		return lkey + ":" + val

	case "from", "to", "cc", "reply-to", "in-reply-to":
		return lkey + ":" + getAddresses(val)

	default:
		return lkey + ":" + val
	}
}

// sectionPath returns the dotted MIME path of a section; the root section
// has an empty identifier and returns "".
func sectionPath(section *Section) string {
	identifier := section.Identifier()

	parts := make([]string, len(identifier))
	for i, id := range identifier {
		parts[i] = strconv.Itoa(id)
	}

	return strings.Join(parts, ".")
}

// LiteralsHeadersEqual reports whether two message literals have equivalent headers
// across the MIME tree, ignoring MIME boundary drift and the Gluon-internal header.
// See GetMessageHeaderFingerprint.
func LiteralsHeadersEqual(a, b []byte) (bool, error) {
	fa, err := GetMessageHeaderFingerprint(a)
	if err != nil {
		return false, err
	}

	fb, err := GetMessageHeaderFingerprint(b)
	if err != nil {
		return false, err
	}

	return fa == fb, nil
}

// CompareLiterals compares two message literals and returns true if they are structurally and header-wise equivalent.
func CompareLiterals(a, b []byte) (bool, error) {
	structEq, err := LiteralsStructurallyEqual(a, b)
	if err != nil {
		return false, err
	}

	if !structEq {
		return false, nil
	}

	return LiteralsHeadersEqual(a, b)
}

func hashBody(writer io.Writer, body []byte, mimeType MIMEType, encoding string) error {
	if mimeType != TextHTML && mimeType != TextPlain {
		body = bytes.ReplaceAll(body, []byte{'\r'}, nil)
		body = bytes.TrimSpace(body)
		_, err := writer.Write(body)

		return err
	}

	// We need to remove the transfer encoding from the text part as it is possible the that encoding sent to SMTP
	// is different than the one sent to the IMAP client.
	var decoded []byte

	switch strings.ToLower(encoding) {
	case "quoted-printable":
		d, err := io.ReadAll(quotedprintable.NewReader(bytes.NewReader(body)))
		if err != nil {
			return err
		}

		decoded = d

	case "base64":
		d, err := io.ReadAll(base64.NewDecoder(base64.StdEncoding, bytes.NewReader(body)))
		if err != nil {
			return err
		}

		decoded = d

	default:
		decoded = body
	}

	decoded = bytes.ReplaceAll(decoded, []byte{'\r'}, nil)
	decoded = bytes.TrimSpace(decoded)

	_, err := writer.Write(decoded)

	return err
}

func getAddresses(fieldAddr string) string {
	var addresses strings.Builder

	addrList, err := rfc5322.ParseAddressList(fieldAddr)
	if err != nil {
		return fieldAddr
	}

	for _, addr := range addrList {
		addresses.WriteString(addr.Address)
	}

	return addresses.String()
}
