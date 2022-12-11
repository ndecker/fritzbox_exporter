package fritzbox_upnp

import (
	"crypto/tls"
	"net/http"

	dac "github.com/ndecker/go-http-digest-auth-client"
)

func setupClient(username string, password string, allowSelfsigned bool) *http.Client {
	var t http.RoundTripper
	t = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: allowSelfsigned,
		},
	}

	if username != "" {
		client := &http.Client{
			Transport: t,
		}

		t = &dac.DigestTransport{
			Client:   client,
			Username: username,
			Password: password,
		}
	}

	client := &http.Client{
		Transport: t,
	}
	return client
}
