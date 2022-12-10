// Query UPNP variables from Fritz!Box devices.
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
	"log"
	"net/http"
	"strconv"
	"strings"

	dac "github.com/123Haynes/go-http-digest-auth-client"
)

// curl http://fritz.box:49000/igddesc.xml
// curl http://fritz.box:49000/any.xml
// curl http://fritz.box:49000/igdconnSCPD.xml
// curl http://fritz.box:49000/igdicfgSCPD.xml
// curl http://fritz.box:49000/igddslSCPD.xml
// curl http://fritz.box:49000/igd2ipv6fwcSCPD.xml

// For TR64: curl http://fritz.box:49000/tr64desc.xmll

const text_xml = `text/xml; charset="utf-8"`

var ErrInvalidSOAPResponse = errors.New("invalid SOAP response")

// Root of the UPNP tree
type Root struct {
	BaseUrl  string
	Username string
	Password string
	Device   Device              `xml:"device"`
	Services map[string]*Service // Map of all services indexed by .ServiceType
}

// An UPNP Device
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

// An UPNP Service
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

// An UPNP Acton on a service
type Action struct {
	service *Service

	Name        string               `xml:"name"`
	Arguments   []*Argument          `xml:"argumentList>argument"`
	ArgumentMap map[string]*Argument // Map of arguments indexed by .Name
}

// Returns if the action seems to be a query for information.
// This is determined by checking if the action has no input arguments and at least one output argument.
func (a *Action) IsGetOnly() bool {
	for _, a := range a.Arguments {
		if a.Direction == "in" {
			return false
		}
	}
	return len(a.Arguments) > 0

	return false

}

// An Argument to an action
type Argument struct {
	Name                 string `xml:"name"`
	Direction            string `xml:"direction"`
	RelatedStateVariable string `xml:"relatedStateVariable"`
	StateVariable        *StateVariable
}

// A state variable that can be manipulated through actions
type StateVariable struct {
	Name         string `xml:"name"`
	DataType     string `xml:"dataType"`
	DefaultValue string `xml:"defaultValue"`
}

// The result of a Call() contains all output arguments of the call.
// The map is indexed by the name of the state variable.
// The type of the value is string, uint64 or bool depending of the DataType of the variable.
type Result map[string]interface{}

// load the whole tree
func (r *Root) load() error {
	igddesc, err := http.Get(
		fmt.Sprintf("%s/igddesc.xml", r.BaseUrl),
	)

	if err != nil {
		return err
	}

	dec := xml.NewDecoder(igddesc.Body)

	err = dec.Decode(r)
	if err != nil {
		return fmt.Errorf("failed to decode igdesc.xml: %w", err)
	}

	r.Services = make(map[string]*Service)
	return r.Device.fillServices(r)
}

func (r *Root) loadTr64() error {
	igddesc, err := http.Get(
		fmt.Sprintf("%s/tr64desc.xml", r.BaseUrl),
	)

	if err != nil {
		return fmt.Errorf("failed to decode tr64desc.xml: %w", err)
	}

	dec := xml.NewDecoder(igddesc.Body)

	err = dec.Decode(r)
	if err != nil {
		return err
	}

	r.Services = make(map[string]*Service)
	return r.Device.fillServices(r)
}

// load all service descriptions
func (d *Device) fillServices(r *Root) error {
	d.root = r

	for _, s := range d.Services {
		s.Device = d

		response, err := http.Get(r.BaseUrl + s.SCPDUrl)
		if err != nil {
			return err
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

// Call an action.
// Currently only actions without input arguments are supported.
func (a *Action) Call() (Result, error) {
	bodystr := fmt.Sprintf(`
        <?xml version='1.0' encoding='utf-8'?>
        <s:Envelope s:encodingStyle='http://schemas.xmlsoap.org/soap/encoding/' xmlns:s='http://schemas.xmlsoap.org/soap/envelope/'>
            <s:Body>
                <u:%s xmlns:u='%s' />
            </s:Body>
        </s:Envelope>
    `, a.Name, a.service.ServiceType)

	url := a.service.Device.root.BaseUrl + a.service.ControlUrl
	body := strings.NewReader(bodystr)

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	action := fmt.Sprintf("%s#%s", a.service.ServiceType, a.Name)

	req.Header.Set("Content-Type", text_xml)
	req.Header.Set("SoapAction", action)

	t := dac.NewTransport(a.service.Device.root.Username, a.service.Device.root.Password)

	resp, err := t.RoundTrip(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		return nil, fmt.Errorf("cannot read service %s: status 401 unauthorized", a.Name)
	}

	data,err  := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot read request body: %w", err)
	}

	return a.parseSoapResponse(data)

}

func (a *Action) parseSoapResponse(data []byte) (Result, error) {
	res := make(Result)
	dec := xml.NewDecoder(bytes.NewReader(data))

	for {
		t, err := dec.Token()
		if err == io.EOF {
			return res, nil
		}

		if err != nil {
			return nil, fmt.Errorf("cannot parse soap response: %w; body: %s", err, data)
		}

		if se, ok := t.(xml.StartElement); ok {
			arg, ok := a.ArgumentMap[se.Name.Local]

			if ok {
				t2, err := dec.Token()
				if err != nil {
					return nil, err
				}

				var val string
				switch element := t2.(type) {
				case xml.EndElement:
					val = ""
				case xml.CharData:
					val = string(element)
				default:
					return nil, ErrInvalidSOAPResponse
				}

				converted, err := convertResult(val, arg)
				if err != nil {
					return nil, err
				}
				res[arg.StateVariable.Name] = converted
			}
		}

	}
}

func convertResult(val string, arg *Argument) (interface{}, error) {
	switch arg.StateVariable.DataType {
	case "string":
		return val, nil
	case "boolean":
		return bool(val == "1"), nil

	case "ui1", "ui2", "ui4":
		// type ui4 can contain values greater than 2^32!
		res, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return nil, err
		}
		return uint64(res), nil
	default:
		return nil, fmt.Errorf("unknown datatype: %s", arg.StateVariable.DataType)

	}
}

// Load the services tree from an device.
func LoadServices(device string, port uint16, username string, password string) (*Root, error) {
	var root = &Root{
		BaseUrl:  fmt.Sprintf("http://%s:%d", device, port),
		Username: username,
		Password: password,
	}

	err := root.load()
	if err != nil {
		return nil, fmt.Errorf("cannot load services: %w", err)
	}

	var rootTr64 = &Root{
		BaseUrl:  fmt.Sprintf("http://%s:%d", device, port),
		Username: username,
		Password: password,
	}

	err = rootTr64.loadTr64()
	if err != nil {
		return nil, fmt.Errorf("cannot load TR64 services: %w", err)
	}

	for k, v := range rootTr64.Services {
		root.Services[k] = v
	}

	return root, nil
}
