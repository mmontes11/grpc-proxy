package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"

	"github.com/gorilla/mux"
	"golang.org/x/net/http2"
	"google.golang.org/grpc/testdata"
)

var (
	port       = flag.Int("port", 8000, "The port where the proxy will listen")
	serverAddr = flag.String("server_addr", "localhost:11000", "The server address in the format of host:port")
)

func proxyGRPC(backendAddr string) (*httputil.ReverseProxy, error) {
	proxy := &httputil.ReverseProxy{
		Director: func(r *http.Request) {
			r.URL.Scheme = "https"
			r.URL.Host = backendAddr
		},
		Transport: &http2.Transport{
			DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
				return net.Dial(network, addr)
			},
		},
	}
	return proxy, nil
}

func main() {
	flag.Parse()

	router := mux.NewRouter()
	router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, err := proxyGRPC(*serverAddr)
		if err != nil {
			log.Print(err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		p.ServeHTTP(w, r)
	})

	log.Printf("Proxy listening on port %d", *port)
	addr := fmt.Sprintf(":%d", *port)
	err := http.ListenAndServeTLS(addr, testdata.Path("server1.pem"), testdata.Path("server1.key"), router)
	if err != nil {
		log.Fatal(err)
	}
}
