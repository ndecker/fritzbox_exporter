package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	upnp "github.com/ndecker/fritzbox_exporter/fritzbox_upnp"
	"github.com/prometheus/client_golang/prometheus"
)

const serviceLoadRetryTime = 1 * time.Minute

var collectErrors = prometheus.NewCounter(prometheus.CounterOpts{
	Name: "fritzbox_exporter_collect_errors",
	Help: "Number of collection errors.",
})

type FritzboxCollector struct {
	Parameters upnp.ConnectionParameters
	sync.Mutex // protects Roots
	IGDRoot    *upnp.Root
	TR64Root   *upnp.Root
}

// LoadServices tries to load the service information. Retries until success.
func (fc *FritzboxCollector) LoadServices() {
	igdRoot := fc.loadService(upnp.IGDServiceDescriptor)
	log.Printf("%d IGD services loaded\n", len(igdRoot.Services))
	fc.Lock()
	fc.IGDRoot = igdRoot
	fc.Unlock()

	if fc.Parameters.Username == "" {
		log.Printf("no username set: not loading TR64 services")
		return
	}
	tr64Root := fc.loadService(upnp.TR64ServiceDescriptor)
	log.Printf("%d TR64 services loaded\n", len(tr64Root.Services))
	fc.Lock()
	fc.TR64Root = tr64Root
	fc.Unlock()
}

func (fc *FritzboxCollector) loadService(desc string) *upnp.Root {
	for {
		root, err := upnp.LoadServiceRoot(fc.Parameters, desc)
		if err != nil {
			fmt.Printf("cannot load services: %s\n", err)

			time.Sleep(serviceLoadRetryTime)
			continue
		}
		return root
	}
}

func (fc *FritzboxCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range IGDMetrics {
		ch <- m.Desc
	}
	for _, m := range TR64Metrics {
		ch <- m.Desc
	}
}

func (fc *FritzboxCollector) Collect(ch chan<- prometheus.Metric) {
	fc.Lock()
	root := fc.IGDRoot
	fc.Unlock()
	fc.collectRoot(ch, root, IGDMetrics)

	fc.Lock()
	root = fc.TR64Root
	fc.Unlock()
	fc.collectRoot(ch, root, TR64Metrics)
}

func (fc *FritzboxCollector) collectRoot(ch chan<- prometheus.Metric, root *upnp.Root, metrics []*Metric) {
	if root == nil {
		return // Services not loaded yet
	}

	var err error
	var lastService string
	var lastMethod string
	var lastResult upnp.Result

	for _, m := range metrics {
		if m.Service != lastService || m.Action != lastMethod {
			service, ok := root.Services[m.Service]
			if !ok {
				// TODO
				fmt.Println("cannot find service", m.Service)
				fmt.Println(root.Services)
				continue
			}
			action, ok := service.Actions[m.Action]
			if !ok {
				// TODO
				fmt.Println("cannot find action", m.Action)
				continue
			}

			lastResult, err = action.Call()
			if err != nil {
				fmt.Println(err)
				collectErrors.Inc()
				continue
			}
		}

		val, ok := lastResult[m.Result]
		if !ok {
			fmt.Println("result not found", m.Result)
			collectErrors.Inc()
			continue
		}

		floatVal, ok := toFloat(val, m.OkValue)
		if !ok {
			fmt.Println("cannot convert to float:", val)
			collectErrors.Inc()
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			m.Desc,
			m.MetricType,
			floatVal,
			fc.Parameters.Device,
		)
	}
}

func toFloat(val any, okValue string) (float64, bool) {
	switch tval := val.(type) {
	case uint64:
		return float64(tval), true
	case bool:
		if tval {
			return 1, true
		} else {
			return 0, true
		}
	case string:
		if tval == okValue {
			return 1, true
		} else {
			return 0, true
		}
	default:
		return 0, false

	}

}
