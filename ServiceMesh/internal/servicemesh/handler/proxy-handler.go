package handler

import (
	prom "ServiceMesh/internal/servicemesh/prometheus"
	"ServiceMesh/internal/servicemesh/routes"
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	localServiceName string
	localServicePort int
)

func init() {
	var err error
	localServiceName = os.Getenv("LOCAL_SERVICE_NAME")
	if localServiceName == "" {
		log.Fatal("LOCAL_SERVICE_NAME environment variable is required")
	}
	localServicePort, err = strconv.Atoi(os.Getenv("LOCAL_SERVICE_PORT"))
	if err != nil || localServicePort == 0 {
		log.Fatal("LOCAL_SERVICE_PORT environment variable is required and must be a valid port number")
	}
}

func ForwardRequest(ctx context.Context, w http.ResponseWriter, r *http.Request, targetHost, service string, metricsMap prom.Metrics) {
	timer := prometheus.NewTimer(metricsMap.E2EDelay.With(prometheus.Labels{"service_from": localServiceName, "service_to": service}))
	defer timer.ObserveDuration()
	if !strings.Contains(targetHost, "localhost") {
		metricsMap.HttpRequests.With(prometheus.Labels{"service_from": localServiceName, "service_to": service}).Inc()
	}
	log.Printf("Making a request on %s", targetHost+service)
	req, err := http.NewRequestWithContext(ctx, r.Method, targetHost+service, r.Body)
	if err != nil {
		log.Fatal(err)
	}
	for k, v := range r.Header {
		req.Header[k] = v
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		metricsMap.HttpErrors.With(prometheus.Labels{"service_from": localServiceName, "service_to": service}).Inc()
		log.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 && resp.StatusCode >= 300 {
		metricsMap.HttpErrors.With(prometheus.Labels{"service_from": localServiceName, "service_to": service}).Inc()
		http.Error(w, "Error has occurred", resp.StatusCode)
	}
	for k, v := range resp.Header {
		w.Header()[k] = v
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func ProxyHandler(w http.ResponseWriter, r *http.Request) {
	service := r.URL.Path[1:]
	log.Printf("Got an incoming request for %s", service)
	found := routes.ServiceExists(service)
	if !found {
		log.Printf("Could not find any route object for service: %s", service)
		http.NotFound(w, r)
		return
	}
	log.Printf("Found route object for service: %s", service)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	targetHost, err := routes.GetTargetHost(service)
	if err != nil {
		log.Printf(err.Error())
		return
	}
	metricsMap := prom.GetMetricsMap()
	if service == localServiceName {
		// this means it is an incomming request as the final destination is the local service
		targetHost = "http://localhost:" + strconv.Itoa(localServicePort) + "/"
	}
	ForwardRequest(ctx, w, r, targetHost, service, metricsMap)
}
