package rfc822

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMultipartWriter(t *testing.T) {
	b := new(bytes.Buffer)

	w := NewMultipartWriter(b, "boundary")

	require.NoError(t, w.AddPart(func(w io.Writer) error {
		if _, err := fmt.Fprintf(w, "1"); err != nil {
			return err
		}

		return nil
	}))

	require.NoError(t, w.AddPart(func(w io.Writer) error {
		if _, err := fmt.Fprintf(w, "2"); err != nil {
			return err
		}

		return nil
	}))

	require.NoError(t, w.AddPart(func(w io.Writer) error {
		if _, err := fmt.Fprintf(w, "3"); err != nil {
			return err
		}

		return nil
	}))

	require.NoError(t, w.Done())

	assert.Equal(t, "--boundary\r\n1\r\n--boundary\r\n2\r\n--boundary\r\n3\r\n--boundary--\r\n", b.String())
}
