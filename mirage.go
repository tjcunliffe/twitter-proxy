package main

import (
	"net/http"
	log "github.com/Sirupsen/logrus"
	"io/ioutil"
    "bytes"
	"errors"
)

// contains is a structure to store multiple request matchers
type contains struct {
	Contains []string `json:contains`
}

// req structure holds original request to proxy structure which is a part of Mirage payload
type req struct {
	Method string `json:method`
	BodyPatterns string `json:bodyPatterns`
	Contains []contains `json:contains`
	Headers map[string]string `json:headers`
}

// res structure hold response body from external service, body is not decoded and is supposed
// to be bytes, however headers should provide all required information for later decoding
// by the client.
type res struct {
	Status int `json:status`
	Body []bytes `json:body`
	Headers map[string]string `json:headers`
}

// Mirage structure holds whole payload that Mirage system will understand during request recording
type Mirage struct {
	Request req `json:request`
	Response res `json:response`
}


type params struct {
	url, body, method string
	bodyBytes          []byte
	headers            map[string]string
}

// putStub transparently passes request body to Mirage
func (c *Client) putStub(scenario, session, args string, body []byte, headers map[string]string) ([]byte, int, error) {



	if session, ok := headers["session"]; ok {
		if scenario != "" && session != "" {
			var s params

			path := AppConfig.MirageEndpoint + "/stubo/api/v2/scenarios/objects/" + scenario + "/stubs?" + args

			s.url = path
			s.headers = headers
			s.method = "PUT"

			// assigning body in bytes
			s.bodyBytes = body
			// setting logger
			log.WithFields(log.Fields{
				"scenario":      scenario,
				"session":       headers["session"],
				"urlPath":       path,
				"headers":       "",
				"requestMethod": s.method,
			}).Debug("Adding stub to scenario")

			return c.makeRequest(s)
		}
		return []byte(""), http.StatusBadRequest, errors.New("mirage.putStub error: scenario or session not supplied")
	}
	return []byte(""), http.StatusBadRequest, errors.New("mirage.putStub error: session key not supplied")
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
