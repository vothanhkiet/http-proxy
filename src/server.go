package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
)

// Hop-by-hop headers. These are removed when sent to the backend.
// http://www.w3.org/Protocols/rfc2616/rfc2616-sec13.html
var hopHeaders = []string{
	"Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te", // canonicalized version of "TE"
	"Trailers",
	"Transfer-Encoding",
	"Upgrade",
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func delHopHeaders(header http.Header) {
	for _, h := range hopHeaders {
		header.Del(h)
	}
}

func appendHostToXForwardHeader(header http.Header, host string) {
	// If we aren't the first proxy retain prior
	// X-Forwarded-For information as a comma+space
	// separated list and fold multiple headers into one.
	if prior, ok := header["X-Forwarded-For"]; ok {
		host = strings.Join(prior, ", ") + ", " + host
	}
	header.Set("X-Forwarded-For", host)
}

type proxy struct {
	url *url.URL
}

var cors *bool
var allowHeaders *string
var allowMethods *string
var allowOrigin *string
var exposeHeaders *string
var maxAge *int

func (p *proxy) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	log.Println(req.RemoteAddr, " ", req.Method, " ", req.URL)
	delHopHeaders(req.Header)

	if clientIP, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
		appendHostToXForwardHeader(req.Header, clientIP)
	}

	if req.Method == http.MethodOptions {
		wr.Header().Set("Access-Control-Allow-Headers", *allowHeaders)
		wr.Header().Set("Access-Control-Allow-Methods", *allowMethods)
		wr.Header().Set("Access-Control-Allow-Origin", *allowOrigin)
		wr.Header().Set("Access-Control-Expose-Headers", *exposeHeaders)
		wr.Header().Set("Access-Control-Max-Age", strconv.Itoa(*maxAge))
		wr.Header().Set("Cache-Control", "public, max-age="+strconv.Itoa(*maxAge))
		wr.Header().Set("Pragma", "Public")
		wr.WriteHeader(http.StatusNoContent)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(p.url)

	// Update the headers to allow for SSL redirection
	req.URL.Host = p.url.Host
	req.URL.Scheme = p.url.Scheme
	req.Header.Set("X-Forwarded-Host", req.Host)
	req.Host = p.url.Host

	// Note that ServeHttp is non blocking and uses a go routine under the hood
	proxy.ServeHTTP(wr, req)
}

func main() {
	var addr = flag.String("addr", "0.0.0.0:8080", "The addr of the application.")
	var target = flag.String("target", "https://www.google.com", "The addr of the application.")
	cors = flag.Bool("cors", false, "Enable cors.")
	allowMethods = flag.String("allowMethods", "GET, POST, PUT, PATCH, DELETE", "Allowed method.")
	allowHeaders = flag.String("allowHeaders", "Origin, Content-Type, Accept, Authorization", "Allowed header")
	allowOrigin = flag.String("allowOrigin", "*", "Allowed origin")
	exposeHeaders = flag.String("exposeHeaders", "Limit, Offset, Total", "Allowed header")
	maxAge = flag.Int("maxAge", 3600, "Allowed origin")
	flag.Parse()

	url, _ := url.Parse(*target)

	handler := &proxy{url: url}

	log.Println("Starting proxy server on", *addr)
	log.Println("Forwarding connection to ", *target)
	if *cors {
		log.Println("CORS is enable")
		log.Println("Allowed headers: ", *allowHeaders)
		log.Println("Allowed methods: ", *allowMethods)
		log.Println("Allowed origin: ", *allowOrigin)
		log.Println("Expose headers: ", *exposeHeaders)
		log.Println("Max age: ", *maxAge)
	}
	if err := http.ListenAndServe(*addr, handler); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
