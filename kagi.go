package main

import (
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
	//queryURL = ""
	req, err := http.NewRequest("GET", queryURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bot %s", k.apiKey))

	res, err := k.client.Do(req)
	if err != nil {
		return "", err
	}
	fmt.Println(res.Status)

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	fmt.Sprintf("response: %v", string(resBody))

	return string(resBody), nil
}
