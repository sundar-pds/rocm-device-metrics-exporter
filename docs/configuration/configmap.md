# Kubernetes configuration

When deploying AMD Device Metrics Exporter on Kubernetes, a `ConfigMap` is deployed in the exporter namespace.

## Configuration parameters

- `ServerPort`: this field is ignored when Device Metrics Exporter is deployed by the [GPU Operator](https://instinct.docs.amd.com/projects/gpu-operator/en/latest/) to avoid conflicts with the service node port config.
- `GPUConfig`:
  - `Fields`: An array of strings specifying what metrics field to be exported.
  - `Labels`: `SERIAL_NUMBER`, `GPU_ID`, `POD`, `NAMESPACE`, `CONTAINER`, `JOB_ID`, `JOB_USER`, `JOB_PARTITION`, `CARD_MODEL`, `HOSTNAME`, `GPU_PARTITION_ID`, `GPU_COMPUTE_PARTITION_TYPE`, and `GPU_MEMORY_PARTITION_TYPE` are always set and cannot be removed. Labels supported are available in the provided example `configmap.yml`.
  - `CustomLabels`: A map of user-defined labels and their values. Users can set up to 10 custom labels. From the `GPUMetricLabel` list, only `CLUSTER_NAME` is allowed to be set in `CustomLabels`. Any other labels from this list cannot be set. Users can define other custom labels outside of this restriction. These labels will be exported with every metric, ensuring consistent metadata across all metrics.
  - `ExtraPodLabels`: This defines a map that links Prometheus label names to Kubernetes pod labels. Each key is the Prometheus label that will be exposed in metrics, and the value is the pod label to pull the data from. This lets you expose pod metadata as Prometheus labels for easier filtering and querying.<br>(e.g. Considering an entry like `"WORKLOAD_ID"   : "amd-workload-id"`, where `WORKLOAD_ID` is a label visible in metrics and its value is the pod label value of a pod label key set as `amd-workload-id`).
  - `ProfilerMetrics`: A map of toggle to enable Profiler Metrics either for `all` nodes or a specific hostname with desired state. Key with specific hostname `$HOSTNAME` takes precedense over a `all` key. This only controls the Profiler Metrics which has prefix of GPU_PROF_ from the metrics list.
- `CommonConfig`: 
  - `MetricsFieldPrefix`: Add prefix string for all the fields exporter. [Premetheus Metric Label formatted](https://prometheus.io/docs/concepts/data_model/#metric-names-and-labels) string prefix will be accepted, on any invalid prefix will default to empty prefix to allow exporting of the fields.
  - `HealthService` : Health Service configurations for the exproter.
    - `Enable` : false to disable, otherwise enabled by default
   
## Setting custom values

To use a custom configuration when deploying the Metrics Exporter:

1. Create a `ConfigMap` based on the provided example [configmap.yml](../examples/configmap.yml) file.
2. Change the `configMap` property in `values.yaml` to `configmap.yml`
3. Run `helm install`:

```bash
helm install exporter https://github.com/ROCm/device-metrics-exporter/releases/download/v1.4.0/device-metrics-exporter-charts-v1.4.0.tgz -n metrics-exporter -f values.yaml --create-namespace
```

Device Metrics Exporter polls for configuration changes every minute, so updates take effect without container restarts.
