package main

// Copyright 2016 Nils Decker
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/mxschmitt/fritzbox_exporter/pkg/fritzboxmetrics"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

const serviceLoadRetryTime = 1 * time.Minute

var (
	collectErrors = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "fritzbox_exporter_collect_errors",
		Help: "Number of collection errors.",
	})
)

type Metric struct {
	Service string
	Action  string
	Result  string
	OkValue string

	Desc       *prometheus.Desc
	MetricType prometheus.ValueType
}

var metrics = []*Metric{
	{
		Service: "urn:schemas-upnp-org:service:WANCommonInterfaceConfig:1",
		Action:  "GetTotalPacketsReceived",
		Result:  "TotalPacketsReceived",
		Desc: prometheus.NewDesc(
			"gateway_wan_packets_received",
			"packets received on gateway WAN interface",
			[]string{"gateway"},
			nil,
		),
		MetricType: prometheus.CounterValue,
	},
	{
		Service: "urn:schemas-upnp-org:service:WANCommonInterfaceConfig:1",
		Action:  "GetTotalPacketsSent",
		Result:  "TotalPacketsSent",
		Desc: prometheus.NewDesc(
			"gateway_wan_packets_sent",
			"packets sent on gateway WAN interface",
			[]string{"gateway"},
			nil,
		),
		MetricType: prometheus.CounterValue,
	},
	{
		Service: "urn:schemas-upnp-org:service:WANCommonInterfaceConfig:1",
		Action:  "GetAddonInfos",
		Result:  "TotalBytesReceived",
		Desc: prometheus.NewDesc(
			"gateway_wan_bytes_received",
			"bytes received on gateway WAN interface",
			[]string{"gateway"},
			nil,
		),
		MetricType: prometheus.CounterValue,
	},
	{
		Service: "urn:schemas-upnp-org:service:WANCommonInterfaceConfig:1",
		Action:  "GetAddonInfos",
		Result:  "TotalBytesSent",
		Desc: prometheus.NewDesc(
			"gateway_wan_bytes_sent",
			"bytes sent on gateway WAN interface",
			[]string{"gateway"},
			nil,
		),
		MetricType: prometheus.CounterValue,
	},
	{
		Service: "urn:schemas-upnp-org:service:WANCommonInterfaceConfig:1",
		Action:  "GetAddonInfos",
		Result:  "ByteSendRate",
		Desc: prometheus.NewDesc(
			"gateway_wan_bytes_send_rate",
			"byte send rate on gateway WAN interface",
			[]string{"gateway"},
			nil,
		),
		MetricType: prometheus.GaugeValue,
	},
	{
		Service: "urn:schemas-upnp-org:service:WANCommonInterfaceConfig:1",
		Action:  "GetAddonInfos",
		Result:  "ByteReceiveRate",
		Desc: prometheus.NewDesc(
			"gateway_wan_bytes_receive_rate",
			"byte receive rate on gateway WAN interface",
			[]string{"gateway"},
			nil,
		),
		MetricType: prometheus.GaugeValue,
	},
	{
		Service: "urn:schemas-upnp-org:service:WANCommonInterfaceConfig:1",
		Action:  "GetCommonLinkProperties",
		Result:  "Layer1UpstreamMaxBitRate",
		Desc: prometheus.NewDesc(
			"gateway_wan_layer1_upstream_max_bitrate",
			"Layer1 upstream max bitrate",
			[]string{"gateway"},
			nil,
		),
		MetricType: prometheus.GaugeValue,
	},
	{
		Service: "urn:schemas-upnp-org:service:WANCommonInterfaceConfig:1",
		Action:  "GetCommonLinkProperties",
		Result:  "Layer1DownstreamMaxBitRate",
		Desc: prometheus.NewDesc(
			"gateway_wan_layer1_downstream_max_bitrate",
			"Layer1 downstream max bitrate",
			[]string{"gateway"},
			nil,
		),
		MetricType: prometheus.GaugeValue,
	},
	{
		Service: "urn:schemas-upnp-org:service:WANCommonInterfaceConfig:1",
		Action:  "GetCommonLinkProperties",
		Result:  "PhysicalLinkStatus",
		OkValue: "Up",
		Desc: prometheus.NewDesc(
			"gateway_wan_layer1_link_status",
			"Status of physical link (Up = 1)",
			[]string{"gateway"},
			nil,
		),
		MetricType: prometheus.GaugeValue,
	},
	{
		Service: "urn:schemas-upnp-org:service:WANIPConnection:1",
		Action:  "GetStatusInfo",
		Result:  "ConnectionStatus",
		OkValue: "Connected",
		Desc: prometheus.NewDesc(
			"gateway_wan_connection_status",
			"WAN connection status (Connected = 1)",
			[]string{"gateway"},
			nil,
		),
		MetricType: prometheus.GaugeValue,
	},
	{
		Service: "urn:schemas-upnp-org:service:WANIPConnection:1",
		Action:  "GetStatusInfo",
		Result:  "Uptime",
		Desc: prometheus.NewDesc(
			"gateway_wan_connection_uptime_seconds",
			"WAN connection uptime",
			[]string{"gateway"},
			nil,
		),
		MetricType: prometheus.GaugeValue,
	},
	{
		Service: "urn:dslforum-org:service:WLANConfiguration:1",
		Action:  "GetTotalAssociations",
		Result:  "TotalAssociations",
		Desc: prometheus.NewDesc(
			"gateway_wlan_current_connections",
			"current WLAN connections",
			[]string{"gateway"},
			nil,
		),
		MetricType: prometheus.GaugeValue,
	},
}

