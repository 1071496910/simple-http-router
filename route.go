package router

import (
	"github.com/1071496910/simple-http-router/lib/dispatcher"
	"net/http"
	"path/filepath"
	"sync"
)

type Route struct {
	route   map[string]map[string]http.Handler
	dps     map[string]dispatcher.Dispatcher
	mtx     sync.Mutex
	filters []filterFunc
}

type filterFunc func(rw http.ResponseWriter, r *http.Request) bool

func New() *Route {
	routeMap := make(map[string]map[string]http.Handler)
	dps := make(map[string]dispatcher.Dispatcher)
	for _, method := range []string{"GET", "POST", "PUT", "DELETE", "HEAD"} {
		routeMap[method] = make(map[string]http.Handler)
		dps[method] = dispatcher.NewDispatcher()
	}
	return &Route{
		route:   routeMap,
		dps:     dps,
		filters: make([]filterFunc, 0),
	}
}

func (rt *Route) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	for _, filter := range rt.filters {
		accept := filter(rw, r)
		if !accept {
			return
		}
	}
	url := filepath.Join(r.URL.String())
	location, err := rt.dps[r.Method].Dispatch(url)
	if err != nil {
		if err == dispatcher.ErrNoRoute {
			http.NotFound(rw, r)
			return
		}
		http.Error(rw, "unknow err", 503)
		return
	}
	rt.route[r.Method][location].ServeHTTP(rw, r)
}

func (r *Route) Filter(fn filterFunc) {
	r.filters = append(r.filters, fn)
}

func (r *Route) Head(path string, handlerFunc http.HandlerFunc) {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	r.dps["HEAD"].Register(path)
	r.route["HEAD"][filepath.Join(path)] = handlerFunc
	return
}

//Handle for all method
func (r *Route) Handle(path string, handlerFunc http.HandlerFunc) {
	r.Get(path, handlerFunc)
	r.Post(path, handlerFunc)
	r.Put(path, handlerFunc)
	r.Delete(path, handlerFunc)
	r.Head(path, handlerFunc)
}

func (r *Route) Get(path string, handlerFunc http.HandlerFunc) {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	r.dps["GET"].Register(path)
	r.route["GET"][filepath.Join(path)] = handlerFunc
	return
}

func (r *Route) Put(path string, handlerFunc http.HandlerFunc) {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	r.dps["PUT"].Register(path)
	r.route["PUT"][filepath.Join(path)] = handlerFunc
	return
}

func (r *Route) Post(path string, handlerFunc http.HandlerFunc) {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	r.dps["POST"].Register(path)
	r.route["POST"][filepath.Join(path)] = handlerFunc
	return
}

func (r *Route) Delete(path string, handlerFunc http.HandlerFunc) {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	r.dps["DELETE"].Register(path)
	r.route["DELETE"][filepath.Join(path)] = handlerFunc
	return
}
