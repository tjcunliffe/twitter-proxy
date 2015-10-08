package main
import (
	"net/http"
	log "github.com/Sirupsen/logrus"
	"io/ioutil"

)

func (h HTTPClientHandler) tweetSearchEndpoint(w http.ResponseWriter, r *http.Request) {
	// getting query
	urlQuery := r.URL.Query()
	// getting submitted query string
	queryString := urlQuery["q"]

	client := h.http.HTTPClient

	if Record {
		log.Info("*RECORD* mode detected")

		// here we could probably reuse url path
		externalSystemEndpoint := AppConfig.ExternalSystem + "/1.1/search/tweets.json?q=" + queryString[0]
		// logging full URL and path
		log.WithFields(log.Fields{
			"targetedExternalSystem": AppConfig.ExternalSystem,
			"query": queryString[0],
			"finalTwitterEndpoint": externalSystemEndpoint,
		}).Info("Endpoint created, performing query...")

		// preparing request
		req, err := http.NewRequest("GET", externalSystemEndpoint, nil)

		if err != nil {
			log.Error("Failed to prepare NewRequest", err)
		}
		// proxy is in the record mode so we should get headers for authentication
		// to the real twitter API
		for k, v := range r.Header {
			log.WithFields(log.Fields{
				"key": k,
				"value": v[0],
			}).Info("Reading headers...")

			// adding key/value pairs
			req.Header.Add(k, v[0])
		}
        // performing request
		resp, err := client.Do(req)

		if err != nil {
			log.Error("Failed to get response!", err)
		}

		defer resp.Body.Close()
		// reading body
		body, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			log.Error("Failed to read response body", err)
		}
		// now setting all headers from external response
		// back to the "our" response so it looks as if nothing has
		// tampered with it
		for k, v := range resp.Header {
			w.Header().Set(k, v[0])
		}
		w.Write(body)
	}
}