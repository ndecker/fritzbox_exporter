package fritzbox_upnp

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Action is an UPNP Action on a service
type Action struct {
	service *Service

	Name        string               `xml:"name"`
	Arguments   []*Argument          `xml:"argumentList>argument"`
	ArgumentMap map[string]*Argument // Map of arguments indexed by .Name
}

// IsGetOnly returns if the action seems to be a query for information.
// This is determined by checking if the action has no input arguments and at least one output argument.
func (a *Action) IsGetOnly() bool {
	for _, a := range a.Arguments {
		if a.Direction == "in" {
			return false
		}
	}
	return len(a.Arguments) > 0
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

	url := a.service.Device.root.baseUrl + a.service.ControlUrl
	body := strings.NewReader(bodystr)

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	action := fmt.Sprintf("%s#%s", a.service.ServiceType, a.Name)

	req.Header.Set("Content-Type", textXml)
	req.Header.Set("SoapAction", action)

	client := a.service.Device.root.client

	resp, err := client.Transport.RoundTrip(req)
	if err != nil {
		return nil, fmt.Errorf("cannod call %s: %w", a.Name, err)
	}
	defer closeIgnoringError(resp.Body)

	if resp.StatusCode == 401 {
		return nil, fmt.Errorf("cannot read service %s: status 401 unauthorized", a.Name)
	}

	data, err := io.ReadAll(resp.Body)
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
		return val == "1", nil

	case "ui1", "ui2", "ui4":
		// type ui4 can contain values greater than 2^32!
		res, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return nil, err
		}
		return res, nil
	case "i1", "i2", "i4":
		res, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return nil, err
		}
		return res, nil
	case "dateTime":
		const timeLayout = "2006-01-02T15:04:05"
		res, err := time.Parse(timeLayout, val)
		if err != nil {
			return nil, err
		}
		return res, nil
	default:
		return nil, fmt.Errorf("unknown datatype: %s: %s", arg.StateVariable.DataType, val)

	}
}
