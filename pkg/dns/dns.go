package dns

import (
	"fmt"
	"net"
	"time"

	"github.com/miekg/dns"
)

type Query struct {
	ServerAddress string

	Recurse    bool
	QueryName  string
	QueryType  QueryType
	QueryClass QueryClass
}

var defaultServerAddress = "127.0.0.1:53"

func init() {
	if cc, err := dns.ClientConfigFromFile("/etc/resolv.conf"); err == nil {
		if len(cc.Servers) > 0 {
			ip := net.ParseIP(cc.Servers[0])
			if ip4 := ip.To4(); len(ip4) == net.IPv4len {
				defaultServerAddress = cc.Servers[0] + ":53"
			} else if len(ip) == net.IPv6len {
				defaultServerAddress = "[" + ip.String() + "]:53"
			}
		}
	}
}

func New() *Query {
	return &Query{
		ServerAddress: defaultServerAddress,

		Recurse:    true,
		QueryName:  "",
		QueryType:  "A",
		QueryClass: "IN",
	}
}

func (q *Query) Run() error {
	m := &dns.Msg{
		MsgHdr: dns.MsgHdr{
			Id:               dns.Id(),
			RecursionDesired: q.Recurse,
		},
		Question: []dns.Question{
			{
				Name:   q.QueryName,
				Qtype:  q.QueryType.DNSType(),
				Qclass: q.QueryClass.DNSClass(),
			},
		},
	}

	c := dns.Client{
		Dialer: &net.Dialer{
			Timeout: 5 * time.Second,
		},
	}

	in, rtt, err := c.Exchange(m, q.ServerAddress)
	if err != nil {
		return err
	}

	fmt.Printf("%+v\n", in)
	fmt.Printf(";; rtt: %+v\n", rtt)
	return nil
}
