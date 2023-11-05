package unifi

import (
	"context"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/miekg/dns"
	"github.com/xruins/unifi-coredns/pkg/test"
	"github.com/xruins/unifi-coredns/pkg/unifi"
	"net"
	"testing"
)

func initMockedClient() *Unifi {
	mock := test.MockClient{}
	return &Unifi{
		unifiClient: &unifi.Client{
			Client: &mock,
		},
	}
}

func TestUnifi_UpdateHosts(t *testing.T) {
	client := initMockedClient()
	ctx := context.Background()
	err := client.updateHosts(ctx)
	if err != nil {
		t.Fatalf("updateHosts failed: %s", err)
	}

	got := client.hostMap
	ip := net.IPv4(192, 168, 1, 1)
	want := map[string]*net.IP{
		"parsable": &ip,
	}
	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("updateHosts() mismatch (-want, +got): %s", diff)
	}
}

func TestUnifi_ServeDNS(t *testing.T) {
	client := initMockedClient()
	ctx := context.Background()
	err := client.updateHosts(ctx)
	if err != nil {
		t.Fatalf("updateHosts failed: %s", err)
	}
	client.options.aaaa = true
	client.options.ttl = 60

	tests := []struct {
		description string
		name        string
		qtype       uint16
		wantCode    int
		wantErr     bool
		want        *dns.Msg
	}{
		{
			description: "successful query (A record)",
			name:        "parsable",
			qtype:       dns.TypeA,
			wantCode:    dns.RcodeSuccess,
			want: &dns.Msg{
				Answer: []dns.RR{
					&dns.A{
						Hdr: dns.RR_Header{Name: "parsable", Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60},
						A:   net.IPv4(192, 168, 1, 1),
					},
				},
			},
		},
	}

	for _, tst := range tests {
		query := new(dns.Msg)
		query.SetQuestion(tst.name, tst.qtype)
		rw := &test.ResponseWriter{}
		gotCode, err := client.ServeDNS(ctx, rw, query)
		if !tst.wantErr && err != nil {
			t.Errorf("ServeDNS() unexpected error: %s", err)
		}
		if gotCode != tst.wantCode {
			t.Errorf("got unexpected code. got: %d, want: %d", gotCode, tst.wantCode)
		}
		got := rw.Msg
		if diff := cmp.Diff(got, tst.want, cmpopts.IgnoreFields(dns.Msg{}, "Question", "MsgHdr")); diff != "" {
			t.Errorf("ServeDNS() returned unexpected DNS record (-want, +got): %s", diff)
		}
	}
}
