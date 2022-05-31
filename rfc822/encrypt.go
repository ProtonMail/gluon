package rfc822

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/quotedprintable"
	"strings"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"golang.org/x/text/encoding/htmlindex"
	"golang.org/x/text/encoding/ianaindex"
)

func Encrypt(kr *crypto.KeyRing, r io.Reader) ([]byte, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	header, body, err := Split(b)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)

	result, err := writeEncryptedPart(kr, ParseHeader(header), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	if _, err := buf.Write(header); err != nil {
		return nil, err
	}

	if _, err := result.WriteTo(buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func writeEncryptedPart(kr *crypto.KeyRing, header *Header, r io.Reader) (io.WriterTo, error) {
	decoder := getTransferDecoder(r, string(header.Get("Content-Transfer-Encoding")))
	encoded := new(bytes.Buffer)

	contentType, contentParams, err := ParseContentType(header.Get("Content-Type"))
	if err != nil && !errors.Is(err, mime.ErrInvalidMediaParameter) {
		return nil, err
	}

	switch {
	case contentType == "", strings.HasPrefix(contentType, "text/"), strings.HasPrefix(contentType, "message/"):
		header.Del("Content-Transfer-Encoding")

		if charset, ok := contentParams["charset"]; ok {
			decoder = getCharsetDecoder(decoder, charset)

			contentParams["charset"] = "utf-8"

			header.Set("Content-Type", mime.FormatMediaType(contentType, contentParams))
		}

		if err := encode(&writeCloser{encoded}, func(w io.Writer) error {
			return writeEncryptedTextPart(w, decoder, kr)
		}); err != nil {
			return nil, err
		}

	case contentType == "multipart/encrypted":
		if _, err := encoded.ReadFrom(decoder); err != nil {
			return nil, err
		}

	case strings.HasPrefix(contentType, "multipart/"):
		if err := encode(&writeCloser{encoded}, func(w io.Writer) error {
			return writeEncryptedMultiPart(kr, w, header, decoder)
		}); err != nil {
			return nil, err
		}

	default:
		header.Set("Content-Transfer-Encoding", "base64")

		if err := encode(base64.NewEncoder(base64.StdEncoding, encoded), func(w io.Writer) error {
			return writeEncryptedAttachmentPart(w, decoder, kr)
		}); err != nil {
			return nil, err
		}
	}

	return encoded, nil
}

func writeEncryptedTextPart(w io.Writer, r io.Reader, kr *crypto.KeyRing) error {
	dec, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	var arm string

	if msg, err := crypto.NewPGPMessageFromArmored(string(dec)); err != nil {
		enc, err := kr.Encrypt(crypto.NewPlainMessage(dec), kr)
		if err != nil {
			return err
		}

		if arm, err = enc.GetArmored(); err != nil {
			return err
		}
	} else {
		if arm, err = msg.GetArmored(); err != nil {
			return err
		}
	}

	if _, err := io.WriteString(w, arm); err != nil {
		return err
	}

	return nil
}

func writeEncryptedAttachmentPart(w io.Writer, r io.Reader, kr *crypto.KeyRing) error {
	dec, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	enc, err := kr.Encrypt(crypto.NewPlainMessage(dec), kr)
	if err != nil {
		return err
	}

	if _, err := w.Write(enc.GetBinary()); err != nil {
		return err
	}

	return nil
}

func writeEncryptedMultiPart(kr *crypto.KeyRing, w io.Writer, header *Header, r io.Reader) error {
	_, contentParams, err := ParseContentType(header.Get("Content-Type"))
	if err != nil {
		return err
	}

	scanner, err := NewScanner(r, contentParams["boundary"])
	if err != nil {
		return err
	}

	parts, err := scanner.ScanAll()
	if err != nil {
		return err
	}

	multipartWriter := NewMultipartWriter(w, contentParams["boundary"])

	for _, part := range parts {
		header, body, err := Split(part.Data)
		if err != nil {
			return err
		}

		result, err := writeEncryptedPart(kr, ParseHeader(header), bytes.NewReader(body))
		if err != nil {
			return err
		}

		if err := multipartWriter.AddPart(func(w io.Writer) error {
			if _, err := w.Write(header); err != nil {
				return err
			}

			if _, err := result.WriteTo(w); err != nil {
				return err
			}

			return nil
		}); err != nil {
			return err
		}
	}

	return multipartWriter.Done()
}

func getTransferDecoder(r io.Reader, encoding string) io.Reader {
	switch strings.ToLower(encoding) {
	case "base64":
		return base64.NewDecoder(base64.StdEncoding, r)

	case "quoted-printable":
		return quotedprintable.NewReader(r)

	default:
		return r
	}
}

func getCharsetDecoder(r io.Reader, charset string) io.Reader {
	if enc, err := ianaindex.MIME.Encoding(strings.ToLower(charset)); err == nil {
		return enc.NewDecoder().Reader(r)
	}

	if enc, err := ianaindex.MIME.Encoding("cs" + strings.ToLower(charset)); err == nil {
		return enc.NewDecoder().Reader(r)
	}

	if enc, err := htmlindex.Get(strings.ToLower(charset)); err == nil {
		return enc.NewDecoder().Reader(r)
	}

	panic(fmt.Errorf("unsupported charset: %v", charset))
}

func encode(wc io.WriteCloser, fn func(io.Writer) error) error {
	if err := fn(wc); err != nil {
		return err
	}

	return wc.Close()
}

type writeCloser struct {
	io.Writer
}

func (writeCloser) Close() error { return nil }
