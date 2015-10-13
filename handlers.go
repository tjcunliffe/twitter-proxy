package main


import (
	"net/http"
	log "github.com/Sirupsen/logrus"
	"io/ioutil"
	"strings"
	"github.com/garyburd/redigo/redis"
	"encoding/json"
)

type StateRequest struct {
	record bool `json:"record"`
}

type StateResponse struct {
	newState bool `json:"newState"`
}

func (h HTTPClientHandler) tweetSearchEndpoint(w http.ResponseWriter, r *http.Request) {

	mirageSession := r.Header.Get("MirageScenarioSession")
	// session format should be "scenario:session"
	slices := strings.Split(mirageSession, ":")
	if len(slices) < 2 {
		msg := "Bad request, missing session or scenario name. When under proxy, please use 'scenario:session' format in your" +
		"URL query, such as '/stubo/api/put/stub?session=scenario:session_name' "
		log.Warn(msg)
		http.Error(w, msg, 400)
		return
	}
	scenario := slices[0]
	session := slices[1]

	log.WithFields(log.Fields{
		"Scenario": scenario,
		"Session": session,
	}).Info("Got scenario and session!...")

	// getting query
	urlQuery := r.URL.Query()
	// getting submitted query string
	queryString := urlQuery["q"]

	client := h.http.HTTPClient

	// getting current proxy state
	record := h.getCurrentState()

	if record {
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

		// recording request to mirage
		go h.http.recordRequest(scenario, session, queryString[0], r, resp, body)

		// returning original response back to client app (system under test)


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
	} else {
		// playback time!!
		log.Info("PLAYBACK MODE")
		data := h.http.playbackResponse(scenario, session, queryString[0])
		for k, v := range data.Headers {
			w.Header().Set(k, v)
		}
		w.WriteHeader(data.StatusCode)
		w.Write(data.Body)
	}
}

func (h HTTPClientHandler) adminHandler(w http.ResponseWriter, r *http.Request) {
	h.r.HTML(w, http.StatusOK, "adminHome", nil)
}

func (h HTTPClientHandler) stateHandler(w http.ResponseWriter, r *http.Request) {
	var stateRequest StateRequest

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)

	if err := json.Unmarshal(body, &stateRequest); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(422) // can' process this entity

		if err := json.NewEncoder(w).Encode(err); err != nil {
			panic(err)
		}
	}
	log.WithFields(log.Fields{
		"newState": stateRequest.record,
		"body": string(body),
	}).Info("Handling state change request!")
	// setting state to redis
	err = h.setState(stateRequest.record)
	if(err != nil){
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(400) // failed to change it
	} else {
		w.WriteHeader(200)
	}

}

// getCurrentState returns current proxy state (record is default one since if Mirage is not around it will get response
// from external service and return it to the client
func (h HTTPClientHandler) getCurrentState() (bool) {
	// default state
	record := true

	c := h.pool.Get()
	defer c.Close()

	state, err := redis.Bool(c.Do("GET", "twproxy"))
	if err != nil {
		log.Warning("State not found, switching to recording mode")
		c.Do("SET", "twproxy", record)
		return record
	} else {
		log.WithFields(log.Fields{
			"state": state,
		}).Info("Proxy configuration found in Redis...")
		return state
	}
}

// setState sets new state for proxy inside redis. Supply 1 for "Recording" state or 0 for "Playback"
func (h HTTPClientHandler) setState(state bool) (error) {
	c := h.pool.Get()
	defer c.Close()

	status, err := c.Do("SET", "twproxy", state)

	if err != nil {
		log.WithFields(log.Fields{
			"record": state,
		}).Error("Failed to update proxy state...")
		return err
	} else {
		log.WithFields(log.Fields{
			"record": state,
			"status": status,
		}).Info("State updated!")
		return nil
	}
}