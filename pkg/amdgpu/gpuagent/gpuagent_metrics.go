/**
# Copyright (c) Advanced Micro Devices, Inc. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the \"License\");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an \"AS IS\" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
**/

package gpuagent

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/ROCm/device-metrics-exporter/pkg/amdgpu/gen/amdgpu"
	"github.com/ROCm/device-metrics-exporter/pkg/exporter/gen/exportermetrics"
	"github.com/ROCm/device-metrics-exporter/pkg/exporter/gen/metricssvc"
	"github.com/ROCm/device-metrics-exporter/pkg/exporter/globals"
	"github.com/ROCm/device-metrics-exporter/pkg/exporter/logger"
	"github.com/ROCm/device-metrics-exporter/pkg/exporter/parserutil"
	"github.com/ROCm/device-metrics-exporter/pkg/exporter/scheduler"
	"github.com/ROCm/device-metrics-exporter/pkg/exporter/utils"
)

type FieldMeta struct {
	Metric prometheus.GaugeVec
	Alias  string
}

// local variables
var (
	mandatoryLables = []string{
		exportermetrics.GPUMetricLabel_GPU_ID.String(),
		exportermetrics.MetricLabel_CARD_MODEL.String(),
		exportermetrics.GPUMetricLabel_GPU_PARTITION_ID.String(),
		exportermetrics.GPUMetricLabel_GPU_COMPUTE_PARTITION_TYPE.String(),
		exportermetrics.GPUMetricLabel_GPU_MEMORY_PARTITION_TYPE.String(),
		exportermetrics.GPUMetricLabel_KFD_PROCESS_ID.String(),
		exportermetrics.GPUMetricLabel_DEPLOYMENT_MODE.String(),
		exportermetrics.MetricLabel_SERIAL_NUMBER.String(),
		exportermetrics.MetricLabel_POD.String(),
		exportermetrics.MetricLabel_NAMESPACE.String(),
		exportermetrics.MetricLabel_CONTAINER.String(),
		exportermetrics.MetricLabel_JOB_ID.String(),
		exportermetrics.MetricLabel_JOB_USER.String(),
		exportermetrics.MetricLabel_JOB_PARTITION.String(),
		exportermetrics.MetricLabel_HOSTNAME.String(),
	}
	// List of supported labels that can be customized
	allowedCustomLabels = []string{
		exportermetrics.MetricLabel_CLUSTER_NAME.String(),
	}
	exportLabels      map[string]bool
	exportFieldMap    map[string]bool // all upper case keys
	fieldMetricsMap   map[string]FieldMeta
	gpuSelectorMap    map[int]bool
	customLabelMap    map[string]string
	extraPodLabelsMap map[string]string
	k8PodLabelsMap    map[string]map[string]string
)

const (
	// starting and ending should align with Profiler Metrics block of enums
	// from exporterconfig.proto
	profilerStarIndex int32 = 801
	profilerEndIndex  int32 = 1200
)

type metrics struct {
	gpuNodesTotal              prometheus.GaugeVec
	gpuPackagePower            prometheus.GaugeVec
	gpuAvgPkgPower             prometheus.GaugeVec
	gpuEdgeTemp                prometheus.GaugeVec
	gpuJunctionTemp            prometheus.GaugeVec
	gpuMemoryTemp              prometheus.GaugeVec
	gpuHBMTemp                 prometheus.GaugeVec
	gpuGFXActivity             prometheus.GaugeVec
	gpuUMCActivity             prometheus.GaugeVec
	gpuMMAActivity             prometheus.GaugeVec
	gpuVCNActivity             prometheus.GaugeVec
	gpuJPEGActivity            prometheus.GaugeVec
	gpuVoltage                 prometheus.GaugeVec
	gpuGFXVoltage              prometheus.GaugeVec
	gpuMemVoltage              prometheus.GaugeVec
	gpuPCIeSpeed               prometheus.GaugeVec
	gpuPCIeMaxSpeed            prometheus.GaugeVec
	gpuPCIeBandwidth           prometheus.GaugeVec
	gpuEnergyConsumed          prometheus.GaugeVec
	gpuPCIeReplayCount         prometheus.GaugeVec
	gpuPCIeRecoveryCount       prometheus.GaugeVec
	gpuPCIeReplayRolloverCount prometheus.GaugeVec
	gpuPCIeNACKSentCount       prometheus.GaugeVec
	gpuPCIeNACKReceivedCount   prometheus.GaugeVec
	gpuClock                   prometheus.GaugeVec
	gpuPowerUsage              prometheus.GaugeVec

	gpuEccCorrectTotal      prometheus.GaugeVec
	gpuEccUncorrectTotal    prometheus.GaugeVec
	gpuEccCorrectSDMA       prometheus.GaugeVec
	gpuEccUncorrectSDMA     prometheus.GaugeVec
	gpuEccCorrectGFX        prometheus.GaugeVec
	gpuEccUncorrectGFX      prometheus.GaugeVec
	gpuEccCorrectMMHUB      prometheus.GaugeVec
	gpuEccUncorrectMMHUB    prometheus.GaugeVec
	gpuEccCorrectATHUB      prometheus.GaugeVec
	gpuEccUncorrectATHUB    prometheus.GaugeVec
	gpuEccCorrectBIF        prometheus.GaugeVec
	gpuEccUncorrectBIF      prometheus.GaugeVec
	gpuEccCorrectHDP        prometheus.GaugeVec
	gpuEccUncorrectHDP      prometheus.GaugeVec
	gpuEccCorrectXgmiWAFL   prometheus.GaugeVec
	gpuEccUncorrectXgmiWAFL prometheus.GaugeVec
	gpuEccCorrectDF         prometheus.GaugeVec
	gpuEccUncorrectDF       prometheus.GaugeVec
	gpuEccCorrectSMN        prometheus.GaugeVec
	gpuEccUncorrectSMN      prometheus.GaugeVec
	gpuEccCorrectSEM        prometheus.GaugeVec
	gpuEccUncorrectSEM      prometheus.GaugeVec
	gpuEccCorrectMP0        prometheus.GaugeVec
	gpuEccUncorrectMP0      prometheus.GaugeVec
	gpuEccCorrectMP1        prometheus.GaugeVec
	gpuEccUncorrectMP1      prometheus.GaugeVec
	gpuEccCorrectFUSE       prometheus.GaugeVec
	gpuEccUncorrectFUSE     prometheus.GaugeVec
	gpuEccCorrectUMC        prometheus.GaugeVec
	gpuEccUncorrectUMC      prometheus.GaugeVec
	xgmiNbrNopTx0           prometheus.GaugeVec
	xgmiNbrReqTx0           prometheus.GaugeVec
	xgmiNbrRespTx0          prometheus.GaugeVec
	xgmiNbrBeatsTx0         prometheus.GaugeVec
	xgmiNbrNopTx1           prometheus.GaugeVec
	xgmiNbrReqTx1           prometheus.GaugeVec
	xgmiNbrRespTx1          prometheus.GaugeVec
	xgmiNbrBeatsTx1         prometheus.GaugeVec
	xgmiNbrTxTput0          prometheus.GaugeVec
	xgmiNbrTxTput1          prometheus.GaugeVec
	xgmiNbrTxTput2          prometheus.GaugeVec
	xgmiNbrTxTput3          prometheus.GaugeVec
	xgmiNbrTxTput4          prometheus.GaugeVec
	xgmiNbrTxTput5          prometheus.GaugeVec

	gpuTotalVram prometheus.GaugeVec
	gpuUsedVram  prometheus.GaugeVec
	gpuFreeVram  prometheus.GaugeVec

	gpuTotalVisibleVram prometheus.GaugeVec
	gpuUsedVisibleVram  prometheus.GaugeVec
	gpuFreeVisibleVram  prometheus.GaugeVec

	gpuTotalGTT prometheus.GaugeVec
	gpuUsedGTT  prometheus.GaugeVec
	gpuFreeGTT  prometheus.GaugeVec

	gpuEccCorrectMCA   prometheus.GaugeVec
	gpuEccUncorrectMCA prometheus.GaugeVec

	gpuEccCorrectVCN   prometheus.GaugeVec
	gpuEccUncorrectVCN prometheus.GaugeVec

	gpuEccCorrectJPEG   prometheus.GaugeVec
	gpuEccUncorrectJPEG prometheus.GaugeVec

	gpuEccCorrectIH   prometheus.GaugeVec
	gpuEccUncorrectIH prometheus.GaugeVec

	gpuEccCorrectMPIO   prometheus.GaugeVec
	gpuEccUncorrectMPIO prometheus.GaugeVec

	gpuHealth prometheus.GaugeVec

	gpuXgmiLinkStatsRx prometheus.GaugeVec
	gpuXgmiLinkStatsTx prometheus.GaugeVec

	gpuCurrAccCtr prometheus.GaugeVec
	gpuProcHRA    prometheus.GaugeVec
	gpuPPTRA      prometheus.GaugeVec
	gpuSTRA       prometheus.GaugeVec
	gpuVRTRA      prometheus.GaugeVec
	gpuHBMTRA     prometheus.GaugeVec

	gpuGfxBusyInst  prometheus.GaugeVec
	gpuVcnBusyInst  prometheus.GaugeVec
	gpuJpegBusyInst prometheus.GaugeVec

	gpuPcieRx             prometheus.GaugeVec
	gpuPcieTx             prometheus.GaugeVec
	gpuPcieBidirBandwidth prometheus.GaugeVec

	// profiler metrics
	gpuGrbmGuiActivity               prometheus.GaugeVec
	gpuSqWaves                       prometheus.GaugeVec
	gpuGrbmCount                     prometheus.GaugeVec
	gpuCpcStatBusy                   prometheus.GaugeVec
	gpuCpcStatIdle                   prometheus.GaugeVec
	gpuCpcStatStall                  prometheus.GaugeVec
	gpuCpcTciuBusy                   prometheus.GaugeVec
	gpuCpcTciuIdle                   prometheus.GaugeVec
	gpuCpcUtcl2iuBusy                prometheus.GaugeVec
	gpuCpcUtcl2iuIdle                prometheus.GaugeVec
	gpuCpcUtcl2iuStall               prometheus.GaugeVec
	gpuCpcME1BusyForPacketDecode     prometheus.GaugeVec
	gpuCpcME1Dc0SpiBusy              prometheus.GaugeVec
	gpuCpcUtcl1StallOnTranslation    prometheus.GaugeVec
	gpuCpcAlwaysCount                prometheus.GaugeVec
	gpuCpcAdcValidChunkNotAvail      prometheus.GaugeVec
	gpuCpcAdcDispatchAllocDone       prometheus.GaugeVec
	gpuCpcAdcValidChunkEnd           prometheus.GaugeVec
	gpuCpcSynFifoFullLevel           prometheus.GaugeVec
	gpuCpcSynFifoFull                prometheus.GaugeVec
	gpuCpcGdBusy                     prometheus.GaugeVec
	gpuCpcTgSend                     prometheus.GaugeVec
	gpuCpcWalkNextChunk              prometheus.GaugeVec
	gpuCpcStalledBySe0Spi            prometheus.GaugeVec
	gpuCpcStalledBySe1Spi            prometheus.GaugeVec
	gpuCpcStalledBySe2Spi            prometheus.GaugeVec
	gpuCpcStalledBySe3Spi            prometheus.GaugeVec
	gpuCpcLteAll                     prometheus.GaugeVec
	gpuCpcSyncWrreqFifoBusy          prometheus.GaugeVec
	gpuCpcCaneBusy                   prometheus.GaugeVec
	gpuCpcCaneStall                  prometheus.GaugeVec
	gpuCpfCmpUtcl1StallOnTrnsalation prometheus.GaugeVec
	gpuCpfStatBusy                   prometheus.GaugeVec
	gpuCpfStatIdle                   prometheus.GaugeVec
	gpuCpfStatStall                  prometheus.GaugeVec
	gpuCpfStatTciuBusy               prometheus.GaugeVec
	gpuCpfStatTciuIdle               prometheus.GaugeVec
	gpuCpfStatTciuStall              prometheus.GaugeVec

	gpuGPUUtil             prometheus.GaugeVec
	gpuFetchSize           prometheus.GaugeVec
	gpuWriteSize           prometheus.GaugeVec
	gpuTotal16Ops          prometheus.GaugeVec
	gpuTotal32Ops          prometheus.GaugeVec
	gpuTotal64Ops          prometheus.GaugeVec
	gpuOccPercent          prometheus.GaugeVec
	gpuTensorActivePercent prometheus.GaugeVec
	gpuValuPipeIssueUtil   prometheus.GaugeVec
	gpuSMActive            prometheus.GaugeVec
	gpuOccElapsed          prometheus.GaugeVec
	gpuOccPerActiveCU      prometheus.GaugeVec
	gpuMeanOccPerCU        prometheus.GaugeVec
	gpuSimdActive          prometheus.GaugeVec
}

func (ga *GPUAgentClient) ResetMetrics() error {
	// reset all label based fields
	for _, prommetric := range fieldMetricsMap {
		prommetric.Metric.Reset()
	}
	return nil
}

func (ga *GPUAgentClient) GetExporterNonGPULabels() []string {
	labelList := []string{
		strings.ToLower(exportermetrics.MetricLabel_HOSTNAME.String()),
	}
	// Add custom labels
	for label, _ := range customLabelMap {
		labelList = append(labelList, strings.ToLower(label))
	}
	return labelList
}

func (ga *GPUAgentClient) GetExportLabels() []string {
	labelList := []string{}
	for key, enabled := range exportLabels {
		if !enabled {
			continue
		}
		labelList = append(labelList, strings.ToLower(key))
	}

	for key := range extraPodLabelsMap {
		exists := false
		for _, label := range labelList {
			if key == label {
				exists = true
				break
			}
		}
		if !exists {
			labelList = append(labelList, key)
		}
	}

	for key := range customLabelMap {
		exists := false
		for _, label := range labelList {
			if key == label {
				exists = true
				break
			}
		}

		// Add only unique labels to export labels
		if !exists {
			labelList = append(labelList, key)
		}
	}

	return labelList
}

func (ga *GPUAgentClient) initLabelConfigs(config *exportermetrics.GPUMetricConfig) {

	// list of mandatory labels
	exportLabels = make(map[string]bool)

	// common labels
	for _, name := range exportermetrics.MetricLabel_name {
		exportLabels[name] = false
	}
	for _, name := range exportermetrics.GPUMetricLabel_name {
		exportLabels[name] = false
	}
	// only mandatory labels are set for default
	for _, name := range mandatoryLables {
		exportLabels[name] = true
	}

	if config != nil {
		for _, name := range config.GetLabels() {
			name = strings.ToUpper(name)
			if _, ok := exportLabels[name]; ok {
				logger.Log.Printf("label %v enabled", name)
				exportLabels[name] = true
			}
		}
	}
	logger.Log.Printf("export-labels updated to %v", exportLabels)
}

