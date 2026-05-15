package dns

import (
	"testing"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	q := New()
	require.NotNil(t, q)

	assert.True(t, q.Recurse)
	assert.Empty(t, q.QueryName)

	assert.Equal(t, QueryType("A"), q.QueryType)
	assert.Equal(t, dns.TypeA, q.QueryType.DNSType())

	assert.Equal(t, QueryClass("IN"), q.QueryClass)
	assert.Equal(t, uint16(dns.ClassINET), q.QueryClass.DNSClass())

	assert.NotEmpty(t, q.ServerAddress)
}
