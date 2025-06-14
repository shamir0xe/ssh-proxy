package main

import (
	"context"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"ssh_proxy/dependencies"
	"sync"
	"syscall"
	"time"
)

func runSSHTunnel(
	ctx context.Context,
	wg *sync.WaitGroup,
	restartChan <-chan bool,
	url string,
	port string,
	password string,
) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			log.Println("SSH tunnel shutting down due to context cancellation")
			return
		default:
			log.Println("Starting SSH tunnel...")
			tunnelCtx, cancel := context.WithCancel(ctx)

			// Use a goroutine to run the command
			go func() {
				cmd := exec.CommandContext(tunnelCtx, "sshpass", "-p", password,
					"ssh", "-D", port, "-C", "-q", "-N", url)

				err := cmd.Run()
				if err != nil {
					log.Printf("SSH command error: %v", err)
				} else {
					log.Println("SSH command exited unexpectedly")
				}
			}()
			log.Println("SSH tunnel is running, waiting for restart signal")

			<-restartChan
			log.Println("Restart signal received, restarting SSH tunnel")
			cancel()
		}
	}
}

func monitorTunnel(
	ctx context.Context,
	wg *sync.WaitGroup,
	restartChan chan<- bool,
) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			log.Println("Health check stopping due to context cancellation")
			restartChan <- true
			return
		case <-time.After(30 * time.Second):
			log.Println("Checking tunnel health...")

			cmd := exec.Command("proxychains4", "curl", "-4", "icanhazip.com")
			_, err := cmd.CombinedOutput()

			if err != nil {
				log.Printf("Health check failed: %v", err)
				restartChan <- true
			} else {
				log.Printf("Health check success ✅")
			}
		}
	}
}

func main() {
	vp, err := dependencies.NewViperConfig()
	if err != nil {
		panic(err)
	}

	url, err := vp.GetString("server.url")
	if err != nil {
		panic(err)
	}

	port, err := vp.GetString("server.port")
	if err != nil {
		panic(err)
	}

	password, err := vp.GetString("server.password")
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	restartChan := make(chan bool)

	wg.Add(2)
	go runSSHTunnel(ctx, &wg, restartChan, *url, *port, *password)
	go monitorTunnel(ctx, &wg, restartChan)

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
