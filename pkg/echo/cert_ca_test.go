package echo

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	loadCAPair     = sync.OnceValues(generateCACert)
	loadServerPair = sync.OnceValues(func() (Pair, error) {
		ca, err := loadCAPair()
		if err != nil {
			return Pair{}, err
		}
		return generateServerCert(ca)
	})
)

func parseCert(t *testing.T, pemBytes []byte) *x509.Certificate {
	t.Helper()

	block, _ := pem.Decode(pemBytes)
	require.NotNil(t, block, "decoding CertPEM")
	require.Equal(t, "CERTIFICATE", block.Type)

	cert, err := x509.ParseCertificate(block.Bytes)
	require.NoError(t, err)
	return cert
}

func TestGenerateCACert_SelfSigned(t *testing.T) {
	pair, err := loadCAPair()
	require.NoError(t, err)

	cert := parseCert(t, pair.CertPEM)

	assert.True(t, cert.IsCA, "IsCA")
	assert.True(t, cert.BasicConstraintsValid, "BasicConstraintsValid")
	assert.NotZero(t, cert.KeyUsage&x509.KeyUsageCertSign, "KeyUsageCertSign")
	assert.NotZero(t, cert.KeyUsage&x509.KeyUsageDigitalSignature, "KeyUsageDigitalSignature")
	assert.Contains(t, cert.Subject.Organization, "ACME Co.")
	assert.NoError(t, cert.CheckSignatureFrom(cert), "self-signature must verify")

	keyBlock, _ := pem.Decode(pair.PrivatePEM)
	require.NotNil(t, keyBlock, "decoding PrivatePEM")
	assert.Equal(t, "RSA PRIVATE KEY", keyBlock.Type)
	_, err = x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	assert.NoError(t, err)
}

func TestGenerateServerCert_ChainAndSANs(t *testing.T) {
	caPair, err := loadCAPair()
	require.NoError(t, err)
	serverPair, err := loadServerPair()
	require.NoError(t, err)

	caCert := parseCert(t, caPair.CertPEM)
	serverCert := parseCert(t, serverPair.CertPEM)

	pool := x509.NewCertPool()
	pool.AddCert(caCert)

	_, err = serverCert.Verify(x509.VerifyOptions{
		Roots:     pool,
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	})
	assert.NoError(t, err, "server cert must chain to CA")

	assertContainsIP(t, serverCert.IPAddresses, net.IPv4(127, 0, 0, 1))
	assertContainsIP(t, serverCert.IPAddresses, net.IPv6loopback)

	assert.NotZero(t, serverCert.KeyUsage&x509.KeyUsageDigitalSignature, "KeyUsageDigitalSignature")
	assert.Contains(t, serverCert.ExtKeyUsage, x509.ExtKeyUsageClientAuth)
	assert.Contains(t, serverCert.ExtKeyUsage, x509.ExtKeyUsageServerAuth)
	assert.Equal(t, "localhost", serverCert.Subject.CommonName)
}

func TestPair_X509KeyPair_RoundTrip(t *testing.T) {
	caPair, err := loadCAPair()
	require.NoError(t, err)
	serverPair, err := loadServerPair()
	require.NoError(t, err)

	cases := []struct {
		Name string
		Pair Pair
	}{
		{Name: "CA", Pair: caPair},
		{Name: "server", Pair: serverPair},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			cert, err := tc.Pair.X509KeyPair()
			require.NoError(t, err)
			assert.NotEmpty(t, cert.Certificate)
			require.NotNil(t, cert.PrivateKey)
			_, ok := cert.PrivateKey.(*rsa.PrivateKey)
			assert.True(t, ok, "PrivateKey must be *rsa.PrivateKey")
		})
	}
}

func assertContainsIP(t *testing.T, ips []net.IP, want net.IP) {
	t.Helper()
	for _, ip := range ips {
		if ip.Equal(want) {
			return
		}
	}
	assert.Failf(t, "missing IP SAN", "expected %s in %v", want, ips)
}
