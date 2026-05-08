package services

import (
	"context"
	"testing"
	"time"

	"github.com/spf13/viper"
)

func newTestHealthCheck(maxRetries int) *healthCheckStruct {
	hc := &healthCheckStruct{
		timeoutDuration: 1 * time.Second,
		healthCheckURL:  "http://example.invalid",
		socksHost:       "127.0.0.1",
		socksPort:       1080,
		maxRetries:      maxRetries,
		counter:         0,
	}
	hc.doCheck = hc.curlCheck
	return hc
}

func TestNewHealthCheckServiceReadsConfig(t *testing.T) {
	v := viper.New()
	v.Set("health-check.interval-duration", "30s")
	v.Set("health-check.timeout-duration", "5s")
	v.Set("health-check.url", "example.com")
	v.Set("health-check.socks.host", "127.0.0.1")
	v.Set("health-check.socks.port", 1080)
	v.Set("health-check.max-retries", 5)

	svc, err := NewHealthCheckService(v)
	if err != nil {
		t.Fatalf("NewHealthCheckService: %v", err)
	}
	hc := svc.(*healthCheckStruct)
	if hc.maxRetries != 5 {
		t.Errorf("maxRetries = %d, want 5", hc.maxRetries)
	}
	if hc.socksPort != 1080 {
		t.Errorf("socksPort = %d, want 1080", hc.socksPort)
	}
	if hc.counter != 0 {
		t.Errorf("initial counter = %d, want 0", hc.counter)
	}
}

func TestCheckConnectivitySuccessResetsCounter(t *testing.T) {
	hc := newTestHealthCheck(3)
	hc.doCheck = func(_ context.Context) bool { return false }
	hc.checkConnectivity(context.Background()) // counter → 1
	hc.checkConnectivity(context.Background()) // counter → 2

	hc.doCheck = func(_ context.Context) bool { return true }
	if got := hc.checkConnectivity(context.Background()); !got {
		t.Error("checkConnectivity() = false on success, want true")
	}
	if hc.counter != 0 {
		t.Errorf("counter = %d after success, want 0", hc.counter)
	}
}

func TestCheckConnectivityFailureIncrementsCounter(t *testing.T) {
	hc := newTestHealthCheck(3)
	hc.doCheck = func(_ context.Context) bool { return false }

	if got := hc.checkConnectivity(context.Background()); !got {
		t.Error("checkConnectivity() = false on first failure, want true (below max retries)")
	}
	if hc.counter != 1 {
		t.Errorf("counter = %d after first failure, want 1", hc.counter)
	}
}

func TestCheckConnectivityResetsCounterAtMaxRetries(t *testing.T) {
	hc := newTestHealthCheck(3)
	hc.doCheck = func(_ context.Context) bool { return false }

	for i := 0; i < hc.maxRetries-1; i++ {
		if got := hc.checkConnectivity(context.Background()); !got {
			t.Errorf("call %d: checkConnectivity() = false, want true", i+1)
		}
	}

	if got := hc.checkConnectivity(context.Background()); got {
		t.Error("checkConnectivity() = true after max retries, want false")
	}
	if hc.counter != 0 {
		t.Errorf("counter = %d after max retries, want 0 (reset)", hc.counter)
	}
}
