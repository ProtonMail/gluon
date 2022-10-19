package rfc822

import (
	"mime"
	"strings"
)

// ParseMediaType parses a MIME media type.
var ParseMediaType = mime.ParseMediaType

type MIMEType string

const (
	TextPlain        MIMEType = "text/plain"
	TextHTML         MIMEType = "text/html"
	MultipartMixed   MIMEType = "multipart/mixed"
	MultipartRelated MIMEType = "multipart/related"
	MessageRFC822    MIMEType = "message/rfc822"
)

func (mimeType MIMEType) IsMultiPart() bool {
	return strings.HasPrefix(string(mimeType), "multipart/")
}

func (mimeType MIMEType) Type() string {
	if split := strings.SplitN(string(mimeType), "/", 2); len(split) == 2 {
		return split[0]
	}

	return ""
}

func (mimeType MIMEType) SubType() string {
	if split := strings.SplitN(string(mimeType), "/", 2); len(split) == 2 {
		return split[1]
	}

	return ""
}

func ParseMIMEType(val string) (MIMEType, map[string]string, error) {
	if val == "" {
		val = string(TextPlain)
	}

	mimeType, mimeParams, err := ParseMediaType(val)
	if err != nil {
		return "", nil, err
	}

	return MIMEType(mimeType), mimeParams, nil
}