func (ga *GPUAgentClient) initProfilerMetrics(config *exportermetrics.GPUMetricConfig) {
	curNodeName, _ := utils.GetHostName()
	// perf metrics are disabled by default as it has a cost associated
	ga.enableProfileMetrics = false
	// check for disable state else enable profiler metrics
	if config != nil && config.GetProfilerMetrics() != nil {
		profilerConfigMap := config.GetProfilerMetrics()
		// check for hostname entry - higher presedence
		if enabled, ok := profilerConfigMap[curNodeName]; ok {
			ga.enableProfileMetrics = enabled
		} else if enabled, ok := profilerConfigMap["all"]; ok {
			ga.enableProfileMetrics = enabled
		}
	}
	logger.Log.Printf("profiler metric state set for %v -> %v", curNodeName, ga.enableProfileMetrics)
}

func initPodExtraLabels(config *exportermetrics.GPUMetricConfig) {
	// initialize pod labels maps
	k8PodLabelsMap = make(map[string]map[string]string)
	if config != nil {
		extraPodLabelsMap = utils.NormalizeExtraPodLabels(config.GetExtraPodLabels())
	}
	logger.Log.Printf("export-labels updated to %v", extraPodLabelsMap)
}

func initCustomLabels(config *exportermetrics.GPUMetricConfig) {
	customLabelMap = make(map[string]string)
	disallowedLabels := []string{}
	if config != nil && config.GetCustomLabels() != nil {
		for _, name := range exportermetrics.GPUMetricLabel_name {
			found := false
			for _, cname := range allowedCustomLabels {
				if name == cname {
					found = true
					break
				}
			}
			if !found {
				disallowedLabels = append(disallowedLabels, strings.ToLower(name))
			}
		}
		cl := config.GetCustomLabels()
		labelCount := 0

		for l, value := range cl {
			if labelCount >= globals.MaxSupportedCustomLabels {
				logger.Log.Printf("Max custom labels supported: %v, ignoring extra labels.", globals.MaxSupportedCustomLabels)
				break
			}
			label := strings.ToLower(l)

			// Check if custom label is a mandatory label, ignore if true
			found := false
			for _, dlabel := range disallowedLabels {
				if dlabel == label {
					logger.Log.Printf("Label %s cannot be customized, ignoring...", dlabel)
					found = true
					break
				}
			}
			if found {
				continue
			}

			// Store all custom labels
			customLabelMap[label] = value
			labelCount++
		}
	}
	logger.Log.Printf("custom labels being exported: %v", customLabelMap)
}

func initGPUSelectorConfig(config *exportermetrics.GPUMetricConfig) {
	if config != nil && config.GetSelector() != "" {
		selector := config.GetSelector()
		indices, err := parserutil.RangeStrToIntIndices(selector)
		if err != nil {
			logger.Log.Printf("GPUConfig.Selector parsing err :%v", err)
			logger.Log.Printf("monitoring all gpu instances")
			return
		}
		for _, ins := range indices {
			gpuSelectorMap[ins] = true
		}
	}
}

func initFieldConfig(config *exportermetrics.GPUMetricConfig) {
	exportFieldMap = make(map[string]bool)
	// setup metric fields in map to be monitored
	// init the map with all supported strings from enum
	enable_default := true
	if config != nil && len(config.GetFields()) != 0 {
		enable_default = false
	}
	for _, name := range exportermetrics.GPUMetricField_name {
		exportFieldMap[name] = enable_default
	}
	if config == nil || len(config.GetFields()) == 0 {
		return
	}
	for _, fieldName := range config.GetFields() {
		fieldName = strings.ToUpper(fieldName)
		if _, ok := exportFieldMap[fieldName]; ok {
			exportFieldMap[fieldName] = true
		}
	}
	// print disabled short list
	for k, v := range exportFieldMap {
		if !v {
			logger.Log.Printf("%v field is disabled", k)
		}
	}
}

