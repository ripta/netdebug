package dns

import (
	"fmt"
	"github.com/miekg/dns"
)

type QueryClass string

var QueryClassMap = map[string]uint16{
	"IN": dns.ClassINET,
}

func (e *QueryClass) String() string {
	return string(*e)
}

func (e *QueryClass) Set(v string) error {
	if _, ok := QueryClassMap[v]; !ok {
		queryClasses := []string{}
		for k := range QueryClassMap {
			queryClasses = append(queryClasses, k)
		}

		return fmt.Errorf("invalid query class; must be one of %+v", queryClasses)
	}

	*e = QueryClass(v)
	return nil
}

func (e *QueryClass) Type() string {
	return "QueryClass"
}

func (e *QueryClass) DNSClass() uint16 {
	return QueryClassMap[string(*e)]
}

type QueryType string

var QueryTypeMap = map[string]uint16{
	"A":     dns.TypeA,
	"CNAME": dns.TypeCNAME,
	"MX":    dns.TypeMX,
	"NS":    dns.TypeNS,
	"SRV":   dns.TypeSRV,
	"TXT":   dns.TypeTXT,
}

func (e *QueryType) String() string {
	return string(*e)
}

func (e *QueryType) Set(v string) error {
	if _, ok := QueryTypeMap[v]; !ok {
		queryTypes := []string{}
		for k := range QueryTypeMap {
			queryTypes = append(queryTypes, k)
		}

		return fmt.Errorf("invalid query type; must be one of %+v", queryTypes)
	}

	*e = QueryType(v)
	return nil
}

func (e *QueryType) Type() string {
	return "QueryType"
}

func (e *QueryType) DNSType() uint16 {
	return QueryTypeMap[string(*e)]
}
