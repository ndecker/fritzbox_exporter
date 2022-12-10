package main

import (
	"fmt"
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
	sync.Mutex // protects Root
	Root       *upnp.Root
}

// LoadServices tries to load the service information. Retries until success.
func (fc *FritzboxCollector) LoadServices() {
	for {
		root, err := upnp.LoadServices(fc.Parameters)
		if err != nil {
			fmt.Printf("cannot load services: %s\n", err)

			time.Sleep(serviceLoadRetryTime)
			continue
		}

		fmt.Printf("services loaded\n")

		fc.Lock()
		fc.Root = root
		fc.Unlock()
		return
	}
}

func (fc *FritzboxCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range metrics {
		ch <- m.Desc
	}
}

func (fc *FritzboxCollector) Collect(ch chan<- prometheus.Metric) {
	fc.Lock()
	root := fc.Root
	fc.Unlock()

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
