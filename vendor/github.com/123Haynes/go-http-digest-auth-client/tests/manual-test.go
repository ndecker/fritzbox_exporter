package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	dac "github.com/xinsnake/go-http-digest-auth-client"
)

const (
	username = "test"
	password = "test123"
	method   = "GET"
	uri      = "http://172.16.1.5"
)

func main() {
	var resp *http.Response
	var body []byte
	var err error

	dr := dac.NewRequest(username, password, method, uri, "")

	if resp, err = dr.Execute(); err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		log.Fatalln(err)
	}

	fmt.Printf(string(body))
}
