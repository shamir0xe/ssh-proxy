package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"ssh_proxy/dependencies"
	"ssh_proxy/services"
	"sync"
	"syscall"

	"go.uber.org/dig"
)

func newContainer() (*dig.Container, error) {
	container := dig.New()

	if err := container.Provide(services.NewMonitoringService); err != nil {
		return nil, err
	}
	if err := container.Provide(services.NewSSHProxyService); err != nil {
		return nil, err
	}
	if err := container.Provide(dependencies.NewViperConfig); err != nil {
		return nil, err
	}

	return container, nil
}

func main() {
	container, err := newContainer()
	if err != nil {
		panic(err)
	}
	log.Println("Container initialized")

	// Create context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create wait group
	var wg sync.WaitGroup

	// define restart channel
	restartChan := make(chan bool)

	err = container.Invoke(func(sc services.SSHProxyInterface) error {
		go sc.Run(ctx, &wg, restartChan)
		return nil
	})
	if err != nil {
		panic(err)
	}
	err = container.Invoke(func(sc services.MonitoringServiceInterface) error {
		go sc.Run(ctx, &wg, restartChan)
		return nil
	})
	if err != nil {
		panic(err)
	}

	// Create a channel to listen for interrupt signals
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	// Block until an interrupt signal is received
	<-signalChan
	log.Println("Interrupt signal received, shutting down...")

	// Cancel the context to stop goroutines
	cancel()

	// Wait for all goroutines to finish
	wg.Wait()

	log.Println("Application stopped gracefully")
}
