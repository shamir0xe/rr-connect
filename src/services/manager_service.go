package services

import (
	"context"
	"fmt"
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

	go mgr.cfgSwitchService.Run(ctx, &wg, triggerChan)
	go mgr.healthCheckService.Run(ctx, &wg, triggerChan)

	osSignalChan := make(chan os.Signal, 1)
	signal.Notify(osSignalChan, os.Interrupt)
	fmt.Println("Manager is running. Press Ctrl+C to stop.")
	<-osSignalChan

	cancel()

	wg.Wait()
	return nil
}
