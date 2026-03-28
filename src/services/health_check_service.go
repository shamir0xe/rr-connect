package services

import (
	"context"
	"fmt"
	"log"
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
		counter:          0,
	}, nil
}

func (hc *healthCheckStruct) Run(ctx context.Context, wg *sync.WaitGroup, triggerChan chan<- bool) error {
	wg.Add(1)
	defer wg.Done()

	var timer *time.Timer = time.NewTimer(10 * time.Millisecond)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("health-check -> END")
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

func (hc *healthCheckStruct) checkConnectivity(ctx context.Context) bool {
	timeoutCtx, cancel := context.WithTimeout(ctx, hc.timeoutDuration)
	defer cancel()

	log.Printf("health-check -> do\n")
	log.Printf("health-check -> curl --socks5-hostname %s:%d %s\n", hc.socksHost, hc.socksPort, hc.healthCheckURL)
	cmd := exec.CommandContext(timeoutCtx,
		"curl",
		"--socks5-hostname", fmt.Sprintf("%s:%d", hc.socksHost, hc.socksPort),
		hc.healthCheckURL,
	)

	if err := cmd.Run(); err == nil {
		log.Printf("health-check -> GOOD!\n")
		hc.counter = 0
		return true
	}
	hc.counter++
	log.Printf("health-check -> BAD x( (%d)\n", hc.counter)
	if hc.counter < hc.maxRetries {
		log.Println("lower than max-retries, continue")
		return true
	}
	log.Println("reaches max retries, trigger config switch")
	// reaches max retries, trigger config switch
	hc.counter = 0
	return false
}
