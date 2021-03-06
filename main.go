package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
)

type proxy struct {
	reverseproxy *httputil.ReverseProxy
	routeTable   map[string]string
}

func main() {
	var prox proxy
	// Initialize our routeTable map
	prox.routeTable = make(map[string]string)
	// Defaults for command line flags
	const (
		defaultPort                 = "8080"
		defaultPortUsage            = "default server port, '80'; '8080'..."
		defaultHost                 = "localhost"
		defaultHostUsage            = "default server host, 'localhost', '127.0.0.1'; ' 0:0:0:0:0:0:0:1'"
		defualtTls                  = false
		defualtTlsUsage             = "default tls status, 'true', 'false'"
		defualtRouteTablePath       = "routetable.txt"
		defualtRouteTablePathUsage  = "default routetable path, 'routetable.txt'; '/home/declan/work/src/github.com/pietroglyph/go-reverse-prox/routetable.txt'"
		defualtPrivateKeyPath       = "key.pem"
		defualtPrivateKeyPathUsage  = "default private key path, 'key.pem'; '/etc/letsencrypt/live/example.com/privkey.pem'"
		defualtCertificatePath      = "cert.pem"
		defualtCertificatePathUsage = "default path to the *full* certificate chain, 'cert.pem'; '/etc/letsencrypt/live/example.com/fullchain.pem'"
	)
	// Define command line flags
	port := flag.String("port", defaultPort, defaultPortUsage)
	host := flag.String("host", defaultHost, defaultHostUsage)
	tls := flag.Bool("tls", defualtTls, defualtTlsUsage)
	routetable := flag.String("routetable", defualtRouteTablePath, defualtRouteTablePathUsage)
	key := flag.String("key", defualtPrivateKeyPath, defualtPrivateKeyPathUsage)
	cert := flag.String("cert", defualtCertificatePath, defualtCertificatePathUsage)
	// Parse command line flags
	flag.Parse()
	// Print info
	log.Println(fmt.Sprintf("reverse-prox will run on %s:%s", *host, *port))
	/*
		##Read routeTable##
		Each whole route is separated by a newline, and each subroute with source and destination is separated by a space.
		This is simpler than having to map some formatting standard like JSON onto a map.
	*/
	file, err := os.Open(*routetable)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Create a scanner to read the file line-by-line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		split := strings.Split(scanner.Text(), " ")
		if len(split) == 2 {
			prox.routeTable[split[0]] = split[1]
		} else {
			log.Println("coulndn't parse", split, "of routetable.txt, discarding line")
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	// Make sure we know that reading routetable is done
	log.Println("succsessfully read routetable.txt with", len(prox.routeTable), "entries")
	// Make a director to route requests based on routetable
	director := func(req *http.Request) {

		path := strings.TrimPrefix(req.URL.Path, "/") // Trim the leading / from the path of the url (eg. /example->example)
		log.Println(req.URL.Path)
		splitpath := strings.Split(path, "/") // Split path using the string "/" as a deliniator
		routekey := splitpath[0]
		if prox.routeTable[routekey] != "" {
			url, err := url.Parse(prox.routeTable[routekey])
			if err != nil {
				log.Fatal(err)
			}
			strings.TrimSuffix(url.Path, "/") // Trim trailing backslashes so we don't get double slashes in our path from appending the path from our original request
			for key := range splitpath {      // Range over all but the first index of splitpath so we can pass it on to our *http.Request.URL
				if key > 0 {
					url.Path += "/" + splitpath[key]
				}
			}

			// Append any query data
			url.RawQuery = req.URL.RawQuery // ? is inserted for us

			req.URL = url
		} else {
			log.Println("there is no routing entry for ", path)
			req.URL, _ = url.Parse("https://en.wikipedia.org/wiki/List_of_HTTP_status_codes#4xx_Client_Error")
		}

		req.Header.Set("Host", req.URL.Host) // Make sure that we aren't giving the target the wrong host header
		req.Host = req.URL.Host              // **If this isn't set the host header gets ignored**

		if _, ok := req.Header["User-Agent"]; !ok {

			// explicitly disable User-Agent so it's not set to default value

			req.Header.Set("User-Agent", "")

		}

	}

	// Make a Reverseproxy and give it a Director
	prox.reverseproxy = &httputil.ReverseProxy{Director: director}
	// Put our host and port settings into a string
	bind := fmt.Sprintf("%s:%s", *host, *port)
	// Serve the reverse prox
	if *tls {
		log.Printf("Serving with TLS on  %s", bind)
		log.Fatal(http.ListenAndServeTLS(bind, *key, *cert, prox.reverseproxy))
	} else {
		log.Printf("Serving without TLS on %s", bind)
		log.Fatal(http.ListenAndServe(bind, prox.reverseproxy))
	}
}