func (ga *GPUAgentClient) initFieldMetricsMap() {
	//nolint
	fieldMetricsMap = map[string]FieldMeta{
		exportermetrics.GPUMetricField_GPU_NODES_TOTAL.String():                                    FieldMeta{Metric: ga.m.gpuNodesTotal},
		exportermetrics.GPUMetricField_GPU_PACKAGE_POWER.String():                                  FieldMeta{Metric: ga.m.gpuPackagePower},
		exportermetrics.GPUMetricField_GPU_AVERAGE_PACKAGE_POWER.String():                          FieldMeta{Metric: ga.m.gpuAvgPkgPower},
		exportermetrics.GPUMetricField_GPU_EDGE_TEMPERATURE.String():                               FieldMeta{Metric: ga.m.gpuEdgeTemp},
		exportermetrics.GPUMetricField_GPU_JUNCTION_TEMPERATURE.String():                           FieldMeta{Metric: ga.m.gpuJunctionTemp},
		exportermetrics.GPUMetricField_GPU_MEMORY_TEMPERATURE.String():                             FieldMeta{Metric: ga.m.gpuMemoryTemp},
		exportermetrics.GPUMetricField_GPU_HBM_TEMPERATURE.String():                                FieldMeta{Metric: ga.m.gpuHBMTemp},
		exportermetrics.GPUMetricField_GPU_GFX_ACTIVITY.String():                                   FieldMeta{Metric: ga.m.gpuGFXActivity},
		exportermetrics.GPUMetricField_GPU_UMC_ACTIVITY.String():                                   FieldMeta{Metric: ga.m.gpuUMCActivity},
		exportermetrics.GPUMetricField_GPU_MMA_ACTIVITY.String():                                   FieldMeta{Metric: ga.m.gpuMMAActivity},
		exportermetrics.GPUMetricField_GPU_VCN_ACTIVITY.String():                                   FieldMeta{Metric: ga.m.gpuVCNActivity},
		exportermetrics.GPUMetricField_GPU_JPEG_ACTIVITY.String():                                  FieldMeta{Metric: ga.m.gpuJPEGActivity},
		exportermetrics.GPUMetricField_GPU_VOLTAGE.String():                                        FieldMeta{Metric: ga.m.gpuVoltage},
		exportermetrics.GPUMetricField_GPU_GFX_VOLTAGE.String():                                    FieldMeta{Metric: ga.m.gpuGFXVoltage},
		exportermetrics.GPUMetricField_GPU_MEMORY_VOLTAGE.String():                                 FieldMeta{Metric: ga.m.gpuMemVoltage},
		exportermetrics.GPUMetricField_PCIE_SPEED.String():                                         FieldMeta{Metric: ga.m.gpuPCIeSpeed},
		exportermetrics.GPUMetricField_PCIE_MAX_SPEED.String():                                     FieldMeta{Metric: ga.m.gpuPCIeMaxSpeed},
		exportermetrics.GPUMetricField_PCIE_BANDWIDTH.String():                                     FieldMeta{Metric: ga.m.gpuPCIeBandwidth},
		exportermetrics.GPUMetricField_GPU_ENERGY_CONSUMED.String():                                FieldMeta{Metric: ga.m.gpuEnergyConsumed},
		exportermetrics.GPUMetricField_PCIE_REPLAY_COUNT.String():                                  FieldMeta{Metric: ga.m.gpuPCIeReplayCount},
		exportermetrics.GPUMetricField_PCIE_RECOVERY_COUNT.String():                                FieldMeta{Metric: ga.m.gpuPCIeRecoveryCount},
		exportermetrics.GPUMetricField_PCIE_REPLAY_ROLLOVER_COUNT.String():                         FieldMeta{Metric: ga.m.gpuPCIeReplayRolloverCount},
		exportermetrics.GPUMetricField_PCIE_NACK_SENT_COUNT.String():                               FieldMeta{Metric: ga.m.gpuPCIeNACKSentCount},
		exportermetrics.GPUMetricField_PCIE_NACK_RECEIVED_COUNT.String():                           FieldMeta{Metric: ga.m.gpuPCIeNACKReceivedCount},
		exportermetrics.GPUMetricField_GPU_CLOCK.String():                                          FieldMeta{Metric: ga.m.gpuClock},
		exportermetrics.GPUMetricField_GPU_POWER_USAGE.String():                                    FieldMeta{Metric: ga.m.gpuPowerUsage},
		exportermetrics.GPUMetricField_GPU_TOTAL_VRAM.String():                                     FieldMeta{Metric: ga.m.gpuTotalVram},
		exportermetrics.GPUMetricField_GPU_ECC_CORRECT_TOTAL.String():                              FieldMeta{Metric: ga.m.gpuEccCorrectTotal},
		exportermetrics.GPUMetricField_GPU_ECC_UNCORRECT_TOTAL.String():                            FieldMeta{Metric: ga.m.gpuEccUncorrectTotal},
		exportermetrics.GPUMetricField_GPU_ECC_CORRECT_SDMA.String():                               FieldMeta{Metric: ga.m.gpuEccCorrectSDMA},
		exportermetrics.GPUMetricField_GPU_ECC_UNCORRECT_SDMA.String():                             FieldMeta{Metric: ga.m.gpuEccUncorrectSDMA},
		exportermetrics.GPUMetricField_GPU_ECC_CORRECT_GFX.String():                                FieldMeta{Metric: ga.m.gpuEccCorrectGFX},
		exportermetrics.GPUMetricField_GPU_ECC_UNCORRECT_GFX.String():                              FieldMeta{Metric: ga.m.gpuEccUncorrectGFX},
		exportermetrics.GPUMetricField_GPU_ECC_CORRECT_MMHUB.String():                              FieldMeta{Metric: ga.m.gpuEccCorrectMMHUB},
		exportermetrics.GPUMetricField_GPU_ECC_UNCORRECT_MMHUB.String():                            FieldMeta{Metric: ga.m.gpuEccUncorrectMMHUB},
		exportermetrics.GPUMetricField_GPU_ECC_CORRECT_ATHUB.String():                              FieldMeta{Metric: ga.m.gpuEccCorrectATHUB},
		exportermetrics.GPUMetricField_GPU_ECC_UNCORRECT_ATHUB.String():                            FieldMeta{Metric: ga.m.gpuEccUncorrectATHUB},
		exportermetrics.GPUMetricField_GPU_ECC_CORRECT_BIF.String():                                FieldMeta{Metric: ga.m.gpuEccCorrectBIF},
		exportermetrics.GPUMetricField_GPU_ECC_UNCORRECT_BIF.String():                              FieldMeta{Metric: ga.m.gpuEccUncorrectBIF},
		exportermetrics.GPUMetricField_GPU_ECC_CORRECT_HDP.String():                                FieldMeta{Metric: ga.m.gpuEccCorrectHDP},
		exportermetrics.GPUMetricField_GPU_ECC_UNCORRECT_HDP.String():                              FieldMeta{Metric: ga.m.gpuEccUncorrectHDP},
		exportermetrics.GPUMetricField_GPU_ECC_CORRECT_XGMI_WAFL.String():                          FieldMeta{Metric: ga.m.gpuEccCorrectXgmiWAFL},
		exportermetrics.GPUMetricField_GPU_ECC_UNCORRECT_XGMI_WAFL.String():                        FieldMeta{Metric: ga.m.gpuEccUncorrectXgmiWAFL},
		exportermetrics.GPUMetricField_GPU_ECC_CORRECT_DF.String():                                 FieldMeta{Metric: ga.m.gpuEccCorrectDF},
		exportermetrics.GPUMetricField_GPU_ECC_UNCORRECT_DF.String():                               FieldMeta{Metric: ga.m.gpuEccUncorrectDF},
		exportermetrics.GPUMetricField_GPU_ECC_CORRECT_SMN.String():                                FieldMeta{Metric: ga.m.gpuEccCorrectSMN},
		exportermetrics.GPUMetricField_GPU_ECC_UNCORRECT_SMN.String():                              FieldMeta{Metric: ga.m.gpuEccUncorrectSMN},
		exportermetrics.GPUMetricField_GPU_ECC_CORRECT_SEM.String():                                FieldMeta{Metric: ga.m.gpuEccCorrectSEM},
		exportermetrics.GPUMetricField_GPU_ECC_UNCORRECT_SEM.String():                              FieldMeta{Metric: ga.m.gpuEccUncorrectSEM},
		exportermetrics.GPUMetricField_GPU_ECC_CORRECT_MP0.String():                                FieldMeta{Metric: ga.m.gpuEccCorrectMP0},
		exportermetrics.GPUMetricField_GPU_ECC_UNCORRECT_MP0.String():                              FieldMeta{Metric: ga.m.gpuEccUncorrectMP0},
		exportermetrics.GPUMetricField_GPU_ECC_CORRECT_MP1.String():                                FieldMeta{Metric: ga.m.gpuEccCorrectMP1},
		exportermetrics.GPUMetricField_GPU_ECC_UNCORRECT_MP1.String():                              FieldMeta{Metric: ga.m.gpuEccUncorrectMP1},
		exportermetrics.GPUMetricField_GPU_ECC_CORRECT_FUSE.String():                               FieldMeta{Metric: ga.m.gpuEccCorrectFUSE},
		exportermetrics.GPUMetricField_GPU_ECC_UNCORRECT_FUSE.String():                             FieldMeta{Metric: ga.m.gpuEccUncorrectFUSE},
		exportermetrics.GPUMetricField_GPU_ECC_CORRECT_UMC.String():                                FieldMeta{Metric: ga.m.gpuEccCorrectUMC},
		exportermetrics.GPUMetricField_GPU_ECC_UNCORRECT_UMC.String():                              FieldMeta{Metric: ga.m.gpuEccUncorrectUMC},
		exportermetrics.GPUMetricField_GPU_XGMI_NBR_0_NOP_TX.String():                              FieldMeta{Metric: ga.m.xgmiNbrNopTx0},
		exportermetrics.GPUMetricField_GPU_XGMI_NBR_0_REQ_TX.String():                              FieldMeta{Metric: ga.m.xgmiNbrReqTx0},
		exportermetrics.GPUMetricField_GPU_XGMI_NBR_0_RESP_TX.String():                             FieldMeta{Metric: ga.m.xgmiNbrRespTx0},
		exportermetrics.GPUMetricField_GPU_XGMI_NBR_0_BEATS_TX.String():                            FieldMeta{Metric: ga.m.xgmiNbrBeatsTx0},
		exportermetrics.GPUMetricField_GPU_XGMI_NBR_1_NOP_TX.String():                              FieldMeta{Metric: ga.m.xgmiNbrNopTx1},
		exportermetrics.GPUMetricField_GPU_XGMI_NBR_1_REQ_TX.String():                              FieldMeta{Metric: ga.m.xgmiNbrReqTx1},
		exportermetrics.GPUMetricField_GPU_XGMI_NBR_1_RESP_TX.String():                             FieldMeta{Metric: ga.m.xgmiNbrRespTx1},
		exportermetrics.GPUMetricField_GPU_XGMI_NBR_1_BEATS_TX.String():                            FieldMeta{Metric: ga.m.xgmiNbrBeatsTx1},
		exportermetrics.GPUMetricField_GPU_XGMI_NBR_0_TX_THRPUT.String():                           FieldMeta{Metric: ga.m.xgmiNbrTxTput0},
		exportermetrics.GPUMetricField_GPU_XGMI_NBR_1_TX_THRPUT.String():                           FieldMeta{Metric: ga.m.xgmiNbrTxTput1},
		exportermetrics.GPUMetricField_GPU_XGMI_NBR_2_TX_THRPUT.String():                           FieldMeta{Metric: ga.m.xgmiNbrTxTput2},
		exportermetrics.GPUMetricField_GPU_XGMI_NBR_3_TX_THRPUT.String():                           FieldMeta{Metric: ga.m.xgmiNbrTxTput3},
		exportermetrics.GPUMetricField_GPU_XGMI_NBR_4_TX_THRPUT.String():                           FieldMeta{Metric: ga.m.xgmiNbrTxTput4},
		exportermetrics.GPUMetricField_GPU_XGMI_NBR_5_TX_THRPUT.String():                           FieldMeta{Metric: ga.m.xgmiNbrTxTput5},
		exportermetrics.GPUMetricField_GPU_USED_VRAM.String():                                      FieldMeta{Metric: ga.m.gpuUsedVram},
		exportermetrics.GPUMetricField_GPU_FREE_VRAM.String():                                      FieldMeta{Metric: ga.m.gpuFreeVram},
		exportermetrics.GPUMetricField_GPU_TOTAL_VISIBLE_VRAM.String():                             FieldMeta{Metric: ga.m.gpuTotalVisibleVram},
		exportermetrics.GPUMetricField_GPU_USED_VISIBLE_VRAM.String():                              FieldMeta{Metric: ga.m.gpuUsedVisibleVram},
		exportermetrics.GPUMetricField_GPU_FREE_VISIBLE_VRAM.String():                              FieldMeta{Metric: ga.m.gpuFreeVisibleVram},
		exportermetrics.GPUMetricField_GPU_TOTAL_GTT.String():                                      FieldMeta{Metric: ga.m.gpuTotalGTT},
		exportermetrics.GPUMetricField_GPU_USED_GTT.String():                                       FieldMeta{Metric: ga.m.gpuUsedGTT},
		exportermetrics.GPUMetricField_GPU_FREE_GTT.String():                                       FieldMeta{Metric: ga.m.gpuFreeGTT},
		exportermetrics.GPUMetricField_GPU_ECC_CORRECT_MCA.String():                                FieldMeta{Metric: ga.m.gpuEccCorrectMCA},
		exportermetrics.GPUMetricField_GPU_ECC_UNCORRECT_MCA.String():                              FieldMeta{Metric: ga.m.gpuEccUncorrectMCA},
		exportermetrics.GPUMetricField_GPU_ECC_CORRECT_VCN.String():                                FieldMeta{Metric: ga.m.gpuEccCorrectVCN},
		exportermetrics.GPUMetricField_GPU_ECC_UNCORRECT_VCN.String():                              FieldMeta{Metric: ga.m.gpuEccUncorrectVCN},
		exportermetrics.GPUMetricField_GPU_ECC_CORRECT_JPEG.String():                               FieldMeta{Metric: ga.m.gpuEccCorrectJPEG},
		exportermetrics.GPUMetricField_GPU_ECC_UNCORRECT_JPEG.String():                             FieldMeta{Metric: ga.m.gpuEccUncorrectJPEG},
		exportermetrics.GPUMetricField_GPU_ECC_CORRECT_IH.String():                                 FieldMeta{Metric: ga.m.gpuEccCorrectIH},
		exportermetrics.GPUMetricField_GPU_ECC_UNCORRECT_IH.String():                               FieldMeta{Metric: ga.m.gpuEccUncorrectIH},
		exportermetrics.GPUMetricField_GPU_ECC_CORRECT_MPIO.String():                               FieldMeta{Metric: ga.m.gpuEccCorrectMPIO},
		exportermetrics.GPUMetricField_GPU_ECC_UNCORRECT_MPIO.String():                             FieldMeta{Metric: ga.m.gpuEccUncorrectMPIO},
		exportermetrics.GPUMetricField_GPU_HEALTH.String():                                         FieldMeta{Metric: ga.m.gpuHealth},
		exportermetrics.GPUMetricField_GPU_XGMI_LINK_RX.String():                                   FieldMeta{Metric: ga.m.gpuXgmiLinkStatsRx},
		exportermetrics.GPUMetricField_GPU_XGMI_LINK_TX.String():                                   FieldMeta{Metric: ga.m.gpuXgmiLinkStatsTx},
		exportermetrics.GPUMetricField_GPU_VIOLATION_CURRENT_ACCUMULATED_COUNTER.String():          FieldMeta{Metric: ga.m.gpuCurrAccCtr},
		exportermetrics.GPUMetricField_GPU_VIOLATION_PROCESSOR_HOT_RESIDENCY_ACCUMULATED.String():  FieldMeta{Metric: ga.m.gpuProcHRA},
		exportermetrics.GPUMetricField_GPU_VIOLATION_PPT_RESIDENCY_ACCUMULATED.String():            FieldMeta{Metric: ga.m.gpuPPTRA},
		exportermetrics.GPUMetricField_GPU_VIOLATION_SOCKET_THERMAL_RESIDENCY_ACCUMULATED.String(): FieldMeta{Metric: ga.m.gpuSTRA},
		exportermetrics.GPUMetricField_GPU_VIOLATION_VR_THERMAL_RESIDENCY_ACCUMULATED.String():     FieldMeta{Metric: ga.m.gpuVRTRA},
		exportermetrics.GPUMetricField_GPU_VIOLATION_HBM_THERMAL_RESIDENCY_ACCUMULATED.String():    FieldMeta{Metric: ga.m.gpuHBMTRA},
		exportermetrics.GPUMetricField_GPU_GFX_BUSY_INSTANTANEOUS.String():                         FieldMeta{Metric: ga.m.gpuGfxBusyInst},
		exportermetrics.GPUMetricField_GPU_VCN_BUSY_INSTANTANEOUS.String():                         FieldMeta{Metric: ga.m.gpuVcnBusyInst},
		exportermetrics.GPUMetricField_GPU_JPEG_BUSY_INSTANTANEOUS.String():                        FieldMeta{Metric: ga.m.gpuJpegBusyInst},
		exportermetrics.GPUMetricField_PCIE_RX.String():                                            FieldMeta{Metric: ga.m.gpuPcieRx},
		exportermetrics.GPUMetricField_PCIE_TX.String():                                            FieldMeta{Metric: ga.m.gpuPcieTx},
		exportermetrics.GPUMetricField_PCIE_BIDIRECTIONAL_BANDWIDTH.String():                       FieldMeta{Metric: ga.m.gpuPcieBidirBandwidth},
		// profiler entries
		exportermetrics.GPUMetricField_GPU_PROF_GRBM_GUI_ACTIVE.String():                    FieldMeta{Metric: ga.m.gpuGrbmGuiActivity, Alias: "GRBM_GUI_ACTIVE"},
		exportermetrics.GPUMetricField_GPU_PROF_SQ_WAVES.String():                           FieldMeta{Metric: ga.m.gpuSqWaves, Alias: "SQ_WAVES"},
		exportermetrics.GPUMetricField_GPU_PROF_GRBM_COUNT.String():                         FieldMeta{Metric: ga.m.gpuGrbmCount, Alias: "GRBM_COUNT"},
		exportermetrics.GPUMetricField_GPU_PROF_FETCH_SIZE.String():                         FieldMeta{Metric: ga.m.gpuFetchSize, Alias: "FETCH_SIZE"},
		exportermetrics.GPUMetricField_GPU_PROF_WRITE_SIZE.String():                         FieldMeta{Metric: ga.m.gpuWriteSize, Alias: "WRITE_SIZE"},
		exportermetrics.GPUMetricField_GPU_PROF_TOTAL_16_OPS.String():                       FieldMeta{Metric: ga.m.gpuTotal16Ops, Alias: "TOTAL_16_OPS"},
		exportermetrics.GPUMetricField_GPU_PROF_TOTAL_32_OPS.String():                       FieldMeta{Metric: ga.m.gpuTotal32Ops, Alias: "TOTAL_32_OPS"},
		exportermetrics.GPUMetricField_GPU_PROF_TOTAL_64_OPS.String():                       FieldMeta{Metric: ga.m.gpuTotal64Ops, Alias: "TOTAL_64_OPS"},
		exportermetrics.GPUMetricField_GPU_PROF_CPC_CPC_STAT_BUSY.String():                  FieldMeta{Metric: ga.m.gpuCpcStatBusy, Alias: "CPC_CPC_STAT_BUSY"},
		exportermetrics.GPUMetricField_GPU_PROF_CPC_CPC_STAT_IDLE.String():                  FieldMeta{Metric: ga.m.gpuCpcStatIdle, Alias: "CPC_CPC_STAT_IDLE"},
		exportermetrics.GPUMetricField_GPU_PROF_CPC_CPC_STAT_STALL.String():                 FieldMeta{Metric: ga.m.gpuCpcStatStall, Alias: "CPC_CPC_STAT_STALL"},
		exportermetrics.GPUMetricField_GPU_PROF_CPC_CPC_TCIU_BUSY.String():                  FieldMeta{Metric: ga.m.gpuCpcTciuBusy, Alias: "CPC_CPC_TCIU_BUSY"},
		exportermetrics.GPUMetricField_GPU_PROF_CPC_CPC_TCIU_IDLE.String():                  FieldMeta{Metric: ga.m.gpuCpcTciuIdle, Alias: "CPC_CPC_TCIU_IDLE"},
		exportermetrics.GPUMetricField_GPU_PROF_CPC_CPC_UTCL2IU_BUSY.String():               FieldMeta{Metric: ga.m.gpuCpcUtcl2iuBusy, Alias: "CPC_CPC_UTCL2IU_BUSY"},
		exportermetrics.GPUMetricField_GPU_PROF_CPC_CPC_UTCL2IU_IDLE.String():               FieldMeta{Metric: ga.m.gpuCpcUtcl2iuIdle, Alias: "CPC_CPC_UTCL2IU_IDLE"},
		exportermetrics.GPUMetricField_GPU_PROF_CPC_CPC_UTCL2IU_STALL.String():              FieldMeta{Metric: ga.m.gpuCpcUtcl2iuStall, Alias: "CPC_CPC_UTCL2IU_STALL"},
		exportermetrics.GPUMetricField_GPU_PROF_CPC_ME1_BUSY_FOR_PACKET_DECODE.String():     FieldMeta{Metric: ga.m.gpuCpcME1BusyForPacketDecode, Alias: "CPC_ME1_BUSY_FOR_PACKET_DECODE"},
		exportermetrics.GPUMetricField_GPU_PROF_CPC_ME1_DC0_SPI_BUSY.String():               FieldMeta{Metric: ga.m.gpuCpcME1Dc0SpiBusy, Alias: "CPC_ME1_DC0_SPI_BUSY"},
		exportermetrics.GPUMetricField_GPU_PROF_CPC_UTCL1_STALL_ON_TRANSLATION.String():     FieldMeta{Metric: ga.m.gpuCpcUtcl1StallOnTranslation, Alias: "CPC_UTCL1_STALL_ON_TRANSLATION"},
		exportermetrics.GPUMetricField_GPU_PROF_CPC_ALWAYS_COUNT.String():                   FieldMeta{Metric: ga.m.gpuCpcAlwaysCount, Alias: "CPC_ALWAYS_COUNT"},
		exportermetrics.GPUMetricField_GPU_PROF_CPC_ADC_VALID_CHUNK_NOT_AVAIL.String():      FieldMeta{Metric: ga.m.gpuCpcAdcValidChunkNotAvail, Alias: "CPC_ADC_VALID_CHUNK_NOT_AVAIL"},
		exportermetrics.GPUMetricField_GPU_PROF_CPC_ADC_DISPATCH_ALLOC_DONE.String():        FieldMeta{Metric: ga.m.gpuCpcAdcDispatchAllocDone, Alias: "CPC_ADC_DISPATCH_ALLOC_DONE"},
		exportermetrics.GPUMetricField_GPU_PROF_CPC_ADC_VALID_CHUNK_END.String():            FieldMeta{Metric: ga.m.gpuCpcAdcValidChunkEnd, Alias: "CPC_ADC_VALID_CHUNK_END"},
		exportermetrics.GPUMetricField_GPU_PROF_CPC_SYNC_FIFO_FULL_LEVEL.String():           FieldMeta{Metric: ga.m.gpuCpcSynFifoFullLevel, Alias: "CPC_SYNC_FIFO_FULL_LEVEL"},
		exportermetrics.GPUMetricField_GPU_PROF_CPC_SYNC_FIFO_FULL.String():                 FieldMeta{Metric: ga.m.gpuCpcSynFifoFull, Alias: "CPC_SYNC_FIFO_FULL"},
		exportermetrics.GPUMetricField_GPU_PROF_CPC_GD_BUSY.String():                        FieldMeta{Metric: ga.m.gpuCpcGdBusy, Alias: "CPC_GD_BUSY"},
		exportermetrics.GPUMetricField_GPU_PROF_CPC_TG_SEND.String():                        FieldMeta{Metric: ga.m.gpuCpcTgSend, Alias: "CPC_TG_SEND"},
		exportermetrics.GPUMetricField_GPU_PROF_CPC_WALK_NEXT_CHUNK.String():                FieldMeta{Metric: ga.m.gpuCpcWalkNextChunk, Alias: "CPC_WALK_NEXT_CHUNK"},
		exportermetrics.GPUMetricField_GPU_PROF_CPC_STALLED_BY_SE0_SPI.String():             FieldMeta{Metric: ga.m.gpuCpcStalledBySe0Spi, Alias: "CPC_STALLED_BY_SE0_SPI"},
		exportermetrics.GPUMetricField_GPU_PROF_CPC_STALLED_BY_SE1_SPI.String():             FieldMeta{Metric: ga.m.gpuCpcStalledBySe1Spi, Alias: "CPC_STALLED_BY_SE1_SPI"},
		exportermetrics.GPUMetricField_GPU_PROF_CPC_STALLED_BY_SE2_SPI.String():             FieldMeta{Metric: ga.m.gpuCpcStalledBySe2Spi, Alias: "CPC_STALLED_BY_SE2_SPI"},
		exportermetrics.GPUMetricField_GPU_PROF_CPC_STALLED_BY_SE3_SPI.String():             FieldMeta{Metric: ga.m.gpuCpcStalledBySe3Spi, Alias: "CPC_STALLED_BY_SE3_SPI"},
		exportermetrics.GPUMetricField_GPU_PROF_CPC_LTE_ALL.String():                        FieldMeta{Metric: ga.m.gpuCpcLteAll, Alias: "CPC_LTE_ALL"},
		exportermetrics.GPUMetricField_GPU_PROF_CPC_SYNC_WRREQ_FIFO_BUSY.String():           FieldMeta{Metric: ga.m.gpuCpcSyncWrreqFifoBusy, Alias: "CPC_SYNC_WRREQ_FIFO_BUSY"},
		exportermetrics.GPUMetricField_GPU_PROF_CPC_CANE_BUSY.String():                      FieldMeta{Metric: ga.m.gpuCpcCaneBusy, Alias: "CPC_CANE_BUSY"},
		exportermetrics.GPUMetricField_GPU_PROF_CPC_CANE_STALL.String():                     FieldMeta{Metric: ga.m.gpuCpcCaneStall, Alias: "CPC_CANE_STALL"},
		exportermetrics.GPUMetricField_GPU_PROF_CPF_CMP_UTCL1_STALL_ON_TRANSLATION.String(): FieldMeta{Metric: ga.m.gpuCpfCmpUtcl1StallOnTrnsalation, Alias: "CPF_CMP_UTCL1_STALL_ON_TRANSLATION"},
		exportermetrics.GPUMetricField_GPU_PROF_CPF_CPF_STAT_BUSY.String():                  FieldMeta{Metric: ga.m.gpuCpfStatBusy, Alias: "CPF_CPF_STAT_BUSY"},
		exportermetrics.GPUMetricField_GPU_PROF_CPF_CPF_STAT_IDLE.String():                  FieldMeta{Metric: ga.m.gpuCpfStatIdle, Alias: "CPF_CPF_STAT_IDLE"},
		exportermetrics.GPUMetricField_GPU_PROF_CPF_CPF_STAT_STALL.String():                 FieldMeta{Metric: ga.m.gpuCpfStatStall, Alias: "CPF_CPF_STAT_STALL"},
		exportermetrics.GPUMetricField_GPU_PROF_CPF_CPF_TCIU_BUSY.String():                  FieldMeta{Metric: ga.m.gpuCpfStatTciuBusy, Alias: "CPF_CPF_TCIU_BUSY"},
		exportermetrics.GPUMetricField_GPU_PROF_CPF_CPF_TCIU_IDLE.String():                  FieldMeta{Metric: ga.m.gpuCpfStatTciuIdle, Alias: "CPF_CPF_TCIU_IDLE"},
		exportermetrics.GPUMetricField_GPU_PROF_CPF_CPF_TCIU_STALL.String():                 FieldMeta{Metric: ga.m.gpuCpfStatTciuStall, Alias: "CPF_CPF_TCIU_STALL"},
		exportermetrics.GPUMetricField_GPU_PROF_GUI_UTIL_PERCENT.String():                   FieldMeta{Metric: ga.m.gpuGPUUtil, Alias: "GPU_UTIL"},
		exportermetrics.GPUMetricField_GPU_PROF_OCCUPANCY_PERCENT.String():                  FieldMeta{Metric: ga.m.gpuOccPercent, Alias: "OccupancyPercent"},
		exportermetrics.GPUMetricField_GPU_PROF_TENSOR_ACTIVE_PERCENT.String():              FieldMeta{Metric: ga.m.gpuTensorActivePercent, Alias: "MfmaUtil"},
		exportermetrics.GPUMetricField_GPU_PROF_VALU_PIPE_ISSUE_UTIL.String():               FieldMeta{Metric: ga.m.gpuValuPipeIssueUtil, Alias: "ValuPipeIssueUtil"},
		exportermetrics.GPUMetricField_GPU_PROF_SM_ACTIVE.String():                          FieldMeta{Metric: ga.m.gpuSMActive, Alias: "VALUBusy"},
		exportermetrics.GPUMetricField_GPU_PROF_OCCUPANCY_ELAPSED.String():                  FieldMeta{Metric: ga.m.gpuOccElapsed, Alias: "GRBM_GUI_ACTIVE"},
		exportermetrics.GPUMetricField_GPU_PROF_OCCUPANCY_PER_ACTIVE_CU.String():            FieldMeta{Metric: ga.m.gpuOccPerActiveCU, Alias: "MeanOccupancyPerActiveCU"},
		exportermetrics.GPUMetricField_GPU_PROF_OCCUPANCY_PER_CU.String():                   FieldMeta{Metric: ga.m.gpuMeanOccPerCU, Alias: "MeanOccupancyPerCU"},
		exportermetrics.GPUMetricField_GPU_PROF_SIMD_UTILIZATION.String():                   FieldMeta{Metric: ga.m.gpuSimdActive, Alias: "SIMD_UTILIZATION"},
	}
	logger.Log.Printf("Total GPU fields supported : %+v", len(fieldMetricsMap))

}

