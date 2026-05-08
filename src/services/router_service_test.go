package services

import (
	"testing"

	"github.com/spf13/viper"
)

func newTestRouterViper(routers []string) *viper.Viper {
	v := viper.New()
	v.Set("routers", routers)
	return v
}

func TestRouterPickDoesNotAdvance(t *testing.T) {
	r, _ := NewRouter(newTestRouterViper([]string{"a", "b", "c"}))
	r.Pick()
	if got := r.Pick(); got != "a" {
		t.Errorf("Pick() = %q, want %q", got, "a")
	}
}

func TestRouterNextAdvancesIndex(t *testing.T) {
	r, _ := NewRouter(newTestRouterViper([]string{"a", "b", "c"}))
	if got := r.Next(); got != "a" {
		t.Errorf("Next() = %q, want %q", got, "a")
	}
	if got := r.Pick(); got != "b" {
		t.Errorf("Pick() after Next() = %q, want %q", got, "b")
	}
}

func TestRouterPreviousAfterNext(t *testing.T) {
	r, _ := NewRouter(newTestRouterViper([]string{"a", "b", "c"}))
	r.Next()
	if got := r.Previous(); got != "a" {
		t.Errorf("Previous() after one Next() = %q, want %q", got, "a")
	}
}

func TestRouterRoundRobin(t *testing.T) {
	routers := []string{"a", "b", "c"}
	r, _ := NewRouter(newTestRouterViper(routers))
	for i := 0; i < len(routers)*2; i++ {
		got := r.Next()
		want := routers[i%len(routers)]
		if got != want {
			t.Errorf("Next() call %d: got %q, want %q", i+1, got, want)
		}
	}
}

func TestRouterEmptyPickReturnsEmpty(t *testing.T) {
	r, _ := NewRouter(newTestRouterViper([]string{}))
	if got := r.Pick(); got != "" {
		t.Errorf("Pick() on empty router = %q, want empty string", got)
	}
}
