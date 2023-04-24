package rfc822

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
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

		if _, err := h.Write([]byte(header.Get("Content-Type"))); err != nil {
			return err
		}

		if _, err := h.Write([]byte(header.Get("Content-Disposition"))); err != nil {
			return err
		}

		body := section.Body()
		body = bytes.ReplaceAll(body, []byte{'\r'}, nil)
		body = bytes.TrimSpace(body)
		if _, err := h.Write(body); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(h.Sum(nil)), nil
}