func (ga *GPUAgentClient) initProfilerMetricsField() {
	if ga.isProfilerEnabled() {
		// only query enabled fields
		profilerFields := []string{}
		for f, enabled := range exportFieldMap {
			if !enabled {
				continue
			}
			if meta, ok := fieldMetricsMap[f]; ok {
				if meta.Alias != "" {
					profilerFields = append(profilerFields, meta.Alias)
				}
			}
		}
		ga.rocpclient.SetFields(profilerFields)
		return
	}
	// to avoid exporting when disabled
	// update the exporter fields map to disable performance register fields
	for i := profilerStarIndex; i <= profilerEndIndex; i++ {
		if name, exist := exportermetrics.GPUMetricField_name[i]; exist {
			fieldName := strings.ToUpper(name)
			if _, ok := exportFieldMap[fieldName]; ok {
				exportFieldMap[fieldName] = false
				logger.Log.Printf("%v field is disabled", fieldName)
			}
		}
	}
}

func (ga *GPUAgentClient) initPrometheusMetrics() {
	nonGpuLabels := ga.GetExporterNonGPULabels()
	labels := ga.GetExportLabels()
	ga.m = &metrics{
		gpuNodesTotal: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_nodes_total",
			Help: "Number of GPUs in the node",
		},
			nonGpuLabels),
		gpuPackagePower: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_package_power",
			Help: "Current socket power in Watts",
		},
			labels),
		gpuAvgPkgPower: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_average_package_power",
			Help: "Average socket power in Watts",
		},
			labels),
		gpuEdgeTemp: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_edge_temperature",
			Help: "Current edge temperature in Celsius",
		},
			labels),
		gpuJunctionTemp: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_junction_temperature",
			Help: "Current junction/hotspot temperature in Celsius",
		},
			labels),
		gpuMemoryTemp: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_memory_temperature",
			Help: "Current memory temperature in Celsius",
		},
			labels),
		gpuHBMTemp: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_hbm_temperature",
			Help: "List of current HBM temperatures in Celsius",
		},
			append([]string{"hbm_index"}, labels...)),
		gpuGFXActivity: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_gfx_activity",
			Help: "Graphics engine usage in Percentage (0-100)",
		},
			labels),
		gpuUMCActivity: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_umc_activity",
			Help: "Memory engine usage in Percentage (0-100)",
		},
			labels),
		gpuMMAActivity: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_mma_activity",
			Help: "Average MultiMedia (MM) engine usage in Percentage (0-100)",
		},
			labels),
		gpuVCNActivity: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_vcn_activity",
			Help: "List of Video Core Next (VCN) encoe/decode usage in percentage",
		},
			append([]string{"vcn_index"}, labels...)),
		gpuJPEGActivity: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_jpeg_activity",
			Help: "List of JPEG engine usage in Percentage (0-100)",
		},
			append([]string{"jpeg_index"}, labels...)),
		gpuVoltage: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_voltage",
			Help: "Current SoC voltage in mV",
		},
			labels),
		gpuGFXVoltage: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_gfx_voltage",
			Help: "Current gfx voltage in mV",
		},
			labels),
		gpuMemVoltage: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_memory_voltage",
			Help: "Current memory voltage in mV",
		},
			labels),
		gpuPCIeSpeed: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "pcie_speed",
			Help: "Current PCIe speed in GT/s",
		},
			labels),
		gpuPCIeMaxSpeed: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "pcie_max_speed",
			Help: "Maximum PCIe speed in GT/s",
		},
			labels),
		gpuPCIeBandwidth: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "pcie_bandwidth",
			Help: "Current PCIe bandwidth in Mb/s",
		},
			labels),
		gpuEnergyConsumed: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_energy_consumed",
			Help: "Accumulated energy consumed by the GPU in uJ",
		},
			labels),
		gpuPCIeReplayCount: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "pcie_replay_count",
			Help: "Total number of PCIe replays",
		},
			labels),
		gpuPCIeRecoveryCount: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "pcie_recovery_count",
			Help: "Total number of PCIe recoveries",
		},
			labels),
		gpuPCIeReplayRolloverCount: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "pcie_replay_rollover_count",
			Help: "PCIe replay accumulated count",
		},
			labels),
		gpuPCIeNACKSentCount: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "pcie_nack_sent_count",
			Help: "PCIe NAK sent accumulated count",
		},
			labels),
		gpuPCIeNACKReceivedCount: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "pcie_nack_received_count",
			Help: "PCIe NAK received accumulated count",
		},
			labels),
		gpuClock: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_clock",
			Help: "List of current GPU clock frequencies in MHz",
		},
			append([]string{"clock_index", "clock_type"}, labels...)),
		gpuPowerUsage: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_power_usage",
			Help: "GPU Power usage in Watts",
		},
			labels),
		gpuTotalVram: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_total_vram",
			Help: "Total VRAM memory of the GPU (in MB)",
		},
			labels),
		gpuUsedVram: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_used_vram",
			Help: "Used VRAM memory of the GPU (in MB)",
		},
			labels),
		gpuFreeVram: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_free_vram",
			Help: "Free VRAM memory of the GPU (in MB)",
		},
			labels),
		gpuTotalVisibleVram: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_total_visible_vram",
			Help: "Total visible VRAM memory of the GPU (in MB)",
		},
			labels),
		gpuUsedVisibleVram: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_used_visible_vram",
			Help: "Used visible VRAM memory of the GPU (in MB)",
		},
			labels),
		gpuFreeVisibleVram: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_free_visible_vram",
			Help: "Free visible VRAM memory of the GPU (in MB)",
		},
			labels),
		gpuTotalGTT: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_total_gtt",
			Help: "Total graphics translation table memory of the GPU (in MB)",
		},
			labels),
		gpuUsedGTT: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_used_gtt",
			Help: "Used graphics translation table memory of the GPU (in MB)",
		},
			labels),
		gpuFreeGTT: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_free_gtt",
			Help: "Free graphics translation table memory of the GPU (in MB)",
		},
			labels),
		gpuEccCorrectTotal: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_ecc_correct_total",
			Help: "Total Correctable error count",
		},
			labels),
		gpuEccUncorrectTotal: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_ecc_uncorrect_total",
			Help: "Total Uncorrectable error count",
		},
			labels),
		gpuEccCorrectSDMA: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_ecc_correct_sdma",
			Help: "Correctable error count in SDMA block",
		},
			labels),
		gpuEccUncorrectSDMA: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_ecc_uncorrect_sdma",
			Help: "Uncorrectable error count in SDMA block",
		},
			labels),
		gpuEccCorrectGFX: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_ecc_correct_gfx",
			Help: "Correctable error count in GFX block",
		},
			labels),
		gpuEccUncorrectGFX: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_ecc_uncorrect_gfx",
			Help: "Uncorrectable error count in GFX block",
		},
			labels),
		gpuEccCorrectMMHUB: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_ecc_correct_mmhub",
			Help: "Correctable error count in MMHUB block",
		},
			labels),
		gpuEccUncorrectMMHUB: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_ecc_uncorrect_mmhub",
			Help: "Uncorrectable error count in MMHUB block",
		},
			labels),
		gpuEccCorrectATHUB: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_ecc_correct_athub",
			Help: "Correctable error count in ATHUB block",
		},
			labels),
		gpuEccUncorrectATHUB: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_ecc_uncorrect_athub",
			Help: "Uncorrectable error count in ATHUB block",
		},
			labels),
		gpuEccCorrectBIF: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_ecc_correct_bif",
			Help: "Correctable error count in BIF block",
		},
			labels),
		gpuEccUncorrectBIF: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_ecc_uncorrect_bif",
			Help: "Uncorrectable error count in BIF block",
		},
			labels),
		gpuEccCorrectHDP: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_ecc_correct_hdp",
			Help: "Correctable error count in HDP block",
		},
			labels),
		gpuEccUncorrectHDP: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_ecc_uncorrect_hdp",
			Help: "Uncorrectable error count in HDP block",
		},
			labels),
		gpuEccCorrectXgmiWAFL: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_ecc_correct_xgmi_wafl",
			Help: "Correctable error count in WAFL block",
		},
			labels),
		gpuEccUncorrectXgmiWAFL: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_ecc_uncorrect_xgmi_wafl",
			Help: "Uncorrectable error count in WAFL block",
		},
			labels),
		gpuEccCorrectDF: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_ecc_correct_df",
			Help: "Correctable error count in DF block",
		},
			labels),
		gpuEccUncorrectDF: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_ecc_uncorrect_df",
			Help: "Uncorrectable error count in DF block",
		},
			labels),
		gpuEccCorrectSMN: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_ecc_correct_smn",
			Help: "Correctable error count in SMN block",
		},
			labels),
		gpuEccUncorrectSMN: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_ecc_uncorrect_smn",
			Help: "Uncorrectable error count in SMN block",
		},
			labels),
		gpuEccCorrectSEM: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_ecc_correct_sem",
			Help: "Correctable error count in SEM block",
		},
			labels),
		gpuEccUncorrectSEM: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_ecc_uncorrect_sem",
			Help: "Uncorrectable error count in SEM block",
		},
			labels),
		gpuEccCorrectMP0: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_ecc_correct_mp0",
			Help: "Correctable error count in MP0 block",
		},
			labels),
		gpuEccUncorrectMP0: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_ecc_uncorrect_mp0",
			Help: "Uncorrectable error count in MP0 block",
		},
			labels),
		gpuEccCorrectMP1: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_ecc_correct_mp1",
			Help: "Correctable error count in MP1 block",
		},
			labels),
		gpuEccUncorrectMP1: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_ecc_uncorrect_mp1",
			Help: "Uncorrectable error count in MP1 block",
		},
			labels),
		gpuEccCorrectFUSE: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_ecc_correct_fuse",
			Help: "Correctable error count in Fuse block",
		},
			labels),
		gpuEccUncorrectFUSE: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_ecc_uncorrect_fuse",
			Help: "Uncorrectable error count in Fuse block",
		},
			labels),
		gpuEccCorrectUMC: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_ecc_correct_umc",
			Help: "Correctable error count in UMC block",
		},
			labels),
		gpuEccUncorrectUMC: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_ecc_uncorrect_umc",
			Help: "Uncorrectable error count in UMC block",
		},
			labels),
		xgmiNbrNopTx0: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_xgmi_nbr_0_nop_tx",
			Help: "NOPs sent to neighbor 0",
		},
			labels),
		xgmiNbrNopTx1: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_xgmi_nbr_1_nop_tx",
			Help: "NOPs sent to neighbor 1",
		},
			labels),
		xgmiNbrReqTx0: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_xgmi_nbr_0_req_tx",
			Help: "Outgoing requests to neighbor 0",
		},
			labels),
		xgmiNbrReqTx1: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_xgmi_nbr_1_req_tx",
			Help: "Outgoing requests to neighbor 1",
		},
			labels),
		xgmiNbrRespTx0: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_xgmi_nbr_0_resp_tx",
			Help: "Outgoing responses to neighbor 0",
		},
			labels),
		xgmiNbrRespTx1: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_xgmi_nbr_1_resp_tx",
			Help: "Outgoing responses to neighbor 1",
		},
			labels),
		xgmiNbrBeatsTx0: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_xgmi_nbr_0_beats_tx",
			Help: "Data beats sent to neighbor 0; Each beat represents 32 bytes",
		},
			labels),
		xgmiNbrBeatsTx1: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_xgmi_nbr_1_beats_tx",
			Help: "Data beats sent to neighbor 1; Each beat represents 32 bytes",
		},
			labels),
		xgmiNbrTxTput0: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_xgmi_nbr_0_tx_thrput",
			Help: "Represents the number of outbound beats (each representing 32 bytes) on link 0; Throughput = BEATS/time_running * 10^9  bytes/sec",
		},
			labels),
		xgmiNbrTxTput1: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_xgmi_nbr_1_tx_thrput",
			Help: "Represents the number of outbound beats (each representing 32 bytes) on link 1; Throughput = BEATS/time_running * 10^9  bytes/sec",
		},
			labels),
		xgmiNbrTxTput2: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_xgmi_nbr_2_tx_thrput",
			Help: "Represents the number of outbound beats (each representing 32 bytes) on link 2; Throughput = BEATS/time_running * 10^9  bytes/sec",
		},
			labels),
		xgmiNbrTxTput3: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_xgmi_nbr_3_tx_thrput",
			Help: "Represents the number of outbound beats (each representing 32 bytes) on link 3; Throughput = BEATS/time_running * 10^9  bytes/sec",
		},
			labels),
		xgmiNbrTxTput4: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_xgmi_nbr_4_tx_thrput",
			Help: "Represents the number of outbound beats (each representing 32 bytes) on link 4; Throughput = BEATS/time_running * 10^9  bytes/sec",
		},
			labels),
		xgmiNbrTxTput5: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_xgmi_nbr_5_tx_thrput",
			Help: "Represents the number of outbound beats (each representing 32 bytes) on link 5; Throughput = BEATS/time_running * 10^9  bytes/sec",
		},
			labels),
		gpuEccCorrectMCA: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_ecc_correct_mca",
			Help: "Correctable error count in MCA block",
		},
			labels),
		gpuEccUncorrectMCA: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_ecc_uncorrect_mca",
			Help: "Uncorrectable error count in MCA block",
		},
			labels),
		gpuEccCorrectVCN: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_ecc_correct_vcn",
			Help: "Correctable error count in VCN block",
		},
			labels),
		gpuEccUncorrectVCN: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_ecc_uncorrect_vcn",
			Help: "Uncorrectable error count in VCN block",
		},
			labels),
		gpuEccCorrectJPEG: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_ecc_correct_jpeg",
			Help: "Correctable error count in JPEG block",
		},
			labels),
		gpuEccUncorrectJPEG: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_ecc_uncorrect_jpeg",
			Help: "Uncorrectable error count in JPEG block",
		},
			labels),
		gpuEccCorrectIH: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_ecc_correct_ih",
			Help: "Correctable error count in IH block",
		},
			labels),
		gpuEccUncorrectIH: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_ecc_uncorrect_ih",
			Help: "Uncorrectable error count in IH block",
		},
			labels),
		gpuEccCorrectMPIO: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_ecc_correct_mpio",
			Help: "Correctable error count in MPIO block",
		},
			labels),
		gpuEccUncorrectMPIO: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_ecc_uncorrect_mpio",
			Help: "Uncorrectable error count in MPIO block",
		},
			labels),
		gpuHealth: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_health",
			Help: "Health of the GPU (0 = Unhealthy | 1 = Healthy)",
		},
			labels),
		gpuXgmiLinkStatsRx: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_xgmi_link_rx",
			Help: "Accumulated XGMI Link Data Read in KB",
		},
			append([]string{"link_index"}, labels...)),
		gpuXgmiLinkStatsTx: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_xgmi_link_tx",
			Help: "Accumulated XGMI Link Data Write in KB",
		},
			append([]string{"link_index"}, labels...)),
		gpuCurrAccCtr: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_violation_current_accumulated_counter",
			Help: "current accumulated violation counter",
		},
			labels),
		gpuProcHRA: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_violation_processor_hot_residency_accumulated",
			Help: "process hot residency accumulated violation counter",
		},
			labels),
		gpuPPTRA: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_violation_ppt_residency_accumulated",
			Help: "package power tracking accumulated violation counter",
		},
			labels),
		gpuSTRA: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_violation_socket_thermal_residency_accumulated",
			Help: "socket thermal accumulated violation counter",
		},
			labels),
		gpuVRTRA: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_violation_vr_thermal_residency_accumulated",
			Help: "voltage rail accumulated violation counter",
		},
			labels),
		gpuHBMTRA: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_violation_hbm_thermal_residency_accumulated",
			Help: "HBM accumulated violation counter",
		},
			labels),
		gpuGfxBusyInst: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_gfx_busy_instantaneous",
			Help: "Gfx busy instantaneous per accelerated compute processor(xcp) per compute core (xcc), as per partitioning of the system",
		},
			append([]string{"xcc_index"}, labels...)),
		gpuVcnBusyInst: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_vcn_busy_instantaneous",
			Help: "Vcn busy instantaneous per accelerated compute processor(xcp) per compute core (xcc), as per partitioning of the system",
		},
			append([]string{"xcc_index"}, labels...)),
		gpuJpegBusyInst: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_jpeg_busy_instantaneous",
			Help: "Jpeg busy instantaneous per accelerated compute processor(xcp) per compute core (xcc), as per partitioning of the system",
		},
			append([]string{"xcc_index"}, labels...)),
		gpuGrbmGuiActivity: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_prof_grbm_gui_active",
			Help: "Number of GPU active cycles",
		},
			labels),
		gpuSqWaves: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_prof_sq_waves",
			Help: "Number of wavefronts dispatched to sequencers, including both new and restored wavefronts",
		},
			labels),
		gpuGrbmCount: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_prof_grbm_count",
			Help: "Number of free-running GPU cycles",
		},
			labels),
		gpuGPUUtil: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_prof_gui_util_percent",
			Help: "Percentage of the time that GUI is active",
		},
			labels),
		gpuFetchSize: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_prof_fetch_size",
			Help: "The total kilobytes fetched from the video memory. This is measured with all extra fetches and any cache or memory effects taken into account",
		},
			labels),
		gpuWriteSize: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_prof_write_size",
			Help: "The total kilobytes written to the video memory. This is measured with all extra fetches and any cache or memory effects taken into account",
		},
			labels),
		gpuTotal16Ops: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_prof_total_16_ops",
			Help: "The number of 16 bits OPS executed",
		},
			labels),
		gpuTotal32Ops: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_prof_total_32_ops",
			Help: "The number of 32 bits OPS executed",
		},
			labels),
		gpuTotal64Ops: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_prof_total_64_ops",
			Help: "The number of 64 bits OPS executed",
		},
			labels),
		gpuCpcStatBusy: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_prof_cpc_cpc_stat_busy",
			Help: "Number of cycles command processor-compute is busy",
		},
			labels),
		gpuCpcStatIdle: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_prof_cpc_cpc_stat_idle",
			Help: "Number of cycles command processor-compute is idle",
		},
			labels),
		gpuCpcStatStall: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_prof_cpc_cpc_stat_stall",
			Help: "Number of cycles command processor-compute is stalled",
		},
			labels),
		gpuCpcTciuBusy: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_prof_cpc_cpc_tciu_busy",
			Help: "Number of cycles command processor-compute texture cache interface unit interface is busy",
		},
			labels),
		gpuCpcTciuIdle: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_prof_cpc_cpc_tciu_idle",
			Help: "Number of cycles command processor-compute texture cache interface unit interface is idle",
		},
			labels),
		gpuCpcUtcl2iuBusy: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_prof_cpc_cpc_utcl2iu_busy",
			Help: "Number of cycles command processor-compute unified translation cache (L2) interface is busy",
		},
			labels),
		gpuCpcUtcl2iuIdle: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_prof_cpc_cpc_utcl2iu_idle",
			Help: "Number of cycles command processor-compute unified translation cache (L2) interface is idle",
		},
			labels),
		gpuCpcUtcl2iuStall: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_prof_cpc_cpc_utcl2iu_stall",
			Help: "Number of cycles command processor-compute unified translation cache (L2) interface is stalled",
		},
			labels),
		gpuCpcME1BusyForPacketDecode: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_prof_cpc_me1_busy_for_packet_decode",
			Help: "Number of cycles command processor-compute micro engine is busy decoding packets",
		},
			labels),
		gpuCpcME1Dc0SpiBusy: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_prof_cpc_me1_dc0_spi_busy",
			Help: "Number of cycles command processor-compute micro engine processor is busy",
		},
			labels),
		gpuCpcUtcl1StallOnTranslation: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_prof_cpc_utcl1_stall_on_translation",
			Help: "Number of cycles one of the unified translation caches (L1) is stalled waiting on translation",
		},
			labels),
		gpuCpcAlwaysCount: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_prof_cpc_always_count",
			Help: "CPC Always Count",
		},
			labels),
		gpuCpcAdcValidChunkNotAvail: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_prof_cpc_adc_valid_chunk_not_avail",
			Help: "CPC ADC valid chunk not available when dispatch walking is in progress at multi-xcc mode",
		},
			labels),
		gpuCpcAdcDispatchAllocDone: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_prof_cpc_adc_dispatch_alloc_done",
			Help: "CPC ADC dispatch allocation done",
		},
			labels),
		gpuCpcAdcValidChunkEnd: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_prof_cpc_adc_valid_chunk_end",
			Help: "CPC ADC cralwer valid chunk end at multi-xcc mode",
		},
			labels),
		gpuCpcSynFifoFullLevel: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_prof_cpc_sync_fifo_full_level",
			Help: "CPC SYNC FIFO full last cycles",
		},
			labels),
		gpuCpcSynFifoFull: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_prof_cpc_sync_fifo_full",
			Help: "CPC SYNC FIFO full times",
		},
			labels),
		gpuCpcGdBusy: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_prof_cpc_gd_busy",
			Help: "CPC ADC busy",
		},
			labels),
		gpuCpcTgSend: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_prof_cpc_tg_send",
			Help: "CPC ADC thread group send",
		},
			labels),
		gpuCpcWalkNextChunk: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_prof_cpc_walk_next_chunk",
			Help: "CPC ADC walking next valid chunk at multi-xcc mode",
		},
			labels),
		gpuCpcStalledBySe0Spi: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_prof_cpc_stalled_by_se0_spi",
			Help: "CPC ADC csdata stalled by SE0SPI",
		},
			labels),
		gpuCpcStalledBySe1Spi: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_prof_cpc_stalled_by_se1_spi",
			Help: "CPC ADC csdata stalled by SE1SPI",
		},
			labels),
		gpuCpcStalledBySe2Spi: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_prof_cpc_stalled_by_se2_spi",
			Help: "CPC ADC csdata stalled by SE2SPI",
		},
			labels),
		gpuCpcStalledBySe3Spi: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_prof_cpc_stalled_by_se3_spi",
			Help: "CPC ADC csdata stalled by SE3SPI",
		},
			labels),
		gpuCpcLteAll: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_prof_cpc_lte_all",
			Help: "CPC Sync counter LteAll, only Master XCD cares LteAll",
		},
			labels),
		gpuCpcSyncWrreqFifoBusy: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_prof_cpc_sync_wrreq_fifo_busy",
			Help: "CPC Sync Counter Request Fifo is not empty",
		},
			labels),
		gpuCpcCaneBusy: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_prof_cpc_cane_busy",
			Help: "CPC CANE bus busy, means there are inflight sync counter requests",
		},
			labels),
		gpuCpcCaneStall: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_prof_cpc_cane_stall",
			Help: "CPC Sync counter sending is stalled by CANE",
		},
			labels),
		gpuCpfCmpUtcl1StallOnTrnsalation: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_prof_cpf_cmp_utcl1_stall_on_translation",
			Help: "One of the Compute UTCL1s is stalled waiting on translation, XNACK or PENDING response",
		},
			labels),
		gpuCpfStatBusy: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_prof_cpf_cpf_stat_busy",
			Help: "CPF Busy",
		},
			labels),
		gpuCpfStatIdle: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_prof_cpf_cpf_stat_idle",
			Help: "CPF Idle",
		},
			labels),
		gpuCpfStatStall: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_prof_cpf_cpf_stat_stall",
			Help: "CPF Stalled",
		},
			labels),
		gpuCpfStatTciuBusy: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_prof_cpf_cpf_tciu_busy",
			Help: "CPF TCIU interface Busy",
		},
			labels),
		gpuCpfStatTciuIdle: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_prof_cpf_cpf_tciu_idle",
			Help: "CPF TCIU interface Idle",
		},
			labels),
		gpuCpfStatTciuStall: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_prof_cpf_cpf_tciu_stall",
			Help: "CPF TCIU interface Stalled waiting on Free, Tags",
		},
			labels),
		gpuOccPercent: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_prof_occupancy_percent",
			Help: "GPU Occupancy as % of maximum",
		},
			labels),
		gpuTensorActivePercent: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_prof_tensor_active_percent",
			Help: "MFMA Utililization Unit: percent",
		},
			labels),
		gpuValuPipeIssueUtil: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_prof_valu_pipe_issue_util",
			Help: "Percentage of the time that GUI is active",
		},
			labels),
		gpuSMActive: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_prof_sm_active",
			Help: "The percentage of GPUTime vector ALU instructions are processed. Value range: 0% (bad) to 100% (optimal)",
		},
			labels),
		gpuOccElapsed: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_prof_occupancy_elapsed",
			Help: "Number of GPU active cycles",
		},
			labels),
		gpuOccPerActiveCU: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_prof_occupancy_per_active_cu",
			Help: "Mean occupancy per active compute unit",
		},
			labels),
		gpuMeanOccPerCU: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_prof_occupancy_per_cu",
			Help: "Mean occupancy per compute unit",
		},
			labels),
		gpuSimdActive: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gpu_prof_simd_utilization",
			Help: "Fraction of time the SIMDs are being utilized [0,1]",
		},
			labels),
		gpuPcieRx: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "pcie_rx",
			Help: "Accumulated bytes received from the PCIe link",
		},
			labels),
		gpuPcieTx: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "pcie_tx",
			Help: "Accumulated bytes transmitted to the PCIe link",
		},
			labels),
		gpuPcieBidirBandwidth: *prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "pcie_bidirectional_bandwidth",
			Help: "Accumulated bandwidth on PCIe link in GB/sec",
		},
			labels),
	}
	ga.initFieldMetricsMap()

}

