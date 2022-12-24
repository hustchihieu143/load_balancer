package main

import (
	"net/http/httputil"
	"net/url"
	"fmt"
	"flag"
	"log"
	"net/http"
)

type Backend struct {
	URL *url.URL
	Alive bool
	ReverseProxy *httputil.ReverseProxy
}

type ServerPool struct {
	backends []*Backend
	current int64
}

func (s *ServerPool) AddBackend(backend *Backend) {
	s.backends = append(s.backends, backend)
}

func (s *ServerPool) NextIndex() int64 {
	s.current++
	return s.current % int64(len(s.backends))
}

func (s *ServerPool) GetNextBackend() *Backend {
	next := s.NextIndex()
	return s.backends[next]
}

func main() {

	// define list server
	servers := []string {"http://localhost:3001", "http://localhost:3002", "http://localhost:3003"}
	// define pod is listend
	var port int 

	flag.IntVar(&port, "port", 3000, "Port to serve")
	fmt.Println("port: ",port)

	fmt.Println("server: ", servers)

	serverPool := ServerPool{current: -1}

	for _, s := range servers {
		serverUrl, err := url.Parse(s)
		if err != nil {
				log.Fatal(err)
		}

		proxy := httputil.NewSingleHostReverseProxy(serverUrl)
		serverPool.AddBackend(&Backend{
				URL:          serverUrl,
				Alive:        true,
				ReverseProxy: proxy,
		})

	}

	server := http.Server{
		Addr: fmt.Sprintf(":%d", port),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				peer := serverPool.GetNextBackend()
				fmt.Println("per: ", *peer)
				if peer != nil {
						fmt.Println("ok")
						peer.ReverseProxy.ServeHTTP(w, r)
						return
				}

				http.Error(w, "Service not available", http.StatusServiceUnavailable)
		}),
	}

	log.Printf("Load Balancer started at :%d\n", port)
	if err := server.ListenAndServe(); err != nil {
			log.Fatal(err)
	}
}