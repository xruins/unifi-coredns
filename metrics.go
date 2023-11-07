package unifi

import (
	"github.com/coredns/coredns/plugin"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// unifiHostsEntries is the combined number of entries.
	unifiHostsEntries = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: plugin.Namespace,
		Subsystem: "unifi",
		Name:      "entries",
		Help:      "The combined number of entries in hosts and Corefile.",
	}, []string{})
	// hostsReloadTime is the timestamp of the last reload of the records by Unifi API.
	unifiHostsReloadTime = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: plugin.Namespace,
		Subsystem: "unifi",
		Name:      "reload_timestamp_seconds",
		Help:      "The timestamp of the last reload of hosts file.",
	})
)
