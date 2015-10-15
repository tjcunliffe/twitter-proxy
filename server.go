package main

import (
	"github.com/codegangsta/negroni"
	"flag"
	"os"
	"encoding/json"
	"net/http"
	log "github.com/Sirupsen/logrus"

	"github.com/unrolled/render"
	"github.com/meatballhat/negroni-logrus"
	"github.com/go-zoo/bone"
    "github.com/garyburd/redigo/redis"
)

// Initial structure of configuration that is expected from conf.json file
type Configuration struct {
	MirageEndpoint string
	ExternalSystem string
}

// AppConfig stores application configuration
var AppConfig Configuration

// Client structure to be injected into functions to perform HTTP calls
type Client struct {
	HTTPClient *http.Client
}

// HTTPClientHandler used for passing http client connection and template
// information back to handlers, mostly for testing purposes
type HTTPClientHandler struct {
	http Client
	r  *render.Render
	pool *redis.Pool
}

var (
	maxConnections = flag.Int("max-connections", 10, "Max connections to Redis")
	// like ./twitter-proxy -port=":8080" would start on port 8080
	port = flag.String("port", ":8300", "Server port")
)

func main() {
	flag.Parse() // parse the flags
	// Output to stderr instead of stdout, could also be a file.
	log.SetOutput(os.Stderr)
	log.SetFormatter(&log.TextFormatter{})

	// getting app config
	mirageEndpoint := os.Getenv("MirageEndpoint")
	externalSystem := os.Getenv("ExternalSystem")

	if(mirageEndpoint != "" && externalSystem != ""){
		log.Info("Environment variables found.")
		AppConfig.ExternalSystem = externalSystem
		AppConfig.MirageEndpoint = mirageEndpoint
	} else {
		log.Info("Environment variables not found, reading config from file.")
		// env variables not found, getting configuration from file
		file, err := os.Open("conf.json")
		if err != nil {
			log.Panic("Failed to open configuration file, quiting server.")
		}
		decoder := json.NewDecoder(file)
		AppConfig = Configuration{}
		err = decoder.Decode(&AppConfig)
		if err != nil {
			log.WithFields(log.Fields{"Error": err.Error()}).Panic("Failed to read configuration")
		}
	}

	// getting redis connection
	redisAddress := os.Getenv("RedisAddress")
	if(redisAddress == ""){
		redisAddress = ":6379"
	}

	// app starting
	log.WithFields(log.Fields{
		"MirageEndpoint": AppConfig.MirageEndpoint,
		"ExternalSystemEndpoint": AppConfig.ExternalSystem,
	}).Info("app is starting")

	// getting base template and handler struct
	r := render.New(render.Options{Layout: "layout"})

	// getting redis client for state storing
	redisPool := redis.NewPool(func() (redis.Conn, error) {
		c, err := redis.Dial("tcp", redisAddress)

		if err != nil {
			log.WithFields(log.Fields{"Error": err.Error()}).Panic("Failed to create Redis connection pool!")
			return nil, err
		}

		return c, err
	}, *maxConnections)

	defer redisPool.Close()

	h := HTTPClientHandler{http: Client{&http.Client{}}, r: r, pool: redisPool}

	mux := getBoneRouter(h)
	n := negroni.Classic()
	n.Use(negronilogrus.NewMiddleware())
	n.UseHandler(mux)
	n.Run(*port)
}


func getBoneRouter(h HTTPClientHandler) *bone.Mux {
	mux := bone.New()
	mux.Get("/1.1/search/tweets.json", http.HandlerFunc(h.tweetSearchEndpoint))
	mux.Get("/admin", http.HandlerFunc(h.adminHandler))
	mux.Post("/admin/state", http.HandlerFunc(h.stateHandler))
	mux.Get("/admin/state", http.HandlerFunc(h.getStateHandler))
	// handling static files
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	return mux
}

