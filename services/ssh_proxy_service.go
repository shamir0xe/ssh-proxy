package services

import (
	"fmt"
	"context"
	"github.com/shamir0xe/ssh-proxy/dependencies"
	"log"
	"os"
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
			defer func() {
				cancel()
				log.Println("canceling the previous go-routine, waiting %vs", *waitTime)
				<-time.After(time.Second * time.Duration(*waitTime))
				log.Println("Waiting Done!")
			}()

			// Use a goroutine to run the command
			go func() {
				log.Println("NEW GO ROUTINE")
				cmd := exec.CommandContext(
					tunnelCtx, "sshpass",
					"-p", *password,
					"/usr/bin/ssh",
					"-D", *socksPort,
					"-p", *port,
					"-c", "chacha20-poly1305@openssh.com",
					"-o", "Compression=no",
					"-o", "StrictHostKeyChecking=no",
					"-o", "TCPKeepAlive=yes",
					"-o", "ServerAliveInterval=30",
					"-o", "ServerAliveCountMax=6",
					"-o", "IPQoS=throughput",
					"-o", "GSSAPIAuthentication=no",
					"-o", "ControlPersist=10m",
					"-o", "ControlPath=/tmp/ssh_mux_%h_%p_%r",
					"-o", "LogLevel=ERROR",
					"-o", "ProxyCommand=/usr/bin/nc -x 127.0.0.1:10808 %h %p",
					*url,
					"-q",
					"-N",
				)

				log.Printf("Going to start SSH, %v", cmd)
				cmd.Env = os.Environ()
				out, err := cmd.CombinedOutput()
				if err != nil {
					fmt.Println(string(out))
					fmt.Println(err)
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
			log.Println("SHOULD not be PRINTED")
			return false
		}()
		log.Printf("output of the loop: %v", loop)
		if !loop {
			break
		}
	}
	return nil
}
