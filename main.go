package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"time"

	"github.com/CloudAceTW/go-gcs-signedurl/controller"
	"github.com/CloudAceTW/go-gcs-signedurl/otel"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

var notToLogEndpoints = []string{"/api/v1/hck"}

func main() {
	if err := httpServerRun(); err != nil {
		log.Fatalln(err)
	}
	os.Exit(0)
}

func httpServerRun() (err error) {
	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Minute*1, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.Parse()

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	// Set up OpenTelemetry.
	otelShutdown, err := otel.SetupOTelSDK(ctx)
	if err != nil {
		return
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Defaulting to port %s", port)

	h2s := &http2.Server{}
	srv := &http.Server{
		Addr: fmt.Sprintf("0.0.0.0:%s", port),
		// Good practice to set timeouts to avoid Slowloris attacks.
		ReadHeaderTimeout: time.Second * 3,
		WriteTimeout:      time.Second * 30,
		ReadTimeout:       time.Second * 30,
		IdleTimeout:       time.Second * 90,
		Handler:           h2c.NewHandler(newHTTPHandler(), h2s), // Pass our instance of gorilla/mux in.
	}
	// Run our server in a goroutine so that it doesn't block.
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	err = srv.Shutdown(ctx)
	if err != nil {
		log.Printf("shut down err: %+v", err)
	}

	// Handle shutdown properly so nothing leaks.
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	log.Println("shutting down")
	return
}

func newHTTPHandler() http.Handler {
	router := gin.Default()

	router.Use(otelgin.Middleware("go-gcs-signedurl", otelgin.WithFilter(skipHealthCheck)))
	router.Use(cors.New(cors.Config{
		AllowMethods:     []string{"GET", "POST", "OPTIONS", "PUT", "PATCH"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type", "Auth"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
		AllowOriginFunc: func(origin string) bool {
			return true
		},
	}))

	router.GET("/api/v1/upload", controller.Upload)
	router.POST("/api/v1/upload/done", controller.Done)
	router.GET("/api/v1/file/:id", controller.File)
	router.GET("/api/v1/hck", controller.HealthCheck)

	return router
}

func skipHealthCheck(r *http.Request) bool {
	log.Printf("URL: %s", r.URL.Path)
	return slices.Index(notToLogEndpoints, r.URL.Path) == -1
}
