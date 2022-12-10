package main

import "github.com/prometheus/client_golang/prometheus"

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
