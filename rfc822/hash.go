package rfc822

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"io"
	"mime/quotedprintable"
	"strings"

	"github.com/sirupsen/logrus"
)

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

	if _, err := h.Write([]byte(header.Get("From"))); err != nil {
		return "", err
	}

	if _, err := h.Write([]byte(header.Get("To"))); err != nil {
		return "", err
	}

	if _, err := h.Write([]byte(header.Get("Cc"))); err != nil {
		return "", err
	}

	if _, err := h.Write([]byte(header.Get("Reply-To"))); err != nil {
		return "", err
	}

	if _, err := h.Write([]byte(header.Get("In-Reply-To"))); err != nil {
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
		} else {
			if _, err := h.Write([]byte(mimeType)); err != nil {
				return err
			}

			keys := maps.Keys(values)
			slices.Sort(keys)

			for _, k := range keys {
				if strings.EqualFold(k, "boundary") {
					continue
				}

				if _, err := h.Write([]byte(k)); err != nil {
					return err
				}

				if _, err := h.Write([]byte(values[k])); err != nil {
					return err
				}
			}
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
