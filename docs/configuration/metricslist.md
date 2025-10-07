# List of Available Metrics

The following table contains a full list of GPU Metrics that are available using the Device Metrics Exporter. Descriptions of each metric will be included at a later time.

## Cluster Management Metrics

| Hypervisor | Baremetal | Metric                                                | Description                                                                                                                                   |
|------------|-----------|-------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------|
| &check;    | &check;   | GPU_NODES_TOTAL                                       | Number of GPU nodes on the machine                                                                                                            |
| &check;    | &check;   | GPU_PACKAGE_POWER                                     | Current socket power in Watts; not available on guest VM                                                                                      |
| &cross;    | &check;   | GPU_AVERAGE_PACKAGE_POWER                             | Average socket power in Watts; not available on guest VM                                                                                      |
| &check;    | &check;   | GPU_EDGE_TEMPERATURE                                  | Edge temperature value in Celsius  (MI2XX platforms only)                                                                                     |
| &check;    | &check;   | GPU_JUNCTION_TEMPERATURE                              | Hotspot (aka junction) temperature value in Celsius                                                                                           |
| &check;    | &check;   | GPU_MEMORY_TEMPERATURE                                | Memory temperature value in Celsius                                                                                                           |
| &cross;    | &check;   | GPU_HBM_TEMPERATURE                                   | List of hbm temperatures in Celsius (Depricated from 6.14.14 driver)                                                                          |
| &check;    | &check;   | GPU_GFX_ACTIVITY   (Applicable for unpartitioned GPU) | Graphics engine usage percentage (0 - 100)                                                                                                    |
| &check;    | &check;   | GPU_UMC_ACTIVITY                                      | Memory engine usage percentage (0 - 100)                                                                                                      |
| &cross;    | &check;   | GPU_MMA_ACTIVITY                                      | Average multimedia engine usages in percentage (0 - 100)                                                                                      |
| &check;    | &check;   | GPU_VCN_ACTIVITY                                      | List of VCN encode/decode engine utilization per AID                                                                                          |
| &check;    | &check;   | GPU_JPEG_ACTIVITY                                     | List of JPEG engine activity in percentage (0 - 100)                                                                                          |
| &cross;    | &check;   | GPU_VOLTAGE                                           | SoC voltage in mV                                                                                                                             |
| &cross;    | &check;   | GPU_GFX_VOLTAGE                                       | gfx voltage in mV                                                                                                                             |
| &cross;    | &check;   | GPU_MEMORY_VOLTAGE                                    | Mem voltage in mV                                                                                                                             |
| &check;    | &check;   | PCIE_SPEED                                            | Current pcie speed capable in GT/s                                                                                                            |
| &check;    | &check;   | PCIE_MAX_SPEED                                        | Maximum capable pcie speed in GT/s                                                                                                            |
| &check;    | &check;   | PCIE_BANDWIDTH                                        | Current instantaneous bandwidth usage in Mb/s                                                                                                 |
| &check;    | &check;   | GPU_ENERGY_CONSUMED                                   | Accumulated energy consumed by the GPU in Micro Jules (uJ)                                                                                                    |
| &check;    | &check;   | PCIE_REPLAY_COUNT                                     | Total number of PCIe replays (NAKs)                                                                                                           |
| &check;    | &check;   | PCIE_RECOVERY_COUNT                                   | Total number of PCIe replays (NAKs)                                                                                                           |
| &check;    | &check;   | PCIE_REPLAY_ROLLOVER_COUNT                            | PCIe Replay accumulated count                                                                                                                 |
| &check;    | &check;   | PCIE_NACK_SENT_COUNT                                  | PCIe NACK sent accumulated count                                                                                                               |
| &check;    | &check;   | PCIE_NACK_RECEIVED_COUNT                              | PCIe NACK received accumulated count                                                                                                           |
| &cross;    | &cross;   | PCIE_RX                                               | Accumulated bytes received from the PCIe link                                                                                                 |
| &cross;    | &cross;   | PCIE_TX                                               | Accumulated bytes transmitted to the PCIe link                                                                                                |
| &cross;    | &check;   | PCIE_BIDIRECTIONAL_BANDWIDTH                          | Accumulated bandwidth on PCIe link in GB/sec                                                                                                  |
| &check;    | &check;   | GPU_CLOCK                                             | Clock measure of the GPU in Mhz* ([See note below](#gpu_clock-measurements))                                                                  |
| &check;    | &check;   | GPU_POWER_USAGE                                       | GPU power usage in Watts                                                                                                                      |
| &check;    | &check;   | GPU_TOTAL_VRAM                                        | Total VRAM available in MB                                                                                                                    |
| &check;    | &check;   | GPU_ECC_CORRECT_TOTAL                                 | Total Correctable ECC error count                                                                                                             |
| &check;    | &check;   | GPU_ECC_UNCORRECT_TOTAL                               | Total Uncorrectable ECC error count                                                                                                           |
| &check;    | &check;   | GPU_ECC_CORRECT_SDMA                                  | Correctable ECC error in SDMA                                                                                                                 |
| &check;    | &check;   | GPU_ECC_UNCORRECT_SDMA                                | Uncorrectable ECC error in SDMA                                                                                                               |
| &check;    | &check;   | GPU_ECC_CORRECT_GFX                                   | Correctable ECC error in GFX                                                                                                                  |
| &check;    | &check;   | GPU_ECC_UNCORRECT_GFX                                 | Uncorrectable ECC error in GFX                                                                                                                |
| &check;    | &check;   | GPU_ECC_CORRECT_MMHUB                                 | Correctable ECC error in MMHUB                                                                                                                |
| &check;    | &check;   | GPU_ECC_UNCORRECT_MMHUB                               | Uncorrectable ECC error in MMHUB                                                                                                              |
| &cross;    | &check;   | GPU_ECC_CORRECT_ATHUB                                 | Correctable ECC error in ATHUB                                                                                                                |
| &cross;    | &check;   | GPU_ECC_UNCORRECT_ATHUB                               | Uncorrectable ECC error in ATHUB                                                                                                              |
| &check;    | &check;   | GPU_ECC_CORRECT_BIF                                   | Correctable ECC error in BIF                                                                                                                  |
| &check;    | &check;   | GPU_ECC_UNCORRECT_BIF                                 | Uncorrectable ECC error in BIF                                                                                                                |
| &cross;    | &check;   | GPU_ECC_CORRECT_HDP                                   | Correctable ECC error in HDP                                                                                                                  |
| &cross;    | &check;   | GPU_ECC_UNCORRECT_HDP                                 | Uncorrectable ECC error in HDP                                                                                                                |
| &check;    | &check;   | GPU_ECC_CORRECT_XGMI_WAFL                             | Correctable ECC error in XGMI WAFL                                                                                                            |
| &check;    | &check;   | GPU_ECC_UNCORRECT_XGMI_WAFL                           | Uncorrectable ECC error in XGMI WAFL                                                                                                          |
| &cross;    | &check;   | GPU_ECC_CORRECT_DF                                    | Correctable ECC error in DF                                                                                                                   |
| &cross;    | &check;   | GPU_ECC_UNCORRECT_DF                                  | Uncorrectable ECC error in DF                                                                                                                 |
| &cross;    | &check;   | GPU_ECC_CORRECT_SMN                                   | Correctable ECC error in SMN                                                                                                                  |
| &cross;    | &check;   | GPU_ECC_UNCORRECT_SMN                                 | Uncorrectable ECC error in SMN                                                                                                                |
| &cross;    | &check;   | GPU_ECC_CORRECT_SEM                                   | Correctable ECC error in SEM                                                                                                                  |
| &cross;    | &check;   | GPU_ECC_UNCORRECT_SEM                                 | Uncorrectable ECC error in SEM                                                                                                                |
| &cross;    | &check;   | GPU_ECC_CORRECT_MP0                                   | Correctable ECC error in MP0                                                                                                                  |
| &cross;    | &check;   | GPU_ECC_UNCORRECT_MP0                                 | Uncorrectable ECC error in MP0                                                                                                                |
| &cross;    | &check;   | GPU_ECC_CORRECT_MP1                                   | Correctable ECC error in MP1                                                                                                                  |
| &cross;    | &check;   | GPU_ECC_UNCORRECT_MP1                                 | Uncorrectable ECC error in MP1                                                                                                                |
| &cross;    | &check;   | GPU_ECC_CORRECT_FUSE                                  | Correctable ECC error in FUSE                                                                                                                 |
| &cross;    | &check;   | GPU_ECC_UNCORRECT_FUSE                                | Uncorrectable ECC error in FUSE                                                                                                               |
| &check;    | &check;   | GPU_ECC_CORRECT_UMC                                   | Correctable ECC error in UMC                                                                                                                  |
| &check;    | &check;   | GPU_ECC_UNCORRECT_UMC                                 | Uncorrectable ECC error in UMC                                                                                                                |
| &cross;    | &check;   | GPU_XGMI_NBR_0_NOP_TX                                 | NOPs sent to neighbor 0                                                                                                                       |
| &cross;    | &check;   | GPU_XGMI_NBR_0_REQ_TX                                 | Outgoing requests to neighbor 0                                                                                                               |
| &cross;    | &check;   | GPU_XGMI_NBR_0_RESP_TX                                | Outgoing responses to neighbor 0                                                                                                              |
| &cross;    | &check;   | GPU_XGMI_NBR_0_BEATS_TX                               | Data beats sent to neighbor 0; Each beat represents 32 bytes                                                                                  |
| &cross;    | &check;   | GPU_XGMI_NBR_1_NOP_TX                                 | NOPs sent to neighbor 1                                                                                                                       |
| &cross;    | &check;   | GPU_XGMI_NBR_1_REQ_TX                                 | Outgoing requests to neighbor 1                                                                                                               |
| &cross;    | &check;   | GPU_XGMI_NBR_1_RESP_TX                                | Outgoing responses to neighbor 1                                                                                                              |
| &cross;    | &check;   | GPU_XGMI_NBR_1_BEATS_TX                               | Data beats sent to neighbor 1; Each beat represents 32 bytes                                                                                  |
| &cross;    | &check;   | GPU_XGMI_NBR_0_TX_THRPUT                              | Represents the number of outbound beats (each representing 32 bytes) on link 0; Throughput = BEATS/time_running * 10^9  bytes/sec             |
| &cross;    | &check;   | GPU_XGMI_NBR_1_TX_THRPUT                              | Represents the number of outbound beats (each representing 32 bytes) on link 1                                                                |
| &cross;    | &check;   | GPU_XGMI_NBR_2_TX_THRPUT                              | Represents the number of outbound beats (each representing 32 bytes) on link 2                                                                |
| &cross;    | &check;   | GPU_XGMI_NBR_3_TX_THRPUT                              | Represents the number of outbound beats (each representing 32 bytes) on link 3                                                                |
| &cross;    | &check;   | GPU_XGMI_NBR_4_TX_THRPUT                              | Represents the number of outbound beats (each representing 32 bytes) on link 4                                                                |
| &cross;    | &check;   | GPU_XGMI_NBR_5_TX_THRPUT                              | Represents the number of outbound beats (each representing 32 bytes) on link 5                                                                |
| &cross;    | &check;   | GPU_XGMI_LINK_RX                                      | Accumulated XGMI Link Data Read in KB**                                                                                                                   |
| &cross;    | &check;   | GPU_XGMI_LINK_TX                                      | Accumulated XGMI Link Data Write in KB**                                                                                                                  |
| &cross;    | &check;   | GPU_USED_VRAM                                         | Total VRAM memory used in MB                                                                                                                  |
| &cross;    | &check;   | GPU_FREE_VRAM                                         | Total VRAM memory free in MB                                                                                                                  |
| &cross;    | &check;   | GPU_TOTAL_VISIBLE_VRAM                                | Total available visible VRAM memory in MB                                                                                                     |
| &cross;    | &check;   | GPU_USED_VISIBLE_VRAM                                 | Total used VRAM memory in MB                                                                                                                  |
| &cross;    | &check;   | GPU_FREE_VISIBLE_VRAM                                 | Total free VRAM memory in MB                                                                                                                  |
| &cross;    | &check;   | GPU_TOTAL_GTT                                         | Total GTT memory in MB                                                                                                                        |
| &cross;    | &check;   | GPU_USED_GTT                                          | Current GTT memory usage in MB                                                                                                                |
| &cross;    | &check;   | GPU_FREE_GTT                                          | Free GTT memory available in MB                                                                                                               |
| &cross;    | &check;   | GPU_ECC_CORRECT_MCA                                   | Correctable ECC error in MCA                                                                                                                  |
| &cross;    | &check;   | GPU_ECC_UNCORRECT_MCA                                 | Uncorrectable ECC error in MCA                                                                                                                |
| &cross;    | &check;   | GPU_ECC_CORRECT_VCN                                   | Correctable ECC error in VCN                                                                                                                  |
| &cross;    | &check;   | GPU_ECC_UNCORRECT_VCN                                 | Uncorrectable ECC error in VCN                                                                                                                |
| &cross;    | &check;   | GPU_ECC_CORRECT_JPEG                                  | Correctable ECC error in JPEG                                                                                                                 |
| &cross;    | &check;   | GPU_ECC_UNCORRECT_JPEG                                | Uncorrectable ECC error in JPEG                                                                                                               |
| &cross;    | &check;   | GPU_ECC_CORRECT_IH                                    | Correctable ECC error in IH                                                                                                                   |
| &cross;    | &check;   | GPU_ECC_UNCORRECT_IH                                  | Uncorrectable ECC error in IH                                                                                                                 |
| &cross;    | &check;   | GPU_ECC_CORRECT_MPIO                                  | Correctable ECC error in MPIO                                                                                                                 |
| &cross;    | &check;   | GPU_ECC_UNCORRECT_MPIO                                | Uncorrectable ECC error in MPIO                                                                                                               |
| &cross;    | &check;   | GPU_CURRENT_ACCUMULATED_COUNTER                       | Current Accumulated Violation Counter                                                                                                         |
| &cross;    | &check;   | GPU_VIOLATION_PROCESSOR_HOT_RESIDENCY_ACCUMULATED     | Process Hot Residency Accumulated Violation Counter                                                                                           |
| &cross;    | &check;   | GPU_VIOLATION_PPT_RESIDENCY_ACCUMULATED               | Package Power Tracking Accumulated Violation Counter                                                                                          |
| &cross;    | &check;   | GPU_VIOLATION_SOCKET_THERMAL_RESIDENCY_ACCUMULATED    | Socket Thermal accumulated Violation Counter                                                                                                  |
| &cross;    | &check;   | GPU_VIOLATION_VR_THERMAL_RESIDENCY_ACCUMULATED        | Voltage Rail accumulated Violation Counter                                                                                                    |
| &cross;    | &check;   | GPU_VIOLATION_HBM_THERMAL_RESIDENCY_ACCUMULATED       | HBM Accumulated Violation Counter                                                                                                             |
| &cross;    | &check;   | GPU_GFX_BUSY_INSTANTANEOUS                            | GFX Busy Instantaneous Activity Per Accelerator Compute Processor Per Compute Core                                                            |
| &cross;    | &check;   | GPU_VC_BUSY_INSTANTANEOUS                             | VCN Busy Instantaneous Activity Per Accelerator Compute Processor Per Compute Core                                                            |
| &cross;    | &check;   | GPU_JPEG_BUSY_INSTANTANEOUS                           | JPEG Busy Instantaneous Activity Per Accelerator Compute Processor Per Compute Core                                                           |

## Performance Metrics Focused For Application Development

The Device Metrics Exporter now supports a whole list of Performance metrics
to better facilitate the developers to understand more about the application
performance on the GPUs

The `ProfilerMetrics` should be turned off in case of application level
profiling is run as the current hardware limits a single profiler instance to
be run at any given time.

By default the performance metrics are disabled and they can be  enabled through the
ConfigMap `ProfilerMetrics` to enable or disable per host (higer precedence) or `all` key to specify for cluster wide toggle.

The list comprises of all well known metrics supported by MI 200 & 300 platforms. 
Some fields which are not supported by platforms though enabled would not be exported. 
The full list of supported Fields and Registers are available at [Performance Counters](https://rocm.docs.amd.com/en/latest/conceptual/gpu-arch/mi300-mi200-performance-counters.html). 

**_Note_** : Performance metrics are not applicable for hypervisor based deployments

| Metric                                      | Description                                                                                                                                   |
|---------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------|
| GPU_PROF_GRBM_GUI_ACTIVE                    | Number of GPU active cycles                                                                                                                   |
| GPU_PROF_SQ_WAVES                           | Number of wavefronts dispatched to sequencers, including both new and restored wavefronts                                                     |
| GPU_PROF_GRBM_COUNT                         | Number of free-running GPU cycles                                                                                                             |
| GPU_PROF_CPC_CPC_STAT_BUSY                  | Number of cycles command processor-compute is busy                                                                                            |
| GPU_PROF_CPC_CPC_STAT_IDLE                  | Number of cycles command processor-compute is idle                                                                                            |
| GPU_PROF_CPC_CPC_STAT_STALL                 | Number of cycles command processor-compute is stalled                                                                                         |
| GPU_PROF_CPC_CPC_TCIU_BUSY                  | Number of cycles command processor-compute texture cache interface unit interface is busy                                                     |
| GPU_PROF_CPC_CPC_TCIU_IDLE                  | Number of cycles command processor-compute texture cache interface unit interface is idle                                                     |
| GPU_PROF_CPC_CPC_UTCL2IU_BUSY               | Number of cycles command processor-compute unified translation cache (L2) interface is busy                                                   |
| GPU_PROF_CPC_CPC_UTCL2IU_IDLE               | Number of cycles command processor-compute unified translation cache (L2) interface is idle                                                   |
| GPU_PROF_CPC_CPC_UTCL2IU_STALL              | Number of cycles command processor-compute unified translation cache (L2) interface is stalled                                                |
| GPU_PROF_CPC_ME1_BUSY_FOR_PACKET_DECODE     | Number of cycles command processor-compute micro engine is busy decoding packets                                                              |
| GPU_PROF_CPC_ME1_DC0_SPI_BUSY               | Number of cycles command processor-compute micro engine processor is busy                                                                     |
| GPU_PROF_CPC_UTCL1_STALL_ON_TRANSLATION     | Number of cycles one of the unified translation caches (L1) is stalled waiting on translation                                                 |
| GPU_PROF_CPC_ALWAYS_COUNT                   | CPC Always Count                                                                                                                              |
| GPU_PROF_CPC_ADC_VALID_CHUNK_NOT_AVAIL      | CPC ADC valid chunk not available when dispatch walking is in progress at multi-xcc mode                                                      |
| GPU_PROF_CPC_ADC_DISPATCH_ALLOC_DONE        | CPC ADC dispatch allocation done                                                                                                              |
| GPU_PROF_CPC_ADC_VALID_CHUNK_END            | CPC ADC cralwer valid chunk end at multi-xcc mode                                                                                             |
| GPU_PROF_CPC_SYNC_FIFO_FULL_LEVEL           | CPC SYNC FIFO full last cycles                                                                                                                |
| GPU_PROF_CPC_SYNC_FIFO_FULL                 | CPC SYNC FIFO full times                                                                                                                      |
| GPU_PROF_CPC_GD_BUSY                        | CPC ADC busy                                                                                                                                  |
| GPU_PROF_CPC_TG_SEND                        | CPC ADC thread group send                                                                                                                     |
| GPU_PROF_CPC_WALK_NEXT_CHUNK                | CPC ADC walking next valid chunk at multi-xcc mode                                                                                            |
| GPU_PROF_CPC_STALLED_BY_SE0_SPI             | CPC ADC csdata stalled by SE0SPI                                                                                                              |
| GPU_PROF_CPC_STALLED_BY_SE1_SPI             | CPC ADC csdata stalled by SE1SPI                                                                                                              |
| GPU_PROF_CPC_STALLED_BY_SE2_SPI             | CPC ADC csdata stalled by SE2SPI                                                                                                              |
| GPU_PROF_CPC_STALLED_BY_SE3_SPI             | CPC ADC csdata stalled by SE3SPI                                                                                                              |
| GPU_PROF_CPC_LTE_ALL                        | CPC Sync counter LteAll, only Master XCD cares LteAll                                                                                         |
| GPU_PROF_CPC_SYNC_WRREQ_FIFO_BUSY           | CPC Sync Counter Request Fifo is not empty                                                                                                    |
| GPU_PROF_CPC_CANE_BUSY                      | CPC CANE bus busy, means there are inflight sync counter requests                                                                             |
| GPU_PROF_CPC_CANE_STALL                     | CPC Sync counter sending is stalled by CANE                                                                                                   |
| GPU_PROF_CPF_CMP_UTCL1_STALL_ON_TRANSLATION | One of the Compute UTCL1s is stalled waiting on translation, XNACK or PENDING respons                                                         |
| GPU_PROF_CPF_CPF_STAT_BUSY                  | CPF Busy                                                                                                                                      |
| GPU_PROF_CPF_CPF_STAT_IDLE                  | CPF Idle                                                                                                                                      |
| GPU_PROF_CPF_CPF_STAT_STALL                 | CPF Stalled                                                                                                                                   |
| GPU_PROF_CPF_CPF_TCIU_BUSY                  | CPF TCIU interface Busy                                                                                                                       |
| GPU_PROF_CPF_CPF_TCIU_IDLE                  | CPF TCIU interface Idle                                                                                                                       |
| GPU_PROF_CPF_CPF_TCIU_STALL                 | CPF TCIU interface Stalled waiting on Free, Tags                                                                                              |
| GPU_PROF_FETCH_SIZE                         | The total kilobytes fetched from the video memory. This is measured with all extra fetches and any cache or memory effects taken into account |
| GPU_PROF_WRITE_SIZE                         | The total kilobytes written to the video memory. This is measured with all extra fetches and any cache or memory effects taken into account   |
| GPU_PROF_TOTAL_16_OPS                       | The number of 16 bits OPS executed                                                                                                            |
| GPU_PROF_TOTAL_32_OPS                       | The number of 32 bits OPS executed                                                                                                            |
| GPU_PROF_TOTAL_64_OPS                       | The number of 64 bits OPS executed                                                                                                            |
| GPU_PROF_GUI_UTIL_PERCENT                   | Percentage of the time that GUI is active                                                                                                     |
| GPU_PROF_OCCUPANCY_PERCENT                  | GPU Occupancy as Percentage of maximum                                                                                                        |
| GPU_PROF_TENSOR_ACTIVE_PERCENT              | MFMA Utililization Unit percent                                                                                                               |
| GPU_PROF_VALU_PIPE_ISSUE_UTIL               | Percentage of the time that GUI is active                                                                                                     |
| GPU_PROF_SM_ACTIVE                          | The percentage of GPUTime vector ALU instructions are processed. Value range: 0% (bad) to 100% (optimal)                                      |
| GPU_PROF_OCCUPANCY_ELAPSED                  | Number of GPU active cycles                                                                                                                   |
| GPU_PROF_OCCUPANCY_PER_ACTIVE_CU            | Mean occupancy per active compute unit                                                                                                        |
| GPU_PROF_OCCUPANCY_PER_CU                   | Mean occupancy per compute unit                                                                                                               |
| GPU_PROF_SIMD_UTILIZATION                   | Fraction of time the SIMDs are being utilized [0,1]                                                                                           |

## Array-Based Field Descriptions

The following sections describe metrics that return arrays of values, where each element is distinguished by specific label indices to represent different components or links within the GPU hardware.

### GPU_CLOCK measurements

The Device Metrics Exporter `gpu_clock` metric is a common field used for exporting different types of clocks. This metric has a `clock_type` label added to the metric to differentiate the different clock types:

```json
gpu_clock{clock_type="GPU_CLOCK_TYPE_DATA"}
gpu_clock{clock_type="GPU_CLOCK_TYPE_SYSTEM"}
gpu_clock{clock_type="GPU_CLOCK_TYPE_MEMORY"}
gpu_clock{clock_type="GPU_CLOCK_TYPE_VIDEO"}
```

An example of this is shown below:

```json
gpu_clock{card_model="xxxx",clock_index="14",clock_type="GPU_CLOCK_TYPE_DATA",gpu_compute_partition_type="spx",gpu_id="3",gpu_partition_id="0",hostname="xxxx",serial_number="xxxx"} 22
gpu_clock{card_model="xxxx",clock_index="2",clock_type="GPU_CLOCK_TYPE_SYSTEM",gpu_compute_partition_type="spx",gpu_id="0",gpu_partition_id="0",hostname="xxxx",serial_number="xxxx"} 132
gpu_clock{card_model="xxxx",clock_index="8",clock_type="GPU_CLOCK_TYPE_MEMORY",gpu_compute_partition_type="spx",gpu_id="0",gpu_partition_id="0",hostname="xxxx",serial_number="xxxx"} 900
gpu_clock{card_model="xxxx",clock_index="9",clock_type="GPU_CLOCK_TYPE_VIDEO",gpu_compute_partition_type="spx",gpu_id="5",gpu_partition_id="0",hostname="xxxx",serial_number="xxxx"} 29
```

### XGMI Link Read and Write measurements

The Device Metrics Exporter `gpu_xgmi_link_rx` and `gpu_xgmi_link_tx` metrics consist of an array field used for exporting the transfer metrics for each xgmi link connected to a GPU. These metric have a `link_index` label added to the metric to differentiate the different links (usually 8 in an MI300X system):

```json
gpu_xgmi_link_rx{link_index="0"}
gpu_xgmi_link_rx{link_index="1"}
gpu_xgmi_link_rx{link_index="2"}
gpu_xgmi_link_rx{link_index="3"}

gpu_xgmi_link_tx{link_index="0"}
gpu_xgmi_link_tx{link_index="1"}
gpu_xgmi_link_tx{link_index="2"}
gpu_xgmi_link_tx{link_index="3"}
```

An example of this is shown below:

```json
gpu_xgmi_link_rx{card_model="xxxx",gpu_compute_partition_type="spx",gpu_id="0",gpu_partition_id="0",hostname="xxxx",link_index="0",serial_number="xxxx"} 0
gpu_xgmi_link_rx{card_model="xxxx",gpu_compute_partition_type="spx",gpu_id="0",gpu_partition_id="0",hostname="xxxx",link_index="1",serial_number="xxxx"} 2.776148269e+09
gpu_xgmi_link_rx{card_model="xxxx",gpu_compute_partition_type="spx",gpu_id="0",gpu_partition_id="0",hostname="xxxx",link_index="2",serial_number="xxxx"} 2.914491813e+09
gpu_xgmi_link_rx{card_model="xxxx",gpu_compute_partition_type="spx",gpu_id="0",gpu_partition_id="0",hostname="xxxx",link_index="3",serial_number="xxxx"} 2.853215723e+09
gpu_xgmi_link_rx{card_model="xxxx",gpu_compute_partition_type="spx",gpu_id="0",gpu_partition_id="0",hostname="xxxx",link_index="4",serial_number="xxxx"} 2.857943554e+09
gpu_xgmi_link_rx{card_model="xxxx",gpu_compute_partition_type="spx",gpu_id="0",gpu_partition_id="0",hostname="xxxx",link_index="5",serial_number="xxxx"} 2.859773597e+09
gpu_xgmi_link_rx{card_model="xxxx",gpu_compute_partition_type="spx",gpu_id="0",gpu_partition_id="0",hostname="xxxx",link_index="6",serial_number="xxxx"} 2.852296682e+09
gpu_xgmi_link_rx{card_model="xxxx",gpu_compute_partition_type="spx",gpu_id="0",gpu_partition_id="0",hostname="xxxx",link_index="7",serial_number="xxxx"} 2.757052542e+09

gpu_xgmi_link_tx{card_model="xxxx",gpu_compute_partition_type="spx",gpu_id="0",gpu_partition_id="0",hostname="xxxx",link_index="0",serial_number="xxxx"} 0
gpu_xgmi_link_tx{card_model="xxxx",gpu_compute_partition_type="spx",gpu_id="0",gpu_partition_id="0",hostname="xxxx",link_index="1",serial_number="xxxx"} 3.539423344e+09
gpu_xgmi_link_tx{card_model="xxxx",gpu_compute_partition_type="spx",gpu_id="0",gpu_partition_id="0",hostname="xxxx",link_index="2",serial_number="xxxx"} 3.708415141e+09
gpu_xgmi_link_tx{card_model="xxxx",gpu_compute_partition_type="spx",gpu_id="0",gpu_partition_id="0",hostname="xxxx",link_index="3",serial_number="xxxx"} 3.639806555e+09
gpu_xgmi_link_tx{card_model="xxxx",gpu_compute_partition_type="spx",gpu_id="0",gpu_partition_id="0",hostname="xxxx",link_index="4",serial_number="xxxx"} 3.675549728e+09
gpu_xgmi_link_tx{card_model="xxxx",gpu_compute_partition_type="spx",gpu_id="0",gpu_partition_id="0",hostname="xxxx",link_index="5",serial_number="xxxx"} 3.657430314e+09
gpu_xgmi_link_tx{card_model="xxxx",gpu_compute_partition_type="spx",gpu_id="0",gpu_partition_id="0",hostname="xxxx",link_index="6",serial_number="xxxx"} 3.646094607e+09
gpu_xgmi_link_tx{card_model="xxxx",gpu_compute_partition_type="spx",gpu_id="0",gpu_partition_id="0",hostname="xxxx",link_index="7",serial_number="xxxx"} 3.545990503e+0
```

