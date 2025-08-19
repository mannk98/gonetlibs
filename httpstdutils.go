package gonetlibs

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"
)

/*
CLIENT
*/

type HttpClient struct {
	Client *http.Client
}

func HttpClientNewDefaultTransPort() *http.Transport {
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

func NewHttpClient(transport *http.Transport) *HttpClient {

	if transport == nil {
		transport = HttpClientNewDefaultTransPort()
	}
	return &HttpClient{&http.Client{Transport: transport}}
}

// Don't forget add https:// or http
func (client *HttpClient) Get(url string, headers map[string]string) (*http.Response, string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}

	for header, headerval := range headers {
		req.Header.Add(header, headerval)
	}

	resp, err := client.Client.Do(req)
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

func (client *HttpClient) Post(url string, inputBody []byte, headers map[string]string) (*http.Response, string, error) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(inputBody))
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}

	for header, headerval := range headers {
		req.Header.Add(header, headerval)
	}

	resp, err := client.Client.Do(req)
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

/*
SERVER
*/
type HttpServer struct {
	Server *http.Server
}

func NewHttpServer(port string, mux *http.ServeMux) *HttpServer {
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	return &HttpServer{Server: server}
}

func (server *HttpServer) Start() error {
	var err error
	go func() {
		if err := server.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not listen on port %v: %v\n", server.Server.Addr, err)
		}
	}()

	fmt.Println("Server is ready to handle requests at :8080")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	// Block until we receive a signal
	<-stop

	fmt.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	fmt.Println("Server gracefully stopped")
	return err
}
