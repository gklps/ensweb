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
	config         *config.Config
	log            logger.Logger
	address        string
	addr           *url.URL
	hc             *http.Client
	th             TokenHelper
	defaultTimeout time.Duration
	token          string
}

type ClientOptions = func(*Client) error

func SetClientDefaultTimeout(timeout time.Duration) ClientOptions {
	return func(c *Client) error {
		c.defaultTimeout = timeout
		return nil
	}
}

func SetClientTokenHelper(filename string) ClientOptions {
	return func(c *Client) error {
		th, err := NewInternalTokenHelper(filename)
		if err != nil {
			return err
		}
		c.th = th
		return nil
	}
}

// NewClient : Create new client handle
func NewClient(config *config.Config, log logger.Logger, options ...ClientOptions) (Client, error) {
	var address string
	var tr *http.Transport
	clog := log.Named("enswebclient")
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

	hc := &http.Client{
		Transport: tr,
		Timeout:   DefaultTimeout,
	}

	addr, err := url.Parse(address)

	if err != nil {
		clog.Error("failed to parse server address", "err", err)
		return Client{}, err
	}

	tc := Client{
		config:  config,
		log:     clog,
		address: address,
		addr:    addr,
		hc:      hc,
	}

	for _, op := range options {
		err = op(&tc)
		if err != nil {
			clog.Error("failed in setting the option", "err", err)
			return Client{}, err
		}
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

func (c *Client) Do(req *http.Request, timeout ...time.Duration) (*http.Response, error) {
	if timeout != nil {
		c.hc.Timeout = timeout[0]
	} else {
		c.hc.Timeout = c.defaultTimeout
	}
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
