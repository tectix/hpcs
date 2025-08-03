# HPCS - High-Performance Cache System

A distributed cache system written in Go.

## Current Status

Basic foundation implemented:
- TCP server with connection handling
- Thread-safe cache engine
- LRU and LFU eviction policies  
- Consistent hashing implementation
- Configuration and logging systems

## Quick Start

```bash
# Build the server
make build

# Run with default configuration
./bin/hpcs-server --config configs/config.yaml

# Server listens on localhost:6379
```

## Development

```bash
# Install dependencies
make deps

# Build
make build

# Clean
make clean
```
