package conn

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMultiListener(t *testing.T) {
	cert, err := newCert()
	require.NoError(t, err)

	l, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	// Wrap the listener in a MultiListener.
	l = NewListener(l, &tls.Config{Certificates: []tls.Certificate{cert}})

	resCh := make(chan []byte)

	go func() {
		c, err := l.Accept()
		require.NoError(t, err)

		b, err := io.ReadAll(c)
		require.NoError(t, err)

		resCh <- b
	}()

	// Dial the listener with TLS.
	c, err := net.Dial("tcp", l.Addr().String())
	require.NoError(t, err)

	// Write some data.
	n, err := c.Write([]byte("hello"))
	require.NoError(t, err)
	require.Equal(t, 5, n)
	require.NoError(t, c.Close())

	// Read the data back.
	require.Equal(t, "hello", string(<-resCh))
}

func newCert() (tls.Certificate, error) {
	num, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return tls.Certificate{}, err
	}

	certPEM, keyPEM, err := newRawCert(&x509.Certificate{
		SerialNumber:          num,
		IsCA:                  true,
		BasicConstraintsValid: true,
	})
	if err != nil {
		return tls.Certificate{}, err
	}

	return tls.X509KeyPair(certPEM, keyPEM)
}

func newRawCert(template *x509.Certificate) ([]byte, []byte, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, template, template, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	certPEM := new(bytes.Buffer)

	if err := pem.Encode(certPEM, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return nil, nil, err
	}

	keyPEM := new(bytes.Buffer)

	if err := pem.Encode(keyPEM, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)}); err != nil {
		return nil, nil, err
	}

	return certPEM.Bytes(), keyPEM.Bytes(), nil
}
