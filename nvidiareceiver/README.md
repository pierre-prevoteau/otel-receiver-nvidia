# NVIDIA GPU Receiver

The NVIDIA GPU receiver collects GPU utilization and memory metrics for the NVIDIA GPUs installed on the local host.

| Status                |                            |
| --------------------- | -------------------------- |
| Stability             | [alpha]: metrics           |
| Unsupported Platforms | darwin                     |
| Distributions         | none (build with [ocb])    |
| Issues                | [Open issues](https://github.com/pierre-prevoteau/otel-receiver-nvidia/issues) |
| Code Owners           | [@pierre-prevoteau](https://www.github.com/pierre-prevoteau) |

[alpha]: https://github.com/open-telemetry/opentelemetry-collector/blob/main/docs/component-stability.md#alpha
[ocb]: https://opentelemetry.io/docs/collector/custom-collector/

On every collection interval the receiver runs the [`nvidia-smi`](https://developer.nvidia.com/nvidia-system-management-interface) command line tool (shipped with the NVIDIA driver) and converts its output into OpenTelemetry metrics. One set of metrics is emitted per GPU, and each GPU is identified by resource attributes (`nvidia.gpu.index`, `nvidia.gpu.uuid`, `nvidia.gpu.name`).

For example, the following metrics are produced for a GPU at index `0`:

```
nvidia.gpu.utilization{nvidia.gpu.index="0", nvidia.gpu.name="NVIDIA A100-SXM4-40GB", nvidia.gpu.uuid="GPU-..."}        = 42
nvidia.gpu.memory.utilization{nvidia.gpu.index="0", ...}                                                               = 20
nvidia.gpu.memory.total{nvidia.gpu.index="0", ...}                                                                     = 42949672960
nvidia.gpu.memory.used{nvidia.gpu.index="0", ...}                                                                      = 8589934592
nvidia.gpu.memory.free{nvidia.gpu.index="0", ...}                                                                      = 34359738368
```

GPU memory is reported by `nvidia-smi` in MiB and converted to bytes (`MiB * 1024 * 1024`). `nvidia.gpu.memory.utilization` is derived as `memory.used / memory.total * 100`.

`nvidia-smi` must be installed and resolvable on the collector's `PATH` (or referenced explicitly through `binary_path`). The receiver targets Linux hosts with NVIDIA GPUs and is not supported on macOS.

## Configuration

| Field                 | Default      | Description                                                                       |
| --------------------- | ------------ | -------------------------------------------------------------------------------- |
| `collection_interval` | `60s`        | How often `nvidia-smi` is invoked and metrics are produced.                      |
| `initial_delay`       | `1s`         | Time to wait before the first scrape.                                            |
| `binary_path`         | `nvidia-smi` | Path to the `nvidia-smi` executable. A bare file name is resolved using `PATH`.  |

Individual metrics and resource attributes can be toggled under the `metrics` and `resource_attributes` keys. See [documentation.md](./documentation.md) for the full list of emitted metrics and resource attributes.

## Example

### Basic configuration

In its default configuration the receiver scrapes every NVIDIA GPU on the host once per minute:

```yaml
receivers:
  nvidia:
```

### Advanced configuration

```yaml
receivers:
  nvidia:
    collection_interval: 30s
    binary_path: /opt/nvidia/bin/nvidia-smi
    metrics:
      nvidia.gpu.memory.free:
        enabled: false
```
