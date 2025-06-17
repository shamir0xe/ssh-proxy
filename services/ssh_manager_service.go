package services

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type SSHManagerInterface interface {
	Run(context.Context) error
}

type sshManagerService struct {
	sshProxyService   SSHProxyInterface
	monitoringService MonitoringServiceInterface
}

func NewSSHManagerService(
	sshProxyService SSHProxyInterface,
	monitoringService MonitoringServiceInterface,
) (SSHManagerInterface, error) {
	return &sshManagerService{
		sshProxyService:   sshProxyService,
		monitoringService: monitoringService,
	}, nil
}

func (sc *sshManagerService) Run(ctx context.Context) error {
	// Create context
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Create wait group
	var wg sync.WaitGroup

	// define restart channel
	restartChan := make(chan bool)

	err := func() error {
		go sc.sshProxyService.Run(ctx, &wg, restartChan)
		return nil
	}()
	if err != nil {
		return err
	}
	err = func() error {
		go sc.monitoringService.Run(ctx, &wg, restartChan)
		return nil
	}()
	if err != nil {
		return err
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
	return nil
}
