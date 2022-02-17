package httpclient

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	httpClient *http.Client
)

const (
	MaxIdleConnections int = 100
	RequestTimeout     int = 3
)

func init() {
	// use httpClient to send request
	httpClient = createHTTPClient()
}

// createHTTPClient for connection re-use
func createHTTPClient() *http.Client {
	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: MaxIdleConnections,
			IdleConnTimeout:     60 * time.Second,
		}, Timeout: time.Duration(RequestTimeout) * time.Second,
	}

	return client
}

// Request ...
func Request(urlPath string, method string, data map[string]string, header map[string]string, cookie map[string]string) (string, error) {
	req, err := http.NewRequest(method, urlPath, nil)
	if err != nil {
		return "", err
	}

	if method == "GET" {
		q := req.URL.Query()
		for key, value := range data {
			q.Add(key, value)
		}
		req.URL.RawQuery = q.Encode()
	} else if method == "POST" {
		postData := url.Values{}
		for key, value := range data {
			postData.Set(key, value)
		}
		req.Body = ioutil.NopCloser(strings.NewReader(postData.Encode()))
	}

	for key, value := range header {
		req.Header.Set(key, value)
	}
	for key, value := range cookie {
		cookie1 := &http.Cookie{Name: key, Value: value}
		req.AddCookie(cookie1)
	}

	// use httpClient to send request
	response, err := httpClient.Do(req)

	if err != nil || response == nil {
		//todo 上报日志
		return "", err
	} else {
		// Close the connection to reuse it
		defer response.Body.Close()

		// Let's check if the work actually is done
		// We have seen inconsistencies even when we get 200 OK response
		body, err := ioutil.ReadAll(response.Body)
		//todo 上报日志
		if err != nil {
			return "", err
		}
		return string(body), nil
	}

}
