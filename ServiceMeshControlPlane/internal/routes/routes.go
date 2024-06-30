package routes

import (
	"ServiceMeshControlPlane/internal/model"
	"ServiceMeshControlPlane/internal/utils"
	"encoding/json"
	"log"
	"net/http"
	"sync"
)

var (
	Routes      = []model.Route{}
	m           sync.RWMutex
	servicesMap = make(map[string][]string)
)

func DiscardEndpoint(targetHost, service string) {
	m.Lock()
	if targetHosts, exists := servicesMap[service]; exists {
		for idx, host := range targetHosts {
			if targetHost == host {
				servicesMap[service] = utils.Remove(servicesMap[service], idx)
				if len(servicesMap[service]) == 0 {
					delete(servicesMap, service)
				}
				m.Unlock()
				updateRoutes()
			}
		}
	}
}

func updateRoutes() {
	m.RLock()
	defer m.RUnlock()
	Routes = []model.Route{}
	for service, targetHosts := range servicesMap {
		Routes = append(Routes, model.Route{
			Service:     service,
			TargetHosts: targetHosts,
		})
	}
}

func checkSingleEndpoint(targetHost, service string) {
	endpoint := targetHost + service
	log.Printf("Checking health of these endpoints: %s", endpoint)
	resp, err := http.Get(endpoint)
	if err != nil {
		log.Println(err.Error())
		DiscardEndpoint(targetHost, service)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 && resp.StatusCode >= 300 {
		DiscardEndpoint(targetHost, service)
	}
}

func ActiveHealthCheck() {
	log.Println("Checking health of endpoints")
	m.RLock()
	defer m.RUnlock()
	for _, route := range Routes {
		path := route.Service
		log.Printf("Checking health of service: %s", path)
		for _, targetHost := range route.TargetHosts {
			go checkSingleEndpoint(targetHost, path)
		}
	}
}

func EncodeRoutes(w http.ResponseWriter, r *http.Request) {
	m.RLock()
	defer m.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(Routes); err != nil {
		panic(err)
	}
}

func AddNewServices(service, targetHost string, w http.ResponseWriter) {
	m.Lock()
	if targetHosts, exists := servicesMap[service]; exists {
		for _, host := range targetHosts {
			if targetHost == host {
				w.WriteHeader(http.StatusOK)
				return
			}
		}
		servicesMap[service] = append(servicesMap[service], targetHost)
	} else {
		servicesMap[service] = []string{
			targetHost,
		}
	}
	m.Unlock()
	updateRoutes()
}
