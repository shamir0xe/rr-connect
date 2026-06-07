package services

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func newTestConfigSwitch(t *testing.T, placeholder, templatePath, outputPath string, routers []string) *configSwitchStruct {
	t.Helper()
	v := viper.New()
	v.Set("config-switch.placeholders.router", placeholder)
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

// newTestNotify builds a config-switch service wired for notify tests, with a
// runCmd that records the args it is invoked with into *captured.
func newTestNotify(t *testing.T, extraPath string, args []string, captured *[]string) *configSwitchStruct {
	t.Helper()
	v := viper.New()
	v.Set("config-switch.placeholders.router", "__ROUTE__")
	v.Set("config-switch.placeholders.extra", "__EXTRA__")
	v.Set("config-switch.notify.enabled", true)
	v.Set("config-switch.notify.command", "true")
	v.Set("config-switch.notify.delay", "0s")
	v.Set("config-switch.notify.timeout", "5s")
	v.Set("config-switch.notify.extra-path", extraPath)
	v.Set("config-switch.notify.args", args)

	router := &routerStruct{routers: []string{"router-A"}, index: 0}
	svc, err := NewConfigSwitchService(v, router)
	if err != nil {
		t.Fatalf("NewConfigSwitchService: %v", err)
	}
	cs := svc.(*configSwitchStruct)
	cs.runCmd = func(ctx context.Context, name string, cmdArgs ...string) *exec.Cmd {
		*captured = cmdArgs
		return exec.Command("true")
	}
	return cs
}

func TestNotifyReplacesExtraFromFile(t *testing.T) {
	dir := t.TempDir()
	extraPath := filepath.Join(dir, "extra")
	// trailing whitespace must be trimmed
	if err := os.WriteFile(extraPath, []byte("EXTRA_VALUE\n"), 0644); err != nil {
		t.Fatal(err)
	}

	var captured []string
	cs := newTestNotify(t, extraPath, []string{"-d", "route=__ROUTE__ extra=__EXTRA__"}, &captured)

	if err := cs.notify(context.Background(), "router-A"); err != nil {
		t.Fatalf("notify: %v", err)
	}

	want := "route=router-A extra=EXTRA_VALUE"
	if captured[1] != want {
		t.Errorf("arg = %q, want %q", captured[1], want)
	}
}

func TestNotifyLeavesExtraPlaceholderWhenFileMissing(t *testing.T) {
	dir := t.TempDir()
	missingPath := filepath.Join(dir, "does-not-exist")

	var captured []string
	cs := newTestNotify(t, missingPath, []string{"-d", "route=__ROUTE__ extra=__EXTRA__"}, &captured)

	if err := cs.notify(context.Background(), "router-A"); err != nil {
		t.Fatalf("notify: %v", err)
	}

	// router is still substituted; extra placeholder is left untouched
	want := "route=router-A extra=__EXTRA__"
	if captured[1] != want {
		t.Errorf("arg = %q, want %q", captured[1], want)
	}
}

func TestNotifyEmptyExtraPathLeavesPlaceholder(t *testing.T) {
	var captured []string
	cs := newTestNotify(t, "", []string{"-d", "extra=__EXTRA__"}, &captured)

	if err := cs.notify(context.Background(), "router-A"); err != nil {
		t.Fatalf("notify: %v", err)
	}

	want := "extra=__EXTRA__"
	if captured[1] != want {
		t.Errorf("arg = %q, want %q", captured[1], want)
	}
}
