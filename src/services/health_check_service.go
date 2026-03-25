package services

import (
	"context"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"github.com/spf13/viper"
)

type HealthCheckInterface interface {
	Run(context.Context, *sync.WaitGroup, chan<- bool) error
}

type healthCheckStruct struct {
	intervalDuration time.Duration
	timeoutDuration  time.Duration
	healthCheckURL   string
	socksHost        string
	socksPort        int
	counter          int
	maxRetries       int
}

func NewHealthCheckService(cfg *viper.Viper) (HealthCheckInterface, error) {
	sub := cfg.Sub("health-check")
	return &healthCheckStruct{
		intervalDuration: sub.GetDuration("interval-duration"),
		timeoutDuration:  sub.GetDuration("timeout-duration"),
		healthCheckURL:   sub.GetString("url"),
		socksHost:        sub.GetString("socks.host"),
		socksPort:        sub.GetInt("socks.port"),
		maxRetries:       sub.GetInt("max-retries"),
	}, nil
}

func (hc healthCheckStruct) Run(ctx context.Context, wg *sync.WaitGroup, triggerChan chan<- bool) error {
	wg.Add(1)
	defer wg.Done()

	var timer *time.Timer = time.NewTimer(10 * time.Millisecond)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-timer.C:
			if !hc.checkConnectivity(ctx) {
				triggerChan <- true
			}
		}
		timer.Stop()
		timer = time.NewTimer(hc.intervalDuration)
	}
}

func (hc healthCheckStruct) checkConnectivity(ctx context.Context) bool {
	timeoutCtx, cancel := context.WithTimeout(ctx, hc.timeoutDuration)
	defer cancel()

	fmt.Printf("health-check -> do\n")
	fmt.Printf("health-check -> curl --socks5-hostname %s:%d %s\n", hc.socksHost, hc.socksPort, hc.healthCheckURL)
	cmd := exec.CommandContext(timeoutCtx,
		"curl",
		"--socks5-hostname", fmt.Sprintf("%s:%d", hc.socksHost, hc.socksPort),
		hc.healthCheckURL,
	)

	if err := cmd.Run(); err == nil {
		fmt.Printf("health-check -> YES\n")
		hc.counter = 0
		return true
	}
	hc.counter++
	fmt.Printf("health-check -> NO (%d)\n", hc.counter)
	if hc.counter < hc.maxRetries {
		return true
	}
	// reaches max retries, trigger config switch
	hc.counter = 0
	return false
}
