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

	for {
		select {
		case <-ctx.Done():
			log.Println("Health check stopping due to context cancellation")
			return nil
		case <-time.After(interval):
			cmd := exec.Command("proxychains4", "curl", "-4", "icanhazip.com")
			_, err := cmd.CombinedOutput()

			if err != nil {
				log.Printf("Health check failed ❌: %v", err)
				restartChan <- true
			} else {
				log.Printf("Health check success ✅")
			}
		}
	}
}
