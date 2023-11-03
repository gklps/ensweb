package main

import (
	"fmt"
	"net/http"

	"github.com/EnsurityTechnologies/config"
	"github.com/EnsurityTechnologies/helper/jsonutil"
	"github.com/EnsurityTechnologies/logger"
	"github.com/gklps/ensweb"
)

// route declaration
const (
	LoginRoute        string = "/api/login"
	LoginSessionRoute string = "/api/loginsession"
)

type Client struct {
	ensweb.Client
}

func NewClient(cfg *config.Config, log logger.Logger) (*Client, error) {
	c := &Client{}
	var err error
	// th, err := ensweb.NewInternalTokenHelper("token.txt")
	// if err != nil {
	// 	return nil, err
	// }
	c.Client, err = ensweb.NewClient(cfg, log)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Client) Login(userName string, password string) error {
	request := Request{
		UserName: userName,
		Password: password,
	}
	req, err := c.JSONRequest("POST", LoginRoute, request)
	if err != nil {
		return fmt.Errorf("Unable to frame request")
	}
	resp, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("Unable to connect to server")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		errMsg, err := ensweb.JSONDecodeErr(resp)
		if err != nil {
			return fmt.Errorf("Invalid response")
		}
		return fmt.Errorf(errMsg.Error)
	}
	var response Response
	err = jsonutil.DecodeJSONFromReader(resp.Body, &response)
	if err != nil {
		return fmt.Errorf("Invalid response")
	}
	c.SetToken(response.Token)
	return nil
}

func (c *Client) LoginSession() (string, error) {

	req, err := c.JSONRequest("GET", LoginSessionRoute, nil)
	if err != nil {
		return "", fmt.Errorf("Unable to frame request")
	}
	c.SetAuthorization(req, c.GetToken())
	resp, err := c.Do(req)
	if err != nil {
		return "", fmt.Errorf("Unable to connect to server")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		errMsg, err := ensweb.JSONDecodeErr(resp)
		if err != nil {
			return "", fmt.Errorf("Invalid response")
		}
		return "", fmt.Errorf(errMsg.Error)
	}
	var response Response
	err = jsonutil.DecodeJSONFromReader(resp.Body, &response)
	if err != nil {
		return "", fmt.Errorf("Invalid response")
	}
	return response.Message, nil
}
