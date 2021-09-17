package ensweb

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/EnsurityTechnologies/config"
	"github.com/EnsurityTechnologies/logger"
)

// Client : Client struct
type Client struct {
	config  *config.Config
	log     logger.Logger
	address string
	addr    *url.URL
	hc      *http.Client
	th      TokenHelper
	token   string
}

// NewClient : Create new client handle
func NewClient(config *config.Config, log logger.Logger, th TokenHelper) (Client, error) {
	var address string
	var tr *http.Transport
	if config.Production == "true" {
		address = fmt.Sprintf("https://%s", net.JoinHostPort(config.ServerAddress, config.ServerPort))
		tr = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	} else {
		address = fmt.Sprintf("http://%s", net.JoinHostPort(config.ServerAddress, config.ServerPort))
		tr = &http.Transport{
			IdleConnTimeout: 30 * time.Second,
		}
	}

	timeout := time.Duration(ServerTimeout)

	hc := &http.Client{
		Transport: tr,
		Timeout:   timeout,
	}

	addr, err := url.Parse(address)

	if err != nil {
		return Client{}, err
	}

	tc := Client{
		config:  config,
		log:     log,
		address: address,
		addr:    addr,
		hc:      hc,
		th:      th,
	}
	return tc, nil
}

func (c *Client) JSONRequest(method string, requestPath string, model interface{}) (*http.Request, error) {
	var body *bytes.Buffer
	if model != nil {
		j, err := json.Marshal(model)
		if err != nil {
			return nil, err
		}
		body = bytes.NewBuffer(j)
	} else {
		body = bytes.NewBuffer(make([]byte, 0))
	}
	url := &url.URL{
		Scheme: c.addr.Scheme,
		Host:   c.addr.Host,
		User:   c.addr.User,
		Path:   path.Join(c.addr.Path, requestPath),
	}
	req, err := http.NewRequest(method, url.RequestURI(), body)
	req.Host = url.Host
	req.URL.User = url.User
	req.URL.Scheme = url.Scheme
	req.URL.Host = url.Host
	return req, err
}

func (c *Client) SetAuthorization(req *http.Request, token string) {
	var bearer = "Bearer " + token
	req.Header.Set("Authorization", bearer)
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	return c.hc.Do(req)
}

func (c *Client) SetToken(token string) error {
	if c.th != nil {
		return c.th.Store(token)
	}
	c.token = token
	return nil
}

func (c *Client) GetToken() string {
	if c.th != nil {
		tk, err := c.th.Get()
		if err != nil {
			return "InvalidToken"
		} else {
			return tk
		}
	}
	return c.token
}