func (ga *GPUAgentClient) initFieldRegistration() error {
	for field, enabled := range exportFieldMap {
		if !enabled {
			continue
		}
		prommetric, ok := fieldMetricsMap[field]
		if !ok {
			logger.Log.Printf("invalid field found ignore %v", field)
			continue
		}
		if err := ga.mh.RegisterMetric(prommetric.Metric); err != nil {
			logger.Log.Printf("Field %v registration failed with err : %v", field, err)
		}
	}

	return nil
}

func (ga *GPUAgentClient) InitConfigs() error {
	filedConfigs := ga.mh.GetMetricsConfig()

	initPodExtraLabels(filedConfigs)
	initCustomLabels(filedConfigs)
	ga.initLabelConfigs(filedConfigs)
	initFieldConfig(filedConfigs)
	ga.initProfilerMetrics(filedConfigs)
	initGPUSelectorConfig(filedConfigs)
	ga.initPrometheusMetrics()
	ga.initProfilerMetricsField()
	return ga.initFieldRegistration()
}

func getGPURenderId(gpu *amdgpu.GPU) string {
	if gpu != nil && gpu.Status != nil {
		return fmt.Sprintf("%v", gpu.Status.DRMRenderId)
	}
	return ""
}

func getGPUInstanceID(gpu *amdgpu.GPU) int {
	return int(gpu.Status.Index)
}

