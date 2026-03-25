package main

import (
	"context"
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
	log.Println("Welcome to rr-connect!")
	// so basically, it will be a infinite loop that checks the connectivity
	// via curl command, if it works, reset the counter,
	// if it fails, increase the counter by 1, if the counter reaches THRESHOLD,
	// switch to the next config available in the configs.yaml file.
	container, err := createContainer()
	if err != nil {
		log.Fatalf("Failed to create container: %v", err)
	}
	err = container.Invoke(func(mgr services.ManagerInterface) error {
		err := mgr.Run(context.Background())
		if err != nil {
			log.Fatalf("Manager run failed: %v", err)
		}
		log.Println("yoyo")
		return err
	})
	if err != nil {
		panic(err)
	}
}
