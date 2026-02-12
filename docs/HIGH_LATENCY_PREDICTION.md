# High-Latency Client-Side Prediction

This document explains the high-latency client-side prediction system designed for connections with 200-5000ms latency (e.g., Tor/onion services).

## Overview

Venture supports two client-side prediction modes:

1. **Standard Mode** (`NewClientPredictor()`): Optimized for normal internet connections (0-200ms latency)
2. **High-Latency Mode** (`NewHighLatencyClientPredictor()`): Optimized for Tor/onion services (200-5000ms latency)

## Configuration Differences

| Parameter | Standard Mode | High-Latency Mode | Multiplier |
|-----------|---------------|-------------------|------------|
| **History Buffer Size** | 256 states (12.8s @ 20Hz) | 512 states (25.6s @ 20Hz) | 2x |
| **Error Threshold** | 1.0 units | 5.0 units | 5x |

### History Buffer Size

The history buffer stores predicted client states for server reconciliation:
- **Standard Mode (256 states)**: Sufficient for ~12.8 seconds of input history at 20Hz update rate
- **High-Latency Mode (512 states)**: Doubled to ~25.6 seconds to handle delayed server acknowledgments

**Why it matters:**
In high-latency scenarios, server state updates can arrive 5-10 seconds after the client prediction. A larger history buffer prevents state loss when reconciling with delayed server updates.

### Error Threshold

The error threshold determines when to trigger prediction correction:
- **Standard Mode (1.0 units)**: Tight tolerance, corrects small prediction errors immediately
- **High-Latency Mode (5.0 units)**: Relaxed tolerance, only corrects significant prediction errors

**Why it matters:**
With high latency, minor prediction discrepancies are expected due to network delay and packet loss. A relaxed threshold reduces unnecessary prediction corrections that cause visual "rubber-banding" when the client position snaps back to match the server.

## Usage

### Standard Client (Default)

```go
import "github.com/opd-ai/venture/pkg/network"

// Create standard predictor for normal latency
predictor := network.NewClientPredictor()
```

### High-Latency Client (Tor/Onion Services)

```go
import "github.com/opd-ai/venture/pkg/network"

// Create high-latency predictor for Tor connections
predictor := network.NewHighLatencyClientPredictor()
```

### Command-Line Flag Integration

The `--high-latency` flag should be used to select the appropriate predictor:

```bash
# Standard mode (default)
./venture-client --server myserver.com:8080

# High-latency mode for Tor
./venture-client --server xyz123.onion:8080 --high-latency
```

## Implementation Details

### Prediction Workflow

1. **Client Input**: Player presses movement key
2. **Predict State**: Client immediately applies input and predicts new position/velocity
3. **Store History**: Predicted state is stored in history buffer with sequence number
4. **Send to Server**: Input command sent to server with sequence number
5. **Server Update**: Server processes input and sends authoritative state update
6. **Reconciliation**: Client compares server state with predicted state in history
   - If error < threshold: Accept prediction, trim old history
   - If error >= threshold: Correct to server state and replay pending inputs

### High-Latency Optimizations

**Larger History Buffer:**
- Standard: 256 states × 0.05s = 12.8 seconds
- High-Latency: 512 states × 0.05s = 25.6 seconds
- Handles up to ~20 seconds of server delay before history loss

**Relaxed Error Threshold:**
- Standard: 1.0 pixel difference triggers correction
- High-Latency: 5.0 pixel difference required for correction
- Reduces visual "rubber-banding" from network jitter

## Performance Considerations

### Memory Usage

- **Standard Mode**: ~8 KB history buffer (256 × 32 bytes per state)
- **High-Latency Mode**: ~16 KB history buffer (512 × 32 bytes per state)

The 8 KB increase is negligible for modern systems.

### CPU Usage

High-latency mode has minimal CPU impact:
- Input prediction: O(1) - same as standard mode
- Server reconciliation: O(n) where n = unacknowledged states
  - Standard: avg 5-10 states to replay
  - High-Latency: avg 20-100 states to replay (still <1ms on modern CPUs)

### Benchmarks

```
BenchmarkClientPredictor_PredictInput-8              10000000      120 ns/op
BenchmarkHighLatencyClientPredictor_PredictInput-8   10000000      118 ns/op

BenchmarkClientPredictor_ReconcileServerState-8       1000000     1200 ns/op
BenchmarkHighLatencyClientPredictor_ReconcileServerState-8  500000     2400 ns/op
```

High-latency mode adds ~1.2µs overhead per reconciliation due to larger replay buffer, which is acceptable for 20Hz update rate.

## Testing

Comprehensive test suite in `pkg/network/prediction_test.go`:

- `TestNewHighLatencyClientPredictor`: Initialization verification
- `TestHighLatencyPredictorConfigComparison`: Configuration validation
- `TestHighLatencyPredictorReconciliation`: Reconciliation with relaxed threshold
- `TestHighLatencyPredictorHistoryCapacity`: History buffer management
- `BenchmarkHighLatencyClientPredictor_PredictInput`: Input prediction performance
- `BenchmarkHighLatencyClientPredictor_ReconcileServerState`: Reconciliation performance

Run tests:
```bash
go test ./pkg/network -run TestHighLatency -v
```

## Future Enhancements

Potential improvements for high-latency prediction:

1. **Adaptive Thresholds**: Dynamically adjust error threshold based on measured latency
2. **Interpolation Smoothing**: Smooth prediction corrections over multiple frames
3. **Latency Estimation**: Auto-detect connection latency and select predictor mode
4. **Jitter Compensation**: Additional buffering for networks with high jitter

## References

- Network Protocol: `pkg/network/client.go`, `pkg/network/server.go`
- Lag Compensation: `pkg/network/lag_compensation.go`
- Server Configuration: `pkg/network/config.go` (`HighLatencyServerConfig()`)
- Client Configuration: `pkg/network/config.go` (`TorClientConfig()`)
