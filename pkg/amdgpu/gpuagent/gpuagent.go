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
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofrs/uuid"

	"github.com/ROCm/device-metrics-exporter/pkg/exporter/globals"
	"github.com/ROCm/device-metrics-exporter/pkg/exporter/scheduler"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/ROCm/device-metrics-exporter/pkg/amdgpu/fsysdevice"
	"github.com/ROCm/device-metrics-exporter/pkg/amdgpu/gen/amdgpu"
	"github.com/ROCm/device-metrics-exporter/pkg/amdgpu/rocprofiler"
	k8sclient "github.com/ROCm/device-metrics-exporter/pkg/client"
	"github.com/ROCm/device-metrics-exporter/pkg/exporter/gen/metricssvc"
	"github.com/ROCm/device-metrics-exporter/pkg/exporter/logger"
	"github.com/ROCm/device-metrics-exporter/pkg/exporter/metricsutil"
	"github.com/ROCm/device-metrics-exporter/pkg/exporter/utils"
)

const (
	// cachgpuid are updated after this many pull request
	refreshInterval = 30 * time.Second
	queryTimeout    = 15 * time.Second
	cacheTimer      = 15 * time.Second
)

type GPUAgentClient struct {
	sync.Mutex
	conn                   *grpc.ClientConn
	mh                     *metricsutil.MetricsHandler
	gpuclient              amdgpu.GPUSvcClient
	evtclient              amdgpu.EventSvcClient
	rocpclient             *rocprofiler.ROCProfilerClient
	m                      *metrics // client specific metrics
	k8sApiClient           *k8sclient.K8sClient
	k8sScheduler           scheduler.SchedulerClient
	slurmScheduler         scheduler.SchedulerClient
	isKubernetes           bool // pod resource client enabled or not
	enabledK8sApi          bool
	enableZmq              bool
	enableProfileMetrics   bool
	enableAfidMetrics      bool
	enableSriov            bool
	staticHostLabels       map[string]string
	ctx                    context.Context
	cancel                 context.CancelFunc
	healthState            map[string]*metricssvc.GPUState
	mockEccField           map[string]map[string]uint32 // gpuid->fields->count
	computeNodeHealthState bool
	fsysDeviceHandler      *fsysdevice.FsysDevice
	gCache                 *gpuCache
	nodeHealthLabellerCfg  *utils.NodeHealthLabellerConfig
	fl                     *fieldLogger
}

// Cache fields for GPUAgentClient
type gpuCache struct {
	sync.RWMutex
	lastResponse      *amdgpu.GPUGetResponse
	lastCperResponse  *amdgpu.GPUCPERGetResponse
	lastTimestamp     time.Time
	lastCperTimestamp time.Time
}

// cper cache entry
type cperCacheEntry struct {
	entry     *amdgpu.CPEREntry
	timestamp time.Time
}

// GPUAgentClientOptions set desired options
type GPUAgentClientOptions func(ga *GPUAgentClient)

func WithZmq(enableZmq bool) GPUAgentClientOptions {
	return func(ga *GPUAgentClient) {
		logger.Log.Printf("Zmq enable %v", enableZmq)
		ga.enableZmq = enableZmq
	}
}

func WithK8sClient(k8sclient *k8sclient.K8sClient) GPUAgentClientOptions {
	return func(ga *GPUAgentClient) {
		if k8sclient == nil {
			logger.Log.Printf("K8sApiClient disabled")
			ga.k8sApiClient = nil
			return
		}
		ga.enabledK8sApi = true
		logger.Log.Printf("K8sApiClient option set")
		ga.k8sApiClient = k8sclient
	}
}

func WithSRIOV(enableSriov bool) GPUAgentClientOptions {
	return func(ga *GPUAgentClient) {
		logger.Log.Printf("sriov mode set %v", enableSriov)
		ga.enableSriov = enableSriov
	}
}

func WithK8sSchedulerClient(k8sScheduler scheduler.SchedulerClient) GPUAgentClientOptions {
	return func(ga *GPUAgentClient) {
		if utils.IsKubernetes() {
			ga.isKubernetes = true
			logger.Log.Printf("K8sSchedulerClient option set")
			ga.k8sScheduler = k8sScheduler
		}
	}
}

