# Twitter Proxy demo for Mirage

This is an example adaptor (proxy) implementation for service virtualization tool [Mirage](https://github.com/SpectoLabs/mirage) .

## Installation

This application uses vendor strategy to manage dependencies:
* export GO15VENDOREXPERIMENT=1
* go build
* Install Redis (Redis is used for state keeping)

## Configuration

Proxy only needs some basic configuration such as where you keep Mirage and where is the real external system endpoint.

Prefered way is to set environment variables:
export MirageEndpoint=http://localhost:8001
export ExternalSystem=https://api.twitter.com

However, you can also specify them in a file:
Congfiguration file example:

{
  "MirageEndpoint": "http://localhost:8001",
  "ExternalSystem": "https://api.twitter.com"
}

Since proxy is storing state in Redis - you may also want to change default Redis address (it's looking for Redis instance
on localhost:6376. You can override this by exporting environment variable:
export RedisAddress=somehost:6379

or max Redis connections in the pool (10 is default number):
./twitter-proxy -max-connections=20

## Running it

Proxy by default starts on port 8300, you can override it by providing port value during startup:
./twitter-proxy -port=":8888"

If no settings found in Redis - it will by default take recording stance, making calls to external service and then will
try to record results. It returns original response to client application so incremental tests that reuse accumulated 
information can be created.


## Changing state (Record/Playback)

State can be changed using administrator page (_/admin_) or you can do it directly through the API:

To change the state URL path is _/admin/state_, request method: "POST".
Examples:
JSON Body payload to start recording:
```
{
  "record": true
}

```

payload to begin playback:
```
{
  "record": false
}

```

To get current state - path is the same (_/admin/state_), however use method "GET". 
Example response from proxy:

{"record":true}

