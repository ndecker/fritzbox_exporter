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
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	upnp "github.com/ndecker/fritzbox_exporter/fritzbox_upnp"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	listenAddress := getEnv("FRITZBOX_EXPORTER_LISTEN", ":9133")
	flag.StringVar(&listenAddress, "listen-address", listenAddress, "The address to listen on for HTTP requests.")

	flagMetricsYamlFile := flag.String("metrics", os.Getenv("FRITZBOX_EXPORTER_METRICS"), "YAML file for metrics")

	flagTest := flag.Bool("test-metrics", false, "Test which metrics can be read and print YAML metrics file")

	parameters := upnp.ConnectionParameters{
		Device:          getEnv("FRITZBOX_DEVICE", "fritz.box"),
		Port:            getEnvInt("FRITZBOX_PORT", 49000),
		PortTLS:         getEnvInt("FRITZBOX_PORT_TLS", 49443),
		Username:        getEnv("FRITZBOX_USERNAME", ""),
		Password:        getEnv("FRITZBOX_PASSWORD", ""),
		UseTLS:          getEnv("FRITZBOX_USE_TLS", "true") == "true",
		AllowSelfSigned: getEnv("FRITZBOX_ALLOW_SELFSIGNED", "true") == "true",
	}

	flag.StringVar(&parameters.Device, "gateway-address", parameters.Device, "The hostname or IP of the FRITZ!Box")
	flag.IntVar(&parameters.Port, "gateway-port", parameters.Port, "The port of the FRITZ!Box UPnP service")
	flag.IntVar(&parameters.PortTLS, "gateway-port-tls", parameters.PortTLS, "The TLS port of the FRITZ!Box UPnP service")
	flag.StringVar(&parameters.Username, "username", parameters.Username, "The user for the FRITZ!Box UPnP service")
	flag.StringVar(&parameters.Password, "password", parameters.Password, "The password for the FRITZ!Box UPnP service")
	flag.BoolVar(&parameters.UseTLS, "use-tls", parameters.UseTLS, "Use TLS to connect to FRITZ!Box")
	flag.BoolVar(&parameters.AllowSelfSigned, "allow-selfsigned", parameters.AllowSelfSigned, "Allow selfsigned certificate")

	flag.Parse()

	if *flagTest {
		err := testMetrics(parameters, upnp.IGDServiceDescriptor)
		if err != nil {
			return err
		}

		if parameters.Username == "" {
			log.Fatal("no username/password set for TR64")
		}
		err = testMetrics(parameters, upnp.TR64ServiceDescriptor)
		if err != nil {
			return err
		}
		return nil
	}

	var metricsYaml []byte
	if *flagMetricsYamlFile == "" {
		metricsYaml = defaultMetricsYaml
	} else {
		f, err := os.Open(*flagMetricsYamlFile)
		if err != nil {
			return err
		}

		metricsYaml, err = io.ReadAll(f)
		if err != nil {
			return err
		}
		_ = f.Close()
	}

	metrics, err := loadMetrics(metricsYaml)
	if err != nil {
		return err
	}

	log.Printf("loaded %d metrics", len(metrics))

	collector := NewCollector(parameters, metrics)

	prometheus.MustRegister(collector)
	prometheus.MustRegister(collectMetrics...)

	http.Handle("/metrics", promhttp.Handler())
	return http.ListenAndServe(listenAddress, nil)
}

func getEnv(name string, def string) string {
	env := os.Getenv(name)
	if env != "" {
		return env
	} else {
		return def
	}
}

func getEnvInt(name string, def int) int {
	env := os.Getenv(name)
	if env != "" {
		val, err := strconv.Atoi(env)
		if err != nil {
			log.Fatalf("cannot convert %s to int", env)
		}
		return val
	} else {
		return def
	}
}
