package client

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

// API for api struct
type API struct {
	APIAddress string
	Port       string
	Connector  string
	cl         *http.Client
}

// NewAPI for new api
func NewAPI(ipAddress string, port string, connector string) *API {
	return &API{ipAddress, port, connector, &http.Client{}}
}

// Call for validate token
func (client *API) Call(action string) error {
	request, err := http.NewRequest("PUT", fmt.Sprintf("http://%s:%s/connectors/%s/%s", client.APIAddress, client.Port, client.Connector, action), nil)

	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := client.cl.Do(request)

	if err != nil {
		return err
	}

	data, _ := ioutil.ReadAll(response.Body)
	if response.StatusCode != 202 {
		return fmt.Errorf("failed for this request with response: %s", string(data))
	}

	return nil
}