func getGPUNodeID(gpu *amdgpu.GPU) string {
	if gpu != nil && gpu.Status != nil {
		return fmt.Sprintf("%v", gpu.Status.NodeId)
	}
	return ""
}

func (ga *GPUAgentClient) UpdateStaticMetrics() error {
	// send the req to gpuclient
	resp, partitionMap, err := ga.getGPUs()
	if err != nil {
		return err
	}
	if resp != nil && resp.ApiStatus != 0 {
		logger.Log.Printf("resp status :%v", resp.ApiStatus)
		return fmt.Errorf("%v", resp.ApiStatus)
	}
	wls, err := ga.ListWorkloads()
	if err != nil {
		logger.Log.Printf("Error listing workloads: %v", err)
	}

	k8PodLabelsMap, err = ga.FetchPodLabelsForNode()
	nonGpuLabels := ga.populateLabelsFromGPU(nil, nil, nil)
	ga.m.gpuNodesTotal.With(nonGpuLabels).Set(float64(len(resp.Response)))
	// do this only once as the health monitoring thread will
	// update periodically. this is required only for first state
	// of the metrics pull response from prometheus
	newGPUState := ga.processEccErrorMetrics(resp.Response, wls)
	_ = ga.updateNewHealthState(newGPUState)
	for _, gpu := range resp.Response {
		ga.updateGPUInfoToMetrics(wls, gpu, partitionMap, nil)
	}
	return nil
}

func (ga *GPUAgentClient) UpdateMetricsStats() error {
	return ga.getMetricsAll()
}

func (ga *GPUAgentClient) QueryMetrics() (interface{}, error) {
	var resp *amdgpu.GPUGetResponse
	var err error
	resp, _, err = ga.getGPUs()
	if err != nil {
		logger.Log.Printf("querymetrics - get gpus returned error:%v", err)
		return nil, err
	}
	return resp, nil
}

func (ga *GPUAgentClient) GetDeviceType() globals.DeviceType {
	return globals.GPUDevice
}

// getWorkloadsString returns the list of workloads associated with the given GPU in following
// formats
// kubernetes job  - "pod:pod_name, namespace: pod_namespace,container: container_name"
// slurm job       - "id: job_id, user: job_user, partition: job_partition", cluster: job_cluster

func (ga *GPUAgentClient) getWorkloadsListString(wls map[string]scheduler.Workload, gpu *amdgpu.GPU) []string {
	associatedWorkloads := []string{}
	if gpu == nil || gpu.Status == nil {
		return nil
	}

	schedulerJobs := ga.getWorkloadInfo(wls, gpu)
	for _, wl := range schedulerJobs {
		if wl == nil {
			continue
		}
		associatedWorkloads = append(associatedWorkloads, wl.String())
	}
	return associatedWorkloads
}

func (ga *GPUAgentClient) getWorkloadInfo(wls map[string]scheduler.Workload, gpu *amdgpu.GPU) []*scheduler.Workload {
	associatedWorkloads := []*scheduler.Workload{}
	if gpu == nil || gpu.Status == nil {
		return nil
	}
	gpuId := fmt.Sprintf("%v", getGPUInstanceID(gpu))
	gpuRenderId := getGPURenderId(gpu)
	deviceName, _ := ga.fsysDeviceHandler.GetDeviceNameFromRenderID(gpuRenderId)
	// populate with workload info
	if gpu.Status.PCIeStatus != nil {
		if workload, ok := wls[strings.ToLower(gpu.Status.PCIeStatus.PCIeBusId)]; ok {
			associatedWorkloads = append(associatedWorkloads, &workload)
		}
	}
	if workload, ok := wls[deviceName]; ok {
		associatedWorkloads = append(associatedWorkloads, &workload)
	}
	// ignore errors as we always expect slurm deployment as default
	if workload, ok := wls[gpuRenderId]; ok {
		associatedWorkloads = append(associatedWorkloads, &workload)
	}

	if workload, ok := wls[gpuId]; ok {
		associatedWorkloads = append(associatedWorkloads, &workload)
	}
	return associatedWorkloads
}

func (ga *GPUAgentClient) populateLabelsFromGPU(
	wls map[string]scheduler.Workload,
	gpu *amdgpu.GPU,
	partitionMap map[string]*amdgpu.GPU) map[string]string {
	var podInfo scheduler.PodResourceInfo
	var jobInfo scheduler.JobInfo

	if jobInfos := ga.getWorkloadInfo(wls, gpu); jobInfos != nil {
		for _, wl := range jobInfos {
			if wl == nil {
				continue
			}
			switch wl.Type {
			case scheduler.Kubernetes:
				podInfo = wl.Info.(scheduler.PodResourceInfo)
			case scheduler.Slurm:
				jobInfo = wl.Info.(scheduler.JobInfo)
			}
		}
	}

	labels := make(map[string]string)
	var parentPartition *amdgpu.GPU

	if partitionMap != nil && gpu != nil && gpu.Status.PCIeStatus != nil {
		gpuPcieAddr := strings.ToLower(gpu.Status.PCIeStatus.PCIeBusId)
		pcieBaseAddr := utils.GetPCIeBaseAddress(gpuPcieAddr)
		if p, ok := partitionMap[pcieBaseAddr]; ok {
			parentPartition = p
		}
	}

	for ckey, enabled := range exportLabels {
		if !enabled {
			continue
		}
		key := strings.ToLower(ckey)
		switch ckey {
		case exportermetrics.GPUMetricLabel_GPU_UUID.String():
			if gpu != nil {
				guuid, _ := uuid.FromBytes(gpu.Spec.Id)
				labels[key] = guuid.String()
			}
		case exportermetrics.GPUMetricLabel_GPU_ID.String():
			if gpu != nil {
				labels[key] = fmt.Sprintf("%v", getGPUInstanceID(gpu))
			}
		case exportermetrics.MetricLabel_CARD_SERIES.String():
			if gpu != nil {
				if parentPartition != nil {
					labels[key] = parentPartition.Status.CardSeries
				} else {
					labels[key] = gpu.Status.CardSeries
				}
			}
		case exportermetrics.MetricLabel_CARD_MODEL.String():
			if gpu != nil {
				if parentPartition != nil {
					labels[key] = parentPartition.Status.CardModel
				} else {
					labels[key] = gpu.Status.CardModel
				}
			}
		case exportermetrics.MetricLabel_CARD_VENDOR.String():
			if gpu != nil {
				if parentPartition != nil {
					labels[key] = parentPartition.Status.CardVendor
				} else {
					labels[key] = gpu.Status.CardVendor
				}
			}
		case exportermetrics.MetricLabel_DRIVER_VERSION.String():
			if gpu != nil {
				labels[key] = gpu.Status.DriverVersion
			}
		case exportermetrics.MetricLabel_VBIOS_VERSION.String():
			if gpu != nil {
				labels[key] = gpu.Status.VBIOSVersion
			}
		case exportermetrics.MetricLabel_POD.String():
			if gpu != nil {
				labels[key] = podInfo.Pod
			}
		case exportermetrics.MetricLabel_NAMESPACE.String():
			if gpu != nil {
				labels[key] = podInfo.Namespace
			}
		case exportermetrics.MetricLabel_CONTAINER.String():
			if gpu != nil {
				labels[key] = podInfo.Container
			}
		case exportermetrics.MetricLabel_JOB_ID.String():
			if gpu != nil {
				labels[key] = jobInfo.Id
			}
		case exportermetrics.MetricLabel_JOB_USER.String():
			if gpu != nil {
				labels[key] = jobInfo.User
			}
		case exportermetrics.MetricLabel_JOB_PARTITION.String():
			if gpu != nil {
				labels[key] = jobInfo.Partition
			}
		case exportermetrics.MetricLabel_CLUSTER_NAME.String():
			if gpu != nil {
				labels[key] = jobInfo.Cluster
			}
		case exportermetrics.MetricLabel_SERIAL_NUMBER.String():
			if gpu != nil {
				if parentPartition != nil {
					labels[key] = parentPartition.Status.SerialNum
				} else {
					labels[key] = gpu.Status.SerialNum
				}
			}
		case exportermetrics.MetricLabel_HOSTNAME.String():
			labels[key] = ga.staticHostLabels[exportermetrics.MetricLabel_HOSTNAME.String()]
		case exportermetrics.GPUMetricLabel_GPU_PARTITION_ID.String():
			if gpu != nil {
				if gpu.Status.PartitionId == math.MaxUint32 {
					labels[key] = "NA"
				} else {
					labels[key] = fmt.Sprintf("%v", gpu.Status.PartitionId)
				}
			}
		case exportermetrics.GPUMetricLabel_GPU_COMPUTE_PARTITION_TYPE.String():
			if gpu != nil {
				partitionType := gpu.Spec.ComputePartitionType
				if parentPartition != nil {
					partitionType = parentPartition.Spec.ComputePartitionType
				}
				trimmedValue := strings.TrimPrefix(partitionType.String(), "GPU_COMPUTE_PARTITION_TYPE_")
				labels[key] = strings.ToLower(trimmedValue)
			}
		case exportermetrics.GPUMetricLabel_GPU_MEMORY_PARTITION_TYPE.String():
			if gpu != nil {
				partitionType := gpu.Spec.MemoryPartitionType
				if parentPartition != nil {
					partitionType = parentPartition.Spec.MemoryPartitionType
				}
				trimmedValue := strings.TrimPrefix(partitionType.String(), "GPU_MEMORY_PARTITION_TYPE_")
				labels[key] = strings.ToLower(trimmedValue)
			}
		case exportermetrics.GPUMetricLabel_KFD_PROCESS_ID.String():
			if gpu != nil {
				if len(gpu.Status.KFDProcessId) > 0 {
					sort.SliceStable(gpu.Status.KFDProcessId, func(i, j int) bool {
						return gpu.Status.KFDProcessId[i] < gpu.Status.KFDProcessId[j]
					})
					processIds := make([]string, len(gpu.Status.KFDProcessId))
					for i, pid := range gpu.Status.KFDProcessId {
						processIds[i] = fmt.Sprintf("%v", pid)
					}
					labels[key] = strings.Join(processIds, ",")
				} else {
					labels[key] = ""
				}
			}
		case exportermetrics.GPUMetricLabel_DEPLOYMENT_MODE.String():
			if gpu != nil {
				virtualizationMode := gpu.Status.VirtualizationMode
				trimmedValue := strings.TrimPrefix(virtualizationMode.String(), "GPU_VIRTUALIZATION_MODE_")
				labels[key] = utils.VirtualizationModeToDeploymentMode(trimmedValue)
			}
		default:
			logger.Log.Printf("Invalid label is ignored %v", key)
		}
	}

	// Add extra pod labels only if config has mapped any
	if gpu != nil && len(extraPodLabelsMap) > 0 {
		podLabels := utils.GetPodLabels(&podInfo, k8PodLabelsMap)
		for prometheusPodlabel, k8Podlabel := range extraPodLabelsMap {
			label := strings.ToLower(prometheusPodlabel)
			labels[label] = podLabels[k8Podlabel]
		}
	}

	// Add custom labels
	for label, value := range customLabelMap {
		labels[label] = value
	}
	return labels
}

func (ga *GPUAgentClient) exporterEnabledGPU(instance int) bool {
	if gpuSelectorMap == nil {
		return true
	}
	_, enabled := gpuSelectorMap[instance]
	return enabled

}

