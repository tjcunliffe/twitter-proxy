package main

import (
	"net/http"
	log "github.com/Sirupsen/logrus"
	"io/ioutil"
    "bytes"
	"encoding/json"
	"fmt"
)

// contains is a structure to store multiple request matchers
//type contains struct {
//	Contains []string `json:contains`
//}



// req structure holds original request to proxy structure which is a part of Mirage payload
// BodyPatterns usually holds "contains" key..
type req struct {
	Method string `json:method`
	BodyPatterns []map[string][]string `json:bodyPatterns`
	Headers map[string]string `json:headers`
}

// res structure hold response body from external service, body is not decoded and is supposed
// to be bytes, however headers should provide all required information for later decoding
// by the client.
type res struct {
	Status int `json:status`
	Body []byte `json:body`
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

// getBodyBytes is a helper method to read re
func getBodyBytes(req *http.Request) ([]byte, error) {
	defer req.Body.Close()
	// reading body
	body, err := ioutil.ReadAll(req.Body)
	if(err != nil) {
		log.WithFields(log.Fields{
			"Error": err,
		}).Error("Failed to read request body bytes and log error if there is one")
	}
	return body, err
}

// getHeaders forms a map of headers from http request headers
func getHeaders(req *http.Request) (map[string]string){
	headers :=make(map[string]string)
	for key, value := range req.Header{
		headers[key] = value[0]
	}
	return headers
}

// createMiragePayload is used to create JSON payload that could be delivered to Mirage during record
func createMiragePayload(matcher string, request *http.Request) (Mirage){
    // defining payload
	var mirageObj Mirage

	// formatting request part
    // assigning headers
	headers := getHeaders(request)
	mirageObj.Request.Headers = headers
    // assigning request method
	mirageObj.Request.Method = request.Method
	// getting contains matcher
	bodyPatterns := make(map[string][]string)
	matchers := []string{matcher}
	bodyPatterns["contains"] = matchers

	// assigning matcher to body patterns
	mirageObj.Request.BodyPatterns = []map[string][]string{bodyPatterns}

	return mirageObj

}

// helper function for development
func prettyprint(b []byte) ([]byte, error) {
	var out bytes.Buffer
	err := json.Indent(&out, b, "", "  ")
	return out.Bytes(), err
}

// recordRequest records supplied request to Mirage. Parameters taken:
// scenario - scenario name, usually you have to create scenario manually
// session - session name should be in "Record" mode for a recording to be successful
// matcher - matcher is a parameter that will be the key for retrieving response during playback
func (c *Client) recordRequest(scenario, session, matcher string, request *http.Request) () {

	if(scenario != "" && session != "") {
		var s params

		path := AppConfig.MirageEndpoint + "/stubo/api/v2/scenarios/objects/" + scenario + "/stubs"
		s.url = path
		s.method = "PUT"

		miragePayload := createMiragePayload(matcher, request)

		bts, _ := json.Marshal(miragePayload)
		b, _ := prettyprint(bts)
		fmt.Printf("%s", b)

	} else{
		log.Error("Scenario or session not supplied.")
	}
    return
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

	defer resp.Body.Close()
	// reading body
	body, err := ioutil.ReadAll(resp.Body)

	return body, resp.StatusCode, err
}
