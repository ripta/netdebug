package echo

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"time"
)

type Pair struct {
	PrivatePEM []byte
	CertPEM    []byte
}

func (p Pair) X509KeyPair() (tls.Certificate, error) {
	return tls.X509KeyPair(p.CertPEM, p.PrivatePEM)
}

func generateCACert() (Pair, error) {
	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return Pair{}, fmt.Errorf("generating CA private key: %w", err)
	}

	crt := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"ACME Co."},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(0, 3, 0),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	ca, err := x509.CreateCertificate(rand.Reader, crt, crt, &key.PublicKey, key)
	if err != nil {
		return Pair{}, fmt.Errorf("creating X509 certificate: %w", err)
	}

	crtBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: ca,
	}

	crtPem := &bytes.Buffer{}
	if err := pem.Encode(crtPem, crtBlock); err != nil {
		return Pair{}, fmt.Errorf("PEM-encoding certificate: %w", err)
	}

	keyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}

	keyPem := &bytes.Buffer{}
	if err := pem.Encode(keyPem, keyBlock); err != nil {
		return Pair{}, fmt.Errorf("PEM-encoding private key: %w", err)
	}

	return Pair{
		CertPEM:    crtPem.Bytes(),
		PrivatePEM: keyPem.Bytes(),
	}, nil
}

func generateServerCert(caPair Pair) (Pair, error) {
	caCertPem, _ := pem.Decode(caPair.CertPEM)
	caCert, err := x509.ParseCertificate(caCertPem.Bytes)
	if err != nil {
		return Pair{}, fmt.Errorf("parsing CA cert: %w (PEM block type %s)", err, caCertPem.Type)
	}

	caKeyPem, _ := pem.Decode(caPair.PrivatePEM)
	caKey, err := x509.ParsePKCS1PrivateKey(caKeyPem.Bytes)
	if err != nil {
		return Pair{}, fmt.Errorf("parsing CA key: %w (PEM block type %s)", err, caKeyPem.Type)
	}

	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return Pair{}, fmt.Errorf("generating private key: %w", err)
	}

	crt := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:   "localhost",
			Organization: []string{"ACME Co."},
		},
		IPAddresses: []net.IP{
			net.IPv4(127, 0, 0, 1),
			net.IPv6loopback,
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(0, 0, 14),
		KeyUsage:  x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageClientAuth,
			x509.ExtKeyUsageServerAuth,
		},
	}

	ca, err := x509.CreateCertificate(rand.Reader, crt, caCert, &key.PublicKey, caKey)
	if err != nil {
		return Pair{}, fmt.Errorf("creating X509 certificate: %w", err)
	}

	crtBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: ca,
	}

	crtPem := &bytes.Buffer{}
	if err := pem.Encode(crtPem, crtBlock); err != nil {
		return Pair{}, fmt.Errorf("PEM-encoding certificate: %w", err)
	}

	keyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}

	keyPem := &bytes.Buffer{}
	if err := pem.Encode(keyPem, keyBlock); err != nil {
		return Pair{}, fmt.Errorf("PEM-encoding private key: %w", err)
	}

	return Pair{
		PrivatePEM: keyPem.Bytes(),
		CertPEM:    crtPem.Bytes(),
	}, nil
}