func (ga *GPUAgentClient) updateGPUInfoToMetrics(
	wls map[string]scheduler.Workload,
	gpu *amdgpu.GPU,
	partitionMap map[string]*amdgpu.GPU,
	profMetrics map[string]float64,
) {
	if !ga.exporterEnabledGPU(getGPUInstanceID(gpu)) {
		return
	}

	labels := ga.populateLabelsFromGPU(wls, gpu, partitionMap)
	labelsWithIndex := ga.populateLabelsFromGPU(wls, gpu, partitionMap)
	status := gpu.Status
	stats := gpu.Stats

	ga.fl.logWithValidateAndExport(ga.m.gpuPackagePower, exportermetrics.GPUMetricField_GPU_PACKAGE_POWER.String(),
		labels, stats.PackagePower)
	ga.fl.logWithValidateAndExport(ga.m.gpuAvgPkgPower, exportermetrics.GPUMetricField_GPU_AVERAGE_PACKAGE_POWER.String(),
		labels, stats.AvgPackagePower)

	// export health state only if available
	gpuid := fmt.Sprintf("%v", getGPUInstanceID(gpu))
	if hstate, ok := ga.healthState[gpuid]; ok {
		if hstate.Health == strings.ToLower(metricssvc.GPUHealth_HEALTHY.String()) {
			ga.m.gpuHealth.With(labels).Set(1)
		} else {
			ga.m.gpuHealth.With(labels).Set(0)
		}
	}

	// gpu temp stats
	tempStats := stats.Temperature
	if tempStats != nil {
		ga.fl.logWithValidateAndExport(ga.m.gpuEdgeTemp, exportermetrics.GPUMetricField_GPU_EDGE_TEMPERATURE.String(),
			labels, tempStats.EdgeTemperature)
		ga.fl.logWithValidateAndExport(ga.m.gpuJunctionTemp, exportermetrics.GPUMetricField_GPU_JUNCTION_TEMPERATURE.String(),
			labels, tempStats.JunctionTemperature)
		ga.fl.logWithValidateAndExport(ga.m.gpuMemoryTemp, exportermetrics.GPUMetricField_GPU_MEMORY_TEMPERATURE.String(),
			labels, tempStats.MemoryTemperature)
		for j, temp := range tempStats.HBMTemperature {
			labelsWithIndex["hbm_index"] = fmt.Sprintf("%v", j)
			if utils.IsValueApplicable(temp) {
				ga.m.gpuHBMTemp.With(labelsWithIndex).Set(float64(temp))
			}
		}
		delete(labelsWithIndex, "hbm_index")
	}

	// gpu usage
	gpuUsage := stats.Usage
	if gpuUsage != nil {
		ga.fl.logWithValidateAndExport(ga.m.gpuGFXActivity, exportermetrics.GPUMetricField_GPU_GFX_ACTIVITY.String(),
			labels, gpuUsage.GFXActivity)
		ga.fl.logWithValidateAndExport(ga.m.gpuUMCActivity, exportermetrics.GPUMetricField_GPU_UMC_ACTIVITY.String(),
			labels, gpuUsage.UMCActivity)
		ga.fl.logWithValidateAndExport(ga.m.gpuMMAActivity, exportermetrics.GPUMetricField_GPU_MMA_ACTIVITY.String(),
			labels, gpuUsage.MMActivity)
		for j, act := range gpuUsage.VCNActivity {
			labelsWithIndex["vcn_index"] = fmt.Sprintf("%v", j)
			if utils.IsValueApplicable(act) {
				ga.m.gpuVCNActivity.With(labelsWithIndex).Set(float64(act))
			}
		}
		delete(labelsWithIndex, "vcn_index")
		for j, act := range gpuUsage.JPEGActivity {
			labelsWithIndex["jpeg_index"] = fmt.Sprintf("%v", j)
			if utils.IsValueApplicable(act) {
				ga.m.gpuJPEGActivity.With(labelsWithIndex).Set(float64(act))
			}
		}
		delete(labelsWithIndex, "jpeg_index")
		for j, act := range gpuUsage.GFXBusyInst {
			labelsWithIndex["xcc_index"] = fmt.Sprintf("%v", j)
			if utils.IsValueApplicable(act) {
				ga.m.gpuGfxBusyInst.With(labelsWithIndex).Set(float64(act))
			}
		}
		for j, act := range gpuUsage.VCNBusyInst {
			labelsWithIndex["xcc_index"] = fmt.Sprintf("%v", j)
			if utils.IsValueApplicable(act) {
				ga.m.gpuVcnBusyInst.With(labelsWithIndex).Set(float64(act))
			}
		}
		for j, act := range gpuUsage.JPEGBusyInst {
			labelsWithIndex["xcc_index"] = fmt.Sprintf("%v", j)
			if utils.IsValueApplicable(act) {
				ga.m.gpuJpegBusyInst.With(labelsWithIndex).Set(float64(act))
			}
		}
		delete(labelsWithIndex, "xcc_index")
	}

	volt := stats.Voltage
	if volt != nil {
		ga.fl.logWithValidateAndExport(ga.m.gpuVoltage, exportermetrics.GPUMetricField_GPU_VOLTAGE.String(),
			labels, volt.Voltage)
		ga.fl.logWithValidateAndExport(ga.m.gpuGFXVoltage, exportermetrics.GPUMetricField_GPU_GFX_VOLTAGE.String(),
			labels, volt.GFXVoltage)
		ga.fl.logWithValidateAndExport(ga.m.gpuMemVoltage, exportermetrics.GPUMetricField_GPU_MEMORY_VOLTAGE.String(),
			labels, volt.MemoryVoltage)
	}

	// pcie status
	pcieStatus := status.PCIeStatus
	if pcieStatus != nil {
		ga.fl.logWithValidateAndExport(ga.m.gpuPCIeSpeed, exportermetrics.GPUMetricField_PCIE_SPEED.String(),
			labels, pcieStatus.Speed)
		ga.fl.logWithValidateAndExport(ga.m.gpuPCIeMaxSpeed, exportermetrics.GPUMetricField_PCIE_MAX_SPEED.String(),
			labels, pcieStatus.MaxSpeed)
		ga.fl.logWithValidateAndExport(ga.m.gpuPCIeBandwidth, exportermetrics.GPUMetricField_PCIE_BANDWIDTH.String(),
			labels, pcieStatus.Bandwidth)
	}

	// pcie stats
	pcieStats := stats.PCIeStats
	if pcieStats != nil {
		ga.fl.logWithValidateAndExport(ga.m.gpuPCIeReplayCount, exportermetrics.GPUMetricField_PCIE_REPLAY_COUNT.String(),
			labels, pcieStats.ReplayCount)
		ga.fl.logWithValidateAndExport(ga.m.gpuPCIeRecoveryCount, exportermetrics.GPUMetricField_PCIE_RECOVERY_COUNT.String(),
			labels, pcieStats.RecoveryCount)
		ga.fl.logWithValidateAndExport(ga.m.gpuPCIeReplayRolloverCount, exportermetrics.GPUMetricField_PCIE_REPLAY_ROLLOVER_COUNT.String(),
			labels, pcieStats.ReplayRolloverCount)
		ga.fl.logWithValidateAndExport(ga.m.gpuPCIeNACKSentCount, exportermetrics.GPUMetricField_PCIE_NACK_SENT_COUNT.String(),
			labels, pcieStats.NACKSentCount)
		ga.fl.logWithValidateAndExport(ga.m.gpuPCIeNACKReceivedCount, exportermetrics.GPUMetricField_PCIE_NACK_RECEIVED_COUNT.String(),
			labels, pcieStats.NACKReceivedCount)
		ga.fl.logWithValidateAndExport(ga.m.gpuPcieRx, exportermetrics.GPUMetricField_PCIE_RX.String(),
			labels, pcieStats.RxBytes)
		ga.fl.logWithValidateAndExport(ga.m.gpuPcieTx, exportermetrics.GPUMetricField_PCIE_TX.String(),
			labels, pcieStats.TxBytes)
		ga.fl.logWithValidateAndExport(ga.m.gpuPcieBidirBandwidth, exportermetrics.GPUMetricField_PCIE_BIDIRECTIONAL_BANDWIDTH.String(),
			labels, pcieStats.BiDirBandwidth)
	}

	ga.fl.logWithValidateAndExport(ga.m.gpuEnergyConsumed, exportermetrics.GPUMetricField_GPU_ENERGY_CONSUMED.String(),
		labels, stats.EnergyConsumed)

	// clock status
	clockStatus := status.ClockStatus
	if clockStatus != nil {
		for j, clock := range clockStatus {
			labelsWithIndex["clock_index"] = fmt.Sprintf("%v", j)
			labelsWithIndex["clock_type"] = fmt.Sprintf("%v", clock.Type.String())
			if utils.IsValueApplicable(clock.Frequency) {
				ga.m.gpuClock.With(labelsWithIndex).Set(float64(clock.Frequency))
			}
		}
		delete(labelsWithIndex, "clock_index")
		delete(labelsWithIndex, "clock_type")
	}

	ga.fl.logWithValidateAndExport(ga.m.gpuPowerUsage, exportermetrics.GPUMetricField_GPU_POWER_USAGE.String(), labels, stats.PowerUsage)

	ga.fl.logWithValidateAndExport(ga.m.gpuEccCorrectTotal, exportermetrics.GPUMetricField_GPU_ECC_CORRECT_TOTAL.String(),
		labels, stats.TotalCorrectableErrors)
	ga.fl.logWithValidateAndExport(ga.m.gpuEccUncorrectTotal, exportermetrics.GPUMetricField_GPU_ECC_UNCORRECT_TOTAL.String(),
		labels, stats.TotalUncorrectableErrors)
	ga.fl.logWithValidateAndExport(ga.m.gpuEccCorrectSDMA, exportermetrics.GPUMetricField_GPU_ECC_CORRECT_SDMA.String(),
		labels, stats.SDMACorrectableErrors)
	ga.fl.logWithValidateAndExport(ga.m.gpuEccUncorrectSDMA, exportermetrics.GPUMetricField_GPU_ECC_UNCORRECT_SDMA.String(),
		labels, stats.SDMAUncorrectableErrors)
	ga.fl.logWithValidateAndExport(ga.m.gpuEccCorrectGFX, exportermetrics.GPUMetricField_GPU_ECC_CORRECT_GFX.String(),
		labels, stats.GFXCorrectableErrors)
	ga.fl.logWithValidateAndExport(ga.m.gpuEccUncorrectGFX, exportermetrics.GPUMetricField_GPU_ECC_UNCORRECT_GFX.String(),
		labels, stats.GFXUncorrectableErrors)
	ga.fl.logWithValidateAndExport(ga.m.gpuEccCorrectMMHUB, exportermetrics.GPUMetricField_GPU_ECC_CORRECT_MMHUB.String(),
		labels, stats.MMHUBCorrectableErrors)
	ga.fl.logWithValidateAndExport(ga.m.gpuEccUncorrectMMHUB, exportermetrics.GPUMetricField_GPU_ECC_UNCORRECT_MMHUB.String(),
		labels, stats.MMHUBUncorrectableErrors)
	ga.fl.logWithValidateAndExport(ga.m.gpuEccCorrectATHUB, exportermetrics.GPUMetricField_GPU_ECC_CORRECT_ATHUB.String(),
		labels, stats.ATHUBCorrectableErrors)
	ga.fl.logWithValidateAndExport(ga.m.gpuEccUncorrectATHUB, exportermetrics.GPUMetricField_GPU_ECC_UNCORRECT_ATHUB.String(),
		labels, stats.ATHUBUncorrectableErrors)

	ga.fl.logWithValidateAndExport(ga.m.gpuEccCorrectBIF, exportermetrics.GPUMetricField_GPU_ECC_CORRECT_BIF.String(),
		labels, stats.BIFCorrectableErrors)
	ga.fl.logWithValidateAndExport(ga.m.gpuEccUncorrectBIF, exportermetrics.GPUMetricField_GPU_ECC_UNCORRECT_BIF.String(),
		labels, stats.BIFUncorrectableErrors)
	ga.fl.logWithValidateAndExport(ga.m.gpuEccCorrectHDP, exportermetrics.GPUMetricField_GPU_ECC_CORRECT_HDP.String(),
		labels, stats.HDPCorrectableErrors)
	ga.fl.logWithValidateAndExport(ga.m.gpuEccUncorrectHDP, exportermetrics.GPUMetricField_GPU_ECC_UNCORRECT_HDP.String(),
		labels, stats.HDPUncorrectableErrors)
	ga.fl.logWithValidateAndExport(ga.m.gpuEccCorrectXgmiWAFL, exportermetrics.GPUMetricField_GPU_ECC_CORRECT_XGMI_WAFL.String(),
		labels, stats.XGMIWAFLCorrectableErrors)
	ga.fl.logWithValidateAndExport(ga.m.gpuEccUncorrectXgmiWAFL, exportermetrics.GPUMetricField_GPU_ECC_UNCORRECT_XGMI_WAFL.String(),
		labels, stats.XGMIWAFLUncorrectableErrors)
	ga.fl.logWithValidateAndExport(ga.m.gpuEccCorrectDF, exportermetrics.GPUMetricField_GPU_ECC_CORRECT_DF.String(),
		labels, stats.DFCorrectableErrors)
	ga.fl.logWithValidateAndExport(ga.m.gpuEccUncorrectDF, exportermetrics.GPUMetricField_GPU_ECC_UNCORRECT_DF.String(),
		labels, stats.DFUncorrectableErrors)
	ga.fl.logWithValidateAndExport(ga.m.gpuEccCorrectSMN, exportermetrics.GPUMetricField_GPU_ECC_CORRECT_SMN.String(),
		labels, stats.SMNCorrectableErrors)
	ga.fl.logWithValidateAndExport(ga.m.gpuEccUncorrectSMN, exportermetrics.GPUMetricField_GPU_ECC_UNCORRECT_SMN.String(),
		labels, stats.SMNUncorrectableErrors)
	ga.fl.logWithValidateAndExport(ga.m.gpuEccCorrectSEM, exportermetrics.GPUMetricField_GPU_ECC_CORRECT_SEM.String(),
		labels, stats.SEMCorrectableErrors)
	ga.fl.logWithValidateAndExport(ga.m.gpuEccUncorrectSEM, exportermetrics.GPUMetricField_GPU_ECC_UNCORRECT_SEM.String(),
		labels, stats.SEMUncorrectableErrors)

	ga.fl.logWithValidateAndExport(ga.m.gpuEccCorrectMP0, exportermetrics.GPUMetricField_GPU_ECC_CORRECT_MP0.String(),
		labels, stats.MP0CorrectableErrors)
	ga.fl.logWithValidateAndExport(ga.m.gpuEccUncorrectMP0, exportermetrics.GPUMetricField_GPU_ECC_UNCORRECT_MP0.String(),
		labels, stats.MP0UncorrectableErrors)
	ga.fl.logWithValidateAndExport(ga.m.gpuEccCorrectMP1, exportermetrics.GPUMetricField_GPU_ECC_CORRECT_MP1.String(),
		labels, stats.MP1CorrectableErrors)
	ga.fl.logWithValidateAndExport(ga.m.gpuEccUncorrectMP1, exportermetrics.GPUMetricField_GPU_ECC_UNCORRECT_MP1.String(),
		labels, stats.MP1UncorrectableErrors)
	ga.fl.logWithValidateAndExport(ga.m.gpuEccCorrectFUSE, exportermetrics.GPUMetricField_GPU_ECC_CORRECT_FUSE.String(),
		labels, stats.FUSECorrectableErrors)
	ga.fl.logWithValidateAndExport(ga.m.gpuEccUncorrectFUSE, exportermetrics.GPUMetricField_GPU_ECC_UNCORRECT_FUSE.String(),
		labels, stats.FUSEUncorrectableErrors)
	ga.fl.logWithValidateAndExport(ga.m.gpuEccCorrectUMC, exportermetrics.GPUMetricField_GPU_ECC_CORRECT_UMC.String(),
		labels, stats.UMCCorrectableErrors)
	ga.fl.logWithValidateAndExport(ga.m.gpuEccUncorrectUMC, exportermetrics.GPUMetricField_GPU_ECC_UNCORRECT_UMC.String(),
		labels, stats.UMCUncorrectableErrors)

	ga.fl.logWithValidateAndExport(ga.m.gpuEccCorrectMCA, exportermetrics.GPUMetricField_GPU_ECC_CORRECT_MCA.String(),
		labels, stats.MCACorrectableErrors)
	ga.fl.logWithValidateAndExport(ga.m.gpuEccUncorrectMCA, exportermetrics.GPUMetricField_GPU_ECC_UNCORRECT_MCA.String(),
		labels, stats.MCAUncorrectableErrors)

	ga.fl.logWithValidateAndExport(ga.m.gpuEccCorrectVCN, exportermetrics.GPUMetricField_GPU_ECC_CORRECT_VCN.String(),
		labels, stats.VCNCorrectableErrors)
	ga.fl.logWithValidateAndExport(ga.m.gpuEccUncorrectVCN, exportermetrics.GPUMetricField_GPU_ECC_UNCORRECT_VCN.String(),
		labels, stats.VCNUncorrectableErrors)

	ga.fl.logWithValidateAndExport(ga.m.gpuEccCorrectJPEG, exportermetrics.GPUMetricField_GPU_ECC_CORRECT_JPEG.String(),
		labels, stats.JPEGCorrectableErrors)
	ga.fl.logWithValidateAndExport(ga.m.gpuEccUncorrectJPEG, exportermetrics.GPUMetricField_GPU_ECC_UNCORRECT_JPEG.String(),
		labels, stats.JPEGUncorrectableErrors)

	ga.fl.logWithValidateAndExport(ga.m.gpuEccCorrectIH, exportermetrics.GPUMetricField_GPU_ECC_CORRECT_IH.String(),
		labels, stats.IHCorrectableErrors)
	ga.fl.logWithValidateAndExport(ga.m.gpuEccUncorrectIH, exportermetrics.GPUMetricField_GPU_ECC_UNCORRECT_IH.String(),
		labels, stats.IHUncorrectableErrors)

	ga.fl.logWithValidateAndExport(ga.m.gpuEccCorrectMPIO, exportermetrics.GPUMetricField_GPU_ECC_CORRECT_MPIO.String(),
		labels, stats.MPIOCorrectableErrors)
	ga.fl.logWithValidateAndExport(ga.m.gpuEccUncorrectMPIO, exportermetrics.GPUMetricField_GPU_ECC_UNCORRECT_MPIO.String(),
		labels, stats.MPIOUncorrectableErrors)

	ga.fl.logWithValidateAndExport(ga.m.xgmiNbrNopTx0, exportermetrics.GPUMetricField_GPU_XGMI_NBR_0_NOP_TX.String(),
		labels, stats.XGMINeighbor0TxNOPs)
	ga.fl.logWithValidateAndExport(ga.m.xgmiNbrReqTx0, exportermetrics.GPUMetricField_GPU_XGMI_NBR_0_REQ_TX.String(),
		labels, stats.XGMINeighbor0TxRequests)
	ga.fl.logWithValidateAndExport(ga.m.xgmiNbrRespTx0, exportermetrics.GPUMetricField_GPU_XGMI_NBR_0_RESP_TX.String(),
		labels, stats.XGMINeighbor0TxResponses)
	ga.fl.logWithValidateAndExport(ga.m.xgmiNbrBeatsTx0, exportermetrics.GPUMetricField_GPU_XGMI_NBR_0_BEATS_TX.String(),
		labels, stats.XGMINeighbor0TXBeats)

	ga.fl.logWithValidateAndExport(ga.m.xgmiNbrNopTx1, exportermetrics.GPUMetricField_GPU_XGMI_NBR_1_NOP_TX.String(),
		labels, stats.XGMINeighbor1TxNOPs)
	ga.fl.logWithValidateAndExport(ga.m.xgmiNbrReqTx1, exportermetrics.GPUMetricField_GPU_XGMI_NBR_1_REQ_TX.String(),
		labels, stats.XGMINeighbor1TxRequests)
	ga.fl.logWithValidateAndExport(ga.m.xgmiNbrRespTx1, exportermetrics.GPUMetricField_GPU_XGMI_NBR_1_RESP_TX.String(),
		labels, stats.XGMINeighbor1TxResponses)
	ga.fl.logWithValidateAndExport(ga.m.xgmiNbrBeatsTx1, exportermetrics.GPUMetricField_GPU_XGMI_NBR_1_BEATS_TX.String(),
		labels, stats.XGMINeighbor1TXBeats)

	ga.fl.logWithValidateAndExport(ga.m.xgmiNbrTxTput0, exportermetrics.GPUMetricField_GPU_XGMI_NBR_0_TX_THRPUT.String(),
		labels, stats.XGMINeighbor0TxThroughput)
	ga.fl.logWithValidateAndExport(ga.m.xgmiNbrTxTput1, exportermetrics.GPUMetricField_GPU_XGMI_NBR_1_TX_THRPUT.String(),
		labels, stats.XGMINeighbor1TxThroughput)
	ga.fl.logWithValidateAndExport(ga.m.xgmiNbrTxTput2, exportermetrics.GPUMetricField_GPU_XGMI_NBR_2_TX_THRPUT.String(),
		labels, stats.XGMINeighbor2TxThroughput)
	ga.fl.logWithValidateAndExport(ga.m.xgmiNbrTxTput3, exportermetrics.GPUMetricField_GPU_XGMI_NBR_3_TX_THRPUT.String(),
		labels, stats.XGMINeighbor3TxThroughput)
	ga.fl.logWithValidateAndExport(ga.m.xgmiNbrTxTput4, exportermetrics.GPUMetricField_GPU_XGMI_NBR_4_TX_THRPUT.String(),
		labels, stats.XGMINeighbor4TxThroughput)
	ga.fl.logWithValidateAndExport(ga.m.xgmiNbrTxTput5, exportermetrics.GPUMetricField_GPU_XGMI_NBR_5_TX_THRPUT.String(),
		labels, stats.XGMINeighbor5TxThroughput)

	vramUsage := stats.VRAMUsage
	vramStatus := status.GetVRAMStatus()
	var totalVRAM, usedVRAM, freeVRAM float64
	if vramUsage != nil {
		ga.fl.logWithValidateAndExport(ga.m.gpuTotalVisibleVram, exportermetrics.GPUMetricField_GPU_TOTAL_VISIBLE_VRAM.String(),
			labels, vramUsage.TotalVisibleVRAM)
		ga.fl.logWithValidateAndExport(ga.m.gpuUsedVisibleVram, exportermetrics.GPUMetricField_GPU_USED_VISIBLE_VRAM.String(),
			labels, vramUsage.UsedVisibleVRAM)
		ga.fl.logWithValidateAndExport(ga.m.gpuFreeVisibleVram, exportermetrics.GPUMetricField_GPU_FREE_VISIBLE_VRAM.String(),
			labels, vramUsage.FreeVisibleVRAM)

		ga.fl.logWithValidateAndExport(ga.m.gpuTotalGTT, exportermetrics.GPUMetricField_GPU_TOTAL_GTT.String(),
			labels, vramUsage.TotalGTT)
		ga.fl.logWithValidateAndExport(ga.m.gpuUsedGTT, exportermetrics.GPUMetricField_GPU_USED_GTT.String(),
			labels, vramUsage.UsedGTT)
		ga.fl.logWithValidateAndExport(ga.m.gpuFreeGTT, exportermetrics.GPUMetricField_GPU_FREE_GTT.String(),
			labels, vramUsage.FreeGTT)
	}
	if vramStatus != nil {
		totalVRAM = utils.NormalizeUint64(vramStatus.Size)
	}
	if vramUsage != nil {
		usedVRAM = utils.NormalizeUint64(vramUsage.UsedVRAM)
	}
	freeVRAM = totalVRAM - usedVRAM
	if totalVRAM != 0 {
		ga.fl.logWithValidateAndExport(ga.m.gpuTotalVram, exportermetrics.GPUMetricField_GPU_TOTAL_VRAM.String(), labels, totalVRAM)
		ga.fl.logWithValidateAndExport(ga.m.gpuUsedVram, exportermetrics.GPUMetricField_GPU_USED_VRAM.String(), labels, usedVRAM)
		ga.fl.logWithValidateAndExport(ga.m.gpuFreeVram, exportermetrics.GPUMetricField_GPU_FREE_VRAM.String(), labels, freeVRAM)
	}
	xgmiStats := stats.XGMILinkStats
	if xgmiStats != nil {
		for j, linkStat := range xgmiStats {
			labelsWithIndex["link_index"] = fmt.Sprintf("%v", j)
			if utils.IsValueApplicable(linkStat.DataRead) {
				ga.m.gpuXgmiLinkStatsRx.With(labelsWithIndex).Set(float64(linkStat.DataRead))
			}
			if utils.IsValueApplicable(linkStat.DataWrite) {
				ga.m.gpuXgmiLinkStatsTx.With(labelsWithIndex).Set(float64(linkStat.DataWrite))
			}
		}
		delete(labelsWithIndex, "link_index")
	}
	violationStats := stats.ViolationStats
	if violationStats != nil {
		ga.fl.logWithValidateAndExport(ga.m.gpuCurrAccCtr, exportermetrics.GPUMetricField_GPU_VIOLATION_CURRENT_ACCUMULATED_COUNTER.String(),
			labels, violationStats.CurrentAccumulatedCounter)
		ga.fl.logWithValidateAndExport(ga.m.gpuProcHRA, exportermetrics.GPUMetricField_GPU_VIOLATION_PROCESSOR_HOT_RESIDENCY_ACCUMULATED.String(),
			labels, violationStats.ProcessorHotResidencyAccumulated)
		ga.fl.logWithValidateAndExport(ga.m.gpuPPTRA, exportermetrics.GPUMetricField_GPU_VIOLATION_PPT_RESIDENCY_ACCUMULATED.String(),
			labels, violationStats.PPTResidencyAccumulated)
		ga.fl.logWithValidateAndExport(ga.m.gpuSTRA, exportermetrics.GPUMetricField_GPU_VIOLATION_SOCKET_THERMAL_RESIDENCY_ACCUMULATED.String(),
			labels, violationStats.SocketThermalResidencyAccumulated)
		ga.fl.logWithValidateAndExport(ga.m.gpuVRTRA, exportermetrics.GPUMetricField_GPU_VIOLATION_VR_THERMAL_RESIDENCY_ACCUMULATED.String(),
			labels, violationStats.VRThermalResidencyAccumulated)
		ga.fl.logWithValidateAndExport(ga.m.gpuHBMTRA, exportermetrics.GPUMetricField_GPU_VIOLATION_HBM_THERMAL_RESIDENCY_ACCUMULATED.String(),
			labels, violationStats.HBMThermalResidencyAccumulated)
	}

	// populate prof metrics if available
	if profMetrics == nil {
		return
	}
	// map list entry to exporter fields rocprofilerclient/rocpclient.cpp
	// (all_fields)
	// case sensitive
	for mkey, value := range profMetrics {
		switch mkey {
		case "GRBM_GUI_ACTIVE":
			ga.m.gpuGrbmGuiActivity.With(labels).Set(value)
			ga.m.gpuOccElapsed.With(labels).Set(value)
		case "SQ_WAVES":
			ga.m.gpuSqWaves.With(labels).Set(value)
		case "GRBM_COUNT":
			ga.m.gpuGrbmCount.With(labels).Set(value)
		case "GPU_UTIL":
			ga.m.gpuGPUUtil.With(labels).Set(value)
		case "FETCH_SIZE":
			ga.m.gpuFetchSize.With(labels).Set(value)
		case "WRITE_SIZE":
			ga.m.gpuWriteSize.With(labels).Set(value)
		case "TOTAL_16_OPS":
			ga.m.gpuTotal16Ops.With(labels).Set(value)
		case "TOTAL_32_OPS":
			ga.m.gpuTotal32Ops.With(labels).Set(value)
		case "TOTAL_64_OPS":
			ga.m.gpuTotal64Ops.With(labels).Set(value)
		case "CPC_CPC_STAT_BUSY":
			ga.m.gpuCpcStatBusy.With(labels).Set(value)
		case "CPC_CPC_STAT_IDLE":
			ga.m.gpuCpcStatIdle.With(labels).Set(value)
		case "CPC_CPC_STAT_STALL":
			ga.m.gpuCpcStatStall.With(labels).Set(value)
		case "CPC_CPC_TCIU_BUSY":
			ga.m.gpuCpcTciuBusy.With(labels).Set(value)
		case "CPC_CPC_TCIU_IDLE":
			ga.m.gpuCpcTciuIdle.With(labels).Set(value)
		case "CPC_CPC_UTCL2IU_BUSY":
			ga.m.gpuCpcUtcl2iuBusy.With(labels).Set(value)
		case "CPC_CPC_UTCL2IU_IDLE":
			ga.m.gpuCpcUtcl2iuIdle.With(labels).Set(value)
		case "CPC_CPC_UTCL2IU_STALL":
			ga.m.gpuCpcUtcl2iuStall.With(labels).Set(value)
		case "CPC_ME1_BUSY_FOR_PACKET_DECODE":
			ga.m.gpuCpcME1BusyForPacketDecode.With(labels).Set(value)
		case "CPC_ME1_DC0_SPI_BUSY":
			ga.m.gpuCpcME1Dc0SpiBusy.With(labels).Set(value)
		case "CPC_UTCL1_STALL_ON_TRANSLATION":
			ga.m.gpuCpcUtcl1StallOnTranslation.With(labels).Set(value)
		case "CPC_ALWAYS_COUNT":
			ga.m.gpuCpcAlwaysCount.With(labels).Set(value)
		case "CPC_ADC_VALID_CHUNK_NOT_AVAIL":
			ga.m.gpuCpcAdcValidChunkNotAvail.With(labels).Set(value)
		case "CPC_ADC_DISPATCH_ALLOC_DONE":
			ga.m.gpuCpcAdcDispatchAllocDone.With(labels).Set(value)
		case "CPC_ADC_VALID_CHUNK_END":
			ga.m.gpuCpcAdcValidChunkEnd.With(labels).Set(value)
		case "CPC_SYNC_FIFO_FULL_LEVEL":
			ga.m.gpuCpcSynFifoFullLevel.With(labels).Set(value)
		case "CPC_SYNC_FIFO_FULL":
			ga.m.gpuCpcSynFifoFull.With(labels).Set(value)
		case "CPC_GD_BUSY":
			ga.m.gpuCpcGdBusy.With(labels).Set(value)
		case "CPC_TG_SEND":
			ga.m.gpuCpcTgSend.With(labels).Set(value)
		case "CPC_WALK_NEXT_CHUNK":
			ga.m.gpuCpcWalkNextChunk.With(labels).Set(value)
		case "CPC_STALLED_BY_SE0_SPI":
			ga.m.gpuCpcStalledBySe0Spi.With(labels).Set(value)
		case "CPC_STALLED_BY_SE1_SPI":
			ga.m.gpuCpcStalledBySe1Spi.With(labels).Set(value)
		case "CPC_STALLED_BY_SE2_SPI":
			ga.m.gpuCpcStalledBySe2Spi.With(labels).Set(value)
		case "CPC_STALLED_BY_SE3_SPI":
			ga.m.gpuCpcStalledBySe3Spi.With(labels).Set(value)
		case "CPC_LTE_ALL":
			ga.m.gpuCpcLteAll.With(labels).Set(value)
		case "CPC_SYNC_WRREQ_FIFO_BUSY":
			ga.m.gpuCpcSyncWrreqFifoBusy.With(labels).Set(value)
		case "CPC_CANE_BUSY":
			ga.m.gpuCpcCaneBusy.With(labels).Set(value)
		case "CPC_CANE_STALL":
			ga.m.gpuCpcCaneStall.With(labels).Set(value)
		case "CPF_CMP_UTCL1_STALL_ON_TRANSLATION":
			ga.m.gpuCpfCmpUtcl1StallOnTrnsalation.With(labels).Set(value)
		case "CPF_CPF_STAT_BUSY":
			ga.m.gpuCpfStatBusy.With(labels).Set(value)
		case "CPF_CPF_STAT_IDLE":
			ga.m.gpuCpfStatIdle.With(labels).Set(value)
		case "CPF_CPF_STAT_STALL":
			ga.m.gpuCpfStatStall.With(labels).Set(value)
		case "CPF_CPF_TCIU_BUSY":
			ga.m.gpuCpfStatTciuBusy.With(labels).Set(value)
		case "CPF_CPF_TCIU_IDLE":
			ga.m.gpuCpfStatTciuIdle.With(labels).Set(value)
		case "CPF_CPF_TCIU_STALL":
			ga.m.gpuCpfStatTciuStall.With(labels).Set(value)
		case "OccupancyPercent":
			ga.m.gpuOccPercent.With(labels).Set(value)
		case "MfmaUtil":
			ga.m.gpuTensorActivePercent.With(labels).Set(value)
		case "ValuPipeIssueUtil":
			ga.m.gpuValuPipeIssueUtil.With(labels).Set(value)
		case "VALUBusy":
			ga.m.gpuSMActive.With(labels).Set(value)
		case "MeanOccupancyPerActiveCU":
			ga.m.gpuOccPerActiveCU.With(labels).Set(value)
		case "MeanOccupancyPerCU":
			ga.m.gpuMeanOccPerCU.With(labels).Set(value)
		case "SIMD_UTILIZATION":
			ga.m.gpuSimdActive.With(labels).Set(value)
		}
	}
}

func (ga *GPUAgentClient) populateStaticHostLabels() error {
	ga.staticHostLabels = map[string]string{}
	hostname, err := utils.GetHostName()
	if err != nil {
		return err
	}
	logger.Log.Printf("hostame %v", hostname)
	ga.staticHostLabels[exportermetrics.MetricLabel_HOSTNAME.String()] = hostname
	return nil
}

func GetGPUAgentMandatoryLabels() []string {
	return mandatoryLables
}
