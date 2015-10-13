# Twitter Proxy demo for Mirage

This is an example adaptor (proxy) implementation for service virtualization tool [Mirage](https://github.com/SpectoLabs/mirage) .

## Installation

This application uses vendor strategy to manage dependencies:
* export GO15VENDOREXPERIMENT=1
* go build

## Running it

Proxy by default starts on port 8300, you can override it by providing port value during startup:
./twitter-proxy -port=":8888"

If no settings found in Redis - it will by default take recording stance, making calls to external service and then will
try to record results. It returns original response to client application so incremental tests that reuse accumulated 
information can be created. 