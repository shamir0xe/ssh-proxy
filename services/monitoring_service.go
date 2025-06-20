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

	timer := time.NewTimer(0)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Health check stopping due to context cancellation")
			return ctx.Err()

		case <-timer.C:
			func() {
				// Run the health check logic
				timedCtx, cancel := context.WithTimeout(ctx, timeoutDuration)
				defer cancel()

				log.Printf("Health check:")
				cmd := exec.CommandContext(timedCtx, "proxychains4", "curl", "-4", "icanhazip.com")
				output, err := cmd.CombinedOutput()

				if err != nil {
					log.Printf("failed ❌: %v", err)
					if timedCtx.Err() == context.DeadlineExceeded {
						log.Printf("Timeout exceeded after %s", timeoutDuration)
					} else {
						log.Printf("Output from failed command: %s", string(output))
					}
					restartChan <- true
				} else {
					log.Printf("success ✅")
				}
			}()

			timer.Reset(interval)
		}
	}
}
