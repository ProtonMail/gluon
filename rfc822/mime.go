package rfc822

import "mime"

type MIMEType string

const (
	TextPlain        MIMEType = "text/plain"
	TextHTML         MIMEType = "text/html"
	MultipartMixed   MIMEType = "multipart/mixed"
	MultipartRelated MIMEType = "multipart/related"
	MessageRFC822    MIMEType = "message/rfc822"
)

func ParseContentType(val string) (string, map[string]string, error) {
	if val == "" {
		val = string(TextPlain)
	}

	return mime.ParseMediaType(val)
}
