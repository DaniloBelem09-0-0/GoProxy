package balancer

import (
	"net/url"
	"sync"
	"sync/atomic"
)

type Backend struct {
	URL   *url.URL
	Alive bool
	mux   sync.RWMutex
}

type RoundRobinBalancer struct {
	targets []*Backend
	current uint64
	mux     sync.RWMutex
}

type Balancer interface {
	GetNext() *Backend
}

func NewRoundRobinBalancer(targets []string) *RoundRobinBalancer {
	var urls []*url.URL
	for _, target := range targets {
		u, err := url.Parse(target)
		if err != nil {
			continue
		}
		urls = append(urls, u)
	}

	backends := make([]*Backend, len(urls))

	for i, u := range urls {
		backends[i] = &Backend{URL: u, Alive: true}
	}

	return &RoundRobinBalancer{
		targets: backends,
	}
}

func (rr *RoundRobinBalancer) GetNext() *Backend {
	rr.mux.RLock()
	defer rr.mux.RUnlock()

	if len(rr.targets) == 0 {
		return nil
	}
	index := atomic.AddUint64(&rr.current, 1)
	return rr.targets[(index-1)%uint64(len(rr.targets))]
}

func (rr *RoundRobinBalancer) GetTargets() []*Backend {
	rr.mux.RLock()
	defer rr.mux.RUnlock()
	return rr.targets
}

func (b *Backend) SetStatus(alive bool) {
	b.mux.Lock()
	b.Alive = alive
	b.mux.Unlock()
}

func (b *Backend) IsAlive() bool {
	b.mux.RLock()
	alive := b.Alive
	b.mux.RUnlock()
	return alive
}

func (rr *RoundRobinBalancer) UpdateTargets(newTargets []string) {
	var backends []*Backend
	for _, t := range newTargets {
		u, _ := url.Parse(t)
		backends = append(backends, &Backend{URL: u, Alive: true})
	}

	rr.mux.Lock()
	rr.targets = backends
	rr.mux.Unlock()
}