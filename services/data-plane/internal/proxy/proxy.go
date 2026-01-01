package proxy

import (
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

type Proxy struct {
	target *url.URL
	engine *httputil.ReverseProxy
}

func NewProxy(target *url.URL) *Proxy {
	p := &Proxy{target: target}

	director := func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		
		// Preserve the original request path and query
		// It's critical for the backend to know who made the request, ensuring to have registered the client IP
		if clientIP, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
			req.Header.Set("X-Forwarded-For", clientIP)
		}
	}

	transport := &http.Transport{
		ResponseHeaderTimeout: 10 * time.Second,  // Timeout for reading response headers
		DialContext: (&net.Dialer{
			Timeout: 5 * time.Second, // Timeout for establishing connections
		}).DialContext,
	}

	p.engine = &httputil.ReverseProxy{
		Director:  director,
		Transport: transport,
		ErrorHandler: p.handleError,
	}

	return p
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.engine.ServeHTTP(w, r)
}

func (p *Proxy) handleError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("[PROXY ERROR] Target %s failed: %v", p.target, err)
	w.WriteHeader(http.StatusBadGateway)
	w.Write([]byte("Nexus-Mesh: Backend is unreachable (502 Bad Gateway)"))
}