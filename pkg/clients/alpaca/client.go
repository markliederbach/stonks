package alpaca

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/markliederbach/stonks/pkg/config"
)

type Endpoint string

const (
	headerAPIKeyID     string = "APCA-API-KEY-ID"
	headerAPISecretKey string = "APCA-API-SECRET-KEY"

	lastQuoteEndpoint Endpoint = "/v1/last_quote/stocks"
)

type Alpaca struct {
	baseURL      string
	apiKeyID     string
	apiSecretKey string

	client *http.Client
}

func NewAlpaca(config config.Config) (*Alpaca, error) {
	return &Alpaca{
		baseURL:      config.AlpacaAPIBaseURL,
		apiKeyID:     config.AlpacaAPIKeyID,
		apiSecretKey: config.AlpacaAPISecretKey,

		client: &http.Client{},
	}, nil
}

func (a *Alpaca) setHeaders(request *http.Request, headers map[string]string) {
	for key, value := range headers {
		request.Header.Set(key, value)
	}
}

func (a *Alpaca) setAuthorization(request *http.Request) {
	a.setHeaders(request, map[string]string{
		headerAPIKeyID:     a.apiKeyID,
		headerAPISecretKey: a.apiSecretKey,
	})
}

func (a *Alpaca) buildEndpoint(endpoint Endpoint) string {
	return fmt.Sprintf("%s%s", a.baseURL, endpoint)
}

func (a *Alpaca) get(url string) (*http.Response, error) {
	var err error

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return &http.Response{}, err
	}

	a.setAuthorization(request)
	response, err := a.client.Do(request)
	return response, err
}

func (a *Alpaca) LastQuote(symbol string) (LastQuote, int, error) {
	var result LastQuote
	url := fmt.Sprintf("%s/%s", a.buildEndpoint(lastQuoteEndpoint), symbol)

	response, err := a.get(url)
	if err != nil || response.StatusCode != http.StatusOK {
		return LastQuote{}, response.StatusCode, err
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)

	if err := json.Unmarshal(body, &result); err != nil {
		return LastQuote{}, response.StatusCode, err
	}

	return result, response.StatusCode, nil
}
