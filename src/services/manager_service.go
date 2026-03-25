package services

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
)

type ManagerInterface interface {
	Run(context.Context) error
}

type managerStruct struct {
	cfgSwitchService   ConfigSwitchInterface
	healthCheckService HealthCheckInterface
}

func NewManager(
	cfgSwitchService ConfigSwitchInterface,
	healthCheckService HealthCheckInterface,
) (ManagerInterface, error) {
	return &managerStruct{
		cfgSwitchService:   cfgSwitchService,
		healthCheckService: healthCheckService,
	}, nil
}

func (mgr *managerStruct) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)

	var wg sync.WaitGroup
	var triggerChan = make(chan bool)

	log.Println("Creating goroutines...")

	go mgr.cfgSwitchService.Run(ctx, &wg, triggerChan)
	go mgr.healthCheckService.Run(ctx, &wg, triggerChan)

	osSignalChan := make(chan os.Signal, 1)
	signal.Notify(osSignalChan, os.Interrupt)
	log.Println("Manager is running. Press Ctrl+C to stop.")
	<-osSignalChan

	log.Println("Interrupt signal received, killing processes")
	cancel()

	log.Println("Shutting down gracefully, waiting...")
	wg.Wait()
	return nil
}
