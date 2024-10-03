package netutils

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"
)

func HttpClientNewTransPort() *http.Transport {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 15 * time.Second,
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2:   true,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     30 * time.Second,
		TLSHandshakeTimeout: 6 * time.Second,
		//		ExpectContinueTimeout:  1 * time.Second,
		MaxResponseHeaderBytes: 8192,
		ResponseHeaderTimeout:  time.Millisecond * 5000,
		DisableKeepAlives:      false,
	}
}

func HttpClientNewClient(transport *http.Transport) *http.Client {
	return &http.Client{Transport: transport}
}

// Don't forget add https:// or http
func HttpGet(client *http.Client, url string, headers map[string]string) (*http.Response, string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}

	for header, headerval := range headers {
		req.Header.Add(header, headerval)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, string(body), fmt.Errorf("Error reading response body: %v", err)
	}
	return resp, string(body), err
}

func HttpPost(client *http.Client, url string, inputBody []byte, headers map[string]string) (*http.Response, string, error) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(inputBody))
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}

	for header, headerval := range headers {
		req.Header.Add(header, headerval)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, string(body), fmt.Errorf("Error reading response body: %v", err)
	}
	return resp, string(body), err
}
