/*
*Copyright (c) 2019-2021, Alibaba Group Holding Limited;
*Licensed under the Apache License, Version 2.0 (the "License");
*you may not use this file except in compliance with the License.
*You may obtain a copy of the License at

*   http://www.apache.org/licenses/LICENSE-2.0

*Unless required by applicable law or agreed to in writing, software
*distributed under the License is distributed on an "AS IS" BASIS,
*WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
*See the License for the specific language governing permissions and
*limitations under the License.
 */

package dbclusterwf

import (
	"github.com/ApsaraDB/PolarDB-Stack-Workflow/implement"
	"github.com/ApsaraDB/PolarDB-Stack-Workflow/statemachine"
	"github.com/ApsaraDB/PolarDB-Stack-Workflow/utils/k8sutil"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
)

type DbResource struct {
	implement.KubeResource
	Logger logr.Logger
}

func (s *DbResource) GetDbCluster() *corev1.ConfigMap {
	return s.Resource.(*corev1.ConfigMap)
}

// GetState 获取资源当前状态
func (s *DbResource) GetState() statemachine.State {
	dbCluster := s.Resource.(*corev1.ConfigMap)
	return statemachine.State(dbCluster.Data["clusterStatus"])
}

// UpdateState 更新资源当前状态(string)
func (s *DbResource) UpdateState(state statemachine.State) (statemachine.StateResource, error) {
	so, err := s.fetch()
	if err != nil {
		return nil, err
	}
	dbCluster := so.Resource.(*corev1.ConfigMap)
	if dbCluster.Data == nil {
		dbCluster.Data = map[string]string{}
	}
	dbCluster.Data["clusterStatus"] = string(state)
	if err := k8sutil.ConfigMapUpdate(dbCluster); err != nil {
		return nil, err
	}
	return so, nil
}

// 更新资源信息
func (s *DbResource) Update() error {
	dbCluster := s.Resource.(*corev1.ConfigMap)
	return k8sutil.ConfigMapUpdate(dbCluster)
}

// Fetch 重新获取资源
func (s *DbResource) Fetch() (statemachine.StateResource, error) {
	return s.fetch()
}

// GetScheme ...
func (s *DbResource) GetScheme() *runtime.Scheme {
	return scheme.Scheme
}

func (s *DbResource) IsCancelled() bool {
	sbl, err := s.fetch()
	if err != nil {
		return false
	}
	return sbl.Resource.GetAnnotations()["cancelled"] == "true"
}

func (s *DbResource) fetch() (*DbResource, error) {
	dbCluster, err := k8sutil.ConfigMapFactoryByName(s.Resource.GetName(), s.Resource.GetNamespace())
	if err != nil {
		return nil, err
	}
	return &DbResource{
		KubeResource: implement.KubeResource{
			Resource: dbCluster,
		},
		Logger: s.Logger,
	}, nil
}
