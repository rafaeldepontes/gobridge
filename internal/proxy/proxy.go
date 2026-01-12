package proxy

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/url"
	"sync/atomic"
)

type ReverseProxy struct {
	// Routes is a hash map compose of a key string and a slice of strings, each
	// string need to represent an endpoint/service, so if you have more than
	// one replica running, it should be on this map.
	//
	// It uses a Load Balancer to choose what service gets the request, the
	// LB implementation is done via round robin.
	routes map[string][]string
	id     *uint64
}

func uint64Ptr(num uint64) *uint64 {
	return &num
}

func NewReverseProxy() *ReverseProxy {
	var initialId uint64 = 0
	return &ReverseProxy{
		routes: map[string][]string{
			"/todos/": {
				// The simplicity here is just for demonstration...
				"https://jsonplaceholder.typicode.com",
				"https://jsonplaceholder.typicode.com",
				"https://jsonplaceholder.typicode.com",
			},
		},
		id: uint64Ptr(initialId),
	}
}

// getService uses a Round Robin algorithm to choose what gonna take the request
// and give it back a string and a boolean if none.
func (rp *ReverseProxy) getService(path string) (string, bool) {
	val, ok := rp.routes[path]
	if !ok || len(val) == 0 {
		return "", false
	}
	return RoundRobin(val, rp.id), true
}

// transformBody just changes the CamelCase to snake_case...
func (rp *ReverseProxy) transformBody(body []byte) []byte {
	return bytes.ReplaceAll(body, []byte("userId"), []byte("user_id"))
}

func (rp *ReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	parts := bytes.Split([]byte(r.URL.Path), []byte("/todos/"))
	if len(parts) < 2 {
		http.Error(w, "Not Found", http.StatusBadRequest)
		return
	}

	if len(parts[0]) == 0 {
		parts[0] = []byte("/todos/")
	}

	backend, exists := rp.getService(string(parts[0]))
	if !exists {
		http.Error(w, "Not Found", http.StatusBadRequest)
		return
	}

	remote, err := url.Parse(backend)
	if err != nil {
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}

	pReq, err := http.NewRequest(r.Method, remote.String()+r.URL.Path, r.Body)
	if err != nil {
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}

	pReq.Header = r.Header

	resp, err := http.DefaultClient.Do(pReq)
	if err != nil {
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}

	msg := rp.transformBody(body)

	for k, v := range resp.Header {
		w.Header()[k] = v
	}
	w.WriteHeader(resp.StatusCode)
	_, err = w.Write(msg)
	if err != nil {
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}

	log.Printf("Request: %s, Response: %s", r.URL.Path, string(msg))
}

func RoundRobin(src []string, id *uint64) string {
	idx := atomic.AddUint64(id, 1)
	return src[idx%uint64(len(src))]
}

