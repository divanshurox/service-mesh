package server

import (
	"ServiceMesh/internal/servicemesh/handler"
	"ServiceMesh/internal/servicemesh/middleware"
	"ServiceMesh/internal/servicemesh/models"
	"ServiceMesh/internal/servicemesh/prometheus"
	"ServiceMesh/internal/servicemesh/routes"
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func InitializeServer() {
	http.Handle("/", middleware.AddLogging(http.HandlerFunc(handler.ProxyHandler)))
	prometheus.PrometheusInit()
	server := &http.Server{Addr: ":8080"}
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()
	routesCh := make(chan []models.Route)
	go routes.PollRoutes(routesCh)
	go routes.UpdateRoutes(routesCh)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	log.Println("Server exiting")
}
