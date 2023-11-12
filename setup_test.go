package unifi

import (
	"context"
	"testing"
	"time"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/plugin/pkg/fall"
	"github.com/miekg/dns"
)

func (o *options) equals(other options) bool {
	return o.aaaa == other.aaaa &&
		o.reload == other.reload &&
		o.ttl == other.ttl &&
		o.user == other.user &&
		o.pass == other.pass &&
		o.url == other.url &&
		o.caseSensitive == other.caseSensitive
}

type nopPlugin struct{}

func (h *nopPlugin) Name() string { return "nop" }

func (h *nopPlugin) ServeDNS(_ context.Context, _ dns.ResponseWriter, _ *dns.Msg) (int, error) {
	return 0, nil
}
func TestHostParse(t *testing.T) {
	// need to implement mocking
	t.Skip()
	f := fall.F{}
	f.SetZonesFromArgs([]string{"example.com", "example.org"})
	tests := []struct {
		description string
		input       string
		wantErr     bool
		wantOptions options
	}{
		{
			description: `all parameters are present`,
			input: `
unifi https://unifi unifiuser password {
    aaaa
    reload 1s
    ttl 60
    sites site1 site2
    casesensitive
    fallthrough example.com example.org
}`,
			wantOptions: options{
				aaaa:          true,
				reload:        1 * time.Second,
				ttl:           60,
				sites:         []string{"site1", "site2"},
				caseSensitive: true,
				fall:          f,
				url:           "https://unifi",
				user:          "unifiuser",
				pass:          "password",
			},
		},
		{
			description: `only mandatory arguments are present`,
			input: `
unifi https://unifi unifiuser password`,
			wantOptions: options{
				url:  "https://unifi",
				user: "unifiuser",
				pass: "password",
				// defaults
				reload: 60 * time.Second,
				ttl:    60,
			},
		},
		{
			description: `missing mandatory arguments`,
			input:       `unifi`,
			wantErr:     true,
		},
		{
			description: `malformed arbitrary argument`,
			input: `
unifi https://unifi unifiuser password {
    wrong
}`,
			wantErr: true,
		},
		{
			description: `missing mandatory arguments`,
			input:       `unifi`,
			wantErr:     true,
		},
	}

	for _, test := range tests {
		c := caddy.NewTestController("nop", test.input)
		h, err := hostsParse(c)

		got := h.options
		want := test.wantOptions
		desc := test.description
		if err == nil && test.wantErr {
			t.Fatalf("Test `%s` expected errors, but got no error", desc)
		} else if err != nil && !test.wantErr {
			t.Fatalf("Test `%s` expected no errors, but got '%v'", desc, err)
		} else {
			if !test.wantErr && !got.equals(want) {
				t.Fatalf("Test `%s` expected options to be '%v', but got '%v'", desc, want, got)
			}
		}
	}
}