func initclients(mh *metricsutil.MetricsHandler) (conn *grpc.ClientConn, gpuclient amdgpu.GPUSvcClient, evtclient amdgpu.EventSvcClient, err error) {
	agentAddr := mh.GetAgentAddr()
	logger.Log.Printf("Agent connecting to %v", agentAddr)
	conn, err = grpc.NewClient(agentAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Log.Printf("err :%v", err)
		return
	}
	gpuclient = amdgpu.NewGPUSvcClient(conn)
	evtclient = amdgpu.NewEventSvcClient(conn)
	return
}

func NewAgent(mh *metricsutil.MetricsHandler, opts ...GPUAgentClientOptions) *GPUAgentClient {
	ga := &GPUAgentClient{
		mh:                     mh,
		computeNodeHealthState: true,
		nodeHealthLabellerCfg: &utils.NodeHealthLabellerConfig{
			LabelPrefix: globals.GPUHealthLabelPrefix,
		},
	}
	for _, o := range opts {
		o(ga)
	}

	ga.healthState = make(map[string]*metricssvc.GPUState)
	ga.mockEccField = make(map[string]map[string]uint32)
	ga.fl = NewFieldLogger()
	ga.rocpclient = rocprofiler.NewRocProfilerClient("rocpclient")
	if ga.enableSriov {
		logger.Log.Printf("profiler is disabled on sriov deployment")
		ga.enableProfileMetrics = false
	} else {
		ga.enableProfileMetrics = true
	}
	ga.fsysDeviceHandler = fsysdevice.GetFsysDeviceHandler()
	ga.gCache = &gpuCache{}
	mh.RegisterMetricsClient(ga)
	return ga
}

func (ga *GPUAgentClient) Init() error {
	ga.Lock()
	defer ga.Unlock()
	ga.initializeContext()
	conn, gpuclient, evtclient, err := initclients(ga.mh)
	if err != nil {
		logger.Log.Printf("gpu client init failure err :%v", err)
		return err
	}

	ga.conn = conn
	ga.gpuclient = gpuclient
	ga.evtclient = evtclient

	slurmScl, err := scheduler.NewSlurmClient(ga.ctx, ga.enableZmq)
	if err != nil {
		logger.Log.Printf("gpu client init failure err :%v", err)
		return err
	}
	ga.slurmScheduler = slurmScl

	if err := ga.populateStaticHostLabels(); err != nil {
		return fmt.Errorf("error in populating static host labels, %v", err)
	}

	return nil
}

func (ga *GPUAgentClient) initializeContext() {
	ctx, cancel := context.WithCancel(context.Background())
	ga.ctx = ctx
	ga.cancel = cancel
}

func (ga *GPUAgentClient) reconnect() error {
	ga.Close()
	return ga.Init()
}

func (ga *GPUAgentClient) isActive() bool {
	ga.Lock()
	defer ga.Unlock()
	return ga.gpuclient != nil
}

func (ga *GPUAgentClient) StartMonitor() {
	logger.Log.Printf("GPUAgent monitor started")
	ga.initializeContext()
	pollTimer := time.NewTicker(refreshInterval)
	defer pollTimer.Stop()

	// nolint
	for {
		select {
		case <-pollTimer.C:
			if !ga.isActive() {
				if err := ga.reconnect(); err != nil {
					logger.Log.Printf("gpuagent connection failed %v", err)
					continue
				}
			}
			if err := ga.processHealthValidation(); err != nil {
				logger.Log.Printf("gpuagent health validation failed %v", err)
			}
			if err := ga.sendNodeLabelUpdate(); err != nil {
				logger.Log.Printf("gpuagent failed to send node label update %v", err)
			}
		}
	}
}

