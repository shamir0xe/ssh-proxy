package main

import (
	"context"
	"log"
	"ssh_proxy/dependencies"
	"ssh_proxy/services"

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
	if err := container.Provide(services.NewSSHManagerService); err != nil {
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
	ctx := context.Background()

	err = container.Invoke(func(sc services.SSHManagerInterface) error {
		return sc.Run(ctx)
	})
	if err != nil {
		panic(err)
	}

	log.Println("Application stopped gracefully")
}
