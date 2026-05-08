package services

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func newTestConfigSwitch(t *testing.T, placeholder, templatePath, outputPath string, routers []string) *configSwitchStruct {
	t.Helper()
	v := viper.New()
	v.Set("config-switch.placeholder", placeholder)
	v.Set("config-switch.template-cfg", templatePath)
	v.Set("config-switch.output-cfg", outputPath)
	v.Set("config-switch.post-process.enabled", false)
	v.Set("config-switch.notify.enabled", false)

	router := &routerStruct{routers: routers, index: 0}
	svc, err := NewConfigSwitchService(v, router)
	if err != nil {
		t.Fatalf("NewConfigSwitchService: %v", err)
	}
	return svc.(*configSwitchStruct)
}

func TestMakeConfigReplacesPlaceholder(t *testing.T) {
	dir := t.TempDir()
	templatePath := filepath.Join(dir, "config.template.json")
	outputPath := filepath.Join(dir, "config.json")

	if err := os.WriteFile(templatePath, []byte(`{"outbound":"__ROUTE__"}`), 0644); err != nil {
		t.Fatal(err)
	}

	cs := newTestConfigSwitch(t, "__ROUTE__", templatePath, outputPath, []string{"router-A", "router-B"})

	routerName, err := cs.makeConfig(context.Background())
	if err != nil {
		t.Fatalf("makeConfig: %v", err)
	}
	if routerName != "router-A" {
		t.Errorf("returned router = %q, want %q", routerName, "router-A")
	}

	out, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	content := string(out)
	if strings.Contains(content, "__ROUTE__") {
		t.Errorf("output still contains placeholder: %q", content)
	}
	if !strings.Contains(content, "router-A") {
		t.Errorf("output does not contain router name: %q", content)
	}
}

func TestMakeConfigAdvancesRouter(t *testing.T) {
	dir := t.TempDir()
	templatePath := filepath.Join(dir, "config.template.json")
	outputPath := filepath.Join(dir, "config.json")

	if err := os.WriteFile(templatePath, []byte(`__ROUTE__`), 0644); err != nil {
		t.Fatal(err)
	}

	cs := newTestConfigSwitch(t, "__ROUTE__", templatePath, outputPath, []string{"first", "second"})

	first, _ := cs.makeConfig(context.Background())
	second, _ := cs.makeConfig(context.Background())

	if first != "first" {
		t.Errorf("first call returned %q, want %q", first, "first")
	}
	if second != "second" {
		t.Errorf("second call returned %q, want %q", second, "second")
	}
}

func TestMakeConfigMissingTemplateReturnsError(t *testing.T) {
	dir := t.TempDir()
	cs := newTestConfigSwitch(t, "__ROUTE__",
		filepath.Join(dir, "nonexistent.json"),
		filepath.Join(dir, "config.json"),
		[]string{"router-A"},
	)

	_, err := cs.makeConfig(context.Background())
	if err == nil {
		t.Error("expected error for missing template file, got nil")
	}
}
