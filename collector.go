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

var (
	numCalls = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "fritzbox_exporter_calls",
		Help: "Number of calls to a service action.",
	})

	collectErrors = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "fritzbox_exporter_collect_errors",
		Help: "Number of collection errors.",
	})
	serviceNotFound = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "fritzbox_exporter_service_not_found",
		Help: "",
	}, []string{"service"})
	actionNotFound = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "fritzbox_exporter_action_not_found",
		Help: "",
	}, []string{"action"})
	resultNotFound = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "fritzbox_exporter_result_not_found",
		Help: "",
	}, []string{"result"})

	collectMetrics = []prometheus.Collector{numCalls, collectErrors, serviceNotFound, actionNotFound, resultNotFound}
)

type FritzboxCollector struct {
	Parameters upnp.ConnectionParameters
	Metrics    []*Metric

	sync.RWMutex // protects services
	services     map[string]*upnp.Service
}

func NewCollector(params upnp.ConnectionParameters, metrics []*Metric) *FritzboxCollector {
	c := &FritzboxCollector{
		Parameters: params,
		Metrics:    metrics,
		services:   make(map[string]*upnp.Service),
	}
	go c.loadServices()
	return c
}

// LoadServices tries to load the service information. Retries until success.
func (fc *FritzboxCollector) loadServices() {
	igdRoot := fc.loadService(upnp.IGDServiceDescriptor)
	log.Printf("%d IGD services loaded\n", len(igdRoot.Services))
	fc.Lock()
	for _, s := range igdRoot.Services {
		fc.services[s.ServiceType] = s
	}
	fc.Unlock()

	if fc.Parameters.Username == "" {
		log.Printf("no username set: not loading TR64 services")
		return
	}
	tr64Root := fc.loadService(upnp.TR64ServiceDescriptor)
	log.Printf("%d TR64 services loaded\n", len(tr64Root.Services))
	fc.Lock()
	for _, s := range tr64Root.Services {
		fc.services[s.ServiceType] = s
	}
	fc.Unlock()
}

func (fc *FritzboxCollector) loadService(desc string) *upnp.Root {
	for {
		root, err := upnp.LoadServiceRoot(fc.Parameters, desc)
		if err != nil {
			collectErrors.Inc()
			fmt.Printf("cannot load services: %s\n", err)

			time.Sleep(serviceLoadRetryTime)
			continue
		}
		return root
	}
}

func (fc *FritzboxCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range fc.Metrics {
		ch <- m.desc
	}
}

func (fc *FritzboxCollector) Collect(ch chan<- prometheus.Metric) {
	fc.RLock()
	defer fc.RUnlock()

	type cacheKey struct {
		Service string
		Action  string
	}

	// Cache Action call result. Multiple metrics might use different results from a call.
	resultCache := make(map[cacheKey]upnp.Result)

	for _, m := range fc.Metrics {
		result, ok := resultCache[cacheKey{
			Service: m.Service,
			Action:  m.Action,
		}]

		if !ok {
			service, ok := fc.services[m.Service]
			if !ok {
				serviceNotFound.WithLabelValues(m.Service).Inc()
				continue
			}
			action, ok := service.Actions[m.Action]
			if !ok {
				actionNotFound.WithLabelValues(m.Action).Inc()
				continue
			}

			numCalls.Inc()
			var err error
			result, err = action.Call()
			if err != nil {
				fmt.Println(err)
				collectErrors.Inc()
				continue
			}

			resultCache[cacheKey{
				Service: m.Service,
				Action:  m.Action,
			}] = result
		}

		val, ok := result[m.Result]
		if !ok {
			resultNotFound.WithLabelValues(m.Result).Inc()
			continue
		}

		fc.exportMetric(m, ch, val)
	}
}

func (fc *FritzboxCollector) exportMetric(m *Metric, ch chan<- prometheus.Metric, val interface{}) {
	if m.LabelName == "" {
		// normal metric

		floatVal, ok := toFloat(val, m.OkValue)
		if !ok {
			log.Println("cannot convert to float:", val)
			collectErrors.Inc()
		}

		ch <- prometheus.MustNewConstMetric(
			m.desc, m.metricType, floatVal,
			fc.Parameters.Device,
		)
	} else {
		// value as label metric
		stringVal := fmt.Sprintf("%s", val)
		ch <- prometheus.MustNewConstMetric(
			m.desc, m.metricType, 1.0,
			fc.Parameters.Device, stringVal,
		)
	}
}

func toFloat(val any, okValue string) (float64, bool) {
	switch val := val.(type) {
	case uint64:
		return float64(val), true
	case bool:
		if val {
			return 1, true
		} else {
			return 0, true
		}
	case string:
		if okValue == "" {
			return 0, false
		} else if val == okValue {
			return 1, true
		} else {
			return 0, true
		}
	default:
		return 0, false

	}

}
