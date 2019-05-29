package firewall

import (
	"fmt"
	"net"
	"strings"

	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"

	"github.com/mholt/caddy"
	"github.com/miekg/dns"
)

func init() {
	caddy.RegisterPlugin("firewall", caddy.Plugin{
		ServerType: "dns",
		Action:     setup,
	})
}

func setup(c *caddy.Controller) error {
	f := firewall{}
	/*
	 * firewall [ZONES...] {
	 *   ACTION type QTYPE from SOURCE
	 *   ...
	 * }
	 *
	 * ACTION: allow|block
	 */
	for c.Next() {
		// load <ZONES...>.
		f.Zones = c.RemainingArgs()
		if len(f.Zones) == 0 {
			// if empty, the zones from the configuration block are used.
			f.Zones = make([]string, len(c.ServerBlockKeys))
			copy(f.Zones, c.ServerBlockKeys)
		}
		// strip port and transport.
		for i := range f.Zones {
			f.Zones[i] = plugin.Host(f.Zones[i]).Normalize()
		}

		var err error
		// load all tokens in this block.
		for c.NextBlock() {
			rule := Rule{}
			// ACTION type QTYPE from SOURCE
			rule.action = strings.ToLower(c.Val())
			if rule.action != ALLOW && rule.action != BLOCK {
				return c.Errf("Unexpected token '%s'; expect '%s' or '%s'", c.Val(), ALLOW, BLOCK)
			}

			// TODO: simplify the syntax and remove tedious code. (@ihac)
			if !c.NextArg() {
				return c.ArgErr()
			}
			if strings.ToLower(c.Val()) != "type" {
				return c.Errf("Unexpected token '%s'; expect 'type'", c.Val())
			}

			if !c.NextArg() {
				return c.ArgErr()
			}
			rule.qtype, err = parseQype(c.Val())
			if err != nil {
				return err
			}

			if !c.NextArg() {
				return c.ArgErr()
			}
			if strings.ToLower(c.Val()) != "from" {
				return c.Errf("Unexpected token '%s'; expect 'from'", c.Val())
			}

			if !c.NextArg() {
				return c.ArgErr()
			}
			_, rule.source, err = net.ParseCIDR(c.Val())
			if err != nil {
				return c.Errf("Illegal CIDR notation '%s'", c.Val())
			}

			f.Rules = append(f.Rules, rule)
		}
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		f.Next = next
		return f
	})
	return nil
}

// TODO: dns.Type == QType? (@ihac)
func parseQype(raw string) (dns.Type, error) {
	switch raw {
	case "A":
		return dns.Type(dns.TypeA), nil
	case "AAAA":
		return dns.Type(dns.TypeAAAA), nil
	case "AFSDB":
		return dns.Type(dns.TypeAFSDB), nil
	case "CAA":
		return dns.Type(dns.TypeCAA), nil
	case "CDNSKEY":
		return dns.Type(dns.TypeCDNSKEY), nil
	case "CDS":
		return dns.Type(dns.TypeCDS), nil
	case "CERT":
		return dns.Type(dns.TypeCERT), nil
	case "CNAME":
		return dns.Type(dns.TypeCNAME), nil
	case "DHCID":
		return dns.Type(dns.TypeDHCID), nil
	case "DLV":
		return dns.Type(dns.TypeDLV), nil
	case "DNAME":
		return dns.Type(dns.TypeDNAME), nil
	case "DNSKEY":
		return dns.Type(dns.TypeDNSKEY), nil
	case "DS":
		return dns.Type(dns.TypeDS), nil
	case "HIP":
		return dns.Type(dns.TypeHIP), nil
	case "KEY":
		return dns.Type(dns.TypeKEY), nil
	case "KX":
		return dns.Type(dns.TypeKX), nil
	case "LOC":
		return dns.Type(dns.TypeLOC), nil
	case "MX":
		return dns.Type(dns.TypeMX), nil
	case "NAPTR":
		return dns.Type(dns.TypeNAPTR), nil
	case "NS":
		return dns.Type(dns.TypeNS), nil
	case "NSEC":
		return dns.Type(dns.TypeNSEC), nil
	case "NSEC3":
		return dns.Type(dns.TypeNSEC3), nil
	case "NSEC3PARAM":
		return dns.Type(dns.TypeNSEC3PARAM), nil
	case "OPENPGPKEY":
		return dns.Type(dns.TypeOPENPGPKEY), nil
	case "PTR":
		return dns.Type(dns.TypePTR), nil
	case "RRSIG":
		return dns.Type(dns.TypeRRSIG), nil
	case "RP":
		return dns.Type(dns.TypeRP), nil
	case "SIG":
		return dns.Type(dns.TypeSIG), nil
	case "SMIMEA":
		return dns.Type(dns.TypeSMIMEA), nil
	case "SOA":
		return dns.Type(dns.TypeSOA), nil
	case "SRV":
		return dns.Type(dns.TypeSRV), nil
	case "SSHFP":
		return dns.Type(dns.TypeSSHFP), nil
	case "TA":
		return dns.Type(dns.TypeTA), nil
	case "TKEY":
		return dns.Type(dns.TypeTKEY), nil
	case "TLSA":
		return dns.Type(dns.TypeTLSA), nil
	case "TSIG":
		return dns.Type(dns.TypeTSIG), nil
	case "TXT":
		return dns.Type(dns.TypeTXT), nil
	case "URI":
		return dns.Type(dns.TypeURI), nil
	default:
		return 0, fmt.Errorf("Unexpected token '%s'; expect legal QTYPE", raw)
	}
}