type FritzboxCollector struct {
	Gateway  string
	Port     uint16
	Username string
	Password string

	sync.Mutex // protects Root
	Root       *fritzboxmetrics.Root
}

// LoadServices tries to load the service information. Retries until success.
func (fc *FritzboxCollector) LoadServices() {
	for {
		root, err := fritzboxmetrics.LoadServices(fc.Gateway, fc.Port, fc.Username, fc.Password)
		if err != nil {
			fmt.Printf("cannot load services: %v\n", err)
			// Sleep so long how often the metrics should be fetched
			time.Sleep(serviceLoadRetryTime)
			continue
		}

		fmt.Println("services loaded")

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
		// Services not loaded yet
		return
	}

	var lastService string
	var lastMethod string
	var lastResult fritzboxmetrics.Result

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

			var err error
			lastResult, err = action.Call()
			if err != nil {
				log.Printf("could not call action: %v", err)
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

		var floatval float64
		switch tval := val.(type) {
		case uint64:
			floatval = float64(tval)
		case bool:
			if tval {
				floatval = 1
			} else {
				floatval = 0
			}
		case string:
			if tval == m.OkValue {
				floatval = 1
			} else {
				floatval = 0
			}
		default:
			fmt.Println("unknown", val)
			collectErrors.Inc()
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			m.Desc,
			m.MetricType,
			floatval,
			fc.Gateway,
		)
	}
}

func printToStdout(settings *Settings) error {
	root, err := fritzboxmetrics.LoadServices(settings.FritzBoxIP, uint16(settings.FritzBoxPort), settings.FritzBoxUserName, settings.FritzBoxIP)
	if err != nil {
		return errors.Wrap(err, "could not load UPnP service")
	}

	for _, s := range root.Services {
		for _, a := range s.Actions {
			if !a.IsGetOnly() {
				continue
			}

			res, err := a.Call()
			if err != nil {
				log.Printf("unexpected error: %v", err)
				continue
			}

			fmt.Printf("  %s\n", a.Name)
			for _, arg := range a.Arguments {
				fmt.Printf("    %s: %v\n", arg.RelatedStateVariable, res[arg.StateVariable.Name])
			}
		}
	}
	return nil
}

type Settings struct {
	Stdout           bool
	ListenAddr       string
	FritzBoxIP       string
	FritzBoxPort     int
	FritzBoxUserName string
	FritzBoxPassword string
}

func main() {
	settings := &Settings{}
	flag.BoolVar(&settings.Stdout, "stdout", false, "print all available metrics to stdout")
	flag.StringVar(&settings.ListenAddr, "listen-address", ":9133", "The address to listen on for HTTP requests.")

	flag.StringVar(&settings.FritzBoxIP, "gateway-address", "fritz.box", "The hostname or IP of the FRITZ!Box")
	flag.IntVar(&settings.FritzBoxPort, "gateway-port", 49000, "The port of the FRITZ!Box UPnP service")
	flag.StringVar(&settings.FritzBoxUserName, "username", "", "The user for the FRITZ!Box UPnP service")
	flag.StringVar(&settings.FritzBoxPassword, "password", "", "The password for the FRITZ!Box UPnP service")

	flag.Parse()

	if settings.Stdout {
		if err := printToStdout(settings); err != nil {
			log.Printf("could not print metrics to stdout: %v", err)
		}
		return
	}

	collector := &FritzboxCollector{
		Gateway:  settings.FritzBoxIP,
		Port:     uint16(settings.FritzBoxPort),
		Username: settings.FritzBoxUserName,
		Password: settings.FritzBoxPassword,
	}

	go collector.LoadServices()

	prometheus.MustRegister(collector)
	prometheus.MustRegister(collectErrors)

	http.Handle("/metrics", prometheus.Handler())
	log.Fatal(http.ListenAndServe(settings.ListenAddr, nil))
}
