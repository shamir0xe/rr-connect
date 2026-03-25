package main

import (
	"context"
	"fmt"
	"log"

	"github.com/shamir0xe/rr-connect/src/dependencies"
	"github.com/shamir0xe/rr-connect/src/services"
	"go.uber.org/dig"
)

func createContainer() (*dig.Container, error) {
	container := dig.New()
	var err error
	err = container.Provide(dependencies.NewViperConfig)
	if err != nil {
		return nil, err
	}
	err = container.Provide(services.NewRouter)
	if err != nil {
		return nil, err
	}
	err = container.Provide(services.NewConfigSwitchService)
	if err != nil {
		return nil, err
	}
	err = container.Provide(services.NewHealthCheckService)
	if err != nil {
		return nil, err
	}
	err = container.Provide(services.NewManager)
	if err != nil {
		return nil, err
	}
	return container, nil
}

func main() {
	fmt.Println("Welcome to rr-connect!")
	// so basically, it will be a infinite loop that checks the connectivity
	// via curl command, if it works, reset the counter,
	// if it fails, increase the counter by 1, if the counter reaches THRESHOLD,
	// switch to the next config available in the configs.yaml file.
	container, err := createContainer()
	if err != nil {
		log.Fatalf("Failed to create container: %v", err)
	}
	container.Invoke(func(mgr services.ManagerInterface) {
		if err := mgr.Run(context.Background()); err != nil {
			log.Fatalf("Manager run failed: %v", err)
		}
	})
}
