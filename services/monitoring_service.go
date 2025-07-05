package services

import (
	"context"
	"fmt"
	"github.com/shamir0xe/ssh-proxy/dependencies"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"
)

type MonitoringServiceInterface interface {
	Run(context.Context, *sync.WaitGroup, chan<- bool) error
}

type monitoringService struct {
	config dependencies.ConfigInterface
}

func NewMonitoringService(
	config dependencies.ConfigInterface,
) (MonitoringServiceInterface, error) {
	return &monitoringService{config: config}, nil
}

func updateHealthStatusFile(filePath string, isHealthy bool) {
	var status string
	if isHealthy {
		status = "1"
	} else {
		status = "0"
	}

	// The content to write must be a slice of bytes.
	// We convert our status string to []byte.
	content := []byte(status)

	// os.WriteFile handles creating/truncating the file and closing it.
	// 0644 is a standard file permission:
	// - The owner can read and write (6)
	// - The group can only read (4)
	// - Others can only read (4)
	err := os.WriteFile(filePath, content, 0644)
	if err != nil {
		// Using log.Fatalf will print the error and exit the program.
		log.Fatalf("Failed to write to health-check file: %v", err)
	}

	fmt.Printf("Successfully wrote '%s' to %s\n", status, filePath)
}

func (sc *monitoringService) Run(
	ctx context.Context,
	wg *sync.WaitGroup,
	restartChan chan<- bool,
) error {
	wg.Add(1)
	defer wg.Done()
	vp := sc.config

	timerInterval, err := vp.GetInteger("health-check.interval")
	if err != nil {
		return fmt.Errorf("could not find proper filed: %v", err)
	}
	interval := time.Duration(*timerInterval) * time.Second

	timeout, err := vp.GetInteger("health-check.timeout")
	if err != nil {
		return fmt.Errorf("could not find proper filed: %v", err)
	}
	timeoutDuration := time.Duration(*timeout) * time.Second

	conLimit, err := vp.GetInteger("health-check.consecutive-limit")
	if err != nil {
		return fmt.Errorf("Please provide health-check.consecutive-limit")
	}

	tunnelLimit, err := vp.GetInteger("health-check.tunnel-limit")
	if err != nil {
		return fmt.Errorf("Please provide health-check.tunnel-limit")
	}

	healthCheckCommandString, err := vp.GetString("health-check.command")
	if err != nil {
		return fmt.Errorf("Please provide health-check.command")
	}

	healthStatusFile, err := vp.GetString("health-check.file-path")
	if err != nil {
		return fmt.Errorf("Please provide health-check.file-path")
	}

	timer := time.NewTimer(0)
	defer timer.Stop()

	conLoss := 0
	totalRestart := 0
	first := true

	for {
		if !first {
			timer.Reset(interval)
		}
		first = false

		select {
		case <-ctx.Done():
			log.Println("Health check stopping due to context cancellation")
			return ctx.Err()

		case <-timer.C:
			if func() bool {
				// Run the health check logic
				timedCtx, cancel := context.WithTimeout(ctx, timeoutDuration)
				defer cancel()

				log.Printf("Health check:")
				cmd := exec.CommandContext(timedCtx, "bash", []string{"-c", *healthCheckCommandString}...)
				output, err := cmd.CombinedOutput()

				if err != nil {
					log.Printf("failed ❌: %v", err)
					if timedCtx.Err() == context.DeadlineExceeded {
						log.Printf("Timeout exceeded after %s", timeoutDuration)
					} else {
						log.Printf("Output from failed command: %s", string(output))
					}
					conLoss++
					log.Printf("consecutive loss: %d", conLoss)

					// Set the connection limit
					if conLoss > *conLimit {
						conLoss = 0
						totalRestart++
						if totalRestart > 5 {
							panic(fmt.Errorf("Total number of consecutive restarts exceeded"))
						}
						updateHealthStatusFile(*healthStatusFile, false)
					}

					// Check the tunnel limit -- tunnel limit << connection limit
					if conLoss > *tunnelLimit {
						return true
					}
				} else {
					conLoss = 0
					totalRestart = 0
					log.Printf("success ✅")
					updateHealthStatusFile(*healthStatusFile, true)
				}
				return false
			}() {
				restartChan <- true
			}

		}
	}
}
