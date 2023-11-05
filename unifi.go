package unifi

import (
	"context"
	"fmt"
	"github.com/xruins/unifi-coredns/pkg/unifi"
	"net"
	"sync"
	"time"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"

	"github.com/miekg/dns"
)

type options struct {
	aaaa          bool
	reload        time.Duration
	ttl           uint32
	user          string
	pass          string
	url           string
	caseSensitive bool
}

// Unifi is the plugin handler to register Unifi hosts to DNS
type Unifi struct {
	Next    plugin.Handler
	options options

	mu          sync.RWMutex
	hostMap     map[string]*net.IP
	unifiClient *unifi.Client
}

func (h *Unifi) updateHosts(ctx context.Context) error {
	hosts, err := h.unifiClient.GetHosts(ctx)
	if err != nil {
		return fmt.Errorf("failed to get Unifi: %w", err)
	}

	newMap := make(map[string]*net.IP, len(hosts))
	for _, host := range hosts {
		newMap[host.Name] = host.Addr
	}

	h.mu.Lock()
	h.hostMap = newMap
	defer h.mu.Unlock()

	return nil
}

// ServeDNS implements the plugin.Handle interface.
func (h *Unifi) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}
	qname := state.Name()

	var answers []dns.RR

	switch state.QType() {
	case dns.TypeA:
		h.mu.RLock()
		v, ok := h.hostMap[qname]
		h.mu.RUnlock()
		if !ok {
			return plugin.NextOrFailure(h.Name(), h.Next, ctx, w, r)
		}
		answers = a(qname, h.options.ttl, []net.IP{*v})
	case dns.TypeAAAA:
		if !h.options.aaaa {
			return plugin.NextOrFailure(h.Name(), h.Next, ctx, w, r)
		}
		h.mu.RLock()
		v, ok := h.hostMap[qname]
		h.mu.RUnlock()
		if !ok {
			return plugin.NextOrFailure(h.Name(), h.Next, ctx, w, r)
		}
		v6 := v.To16()
		answers = aaaa(qname, h.options.ttl, []net.IP{v6})
	default:
		return plugin.NextOrFailure(h.Name(), h.Next, ctx, w, r)
	}

	// Only on NXDOMAIN we will fallthrough.
	if len(answers) == 0 {
		return plugin.NextOrFailure(h.Name(), h.Next, ctx, w, r)
	}

	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative = true
	m.Answer = answers

	err := w.WriteMsg(m)
	if err != nil {
		return 0, fmt.Errorf("failed to write response message: %w", err)
	}
	return dns.RcodeSuccess, nil
}

// Name implements the plugin.Handle interface.
func (h *Unifi) Name() string { return "unifi" }

// a takes a slice of net.IPs and returns a slice of A RRs.
func a(zone string, ttl uint32, ips []net.IP) []dns.RR {
	answers := make([]dns.RR, len(ips))
	for i, ip := range ips {
		r := new(dns.A)
		r.Hdr = dns.RR_Header{Name: zone, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: ttl}
		r.A = ip
		answers[i] = r
	}
	return answers
}

// aaaa takes a slice of net.IPs and returns a slice of AAAA RRs.
func aaaa(zone string, ttl uint32, ips []net.IP) []dns.RR {
	answers := make([]dns.RR, len(ips))
	for i, ip := range ips {
		r := new(dns.AAAA)
		r.Hdr = dns.RR_Header{Name: zone, Rrtype: dns.TypeAAAA, Class: dns.ClassINET, Ttl: ttl}
		r.AAAA = ip
		answers[i] = r
	}
	return answers
}
