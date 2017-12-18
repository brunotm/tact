package tact

import (
	"fmt"
	"strings"
	"sync"
)

// registry container for Collectors
type registry struct {
	mtx        *sync.RWMutex
	collectors map[string]*Collector
	groups     map[string][]*Collector
}

// Add a Collector path type.collector to the registry
func (r *registry) Add(collector *Collector) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	if collector.Name == "" {
		panic("registry: added collector with empty name")
	}
	if collector.GetData == nil {
		panic("registry: added collector with null collector function")
	}

	if _, ok := r.collectors[collector.Name]; ok {
		panic(fmt.Sprintf("registry: collector already exists: %s", collector.Name))
	}
	r.collectors[collector.Name] = collector

	path := strings.Split(collector.Name, "/")
	group := strings.Join(path[:len(path)-1], "/")
	r.groups[group] = append(r.groups[group], collector)

}

// Get fetches the Collector for the given name
func (r *registry) Get(name string) *Collector {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	collector, ok := r.collectors[name]
	if ok {
		return collector
	}

	panic(fmt.Sprintf("registry: collector %s does not exist", name))
}

// GetGroup fetches the Collector for the given name
func (r *registry) GetGroup(name string) []*Collector {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	collectors, ok := r.groups[name]
	if ok {
		return collectors
	}

	panic(fmt.Sprintf("registry: collector group %s does not exist", name))
}

// Del deletes the CollectorSet for the given name
func (r *registry) del(name string) error {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	_, ok := r.collectors[name]
	if !ok {
		return fmt.Errorf("registry: collector %s does not exist", name)
	}

	delete(r.collectors, name)
	return nil
}
