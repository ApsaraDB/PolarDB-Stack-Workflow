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


package wfengine

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"gitlab.alibaba-inc.com/polar-as/polar-wf-engine/define"
	"gitlab.alibaba-inc.com/polar-as/polar-wf-engine/statemachine"
	"gitlab.alibaba-inc.com/polar-as/polar-wf-engine/utils"
)

const flowSaveResourceNameKey = "_flowSaveResourceNameKey"
const flowSaveResourceNameSpaceKey = "_flowSaveResourceNameSpaceKey"
const flowSaveNumberKey = "_flowSaveNumberKey"

type WfRuntimeMementoStorage interface {
	Save(mementoKey, mementoContent string) error
	LoadMementoMap(careTakerName string) (map[string]string, error)
}

// 工作流运行时数据存储
type WfRuntimeMemento struct {
	name              string
	namespace         string
	content           string
	mementoKey        string
	flowApplyResource string // 这个流程是针对哪个资源的
	isNew             bool   // 是否是新备忘录
	wfRuntime         *WfRuntime
}

// 工作流运行时数据存储管理器
type WfRuntimeMementoCareTaker struct {
	Name              string
	MementoMap        map[string]string
	MementoStorage    WfRuntimeMementoStorage
	Namespace         string
	FlowApplyResource string
}

func CreateWfRuntimeMementoCareTaker(resource statemachine.StateResource, createStorageFunc CreateMementoStorageFunc) (*WfRuntimeMementoCareTaker, error) {
	var (
		storage       WfRuntimeMementoStorage
		mementoMap    map[string]string
		err           error
		careTakerName = resource.GetName() + "-" + "workflow"
	)
	storage, err = createStorageFunc(resource)
	if err != nil {
		return nil, err
	}
	if mementoMap, err = storage.LoadMementoMap(careTakerName); err != nil {
		return nil, err
	}
	return &WfRuntimeMementoCareTaker{
		Name:              careTakerName,
		Namespace:         resource.GetNamespace(),
		FlowApplyResource: resource.GetName(),
		MementoMap:        mementoMap,
		MementoStorage:    storage,
	}, nil
}

// 根据已经存在的工作流运行时创建一个备忘录实例，用于更新备忘录
func (cm *WfRuntimeMementoCareTaker) LoadMemento(wf *WfRuntime) (*WfRuntimeMemento, error) {
	var memento = &WfRuntimeMemento{
		isNew:     false,
		wfRuntime: wf,
	}
	context := wf.InitContext
	err := cm.setMementoPropWithContext(memento, context)
	if err != nil {
		return nil, err
	}
	return memento, nil
}

// 创建一个新的工作流运行时备忘录
func (cm *WfRuntimeMementoCareTaker) CreateMemento(flowName string) (*WfRuntimeMemento, error) {
	mem := &WfRuntimeMemento{
		name:              cm.Name,
		namespace:         cm.Namespace,
		flowApplyResource: cm.FlowApplyResource,
		isNew:             true,
	}

	if cm.MementoMap == nil || len(cm.MementoMap) < 1 {
		mem.mementoKey = fmt.Sprintf("1-%v", flowName)
		return mem, nil
	}

	// 获取当前所有的key, 并且 + 1
	nowMaxId, err := cm.findMaxId()
	if err != nil {
		return nil, err
	}
	mem.mementoKey = fmt.Sprintf("%v-%v", nowMaxId+1, flowName)

	return mem, nil
}

// 保存备忘录
func (cm *WfRuntimeMementoCareTaker) SaveMemento(memento *WfRuntimeMemento) error {
	if err := cm.MementoStorage.Save(memento.mementoKey, memento.content); err != nil {
		return err
	}
	return nil
}

// 获得所有尚未完成的工作流
func (cm *WfRuntimeMementoCareTaker) GetAllUnCompletedWorkflowRuntime() (map[string]*WfRuntime, error) {
	var flowList = make(map[string]*WfRuntime)
	for flowKey, flowValue := range cm.MementoMap {
		flow := &WfRuntime{}
		err := json.Unmarshal([]byte(flowValue), flow)
		if err != nil {
			err = errors.Wrapf(err, "find value illegal: %s", flowValue)
			return nil, err
		}
		//对于失败不再重试的流程，不需要再试尝试
		if flow.FlowStatus != FlowStatusCompleted && flow.FlowStatus != FlowStatusFailedCompleted {
			flowList[flowKey] = flow
		}
	}
	return flowList, nil
}

// 获得最后执行的工作流
func (cm *WfRuntimeMementoCareTaker) GetLastWorkflowRuntime() (string, *WfRuntime, error) {
	if cm.MementoMap == nil || len(cm.MementoMap) < 1 {
		return "", nil, nil
	}

	var (
		lastId              int
		lastKey             string
		lastUnCompletedFlow *WfRuntime
	)
	for key, flowValue := range cm.MementoMap {
		flow := &WfRuntime{}
		err := json.Unmarshal([]byte(flowValue), flow)
		if err != nil {
			err = errors.Wrapf(err, "find value illegal: %s", flowValue)
			return "", nil, err
		}

		id, err := strconv.Atoi(strings.Split(key, "-")[0])
		if err != nil {
			return "", nil, err
		}
		if id > lastId {
			lastId = id
			lastKey = key
			lastUnCompletedFlow = flow
		}
	}
	return lastKey, lastUnCompletedFlow, nil
}

func (cm *WfRuntimeMementoCareTaker) setMementoPropWithContext(memento *WfRuntimeMemento, context map[string]interface{}) error {
	var ok bool
	if memento.name, ok = context[flowSaveResourceNameKey].(string); !ok {
		return errors.New("failed to get flowSaveResourceNameKey by context")
	}
	if memento.namespace, ok = context[flowSaveResourceNameSpaceKey].(string); !ok {
		return errors.New("failed to get flowSaveResourceNameSpaceKey by context")
	}
	if memento.mementoKey, ok = context[flowSaveNumberKey].(string); !ok {
		return errors.New("failed to get flowSaveNumberKey by context")
	}
	if memento.flowApplyResource, ok = context[memento.wfRuntime.ResourceWorkflow.WfManager.GetConfItem(define.WorkFlowResourceName)].(string); !ok {
		return errors.New("failed to get resourceName by context")
	}
	return nil
}

func (cm *WfRuntimeMementoCareTaker) findMaxId() (int, error) {
	flowIdList := make([]int, len(cm.MementoMap))
	for flowKey := range cm.MementoMap {
		flowId, err := strconv.Atoi(strings.Split(flowKey, "-")[0])
		if err != nil {
			return 0, err
		}
		flowIdList = append(flowIdList, flowId)
	}
	return utils.MaxIntSlice(flowIdList), nil
}
