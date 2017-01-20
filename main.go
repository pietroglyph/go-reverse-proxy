package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/httputil"
)

type Proxy struct {
	proxy      *httputil.ReverseProxy
	routeTable map[string]*json.RawMessage
}

func main() {
	// Defualts for command line flags
	const (
		defaultPort        = ":80"
		defaultPortUsage   = "default server port, ':80', ':8080'..."
		defaultTarget      = "http://127.0.0.1:8080"
		defaultTargetUsage = "default redirect url, 'http://127.0.0.1:8080'"
	)
	// Define command line flags
	port := flag.String("port", defaultPort, defaultPortUsage)
	target := flag.String("url", defaultTarget, defaultTargetUsage)
	// Parse command line flags
	flag.Parse()
	// Print info
	fmt.Println("server will run on : %s", *port)
	fmt.Println("redirecting to :%s", *target)\
	// Read routeTable
	rawTable, err := ioutil.ReadFile("routeTable.json")
	if err != nil {
		log.Panic(err)
	}
	// Unmarshall routeTable JSON
	err = json.Unmarshal(rawTable, routeTable)
	if err != nil {
		log.Panic(err)
	}
	// @TODO make a ReverseProxy

}

// @TODO make a ReverseProxy Director function