func (ga *GPUAgentClient) sendNodeLabelUpdate() error {
	if !ga.enabledK8sApi {
		return nil
	}
	// send update to label , reconnect logic tbd
	nodeName := utils.GetNodeName()
	if nodeName == "" {
		logger.Log.Printf("error getting node name on k8s deployment, skip label update")
		return fmt.Errorf("node name not found")
	}
	gpuHealthStates := make(map[string]string)
	ga.Lock()
	for gpuid, hs := range ga.healthState {
		if hs.Health == strings.ToLower(metricssvc.GPUHealth_HEALTHY.String()) {
			continue // skip healthy state
		}
		gpuHealthStates[gpuid] = hs.Health
	}
	ga.Unlock()
	_ = ga.k8sApiClient.UpdateHealthLabel(ga.nodeHealthLabellerCfg, nodeName, gpuHealthStates)
	return nil
}

func (ga *GPUAgentClient) isProfilerEnabled() bool {
	if ga.rocpclient == nil || !ga.enableProfileMetrics {
		// profiler is disabled either at boot time or through configmap
		return false
	}
	return true
}

func strDoubleToFloat(strValue string) float64 {
	floatValue, err := strconv.ParseFloat(strValue, 64)
	if err != nil {
		fmt.Println("Error parsing string:", err)
		return 0.0
	}
	return floatValue
}

// make it easy to parse from json
func (ga *GPUAgentClient) getProfilerMetrics() (map[string]map[string]float64, error) {
	gpuMetrics := make(map[string]map[string]float64)
	// stop exporting fields when disabled
	if !ga.isProfilerEnabled() {
		return gpuMetrics, nil
	}
	gpuProfiler, err := ga.rocpclient.GetMetrics()
	if err != nil {
		return gpuMetrics, err
	}
	for _, gpu := range gpuProfiler.GpuMetrics {
		gpuMetric := make(map[string]float64)
		for _, m := range gpu.Metrics {
			gpuMetric[m.Field] = strDoubleToFloat(m.Value)
		}
		// associate using logical node id
		gpuMetrics[gpu.LogicalNodeId] = gpuMetric
	}
	return gpuMetrics, nil
}

func (ga *GPUAgentClient) getMetricsAll() error {
	// send the req to gpuclient
	resp, partitionMap, err := ga.getGPUs()
	if err != nil {
		return err
	}
	if resp != nil && resp.ApiStatus != 0 {
		logger.Log.Printf("resp status :%v", resp.ApiStatus)
		return fmt.Errorf("%v", resp.ApiStatus)
	}
	cper, err := ga.getLatestCPER()
	if err != nil {
		logger.Log.Printf("getLatestCPER failed with err : %v", err)
		cper = nil
	}
	wls, _ := ga.ListWorkloads()
	pmetrics, err := ga.getProfilerMetrics()
	if err != nil {
		//continue as this may not be available at this time
		pmetrics = nil
	}
	k8PodLabelsMap, err = ga.FetchPodLabelsForNode()
	if err != nil {
		logger.Log.Printf("FetchPodLabelsForNode failed with err : %v", err)
	}
	nonGpuLabels := ga.populateLabelsFromGPU(nil, nil, nil)
	ga.m.gpuNodesTotal.With(nonGpuLabels).Set(float64(len(resp.Response)))
	for _, gpu := range resp.Response {
		var gpuProfMetrics map[string]float64
		// if available use the data
		if pmetrics != nil {
			gpuid := getGPUNodeID(gpu)
			//nolint
			gpuProfMetrics, _ = pmetrics[gpuid]
		}
		ga.updateGPUInfoToMetrics(wls, gpu, partitionMap, gpuProfMetrics, cper)
	}

	return nil
}

