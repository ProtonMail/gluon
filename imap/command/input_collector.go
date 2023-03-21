package command

import (
	"github.com/ProtonMail/gluon/rfcparser"
)

type InputCollector struct {
	source rfcparser.Reader
	bytes  []byte
}

func NewInputCollector(source rfcparser.Reader) *InputCollector {
	return &InputCollector{
		source: source,
		bytes:  make([]byte, 0, 128),
	}
}

func (i *InputCollector) Bytes() []byte {
	return i.bytes
}

func (i *InputCollector) Read(dst []byte) (int, error) {
	n, err := i.source.Read(dst)
	if err == nil {
		i.bytes = append(i.bytes, dst[0:n]...)
	}

	return n, err
}

func (i *InputCollector) ReadByte() (byte, error) {
	b, err := i.source.ReadByte()
	if err == nil {
		i.bytes = append(i.bytes, b)
	}

	return b, err
}

func (i *InputCollector) ReadBytes(delim byte) ([]byte, error) {
	b, err := i.source.ReadBytes(delim)
	if err == nil {
		i.bytes = append(i.bytes, b...)
	}

	return b, err
}

func (i *InputCollector) Reset() {
	i.bytes = i.bytes[:0]
}

func (i *InputCollector) SetSource(source rfcparser.Reader) {
	i.source = source
}
