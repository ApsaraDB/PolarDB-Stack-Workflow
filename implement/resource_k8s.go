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


package implement

import (
	"errors"

	"github.com/ApsaraDB/PolarDB-Stack-Workflow/statemachine"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type IKubeResource interface {
	statemachine.StateResource
	GetScheme() *runtime.Scheme
	GetResource() metav1.Object
	Update() error
}

var NotImplementError = errors.New("methods need to be implemented")

type KubeResource struct {
	Resource metav1.Object
}

func (r *KubeResource) GetName() string {
	return r.Resource.GetName()
}

func (r *KubeResource) GetResource() metav1.Object {
	return r.Resource
}

func (r *KubeResource) GetNamespace() string {
	return r.Resource.GetNamespace()
}

// 获取资源当前状态
func (r *KubeResource) GetState() statemachine.State {
	return statemachine.StateNil
}

// 更新资源当前状态
func (r *KubeResource) UpdateState(statemachine.State) (statemachine.StateResource, error) {
	return nil, NotImplementError
}

// 更新资源信息
func (r *KubeResource) Update() error {
	return NotImplementError
}

// 重新获取资源
func (r *KubeResource) Fetch() (statemachine.StateResource, error) {
	return nil, NotImplementError
}

// 重新获取资源
func (r *KubeResource) GetScheme() *runtime.Scheme {
	return nil
}

// 流程是否被取消
func (r *KubeResource) IsCancelled() bool {
	return true
}
