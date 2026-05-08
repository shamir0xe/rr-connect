# rr-connect

Watches network connectivity through a SOCKS5 proxy and automatically rotates to the next router when the connection fails — designed for use alongside proxy services like Xray/V2Ray.

## How it works

A health-check loop runs `curl` through the configured SOCKS5 proxy at a fixed interval. After `max-retries` consecutive failures it triggers a config switch: the placeholder string in your template config is replaced with the next router name (round-robin), the output file is written, and an optional post-process command (e.g. `systemctl restart`) is executed to apply it. An optional Telegram notification fires asynchronously after the switch.

## Build

```bash
make build          # current platform
make build-linux    # linux/amd64
```

Requires Go 1.25+. The binary looks for `config.yaml` next to itself, then the current directory.

## Configuration

```bash
cp config.sample.yaml config.yaml
```

| Section | Key | Description |
|---|---|---|
| `routers` | — | Ordered list of router names for round-robin rotation |
| `health-check` | `interval-duration` | How often to check (e.g. `25s`) |
| | `timeout-duration` | curl timeout per check |
| | `url` | URL to reach through the proxy |
| | `max-retries` | Consecutive failures before switching |
| | `socks.host/port` | SOCKS5 proxy address |
| `config-switch` | `placeholder` | String to replace in the template (e.g. `__OUTBOUND_ROUTE__`) |
| | `template-cfg` | Input template file |
| | `output-cfg` | Generated output file |
| | `post-process` | Optional command to run after the switch (e.g. `systemctl restart`) |
| | `notify` | Optional curl command for Telegram notification |

The `placeholder` string is also substituted inside `notify.args`, so you can embed the new router name in the notification message.

Environment variables override config values — dots and dashes become underscores (e.g. `HEALTH_CHECK_MAX_RETRIES=5`).

## Run

```bash
./rr-connect
```

Stop with `Ctrl+C` — the process shuts down gracefully.

## Development

```bash
make vet     # go vet
make test    # go test ./...
```
