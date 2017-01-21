package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type Proxy struct {
	reverseProxy *httputil.ReverseProxy
	routeTable   map[string]string
}

func main() {
	var proxy Proxy
	var jsonArbitraryTemplate interface{}
	// Defualts for command line flags
	const (
		defaultPort      = "80"
		defaultPortUsage = "default server port, '80', '8080'..."
		defaultHost      = "localhost"
		defaultHostUsage = "default server host, 'localhost', '127.0.0.1', ' 0:0:0:0:0:0:0:1'"
		defualtTls       = false
		defualtTlsUsage  = "defualt tls status, 'true', 'false'"
	)
	// Define command line flags
	port := flag.String("port", defaultPort, defaultPortUsage)
	host := flag.String("url", defaultHost, defaultHostUsage)
	tls := flag.Bool("tls", defualtTls, defualtTlsUsage)
	// Parse command line flags
	flag.Parse()
	// Print info
	fmt.Printf("reverse-proxy will run on %s:%s\n", *host, *port)
	// Read routeTable
	rawTable, err := ioutil.ReadFile("routeTable.json")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(rawTable))
	// Unmarshall routeTable JSON
	err = json.Unmarshal(rawTable, jsonArbitraryTemplate)
	// Extract top-level map from our json interface for arbitrary data using a type assertion to a map[string]string
	proxy.routeTable = jsonArbitraryTemplate.(map[string]string)
	if err != nil {
		log.Fatal(err)
	}
	// Make a director to route requests based on the JSON table
	director := func(req *http.Request) {

		path := strings.TrimPrefix(req.URL.Path, "/") // Trim the leading / from the path of the url (eg. https://example.com/example->example)

		if proxy.routeTable[path] != "" {
			url, err := url.Parse(proxy.routeTable[path])
			if err != nil {
				log.Fatal(err)
			}
			req.URL = url
		}

		if _, ok := req.Header["User-Agent"]; !ok {

			// explicitly disable User-Agent so it's not set to default value

			req.Header.Set("User-Agent", "")

		}

	}

	// Make a ReverseProxy and give it a Director
	proxy.reverseProxy = &httputil.ReverseProxy{Director: director}
	// Put our host and port settings into a string
	bind := fmt.Sprintf("%s:%s", host, port)
	// Serve the reverse proxy
	if *tls {
		log.Printf("Serving with TLS on  %s.", bind)
		log.Fatal(http.ListenAndServeTLS(bind, "cert.pem", "key.pem", proxy.reverseProxy))
	} else {
		log.Printf("Serving without TLS on %s.", bind)
		log.Fatal(http.ListenAndServe(bind, proxy.reverseProxy))
	}
}
