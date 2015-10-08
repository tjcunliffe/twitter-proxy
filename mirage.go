package main

import (
	"net/http"
	log "github.com/Sirupsen/logrus"
	"io/ioutil"
    "bytes"
)

type params struct {
	url, body, method string
	bodyBytes          []byte
	headers            map[string]string
}


// makeRequest takes Params struct as parameters and makes request to Mirage
// then gets response bytes and returns to caller
func (c *Client) makeRequest(s params) ([]byte, int, error) {

	if s.bodyBytes == nil {
		s.bodyBytes = []byte(s.body)
	}

	log.WithFields(log.Fields{
		"url":           s.url,
		"body":          s.body,
		"headers":       s.headers,
		"requestMethod": s.method,
	}).Info("Transforming URL, preparing for request to Stubo")

	req, err := http.NewRequest(s.method, s.url, bytes.NewBuffer(s.bodyBytes))
	if s.headers != nil {
		for k, v := range s.headers {
			req.Header.Set(k, v)
		}
	}
	//req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		// logging read error
		log.WithFields(log.Fields{
			"error": err.Error(),
			"url":   s.url,
		}).Warn("Failed to get response from Mirage!")

		return []byte(""), http.StatusInternalServerError, err
	}
	defer resp.Body.Close()
	// reading body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// logging read error
		log.WithFields(log.Fields{
			"error": err.Error(),
			"url":   s.url,
		}).Warn("Failed to read response from Mirage!")

		return []byte(""), http.StatusInternalServerError, err
	}
	return body, resp.StatusCode, nil
}
