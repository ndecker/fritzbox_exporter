# FRITZ!Box Upnp statistics exporter for Prometheus

This exporter exports some variables from an [AVM Fritzbox](https://avm.de/produkte/fritzbox/) to Prometheus.


[![Docker Build Status](https://img.shields.io/docker/build/mxschmitt/fritzbox_exporter.svg)](https://store.docker.com/community/images/mxschmitt/fritzbox_exporter)
[![GoDoc](https://godoc.org/github.com/mxschmitt/fritzbox_exporter/pkg/fritzboxmetrics?status.svg)](https://godoc.org/github.com/mxschmitt/fritzbox_exporter/pkg/fritzboxmetrics)

<img src="./docs/fritzbox.svg" alt="FRITZ!Box" height="100px"><img src="./docs/prometheus.svg" alt="FRITZ!Box" height="100px">

## Compatibility

This exporter is known to work with the following models:

| Model          | Firmware |
|----------------|----------|
| FRITZ!Box 4040 | 06.83    |
| FRITZ!Box 7390 | 06.51    |
| FRITZ!Box 7490 | 06.51    |
| FRITZ!Box 6490 | 06.52    |
| FRITZ!Box 7560 | 06.92    |

**FRITZ!OS 7 is not fully supported!**

## Building

```bash
go get github.com/mxschmitt/fritzbox_exporter
cd $GOPATH/src/github.com/mxschmitt/fritzbox_exporter/cmd/exporter
go get ./...
go build
```

## Prerequisites

There has to be UPnP enabled.

## FRITZ!OS 7.00+

`Heimnetz` > `Netzwerk` > `Netwerkeinstellungen` > `Statusinformationen über UPnP übertragen`

## FRITZ!OS 6

`Heimnetz` > `Heimnetzübersicht` > `Netzwerkeinstellungen` > `Statusinformationen über UPnP übertragen`

# Usage

    $GOPATH/src/github.com/mxschmitt/fritzbox_exporter/cmd/exporter/exporter -h
    Usage $GOPATH/src/github.com/mxschmitt/fritzbox_exporter/cmd/exporter/exporter:
      -gateway-address string
        	The hostname or IP of the FRITZ!Box (default "fritz.box")
      -gateway-port int
        	The port of the FRITZ!Box UPnP service (default 49000)
      -listen-address string
        	The address to listen on for HTTP requests. (default ":9133")
      -test
        	print all available metrics to stdout

## Exported metrics

These metrics are exported:

    # HELP fritzbox_exporter_collect_errors Number of collection errors.
    # TYPE fritzbox_exporter_collect_errors counter
    fritzbox_exporter_collect_errors 0
    # HELP gateway_wan_bytes_received bytes received on gateway WAN interface
    # TYPE gateway_wan_bytes_received counter
    gateway_wan_bytes_received{gateway="fritz.box"} 5.037749914e+09
    # HELP gateway_wan_bytes_sent bytes sent on gateway WAN interface
    # TYPE gateway_wan_bytes_sent counter
    gateway_wan_bytes_sent{gateway="fritz.box"} 2.55707479e+08
    # HELP gateway_wan_connection_status WAN connection status (Connected = 1)
    # TYPE gateway_wan_connection_status gauge
    gateway_wan_connection_status{gateway="fritz.box"} 1
    # HELP gateway_wan_connection_uptime_seconds WAN connection uptime
    # TYPE gateway_wan_connection_uptime_seconds gauge
    gateway_wan_connection_uptime_seconds{gateway="fritz.box"} 65259
    # HELP gateway_wan_layer1_downstream_max_bitrate Layer1 downstream max bitrate
    # TYPE gateway_wan_layer1_downstream_max_bitrate gauge
    gateway_wan_layer1_downstream_max_bitrate{gateway="fritz.box"} 1.286e+07
    # HELP gateway_wan_layer1_link_status Status of physical link (Up = 1)
    # TYPE gateway_wan_layer1_link_status gauge
    gateway_wan_layer1_link_status{gateway="fritz.box"} 1
    # HELP gateway_wan_layer1_upstream_max_bitrate Layer1 upstream max bitrate
    # TYPE gateway_wan_layer1_upstream_max_bitrate gauge
    gateway_wan_layer1_upstream_max_bitrate{gateway="fritz.box"} 1.148e+06
    # HELP gateway_wan_packets_received packets received on gateway WAN interface
    # TYPE gateway_wan_packets_received counter
    gateway_wan_packets_received{gateway="fritz.box"} 1.346625e+06
    # HELP gateway_wan_packets_sent packets sent on gateway WAN interface
    # TYPE gateway_wan_packets_sent counter
    gateway_wan_packets_sent{gateway="fritz.box"} 3.05051e+06


## Output of -test

The exporter prints all available Variables to stdout when called with the -test option.
These values are determined by parsing all services from http://fritz.box:49000/igddesc.xml 

    Name: urn:schemas-any-com:service:Any:1
    WANDevice - FRITZ!Box 7490: urn:schemas-upnp-org:service:WANCommonInterfaceConfig:1
      GetCommonLinkProperties
        WANAccessType: DSL
        Layer1UpstreamMaxBitRate: 1148000
        Layer1DownstreamMaxBitRate: 12860000
        PhysicalLinkStatus: Up
      GetTotalBytesSent
        TotalBytesSent: 255710914
      GetTotalBytesReceived
        TotalBytesReceived: 5037753042
      GetTotalPacketsSent
        TotalPacketsSent: 3050536
      GetTotalPacketsReceived
        TotalPacketsReceived: 1346651
      GetAddonInfos
        ByteSendRate: 0
        ByteReceiveRate: 0
        PacketSendRate: 0
        PacketReceiveRate: 0
        TotalBytesSent: 255710914
        TotalBytesReceived: 5037753042
        AutoDisconnectTime: 0
        IdleDisconnectTime: 10
        DNSServer1: 1.1.1.1
        DNSServer2: 2.2.2.2
        VoipDNSServer1: 1.1.1.1
        VoipDNSServer2: 2.2.2.2
        UpnpControlEnabled: false
        RoutedBridgedModeBoth: 1
    WANConnectionDevice - FRITZ!Box 7490: urn:schemas-upnp-org:service:WANDSLLinkConfig:1
      GetDSLLinkInfo
        LinkType: PPPoE
        LinkStatus: Up
      GetModulationType
        ModulationType: ADSL G.lite
      GetDestinationAddress
        DestinationAddress: NONE
      GetATMEncapsulation
        ATMEncapsulation: LLC
      GetFCSPreserved
        FCSPreserved: true
      GetAutoConfig
        AutoConfig: true
    WANConnectionDevice - FRITZ!Box 7490: urn:schemas-upnp-org:service:WANIPConnection:1
      X_AVM_DE_GetDNSServer
        IPv4DNSServer1: 1.1.1.1
        IPv4DNSServer2: 2.2.2.2
      GetAutoDisconnectTime
        AutoDisconnectTime: 0
      GetIdleDisconnectTime
        IdleDisconnectTime: 0
      X_AVM_DE_GetExternalIPv6Address
        ExternalIPv6Address: 
        PrefixLength: 0
        ValidLifetime: 0
        PreferedLifetime: 0
      GetNATRSIPStatus
        RSIPAvailable: false
        NATEnabled: true
      GetExternalIPAddress
        ExternalIPAddress: 1.1.1.1
      X_AVM_DE_GetIPv6Prefix
        IPv6Prefix: 
        PrefixLength: 0
        ValidLifetime: 0
        PreferedLifetime: 0
      X_AVM_DE_GetIPv6DNSServer
        IPv6DNSServer1: 
        ValidLifetime1: 2002000000
        IPv6DNSServer2: 
        ValidLifetime2: 199800000
      GetConnectionTypeInfo
        ConnectionType: IP_Routed
        PossibleConnectionTypes: IP_Routed
      GetStatusInfo
        ConnectionStatus: Connected
        LastConnectionError: ERROR_NONE
        Uptime: 65386
    WANConnectionDevice - FRITZ!Box 7490: urn:schemas-upnp-org:service:WANIPv6FirewallControl:1
      GetFirewallStatus
        FirewallEnabled: true
        InboundPinholeAllowed: false

