package services

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"ssh_proxy/dependencies"
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

func (sc *monitoringService) Run(
	ctx context.Context,
	wg *sync.WaitGroup,
	restartChan chan<- bool,
) error {
	wg.Add(1)
	defer wg.Done()
	vp := sc.config

	timerInterval, err := vp.GetInteger("health_check.interval")
	if err != nil {
		return fmt.Errorf("could not find proper filed: %v", err)
	}
	interval := time.Duration(*timerInterval) * time.Second

	timeout, err := vp.GetInteger("health_check.timeout")
	if err != nil {
		return fmt.Errorf("could not find proper filed: %v", err)
	}
	timeoutDuration := time.Duration(*timeout) * time.Second

	conLimit, err := vp.GetInteger("health_check.consecutive_limit")
	if err != nil {
		return fmt.Errorf("Please provide health_check.consecutive_limit")
	}

	healthCheckCommandString, err := vp.GetString("health_check.command")
	if err != nil {
		return fmt.Errorf("Please provide health_check.command")
	}

	restartCommandString, err := vp.GetString("health_check.restart_command")
	if err != nil {
		return fmt.Errorf("Please provide health_check.restart_command")
	}

	timer := time.NewTimer(0)
	defer timer.Stop()

	conLoss := 0
	totalRestart := 0

	for {
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
					if conLoss > *conLimit {
						conLoss = 0
						totalRestart++
						if totalRestart > 5 {
							panic(fmt.Errorf("Total number of consecutive restarts exceeded"))
						}
						cmd := exec.CommandContext(ctx, "bash", []string{"-c", *restartCommandString}...)
						_, err = cmd.CombinedOutput()
						if err != nil {
							log.Printf("Cannot restart the service ❌")
						}

						return true
					}
				} else {
					conLoss = 0
					totalRestart = 0
					log.Printf("success ✅")
				}
				return false
			}() {
				restartChan <- true
			}

			timer.Reset(interval)
		}
	}
}
