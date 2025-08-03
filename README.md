# HPCS - High-Performance Cache System

A distributed cache system written in Go with Redis protocol compatibility.

## Quick Start

```bash
# Build and run
make build
make run-server

# Test with redis-cli
redis-cli -h localhost -p 6379 ping
redis-cli -h localhost -p 6379 set foo bar
redis-cli -h localhost -p 6379 get foo
```

## Development

```bash
make deps     # Install dependencies
make build    # Build binary
make test     # Run tests
make clean    # Clean build files
```

## Configuration

Server settings are configured via `configs/config.yaml`. Supports environment variable overrides with `HPCS_` prefix.