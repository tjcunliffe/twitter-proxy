package main
import (
	"net/http"
	log "github.com/Sirupsen/logrus"
)

func (h HTTPClientHandler) tweetSearchEndpoint(w *http.ResponseWriter, r *http.Request) {
	// getting query
	urlQuery := r.URL.Query()
	// getting submitted query string
	queryString := urlQuery["q"]

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
				"value": v,
			}).Info("app is starting")

			// adding key/value pairs
			req.Header.Add(k, v)
		}

	}

}