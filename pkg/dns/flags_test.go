package dns

import (
	"testing"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type queryClassSetTest struct {
	Name         string
	Input        string
	WantErr      bool
	WantValue    QueryClass
	WantDNSClass uint16
}

var queryClassSetTests = []queryClassSetTest{
	{
		Name:         "IN accepts known value",
		Input:        "IN",
		WantValue:    "IN",
		WantDNSClass: dns.ClassINET,
	},
	{
		Name:    "lowercase in is rejected",
		Input:   "in",
		WantErr: true,
	},
	{
		Name:    "unknown CH is rejected",
		Input:   "CH",
		WantErr: true,
	},
	{
		Name:    "empty string is rejected",
		Input:   "",
		WantErr: true,
	},
}

func TestQueryClass_Set(t *testing.T) {
	for _, tc := range queryClassSetTests {
		t.Run(tc.Name, func(t *testing.T) {
			var qc QueryClass
			err := qc.Set(tc.Input)
			if tc.WantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.WantValue, qc)
			assert.Equal(t, string(tc.WantValue), qc.String())
			assert.Equal(t, tc.WantDNSClass, qc.DNSClass())
		})
	}
}

func TestQueryClass_Type(t *testing.T) {
	var qc QueryClass
	assert.Equal(t, "QueryClass", qc.Type())
}

func TestQueryClass_SetErrorListsValidValues(t *testing.T) {
	var qc QueryClass
	err := qc.Set("bogus")
	require.Error(t, err)
	for name := range QueryClassMap {
		assert.Contains(t, err.Error(), name)
	}
}

type queryTypeSetTest struct {
	Name        string
	Input       string
	WantErr     bool
	WantValue   QueryType
	WantDNSType uint16
}

var queryTypeSetTests = []queryTypeSetTest{
	{
		Name:        "A round-trips",
		Input:       "A",
		WantValue:   "A",
		WantDNSType: dns.TypeA,
	},
	{
		Name:        "CNAME round-trips",
		Input:       "CNAME",
		WantValue:   "CNAME",
		WantDNSType: dns.TypeCNAME,
	},
	{
		Name:        "MX round-trips",
		Input:       "MX",
		WantValue:   "MX",
		WantDNSType: dns.TypeMX,
	},
	{
		Name:        "NS round-trips",
		Input:       "NS",
		WantValue:   "NS",
		WantDNSType: dns.TypeNS,
	},
	{
		Name:        "SRV round-trips",
		Input:       "SRV",
		WantValue:   "SRV",
		WantDNSType: dns.TypeSRV,
	},
	{
		Name:        "TXT round-trips",
		Input:       "TXT",
		WantValue:   "TXT",
		WantDNSType: dns.TypeTXT,
	},
	{
		Name:        "lowercase mx normalizes to MX",
		Input:       "mx",
		WantValue:   "MX",
		WantDNSType: dns.TypeMX,
	},
	{
		Name:        "mixed-case Mx normalizes to MX",
		Input:       "Mx",
		WantValue:   "MX",
		WantDNSType: dns.TypeMX,
	},
	{
		Name:    "unknown AAAA is rejected",
		Input:   "AAAA",
		WantErr: true,
	},
	{
		Name:    "empty string is rejected",
		Input:   "",
		WantErr: true,
	},
}

func TestQueryType_Set(t *testing.T) {
	for _, tc := range queryTypeSetTests {
		t.Run(tc.Name, func(t *testing.T) {
			var qt QueryType
			err := qt.Set(tc.Input)
			if tc.WantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.WantValue, qt)
			assert.Equal(t, string(tc.WantValue), qt.String())
			assert.Equal(t, tc.WantDNSType, qt.DNSType())
		})
	}
}

func TestQueryType_Type(t *testing.T) {
	var qt QueryType
	assert.Equal(t, "QueryType", qt.Type())
}

func TestQueryType_SetErrorListsValidValues(t *testing.T) {
	var qt QueryType
	err := qt.Set("bogus")
	require.Error(t, err)
	for name := range QueryTypeMap {
		assert.Contains(t, err.Error(), name)
	}
}
