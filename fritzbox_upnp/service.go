// Package fritzbox_upnp queries UPNP variables from Fritz!Box devices.
package fritzbox_upnp

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
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// curl http://fritz.box:49000/igddesc.xml
// curl http://fritz.box:49000/tr64desc.xmll

const textXml = `text/xml; charset="utf-8"`

const (
	IGDServiceDescriptor  = "igddesc.xml"
	TR64ServiceDescriptor = "tr64desc.xml"
)

var ErrInvalidSOAPResponse = errors.New("invalid SOAP response")

type ConnectionParameters struct {
	Device          string // Hostname or IP
	Port            int
	PortTLS         int
	UseTLS          bool
	Username        string
	Password        string
	AllowSelfSigned bool
}

// Root of the UPNP tree
type Root struct {
	client   *http.Client
	baseUrl  string
	params   ConnectionParameters
	Device   Device              `xml:"device"`
	Services map[string]*Service // Map of all services indexed by .ServiceType
}

// Device represents a UPNP Device
type Device struct {
	root *Root

	DeviceType       string `xml:"deviceType"`
	FriendlyName     string `xml:"friendlyName"`
	Manufacturer     string `xml:"manufacturer"`
	ManufacturerUrl  string `xml:"manufacturerURL"`
	ModelDescription string `xml:"modelDescription"`
	ModelName        string `xml:"modelName"`
	ModelNumber      string `xml:"modelNumber"`
	ModelUrl         string `xml:"modelURL"`
	UDN              string `xml:"UDN"`

	Services []*Service `xml:"serviceList>service"` // Service of the device
	Devices  []*Device  `xml:"deviceList>device"`   // Sub-Devices of the device

	PresentationUrl string `xml:"presentationURL"`
}

// Service represents a UPNP Service
type Service struct {
	Device *Device

	ServiceType string `xml:"serviceType"`
	ServiceId   string `xml:"serviceId"`
	ControlUrl  string `xml:"controlURL"`
	EventSubUrl string `xml:"eventSubURL"`
	SCPDUrl     string `xml:"SCPDURL"`

	Actions        map[string]*Action // All actions available on the service
	StateVariables []*StateVariable   // All state variables available on the service
}

type scpdRoot struct {
	Actions        []*Action        `xml:"actionList>action"`
	StateVariables []*StateVariable `xml:"serviceStateTable>stateVariable"`
}

// load all service descriptions
func (d *Device) fillServices(r *Root) error {
	d.root = r

	for _, s := range d.Services {
		s.Device = d

		response, err := r.client.Get(r.baseUrl + s.SCPDUrl)
		if err != nil {
			return fmt.Errorf("cannot load services for %s: %w", s.SCPDUrl, err)
		}

		var scpd scpdRoot
		dec := xml.NewDecoder(response.Body)
		err = dec.Decode(&scpd)
		if err != nil {
			return err
		}

		s.Actions = make(map[string]*Action)
		for _, a := range scpd.Actions {
			s.Actions[a.Name] = a
		}
		s.StateVariables = scpd.StateVariables

		for _, a := range s.Actions {
			a.service = s
			a.ArgumentMap = make(map[string]*Argument)

			for _, arg := range a.Arguments {
				for _, svar := range s.StateVariables {
					if arg.RelatedStateVariable == svar.Name {
						arg.StateVariable = svar
					}
				}

				a.ArgumentMap[arg.Name] = arg
			}
		}

		r.Services[s.ServiceType] = s
	}
	for _, d2 := range d.Devices {
		err := d2.fillServices(r)
		if err != nil {
			return err
		}
	}
	return nil
}

// LoadServiceRoot loads a service descriptor and populates a Service Root
func LoadServiceRoot(params ConnectionParameters, descriptor string) (*Root, error) {
	var baseUrl string
	if params.UseTLS {
		baseUrl = fmt.Sprintf("https://%s:%d", params.Device, params.PortTLS)
	} else {
		baseUrl = fmt.Sprintf("http://%s:%d", params.Device, params.Port)

	}

	var root = &Root{
		params:   params,
		client:   setupClient(params.Username, params.Password, params.AllowSelfSigned),
		baseUrl:  baseUrl,
		Services: make(map[string]*Service),
	}

	descUrl, err := url.JoinPath(root.baseUrl, descriptor)
	if err != nil {
		return nil, err
	}

	igddesc, err := root.client.Get(descUrl)
	if err != nil {
		return nil, err
	}
	defer closeIgnoringError(igddesc.Body)

	if igddesc.StatusCode == 404 {
		return nil, fmt.Errorf("http error 401 when loading service description. Is UPnP activated? (see Readme)")
	}

	body, err := io.ReadAll(igddesc.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	dec := xml.NewDecoder(bytes.NewReader(body))

	err = dec.Decode(root)
	if err != nil {
		return nil, fmt.Errorf("failed to decode igdesc.xml: %w; body: %s", err, body)
	}

	err = root.Device.fillServices(root)
	if err != nil {
		return nil, err
	}

	return root, nil
}

// closeIgnoringError closes c an ignores errors
func closeIgnoringError(c io.Closer) {
	_ = c.Close()
}
