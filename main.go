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
	upnp "github.com/ndecker/fritzbox_exporter/fritzbox_upnp"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"os"
	"strconv"
)

func main() {
	listenAddress := GetEnv("FRITZBOX_EXPORTER_LISTEN", ":9133")
	flag.StringVar(&listenAddress, "listen-address", listenAddress, "The address to listen on for HTTP requests.")

	flagTestIGD := flag.Bool("test-igd", false, "print all available IGD metrics to stdout")
	flagTestTR64 := flag.Bool("test-tr64", false, "print all available TR64 metrics to stdout")

	parameters := upnp.ConnectionParameters{
		Device:          GetEnv("FRITZBOX_DEVICE", "fritz.box"),
		Port:            GetEnvInt("FRITZBOX_PORT", 49000),
		PortTLS:         GetEnvInt("FRITZBOX_PORT_TLS", 49443),
		Username:        GetEnv("FRITZBOX_USERNAME", ""),
		Password:        GetEnv("FRITZBOX_PASSWORD", ""),
		UseTLS:          GetEnv("FRITZBOX_USE_TLS", "true") == "true",
		AllowSelfSigned: GetEnv("FRITZBOX_ALLOW_SELFSIGNED", "true") == "true",
	}

	flag.StringVar(&parameters.Device, "gateway-address", parameters.Device, "The hostname or IP of the FRITZ!Box")
	flag.IntVar(&parameters.Port, "gateway-port", parameters.Port, "The port of the FRITZ!Box UPnP service")
	flag.IntVar(&parameters.PortTLS, "gateway-port-tls", parameters.PortTLS, "The TLS port of the FRITZ!Box UPnP service")
	flag.StringVar(&parameters.Username, "username", parameters.Username, "The user for the FRITZ!Box UPnP service")
	flag.StringVar(&parameters.Password, "password", parameters.Password, "The password for the FRITZ!Box UPnP service")
	flag.BoolVar(&parameters.UseTLS, "use-tls", parameters.UseTLS, "Use TLS to connect to FRITZ!Box")
	flag.BoolVar(&parameters.AllowSelfSigned, "allow-selfsigned", parameters.AllowSelfSigned, "Allow selfsigned certificate")

	flag.Parse()

	if *flagTestIGD {
		test(parameters, upnp.IGDServiceDescriptor)
		return
	}
	if *flagTestTR64 {
		if parameters.Username == "" {
			log.Fatal("no username/password set for TR64")
		}
		test(parameters, upnp.TR64ServiceDescriptor)
		return
	}

	collector := &FritzboxCollector{
		Parameters: parameters,
	}

	go collector.LoadServices()

	prometheus.MustRegister(collector)
	prometheus.MustRegister(collectErrors)

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(listenAddress, nil))
}

func test(p upnp.ConnectionParameters, desc string) {
	root, err := upnp.LoadServiceRoot(p, desc)
	if err != nil {
		panic(err)
	}

	for _, s := range root.Services {
		fmt.Println(s.SCPDUrl)
		for _, a := range s.Actions {
			if !a.IsGetOnly() {
				continue
			}

			res, err := a.Call()
			if err != nil {
				log.Printf("unexpected error: %v\n", err)
				continue
			}

			fmt.Printf("  %s\n", a.Name)
			for _, arg := range a.Arguments {
				fmt.Printf("    %s: %v\n", arg.RelatedStateVariable, res[arg.StateVariable.Name])
			}
		}
	}
}

func GetEnv(name string, def string) string {
	env := os.Getenv(name)
	if env != "" {
		return env
	} else {
		return def
	}
}

func GetEnvInt(name string, def int) int {
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
