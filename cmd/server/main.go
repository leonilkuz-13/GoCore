package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	moviesv1 "github.com/leonilkuz-13/GoCore/api/proto"
	_ "github.com/leonilkuz-13/GoCore/docs"
	"github.com/leonilkuz-13/GoCore/internal/adapter"
	"github.com/leonilkuz-13/GoCore/internal/config"
	movieGRPC "github.com/leonilkuz-13/GoCore/internal/handler/grpc"
	movieHTTP "github.com/leonilkuz-13/GoCore/internal/handler/http"
	"github.com/leonilkuz-13/GoCore/internal/usecase"
	"google.golang.org/grpc"
)

// @title		GoCore Movie API
// @version		1.0
// @description	HTTP API for loading movies from an external movie service.
// @BasePath		/
func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	movieClient := adapter.NewMovieClient(cfg.MovieBaseURL, cfg.RequestTimeout)
	movieService := usecase.NewMovieUsecase(movieClient, cfg.MaxConcurrency)

	httpServer := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           movieHTTP.NewHandler(movieService).Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	grpcListener, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		return err
	}
	grpcServer := grpc.NewServer()
	moviesv1.RegisterMovieServiceServer(grpcServer, movieGRPC.NewServer(movieService))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverErrors := make(chan error, 2)
	go func() {
		log.Printf("HTTP server listening on %s", cfg.HTTPAddr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
			return
		}
		serverErrors <- nil
	}()
	go func() {
		log.Printf("gRPC server listening on %s", cfg.GRPCAddr)
		serverErrors <- grpcServer.Serve(grpcListener)
	}()

	var runErr error
	select {
	case <-ctx.Done():
		log.Print("Shutdown signal received")
	case err = <-serverErrors:
		runErr = err
		if err != nil {
			log.Printf("Server error: %v", err)
		}
	}
	stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	grpcStopped := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(grpcStopped)
	}()

	httpErr := httpServer.Shutdown(shutdownCtx)
	select {
	case <-grpcStopped:
	case <-shutdownCtx.Done():
		grpcServer.Stop()
		<-grpcStopped
	}

	if err := <-serverErrors; err != nil && runErr == nil {
		runErr = err
	}
	if httpErr != nil && runErr == nil {
		runErr = httpErr
	}
	return runErr
}