// cacheRead reads from cache if the last read was successful and within the cacheTimer
// otherwise it reads from hardware
// this ensures that we don't read from hardware too frequently as more clients are added
// and the number of reads increases
func (ga *GPUAgentClient) cacheRead() (*amdgpu.GPUGetResponse, error) {
	now := time.Now()

	// First try fast path with RLock
	ga.gCache.RLock()
	if ga.gCache.lastResponse != nil && now.Sub(ga.gCache.lastTimestamp) < cacheTimer {
		res := ga.gCache.lastResponse
		ga.gCache.RUnlock()
		logger.Log.Printf("returning metrics from cache")
		return res, nil
	}
	ga.gCache.RUnlock()

	// Acquire full Lock to update cache
	ga.gCache.Lock()
	defer ga.gCache.Unlock()

	// Check again after acquiring Lock to handle the case where another goroutine has already updated the cache
	if ga.gCache.lastResponse != nil && time.Since(ga.gCache.lastTimestamp) < cacheTimer {
		logger.Log.Printf("returning metrics from cache (after double-check)")
		return ga.gCache.lastResponse, nil
	}

	// Perform query and update cache
	ctx, cancel := context.WithTimeout(ga.ctx, queryTimeout)
	defer cancel()

	res, err := ga.gpuclient.GPUGet(ctx, &amdgpu.GPUGetRequest{})
	ga.gCache.lastTimestamp = time.Now()
	if err == nil {
		ga.gCache.lastResponse = res
	} else {
		ga.gCache.lastResponse = nil
	}
	return res, err
}

func (ga *GPUAgentClient) cacheCperRead() (*amdgpu.GPUCPERGetResponse, error) {
	now := time.Now()

	// First try fast path with RLock
	ga.gCache.RLock()
	if ga.gCache.lastCperResponse != nil && now.Sub(ga.gCache.lastCperTimestamp) < cacheTimer {
		res := ga.gCache.lastCperResponse
		ga.gCache.RUnlock()
		logger.Log.Printf("returning CPER metrics from cache")
		return res, nil
	}
	ga.gCache.RUnlock()

	// Acquire full Lock to update cache
	ga.gCache.Lock()
	defer ga.gCache.Unlock()

	// Check again after acquiring Lock to handle the case where another goroutine has already updated the cache
	if ga.gCache.lastCperResponse != nil && time.Since(ga.gCache.lastCperTimestamp) < cacheTimer {
		logger.Log.Printf("returning CPER metrics from cache (after double-check)")
		return ga.gCache.lastCperResponse, nil
	}

	// Perform query and update cache
	ctx, cancel := context.WithTimeout(ga.ctx, queryTimeout)
	defer cancel()

	res, err := ga.gpuclient.GPUCPERGet(ctx, &amdgpu.GPUCPERGetRequest{})
	ga.gCache.lastCperTimestamp = time.Now()
	if err == nil {
		ga.gCache.lastCperResponse = res
	} else {
		ga.gCache.lastCperResponse = nil
	}
	return res, err
}

// getLatestCPER fetches the latest CPER entry per GPU from the cached CPER data
func (ga *GPUAgentClient) getLatestCPER() (map[string]*amdgpu.CPEREntry, error) {
	// skip if afid metrics are disabled - saves time on fetching CPER data
	if !ga.enableAfidMetrics {
		return nil, nil
	}
	gpuCpers, err := ga.cacheCperRead()
	if err != nil {
		return nil, err
	}
	if gpuCpers != nil && gpuCpers.ApiStatus != 0 {
		logger.Log.Printf("CPER resp status :%v", gpuCpers.ApiStatus)
		return nil, fmt.Errorf("%v", gpuCpers.ApiStatus)
	}

	latestCPER := make(map[string]*cperCacheEntry)

	for _, c := range gpuCpers.CPER {
		uuid, _ := uuid.FromBytes(c.GPU)
		gpuuid := uuid.String()

		for _, record := range c.CPEREntry {
			ts := record.GetTimestamp() // this is of format Timestamp:"2025-09-12 15:00:27"
			// Parse current timestamp
			currentTS, err := time.Parse("2006-01-02 15:04:05", ts)
			if err != nil {
				logger.Log.Printf("Failed to parse current timestamp: %v", err)
				continue
			}

			if existingCPER, exists := latestCPER[gpuuid]; !exists {
				latestCPER[gpuuid] = &cperCacheEntry{
					entry:     record,
					timestamp: currentTS,
				}
			} else {
				// Update if current is more recent
				if currentTS.After(existingCPER.timestamp) {
					latestCPER[gpuuid] = &cperCacheEntry{
						entry:     record,
						timestamp: currentTS,
					}
				}
			}
		}
	}

	// Convert to return type and log summary
	result := make(map[string]*amdgpu.CPEREntry)
	for gpuuid, cacheEntry := range latestCPER {
		entry := cacheEntry.entry
		result[gpuuid] = entry
	}

	return result, nil
}

