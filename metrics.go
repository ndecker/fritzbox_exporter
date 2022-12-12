package main

import (
	_ "embed"
	"fmt"
	upnp "github.com/ndecker/fritzbox_exporter/fritzbox_upnp"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/yaml.v3"
	"io"
	"log"
	"os"
	"strings"
)

//go:embed default-metrics.yaml
var defaultMetricsYaml []byte

type Metric struct {
	Metric string
	Help   string
	Type   string

	Service   string
	Action    string
	Result    string
	OkValue   string `yaml:",omitempty"`
	LabelName string `yaml:",omitempty"`

	Source       string `yaml:",omitempty"`
	ExampleValue string `yaml:",omitempty"`

	metricType prometheus.ValueType
	desc       *prometheus.Desc
}

func (m *Metric) String() string {
	var res strings.Builder
	if m.Metric != "" {
		res.WriteString(fmt.Sprintf("%s: ", m.Metric))
	}

	res.WriteString(fmt.Sprintf("%s/%s/%s", m.Service, m.Action, m.Result))

	return res.String()
}

func loadMetrics(data []byte) ([]*Metric, error) {
	var metrics []*Metric

	err := yaml.Unmarshal(data, &metrics)
	if err != nil {
		return nil, err
	}

	// Filter valid metrics
	var metrics2 []*Metric
	for _, m := range metrics {
		if m.Metric == "" {
			log.Printf("skipping metric %s: no metric name\n", m)
			continue
		}

		switch m.Type {
		case "counter":
			m.metricType = prometheus.CounterValue
		case "gauge":
			m.metricType = prometheus.GaugeValue
		default:
			log.Printf("skipping metric %s: invalid metric type: %s", m, m.Type)
			continue
		}

		labels := []string{"gateway"}
		if m.LabelName != "" {
			labels = append(labels, m.LabelName)
		}

		m.desc = prometheus.NewDesc(m.Metric, m.Help, labels, nil)
		metrics2 = append(metrics2, m)
	}

	return metrics2, nil
}

func writeMetrics(w io.Writer, metrics []*Metric) error {
	data, err := yaml.Marshal(metrics)
	if err != nil {
		return err
	}

	_, err = w.Write(data)
	return err
}

func testMetrics(p upnp.ConnectionParameters, desc string) error {
	root, err := upnp.LoadServiceRoot(p, desc)
	if err != nil {
		return err
	}

	var metrics []*Metric

	for _, s := range root.Services {
		for _, a := range s.Actions {
			if !a.IsGetOnly() {
				continue
			}

			res, err := a.Call()
			if err != nil {
				log.Printf("unexpected error: %v\n", err)
				continue
			}

			for _, arg := range a.Arguments {
				value := res[arg.StateVariable.Name]

				m := &Metric{
					Metric:       "",
					Help:         "",
					Type:         "",
					Service:      s.ServiceType,
					Action:       a.Name,
					Result:       arg.StateVariable.Name,
					ExampleValue: fmt.Sprintf("%v", value),
					OkValue:      "",
					Source:       desc,
				}
				metrics = append(metrics, m)
			}
		}
	}

	return writeMetrics(os.Stdout, metrics)
}
