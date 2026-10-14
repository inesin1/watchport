package main

import (
	"net/http"
	"strings"
	"time"
)

type checkResult struct {
	url        string
	statusCode int
	elapsed    time.Duration
	err        error
}

func check(rawURL string, client *http.Client) checkResult {
	started := time.Now()
	result := checkResult{url: rawURL}

	request, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		result.err = err
		result.elapsed = time.Since(started)
		return result
	}
	request.URL.Scheme = strings.ToLower(request.URL.Scheme)

	response, err := client.Do(request)
	if err != nil {
		result.err = err
		result.elapsed = time.Since(started)
		return result
	}
	defer response.Body.Close()

	result.statusCode = response.StatusCode
	result.elapsed = time.Since(started)
	return result
}
