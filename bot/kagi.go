package bot

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Kagi struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

func NewKagi(baseURL, apiKey string) *Kagi {
	return &Kagi{
		apiKey:  apiKey,
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 10 * time.Minute,
		},
	}
}

func (k *Kagi) Summarize(sumURL string) (string, error) {
	queryURL := fmt.Sprintf("%s/summarize?engine=muriel&url=%s", k.baseURL, url.QueryEscape(sumURL))
	req, err := http.NewRequest("GET", queryURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bot %s", k.apiKey))

	res, err := k.client.Do(req)
	if err != nil {
		return "", err
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	resp := struct {
		Data struct {
			Output string
		}
	}{}
	if err := json.Unmarshal(resBody, &resp); err != nil {
		return "", err
	}

	return resp.Data.Output, nil
}
