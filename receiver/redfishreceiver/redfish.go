package redfishreceiver // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/redfishreceiver"

import (
	"crypto/tls"
	"encoding/json"
	"net"
	"net/http"
	"net/url"
	"path"
	"time"

	"go.opentelemetry.io/collector/config/configopaque"
)

type redfishClient struct {
	Client           http.Client
	baseURL          *url.URL
	redfishVersion   string
	host             string
	userName         string
	password         configopaque.String
	computerSystemId string
}

func NewRedfishClient(computerSystemId, user string, pwd configopaque.String, addr, redfishVersion string, timeout time.Duration, insecure bool) (*redfishClient, error) {
	baseURL, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}

	address, _, err := net.SplitHostPort(baseURL.Host)
	if err != nil {
		address = baseURL.Host
	}

	return &redfishClient{
		Client: http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure},
				Proxy:           http.ProxyFromEnvironment,
			},
			Timeout: timeout,
		},
		computerSystemId: computerSystemId,
		baseURL:          baseURL,
		host:             address,
		redfishVersion:   redfishVersion,
		userName:         user,
		password:         pwd,
	}, nil
}

func (c *redfishClient) setHeaders(req *http.Request) {
	req.SetBasicAuth(c.userName, string(c.password))
	req.Header.Set("OData-Version", "4.0")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
}

func (c *redfishClient) GetComputerSystem() (*computerSystem, error) {
	url := c.baseURL.ResolveReference(&url.URL{
		Path: path.Join("/redfish/", c.redfishVersion, "/Systems/", c.computerSystemId),
	}).String()
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var system computerSystem
	if err = json.NewDecoder(resp.Body).Decode(&system); err != nil {
		return nil, err
	}
	return &system, nil
}

func (c *redfishClient) GetChassis(ref string) (*chassis, error) {
	url := c.baseURL.ResolveReference(&url.URL{Path: ref}).String()
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var chassis chassis
	if err = json.NewDecoder(resp.Body).Decode(&chassis); err != nil {
		return nil, err
	}
	return &chassis, nil
}

func (c *redfishClient) GetThermal(ref string) (*thermal, error) {
	url := c.baseURL.ResolveReference(&url.URL{Path: ref}).String()
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var thermal thermal
	if err = json.NewDecoder(resp.Body).Decode(&thermal); err != nil {
		return nil, err
	}
	return &thermal, nil
}
