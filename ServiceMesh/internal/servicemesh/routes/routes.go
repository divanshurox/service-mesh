package routes

import (
	"ServiceMesh/internal/servicemesh/models"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"
)

var (
	Routes       = []models.Route{}
	routesMutex  sync.RWMutex
	controlPlane = "http://control-plane-service"
)

var (
	roundRobinIndex = make(map[string]int)
	indexMutex      sync.RWMutex
)

func PollRoutes(ch chan<- []models.Route) {
	for {
		func() {
			resp, err := http.Get(controlPlane + "/get")
			if err != nil {
				log.Println(err.Error())
				return
			}
			defer resp.Body.Close()
			var updatedRoutes []models.Route

			if err := json.NewDecoder(resp.Body).Decode(&updatedRoutes); err != nil {
				log.Fatal(err)
			}
			ch <- updatedRoutes
		}()
		time.Sleep(30 * time.Second)
	}
}

func UpdateRoutes(routesCh <-chan []models.Route) {
	for updatedRoutes := range routesCh {
		routesMutex.Lock()
		Routes = updatedRoutes
		routesMutex.Unlock()
		log.Printf("Updating routes array: %#v", Routes)
		for _, route := range Routes {
			if _, exists := roundRobinIndex[route.Service]; !exists {
				indexMutex.Lock()
				roundRobinIndex[route.Service] = 0
				indexMutex.Unlock()
			}
		}
	}
}

func getRoutesIndexByService(service string) int {
	indexMutex.Lock()
	defer indexMutex.Unlock()
	index := roundRobinIndex[service]
	return index
}

func updateIndexForService(service string, index int) {
	indexMutex.Lock()
	defer indexMutex.Unlock()
	if _, exists := roundRobinIndex[service]; exists {
		roundRobinIndex[service] = index
	}
}

func ServiceExists(service string) bool {
	for _, route := range Routes {
		if route.Service == service {
			return true
		}
	}
	return false
}

func getRouteForService(service string) (models.Route, bool) {
	for _, route := range Routes {
		if route.Service == service {
			return route, true
		}
	}
	return models.Route{}, false
}

func GetTargetHost(service string) (string, error) {
	route, found := getRouteForService(service)
	if !found {
		return "", errors.New("Unable to find route for service: " + service)
	}
	index := getRoutesIndexByService(route.Service)
	log.Printf("Found route object inside getTargetHost %#v for %v index", route, index)
	targetHost := route.TargetHosts[index]
	newIndex := (index + 1) % len(route.TargetHosts)
	updateIndexForService(route.Service, newIndex)
	return targetHost, nil
}
