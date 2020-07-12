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

func proxyGRPC(serverAddr string, r *http.Request) *httputil.ReverseProxy {
	proxy := &httputil.ReverseProxy{
		Director: func(r *http.Request) {
			r.URL.Scheme = "https"
			r.URL.Host = serverAddr
		},
		Transport: &http2.Transport{
			DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
				return net.Dial(network, addr)
			},
		},
		ModifyResponse: func(res *http.Response) error {
			log.Printf("%s %s => %s %d", r.Method, r.URL, serverAddr, res.StatusCode)
			return nil
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			log.Print(err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		},
	}
	return proxy
}

func router(serverAddr string) *mux.Router {
	router := mux.NewRouter()
	router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := proxyGRPC(serverAddr, r)
		p.ServeHTTP(w, r)
	})
	return router
}

func main() {
	flag.Parse()
	router := router(*serverAddr)
	addr := fmt.Sprintf(":%d", *port)
	cert := testdata.Path("server1.pem")
	key := testdata.Path("server1.key")
	log.Printf("Proxy listening on port %d", *port)
	err := http.ListenAndServeTLS(addr, cert, key, router)
	if err != nil {
		log.Fatal(err)
	}
}
