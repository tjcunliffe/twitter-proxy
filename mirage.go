package main

import (
	"net/http"
	log "github.com/Sirupsen/logrus"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// req structure holds original request to proxy structure which is a part of Mirage payload
// BodyPatterns usually holds "contains" key..
type req struct {
	Method       string `json:"method"`
	BodyPatterns []map[string][]string `json:"bodyPatterns"`
	Headers      map[string]string `json:"headers,omitempty"`
}

// res structure hold response body from external service, body is not decoded and is supposed
// to be bytes, however headers should provide all required information for later decoding
// by the client.
type res struct {
	Status  int `json:"status"`
	Body    []byte `json:"body"`
	Headers map[string]string `json:"headers"`
}

// Mirage structure holds whole payload that Mirage system will understand during request recording
type Mirage struct {
	Request  req `json:"request"`
	Response res `json:"response"`
}

type MirageResponse struct {
	Body []byte `json:"body"`
	StatusCode int `json:"statusCode"`
	Headers map[string]string `json:"headers"`
}

// params structure holds information about request to Mirage formation
type params struct {
	url, session, method string
	bodyBytes            []byte
	headers              map[string]string
}

// getHeaders forms a map of headers from http request headers
func getHeaders(req *http.Request) (map[string]string) {
	headers := make(map[string]string)
	for key, value := range req.Header {
		headers[key] = value[0]
	}
	return headers
}

func getHeadersMap(hds map[string][]string) (map[string]string) {
	headers := make(map[string]string)
	for key, value := range hds {
		headers[key] = value[0]
	}
	return headers
}


// createMiragePayload is used to create JSON payload that could be delivered to Mirage during record
func createMiragePayload(matcher string, request *http.Request, response *http.Response, body []byte) (Mirage) {
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

	// formatting response part
	mirageObj.Response.Status = response.StatusCode
	// getting response headers
	mirageObj.Response.Headers = getHeadersMap(response.Header)
	// adding external service response body
	mirageObj.Response.Body = body

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
// request - request from client (system under test)
// response - response from external system (in our case - twitter)
func (c *Client) recordRequest(scenario, session, matcher string, request *http.Request, response *http.Response, body []byte) () {

	if (scenario != "" && session != "") {
		var s params

		path := AppConfig.MirageEndpoint + "/stubo/api/v2/scenarios/objects/" + scenario + "/stubs"
		s.url = path
		s.method = "PUT"
		s.session = session

		// getting mirage payload structure
		miragePayload := createMiragePayload(matcher, request, response, body)
		// converting it to json bytes
		bts, _ := json.Marshal(miragePayload)
		s.bodyBytes = bts

		// pretty printing for debugging
		b, _ := prettyprint(bts)
		fmt.Printf("%s", b)

		// uploading payload to Mirage
        c.makeRequest(s)

	} else {
		log.Error("Scenario or session not supplied.")
	}
	return
}

func (c *Client) playbackResponse(scenario, session, matcher string) (MirageResponse){
	var data MirageResponse
	if (scenario != "" && session != "") {
		mirageMatcherEndpoint := AppConfig.MirageEndpoint + "/api/v2/matcher"


		req, err := http.NewRequest("POST", mirageMatcherEndpoint, bytes.NewBuffer([]byte(matcher)))
		if(err != nil) {
			log.Error(err)
		}
		req.Header.Add("session", fmt.Sprintf("%s:%s", scenario, session))
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.HTTPClient.Do(req)

		// reading mirage response
		defer resp.Body.Close()
		mrgResponseBytes, _ := ioutil.ReadAll(resp.Body)



		err = json.Unmarshal(mrgResponseBytes, &data)
		if(err != nil) {
			log.Error(err)
		}

		return data

	} else {
		log.Error("Scenario or session not supplied during playback, can't send response")
        return data
	}
}


// makeRequest takes Params struct as parameters and makes request to Mirage
// then gets response bytes and returns to caller
func (c *Client) makeRequest(s params) () {

	log.WithFields(log.Fields{
		"url":           s.url,
		"session":       s.session,
		"headers":       s.headers,
		"requestMethod": s.method,
	}).Info("Transforming URL, preparing for request to Mirage")

	req, err := http.NewRequest(s.method, s.url, bytes.NewBuffer(s.bodyBytes))
	req.Header.Set("session", s.session)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTPClient.Do(req)

	if (err != nil) {
		log.Error("Failed to upload Mirage payload! ", err)
	} else {
		// request was successful but it actually doesn't mean that mirage payload is really there
		// adding some logging for that purpose
		if (resp.StatusCode == 201) {
			log.Info("Mirage payload inserted!")
		} else if (resp.StatusCode == 200) {
			log.Info("Mirage payload updated!")
		} else {
			log.Info("Something happened, status code: ", resp.StatusCode)
			// reading resposne body
			defer resp.Body.Close()
			body, _ := ioutil.ReadAll(resp.Body)
			// pretty printing for debugging
			b, _ := prettyprint(body)
			fmt.Printf("%s", b)
		}
	}

}
