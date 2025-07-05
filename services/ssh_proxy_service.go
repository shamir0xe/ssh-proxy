package services

import (
	"context"
	"github.com/shamir0xe/ssh-proxy/dependencies"
	"log"
	"os/exec"
	"sync"
	"time"
)

type SSHProxyInterface interface {
	Run(context.Context,
		*sync.WaitGroup,
		<-chan bool,
	) error
}

type sshProxyService struct {
	config dependencies.ConfigInterface
}

func NewSSHProxyService(
	config dependencies.ConfigInterface,
) (SSHProxyInterface, error) {
	return &sshProxyService{config: config}, nil
}

func (sc *sshProxyService) Run(
	ctx context.Context,
	wg *sync.WaitGroup,
	restartChan <-chan bool,
) error {
	wg.Add(1)
	defer wg.Done()

	vp := sc.config

	url, err := vp.GetString("server.url")
	if err != nil {
		return err
	}

	port, err := vp.GetString("server.port")
	if err != nil {
		return err
	}

	password, err := vp.GetString("server.password")
	if err != nil {
		return err
	}

	socksPort, err := vp.GetString("socks.port")
	if err != nil {
		return err
	}

	waitTime, err := vp.GetInteger("health-check.wait-time")
	if err != nil {
		return err
	}

	for {
		log.Println("Starting SSH tunnel...")
		tunnelCtx, cancel := context.WithCancel(ctx)
		loop := func() bool {
			defer cancel()

			// Use a goroutine to run the command
			go func() {
				cmd := exec.CommandContext(
					tunnelCtx, "sshpass", "-p", *password,
					"ssh", "-D", *socksPort, "-p", *port,
					"-C", "-q", "-N", *url,
				)

				log.Printf("Going to start SSH, %v", cmd)
				err := cmd.Run()
				if err != nil {
					log.Printf("SSH command error: %v", err)
				} else {
					log.Println("SSH command exited unexpectedly")
				}
			}()
			log.Println("SSH task submitted...")

			select {
			case <-restartChan:
				log.Println("Restart signal received, restarting SSH tunnel")
				return true
			case <-ctx.Done():
				log.Println("SSH tunnel shutting down due to context cancellation")
				return false
			}
		}()
		if !loop {
			timer := time.NewTimer(time.Second * time.Duration(*waitTime))
			log.Printf("Wait before start again for %d s", *waitTime)
			<-timer.C
			timer.Stop()
			break
		}
	}
	return nil
}
