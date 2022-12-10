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
)

func main() {
	flagTest := flag.Bool("test", false, "print all available metrics to stdout")
	flagAddr := flag.String("listen-address", ":9133", "The address to listen on for HTTP requests.")

	var parameters upnp.ConnectionParameters
	flag.StringVar(&parameters.Device, "gateway-address", "fritz.box", "The hostname or IP of the FRITZ!Box")
	flag.IntVar(&parameters.Port, "gateway-port", 49000, "The port of the FRITZ!Box UPnP service")
	flag.StringVar(&parameters.Username, "username", "", "The user for the FRITZ!Box UPnP service")
	flag.StringVar(&parameters.Password, "password", "", "The password for the FRITZ!Box UPnP service")

	flag.Parse()

	if *flagTest {
		test(parameters)
		return
	}

	collector := &FritzboxCollector{
		Parameters: parameters,
	}

	go collector.LoadServices()

	prometheus.MustRegister(collector)
	prometheus.MustRegister(collectErrors)

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(*flagAddr, nil))
}

func test(p upnp.ConnectionParameters) {
	root, err := upnp.LoadServices(p)
	if err != nil {
		panic(err)
	}

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

			fmt.Printf("  %s\n", a.Name)
			for _, arg := range a.Arguments {
				fmt.Printf("    %s: %v\n", arg.RelatedStateVariable, res[arg.StateVariable.Name])
			}
		}
	}
}
