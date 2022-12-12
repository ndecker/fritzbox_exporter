# FRITZ!Box Upnp statistics exporter for prometheus

This exporter exports some variables from an [AVM Fritzbox](http://avm.de/produkte/fritzbox/) to prometheus.

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

##  Prerequisites
In the configuration of the Fritzbox the option "Statusinformationen über UPnP übertragen" has to be enabled.

### FRITZ!OS 7.00+

`Heimnetz` > `Netzwerk` > `Netwerkeinstellungen` > `Statusinformationen über UPnP übertragen`

### FRITZ!OS 6

`Heimnetz` > `Heimnetzübersicht` > `Netzwerkeinstellungen` > `Statusinformationen über UPnP übertragen`


## Configuration

| command line parameter | environment variable      | default    |                                                            |
|------------------------|---------------------------|------------|------------------------------------------------------------|
| -metrics               | FRITZBOX_EXPORTER_METRICS | <internal> | YAML file describing exported metrics                      |
| -test-metrics          |                           |            | Test which metrics can be read and print YAML metrics file |
| -listen-address        | FRITZBOX_EXPORTER_LISTEN  | :9133      | The address to listen on for HTTP requests                 |
| -gateway-address       | FRITZBOX_DEVICE           | fritz.box  | The hostname or IP of the FRITZ!Box                        |
| -gateway-port          | FRITZBOX_PORT             | 49000      | The port of the FRITZ!Box UPnP service                     |
| -gateway-port          | FRITZBOX_PORT_TLS         | 49443      | The port of the FRITZ!Box TLS UPnP service                 |
| -username              | FRITZBOX_USERNAME         |            | The user for the FRITZ!Box UPnP service                    |
| -password              | FRITZBOX_PASSWORD         |            | The password for the FRITZ!Box UPnP service                |
| -use-tls               | FRITZBOX_USE_TLS          | true       | Use TLS/HTTPS connection to FRITZ!Box                      |
| -allow-selfsigned      | FRITZBOX_ALLOW_SELFSIGNED | true       | Allow selfsigned certificate from FRITZ!Box                |


## Exported metrics

The default metrics to be exported are described in [default-metrics.yaml](default-metrics.yaml).
This file is compiled into the binary and used by default.
With the `-metrics` option a different file can be specified.

With the `-test-metrics` option all possible metrics of the FRITZ!Box can be queried. This can take a few minutes.
For TR64 metrics username/password must be provided.

     fritzbox_exporter -test-metrics > metrics.yaml
     edit metrics.yaml
     fritzbox_exporter -metrics metrics.yaml

### Examples

This is an example metric as exported by `-test-metrics`

    - metric: ""    # prometheus metric name (required)
      help: ""      # prometheus help text
      type: ""      # metric type: gauge, counter
      service: urn:schemas-upnp-org:service:WANCommonInterfaceConfig:1
      action: GetTotalBytesReceived
      result: TotalBytesReceived
      examplevalue: "325538505" # current value of the metric. Only for info; not used
      source: "igddesc.xml"     # source of the value (iggdesc.xml or tr64desc.xml). Only for info; not used


If you wanted to for example to monitor the number of hosts in you local network, you could use this:

    - metric: "gateway_number_of_hosts"
      help: "Number of hosts in the local network"
      type: "gauge"
      service: urn:dslforum-org:service:Hosts:1
      action: GetHostNumberOfEntries
      result: HostNumberOfEntries

### Metrics with `okvalue`

If the value is a string you can use the `okvalue` field to specify a value to compare the string to.
The metric will be 1 if the value matches okvalue; 0 otherwise.

    - metric: gateway_wan_layer1_link_status
      help: Status of physical link (Up = 1)
      type: gauge
      service: urn:schemas-upnp-org:service:WANCommonInterfaceConfig:1
      action: GetCommonLinkProperties
      result: PhysicalLinkStatus
      okvalue: Up

### Metrics with `labelname`

You can specify `labelname` to set a metric label with the value. The metric value will always be 1.
The following will give a metric `gateway_version{device="fritz.box", version="113.07.29"} = 1`

    - metric: "gateway_version"
      type: "gauge"
      service: urn:dslforum-org:service:DeviceInfo:1
      action: GetInfo
      result: SoftwareVersion
      labelname: version
      source: tr64desc.xml
