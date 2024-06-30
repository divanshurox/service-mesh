package handler

import (
	"ServiceMeshControlPlane/internal/routes"
	"log"
	"net/http"
)

func HandleServiceDiscovery(w http.ResponseWriter, r *http.Request) {
	queryMap := r.URL.Query()
	service := queryMap.Get("service")
	targetHost := queryMap.Get("target_host")
	if service == "" || targetHost == "" {
		log.Println("Query parameters not provided correctly")
		http.Error(w, "Missing query parameters: 'service' and 'target_host' are required", http.StatusBadRequest)
		return
	}
	routes.AddNewServices(service, targetHost, w)
	w.WriteHeader(http.StatusCreated)
}

func HandleServices(w http.ResponseWriter, r *http.Request) {
	routes.EncodeRoutes(w, r)
}
