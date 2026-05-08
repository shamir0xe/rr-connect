# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**rr-connect** is a Go service that monitors network connectivity through a SOCKS5 proxy and automatically rotates between multiple router/proxy configurations when health checks fail. It is designed for proxy failover scenarios, typically used alongside services like Xray/V2Ray.

## Build & Run

```bash
# Build for local platform
go build -o rr-connect main.go

# Cross-compile for Linux amd64
GOOS=linux GOARCH=amd64 go build -o rr-connect-linux-amd64 main.go

# Run (requires config.yaml in the executable directory or current directory)
./rr-connect
```

There is no Makefile, test suite, or CI configuration. `go vet ./...` and `go build ./...` are the primary correctness checks.

## Configuration

Copy `config.sample.yaml` to `config.yaml` and fill in values. The config is loaded by Viper and searched first from the executable's directory, then the current directory. Environment variables override config values using `_` as the separator (e.g., `HEALTH_CHECK_MAX_RETRIES`).

Key config sections:
- `routers`: ordered list of router names used as round-robin rotation
- `health-check`: curl-based connectivity check settings (interval, timeout, SOCKS5 endpoint, max failures before switching)
- `config-switch`: template substitution, optional post-process command (e.g., `systemctl restart`), and optional Telegram notification

The `placeholder` string (e.g., `__OUTBOUND_ROUTE__`) is replaced in `template-cfg` using `sed` to produce `output-cfg`.

## Architecture

Dependency injection is handled by `go.uber.org/dig`. `main.go` registers all providers and invokes `ManagerInterface.Run`.

**Service flow:**

```
Manager
  ├── HealthCheckService  (goroutine) — polls via curl every interval
  │     └── on max-retries failures → sends true to triggerChan
  └── ConfigSwitchService (goroutine) — waits on triggerChan
        ├── makeConfig()   — sed template → output file
        ├── postProcess()  — optional: run command (e.g., systemctl restart)
        └── notify()       — optional async: curl Telegram API after delay
```

`triggerChan` is a `chan bool` created in `Manager.Run` and passed down; it uses a non-blocking send so a backlogged switch is skipped rather than queued.

**Router round-robin:** `Next()` advances the index and returns the current router name (used in `makeConfig`). `Previous()` returns the last-used name (used in `notify`). `Pick()` reads without advancing.

All services are context-aware: they exit when the context is cancelled (Ctrl+C sends `os.Interrupt` → `cancel()` → goroutines drain via `sync.WaitGroup`).

## External Dependencies

The binary shells out to:
- `curl` — health checks and Telegram notifications
- `sed` — template substitution in `makeConfig`
- Configurable post-process command (typically `systemctl`)
