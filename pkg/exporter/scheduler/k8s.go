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

package scheduler

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ROCm/device-metrics-exporter/pkg/exporter/gen/exportermetrics"
	"github.com/ROCm/device-metrics-exporter/pkg/exporter/globals"
	"github.com/ROCm/device-metrics-exporter/pkg/exporter/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	kube "k8s.io/kubelet/pkg/apis/podresources/v1alpha1"

	"os"
)

var KubernetesLabels = map[string]bool{
	exportermetrics.MetricLabel_POD.String():       true,
	exportermetrics.MetricLabel_NAMESPACE.String(): true,
	exportermetrics.MetricLabel_CONTAINER.String(): true,
}

type podResourcesClient struct {
	clientConn *grpc.ClientConn
	ctx        context.Context // parent context
}

// NewKubernetesClient - creates a kubernetes schedler client
func NewKubernetesClient(ctx context.Context) (SchedulerClient, error) {
	if _, err := os.Stat(globals.PodResourceSocket); err != nil {
		logger.Log.Printf("no kubelet found")
		return nil, fmt.Errorf("no kubelet, %v", err)
	}
	client, err := grpc.NewClient("unix://"+globals.PodResourceSocket, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Log.Printf("kubelet socket err: %v", err)
		return nil, fmt.Errorf("kubelet socket error, %v", err)
	}
	logger.Log.Printf("created k8s scheduler client")
	return &podResourcesClient{clientConn: client, ctx: ctx}, nil
}

// ListWorkloads - list all the workloads
// This function will return a map of device id to workload
// The device id is the device id (pcie id or partition xcd string) of the GPU
func (cl *podResourcesClient) ListWorkloads() (map[string]Workload, error) {
	prCl := kube.NewPodResourcesListerClient(cl.clientConn)
	ctx, cancel := context.WithTimeout(cl.ctx, time.Second*10)
	defer cancel()
	resp, err := prCl.List(ctx, &kube.ListPodResourcesRequest{})
	if err != nil {
		logger.Log.Printf("failed to list pod resources, %v", err)
		if cl.isConnectionError(err) {
			logger.Log.Printf("attempting to reconnect to kubelet socket...")
			if recErr := cl.reconnect(); recErr != nil {
				return nil, fmt.Errorf("failed to reconnect: %v (original error: %v)", recErr, err)
			}

			// Retry once after reconnect
			prCl = kube.NewPodResourcesListerClient(cl.clientConn)
			ctx, cancel = context.WithTimeout(cl.ctx, time.Second*10)
			defer cancel()

			resp, err = prCl.List(ctx, &kube.ListPodResourcesRequest{})
			if err != nil {
				logger.Log.Printf("retry after reconnect failed: %v", err)
				return nil, fmt.Errorf("failed to list pod resources after reconnect, %v", err)
			}
		} else {
			return nil, fmt.Errorf("failed to list pod resources, %v", err)
		}
	}

	podInfo := make(map[string]Workload)
	for _, pod := range resp.PodResources {
		for _, container := range pod.Containers {
			for _, devs := range container.GetDevices() {
				if strings.HasPrefix(devs.ResourceName, globals.AMDGPUResourcePrefix) {
					for _, devId := range devs.DeviceIds {
						podInfo[strings.ToLower(devId)] = Workload{
							Type: Kubernetes,
							Info: PodResourceInfo{
								Pod:       pod.Name,
								Namespace: pod.Namespace,
								Container: container.Name,
							},
						}
					}
				}
			}
		}
	}
	return podInfo, nil
}

func (cl *podResourcesClient) CheckExportLabels(labels map[string]bool) bool {
	for k := range KubernetesLabels {
		if ok := labels[k]; ok {
			return true
		}
	}
	return false
}

func (cl *podResourcesClient) Close() error {
	return cl.clientConn.Close()
}

func (cl *podResourcesClient) Type() SchedulerType {
	return Kubernetes
}

// reconnect tries to close the existing connection and dial again.
// You can call this if you detect the connection is broken during usage.
func (cl *podResourcesClient) reconnect() error {
	if cl.clientConn != nil {
		_ = cl.clientConn.Close()
	}

	var err error
	cl.clientConn, err = grpc.NewClient("unix://"+globals.PodResourceSocket, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Log.Printf("failed to reconnect to kubelet socket: %v", err)
		return err
	}
	logger.Log.Printf("reconnected to kubelet socket")
	return nil
}

// isConnectionError checks if the error is related to connection issues
func (cl *podResourcesClient) isConnectionError(err error) bool {
	return strings.Contains(err.Error(), "connection refused") ||
		strings.Contains(err.Error(), "connection reset") ||
		strings.Contains(err.Error(), "transport is closing")
}
