package server

import (
	"ServiceMeshControlPlane/internal/handler"
	"ServiceMeshControlPlane/internal/middleware"
	"ServiceMeshControlPlane/internal/routes"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func InitializeServer() {
	http.Handle("/notify", middleware.AddLogger(http.HandlerFunc(handler.HandleServiceDiscovery)))
	http.Handle("/get", middleware.AddLogger(http.HandlerFunc(handler.HandleServices)))
	server := &http.Server{Addr: ":8090"}
	go func() {
		if err := server.ListenAndServe(); err != nil {
			panic(err)
		}
	}()
	go func() {
		for {
			routes.ActiveHealthCheck()
			time.Sleep(1 * time.Minute)
		}
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGABRT)
	<-stop

	log.Println("Closing server")
	if err := server.Close(); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	log.Println("Server exiting")
}
