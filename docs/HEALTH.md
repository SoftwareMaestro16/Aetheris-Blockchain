# Health Checks

This document outlines the health checks and monitoring endpoints for an Aetra node.

## Node Health Endpoints

### 1. Process Alive
Check if the `aetrad` process is running.
```bash
# Using systemd
systemctl status aetrad

# Using pgrep
pgrep -f aetrad
```

### 2. RPC Status
Check the node status via CometBFT RPC.
```bash
curl -s http://localhost:26657/status | jq '.result.sync_info'
```

### 3. Catching Up
Check if the node is syncing or in sync.
```bash
curl -s http://localhost:26657/status | jq '.result.sync_info.catching_up'
# Expected output: false
```

### 4. Latest Block Height
Ensure the block height is increasing.
```bash
curl -s http://localhost:26657/status | jq '.result.sync_info.latest_block_height'
```

### 5. Peer Count
Check the number of connected peers.
```bash
curl -s http://localhost:26657/net_info | jq '.result.n_peers'
```

## Healthcheck for Docker/CI
For automated checks, you can use the built-in status command:
```bash
aetrad status --node tcp://127.0.0.1:26657 --output json
```