func (ga *GPUAgentClient) getGPUs() (*amdgpu.GPUGetResponse, map[string]*amdgpu.GPU, error) {

	if !ga.isActive() {
		if err := ga.reconnect(); err != nil {
			return nil, nil, err
		}
	}
	res, err := ga.cacheRead()
	if err != nil {
		return nil, nil, err
	}
	// filter out logical GPU
	nres := &amdgpu.GPUGetResponse{
		ApiStatus: res.ApiStatus,
		Response:  []*amdgpu.GPU{},
		ErrorCode: res.ErrorCode,
	}
	partitionMap := make(map[string]*amdgpu.GPU)
	for _, gpu := range res.Response {
		if gpu.Status.PCIeStatus != nil {
			gpuPcieAddr := strings.ToLower(gpu.Status.PCIeStatus.PCIeBusId)
			pcieBaseAddr := utils.GetPCIeBaseAddress(gpuPcieAddr)
			// parent gpu map is created only for partitioned gpu
			if (pcieBaseAddr != gpuPcieAddr) && (gpu.Status.GetPartitionId() == 0) {
				partitionMap[pcieBaseAddr] = gpu
			}
		}
		if len(gpu.Status.GPUPartition) != 0 {
			// skip logical gpu objects
			continue
		}
		nres.Response = append(nres.Response, gpu)
	}
	return nres, partitionMap, err
}

func (ga *GPUAgentClient) getEvents(severity amdgpu.EventSeverity) (*amdgpu.EventResponse, error) {
	req := &amdgpu.EventRequest{}
	if severity != amdgpu.EventSeverity_EVENT_SEVERITY_NONE {
		req.Filter = &amdgpu.EventFilter{
			Filter: &amdgpu.EventFilter_MatchAttrs{
				MatchAttrs: &amdgpu.EventMatchAttrs{
					Severity: severity,
				},
			},
		}
	}
	res, err := ga.evtclient.EventGet(ga.ctx, req)
	return res, err
}

// ListWorkloads - get all workloads from every client , lock must be taken by
// the caller
func (ga *GPUAgentClient) ListWorkloads() (wls map[string]scheduler.Workload, err error) {
	wls = make(map[string]scheduler.Workload)
	if ga.isKubernetes && ga.k8sScheduler != nil {
		var k8sWls map[string]scheduler.Workload
		k8sWls, err = ga.k8sScheduler.ListWorkloads()
		if err != nil {
			return
		}
		for k, wl := range k8sWls {
			wls[k] = wl
		}
	}
	if ga.slurmScheduler == nil {
		return wls, nil
	}
	var swls map[string]scheduler.Workload
	swls, err = ga.slurmScheduler.ListWorkloads()
	if err != nil {
		return
	}
	// return combined list
	for k, wl := range swls {
		wls[k] = wl
	}
	return
}

func (ga *GPUAgentClient) Close() {
	ga.Lock()
	defer ga.Unlock()
	if ga.conn != nil {
		logger.Log.Printf("gpuagent client closing")
		ga.conn.Close()
		ga.gpuclient = nil
		ga.conn = nil
	}
	if ga.k8sScheduler != nil {
		logger.Log.Printf("gpuagent k8s scheduler closing")
		ga.k8sScheduler.Close()
		ga.k8sScheduler = nil
	}

	if ga.slurmScheduler != nil {
		logger.Log.Printf("gpuagent slurm scheduler closing")
		ga.slurmScheduler.Close()
		ga.slurmScheduler = nil
	}
	// cancel all context
	ga.cancel()
}

func (ga *GPUAgentClient) FetchPodLabelsForNode() (map[string]map[string]string, error) {
	listMap := make(map[string]map[string]string)
	if ga.enabledK8sApi && len(extraPodLabelsMap) > 0 {
		return ga.k8sApiClient.GetAllPods()
	}
	return listMap, nil
}
