package redfishreceiver

import (
	"crypto/tls"
	"encoding/json"
	"net"
	"net/http"
	"net/url"
	"path"
	"time"
)

type RedfishClient struct {
	Client           http.Client
	baseURL          *url.URL
	redfishVersion   string
	host             string
	userName         string
	password         string
	computerSystemId string
}

func NewRedfishClient(computerSystemId, user, pwd, addr, redfishVersion string, timeout time.Duration, insecure bool) (*RedfishClient, error) {
	baseURL, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}

	address, _, err := net.SplitHostPort(baseURL.Host)
	if err != nil {
		address = baseURL.Host
	}

	return &RedfishClient{
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

func (c *RedfishClient) setHeaders(req *http.Request) {
	req.SetBasicAuth(c.userName, c.password)
	req.Header.Set("OData-Version", "4.0")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
}

func (c *RedfishClient) GetComputerSystem() (*ComputerSystem, error) {
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
	var system ComputerSystem
	if err = json.NewDecoder(resp.Body).Decode(&system); err != nil {
		return nil, err
	}
	return &system, nil
}

func (c *RedfishClient) GetChassis(ref string) (*Chassis, error) {
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
	var chassis Chassis
	if err = json.NewDecoder(resp.Body).Decode(&chassis); err != nil {
		return nil, err
	}
	return &chassis, nil
}

func (c *RedfishClient) GetThermal(ref string) (*Thermal, error) {
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
	var thermal Thermal
	if err = json.NewDecoder(resp.Body).Decode(&thermal); err != nil {
		return nil, err
	}
	return &thermal, nil
}
