# Fritz!Box Upnp statistics exporter for prometheus

This exporter exports some variables from an 
[AVM Fritzbox](http://avm.de/produkte/fritzbox/)
to prometheus.

## Compatibility

This exporter is known to work with the following models:

| Model             | Firmware   |
|-------------------|------------|
| FRITZ!Box 4040    | 06.83      |
| FRITZ!Box 6490    | 6.51       |
| FRITZ!Box 7390    | 6.51       |
| FRITZ!Box 7490    | 6.51, 7.29 |
| FRITZ!Box 7560    | 6.92       |
| FRITZ!Box 7362 SL | 7.12       |

## Building

### Go install

    go install github.com/ndecker/fritzbox_exporter@latest

### Go build

    git clone https://github.com/ndecker/fritzbox_exporter/
    cd fritzbox_exporter
    go build
    go install

### Docker

    git clone https://github.com/ndecker/fritzbox_exporter/
    docker build -t fritzbox_exporter fritzbox_exporter

## Running

In the configuration of the Fritzbox the option "Statusinformationen über UPnP übertragen" in the dialog "Heimnetz >
Heimnetzübersicht > Netzwerkeinstellungen" has to be enabled.

## Configuration

| command line parameter | environment variable      | default   |                                                |
|------------------------|---------------------------|-----------|------------------------------------------------|
| -listen-address        | FRITZBOX_EXPORTER_LISTEN  | :9133     | The address to listen on for HTTP requests     |
| -gateway-address       | FRITZBOX_DEVICE           | fritz.box | The hostname or IP of the FRITZ!Box            |
| -gateway-port          | FRITZBOX_PORT             | 49000     | The port of the FRITZ!Box UPnP service         |
| -gateway-port          | FRITZBOX_PORT_TLS         | 49443     | The port of the FRITZ!Box TLS UPnP service     |
| -username              | FRITZBOX_USERNAME         |           | The user for the FRITZ!Box UPnP service        |
| -password              | FRITZBOX_PASSWORD         |           | The password for the FRITZ!Box UPnP service    |
| -use-tls               | FRITZBOX_USE_TLS          | true      | Use TLS/HTTPS connection to FRITZ!Box          |
| -allow-selfsigned      | FTITZBOX_ALLOW_SELFSIGNED | true      | Allow selfsigned certificate from FRITZ!Box    |
| -test                  |                           |           | Print all available metrics to stdout and exit |


##  Prerequisites
### FRITZ!OS 7.00+

`Heimnetz` > `Netzwerk` > `Netwerkeinstellungen` > `Statusinformationen über UPnP übertragen`

### FRITZ!OS 6

`Heimnetz` > `Heimnetzübersicht` > `Netzwerkeinstellungen` > `Statusinformationen über UPnP übertragen`

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

      GetFirewallStatus
        FirewallEnabled: true
        InboundPinholeAllowed: false
      GetInfo
        MaxCharsPassword: 32
        MinCharsPassword: 0
        AllowedCharsPassword: 0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz !"#$%&'()*+,-./:;<=>?@[\]^_`{|}~
      X_AVM-DE_GetAnonymousLogin
        X_AVM-DE_AnonymousLoginEnabled: false
        X_AVM-DE_ButtonLoginEnabled: false
      X_AVM-DE_GetUserList
        X_AVM-DE_UserList: <List><Username last_user="1">xxx</Username>
      X_AVM-DE_GetWLANConnectionInfo
        AssociatedDeviceMACAddress: <nil>
        SSID: <nil>
        BSSID: <nil>
        BeaconType: <nil>
        Channel: <nil>
        Standard: <nil>
        X_AVM-DE_SignalStrength: <nil>
        X_AVM-DE_Speed: <nil>
        X_AVM-DE_SpeedRX: <nil>
        X_AVM-DE_SpeedMax: <nil>
        X_AVM-DE_SpeedRXMax: <nil>
      GetHostNumberOfEntries
        HostNumberOfEntries: 5
      X_AVM-DE_GetChangeCounter
        X_AVM-DE_ChangeCounter: 0
      X_AVM-DE_DoUpdate
        UpgradeAvailable: false
        X_AVM-DE_UpdateState: NoUpdate
      GetSecurityPort
        SecurityPort: 49443
      GetAddonInfos
        ByteSendRate: 35
        ByteReceiveRate: 26
        PacketSendRate: 0
        PacketReceiveRate: 0
        TotalBytesSent: 354858561
        TotalBytesReceived: 626712195
        AutoDisconnectTime: 0
        IdleDisconnectTime: 1
        DNSServer1: xxx.xxx.xxx.xxx
        DNSServer2: xxx.xxx.xxx.xxx
        VoipDNSServer1: xxx.xxx.xxx.xxx
        VoipDNSServer2: xxx.xxx.xxx.xxx
        UpnpControlEnabled: false
        RoutedBridgedModeBoth: 1
        X_AVM_DE_TotalBytesSent64: 354858561
        X_AVM_DE_TotalBytesReceived64: 4921679491
        X_AVM_DE_WANAccessType: DSL
      X_AVM_DE_GetDsliteStatus
        X_AVM_DE_DsliteStatus: false
      X_AVM_DE_GetIPTVInfos
        X_AVM_DE_IPTV_Enabled: false
        X_AVM_DE_IPTV_Provider: 
        X_AVM_DE_IPTV_URL: 
      GetCommonLinkProperties
        WANAccessType: DSL
        Layer1UpstreamMaxBitRate: 33251000
        Layer1DownstreamMaxBitRate: 114110000
        PhysicalLinkStatus: Up
      GetTotalBytesSent
        TotalBytesSent: 354858561
      GetTotalBytesReceived
        TotalBytesReceived: 626712195
      GetTotalPacketsSent
        TotalPacketsSent: 245896
      GetTotalPacketsReceived
        TotalPacketsReceived: 69248
      X_AVM_DE_GetDNSServer
        IPv4DNSServer1: xxx.xxx.xxx.xxx
        IPv4DNSServer2: xxx.xxx.xxx.xxx
      GetIdleDisconnectTime
        IdleDisconnectTime: 0
      GetStatusInfo
        ConnectionStatus: Connected
        LastConnectionError: ERROR_NONE
        Uptime: 70993
      GetNATRSIPStatus
        RSIPAvailable: false
        NATEnabled: true
      GetAutoDisconnectTime
        AutoDisconnectTime: 0
      X_AVM_DE_GetIPv6Prefix
        IPv6Prefix: 
        PrefixLength: 0
        ValidLifetime: 0
        PreferedLifetime: 0
      GetConnectionTypeInfo
        ConnectionType: IP_Routed
        PossibleConnectionTypes: IP_Routed
      X_AVM_DE_GetExternalIPv6Address
        ExternalIPv6Address: 
        PrefixLength: 0
        ValidLifetime: 0
        PreferedLifetime: 0
      X_AVM_DE_GetIPv6DNSServer
        IPv6DNSServer1: 
        ValidLifetime1: 0
        IPv6DNSServer2: 
        ValidLifetime2: 2003335812
      GetExternalIPAddress
        ExternalIPAddress: xxx.xxx.xxx.xxx
      X_AVM-DE_GetNightControl
        NightControl: <rule id="0" enabled="0"><item day="127" time="2100" action="0" /><item
    day="127" time="2300" action="1" /></rule>

        NightTimeControlNoForcedOff: false
      X_AVM-DE_GetWLANConnectionInfo
        AssociatedDeviceMACAddress: <nil>
        SSID: <nil>
        BSSID: <nil>
        BeaconType: <nil>
        Channel: <nil>
        Standard: <nil>
        X_AVM-DE_SignalStrength: <nil>
        X_AVM-DE_Speed: <nil>
        X_AVM-DE_SpeedRX: <nil>
        X_AVM-DE_SpeedMax: <nil>
        X_AVM-DE_SpeedRXMax: <nil>
      X_AVM-DE_GetWLANConnectionInfo
        AssociatedDeviceMACAddress: <nil>
        SSID: <nil>
        BSSID: <nil>
        BeaconType: <nil>
        Channel: <nil>
        Standard: <nil>
        X_AVM-DE_SignalStrength: <nil>
        X_AVM-DE_Speed: <nil>
        X_AVM-DE_SpeedRX: <nil>
        X_AVM-DE_SpeedMax: <nil>
        X_AVM-DE_SpeedRXMax: <nil>
      GetAutoConfig
        AutoConfig: false
      GetModulationType
        ModulationType: 
      GetDSLLinkInfo
        LinkType: PPPoE
        LinkStatus: Up
      GetATMEncapsulation
        ATMEncapsulation: 
      GetFCSPreserved
        FCSPreserved: false
      GetDestinationAddress
        DestinationAddress: 
