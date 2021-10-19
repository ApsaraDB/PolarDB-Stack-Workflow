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
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/ApsaraDB/PolarDB-Stack-Workflow/define"
	"github.com/ApsaraDB/PolarDB-Stack-Workflow/statemachine"
	"github.com/ApsaraDB/PolarDB-Stack-Workflow/utils"
)

type ResourceWorkflow struct {
	Logger           logr.Logger
	Resource         statemachine.StateResource
	WfManager        *WfManager
	MementoCareTaker *WfRuntimeMementoCareTaker
}

func CreateResourceWorkflow(
	resource statemachine.StateResource,
	wfManager *WfManager,
) (*ResourceWorkflow, error) {
	var (
		err              error
		mementoCareTaker *WfRuntimeMementoCareTaker
	)
	if resource == nil {
		err = errors.New("resource is nil")
		return nil, err
	}
	if wfManager == nil {
		err = errors.New("WfManager is nil")
		return nil, err
	}
	logger := wfManager.Logger.WithName(fmt.Sprintf("%s/%s", resource.GetNamespace(), resource.GetName()))
	if wfManager.createMementoStorageFunc != nil {
		mementoCareTaker, err = CreateWfRuntimeMementoCareTaker(resource, wfManager.createMementoStorageFunc)
		if err != nil {
			return nil, err
		}
	}
	return &ResourceWorkflow{
		Resource:         resource,
		WfManager:        wfManager,
		MementoCareTaker: mementoCareTaker,
		Logger:           logger,
	}, nil
}

func (m *ResourceWorkflow) Run(
	ctx context.Context,
	flowName string,
	initContext map[string]interface{},
) (err error) {
	baseContext := m.getCommonContext(m.Resource)
	initFlowContext := utils.MapMerge(baseContext, initContext)
	if _, ok := m.WfManager.FlowMetaMap[flowName]; !ok {
		err = errors.New(fmt.Sprintf("flowName: %v not found", flowName))
		m.Logger.Error(err, "")
		return
	} else {
		flowRuntime, err := CreateWfRuntime(flowName, initFlowContext, m)
		if err != nil {
			m.Logger.Error(err, "failed to create workflow runtime", "flowName", flowName, "initFlowContext", initFlowContext)
			return err
		}
		return flowRuntime.Start(ctx)
	}
}

func (m *ResourceWorkflow) GetLastUnCompletedRuntime() (*WfRuntime, error) {
	flows, err := m.MementoCareTaker.GetAllUnCompletedWorkflowRuntime()
	if err != nil {
		err = errors.Wrap(err, "get uncompleted workflow runtime from memento failed!")
		return nil, err
	}
	if flows == nil || len(flows) < 1 {
		return nil, nil
	}
	_, flow, err := getLastUnCompleted(flows)
	if err != nil {
		err = errors.Wrap(err, "get last uncompleted workflow runtime failed!")
		return nil, err
	}
	return flow, nil
}

func (m *ResourceWorkflow) RunLastUnCompletedRuntime(ctx context.Context, flowName string, ignoreOtherUnCompleted bool) (oldWfName string, isWaiting bool, err error) {
	flow, err := m.GetLastUnCompletedRuntime()
	if err != nil {
		err = errors.Wrap(err, "get last uncompleted workflow runtime failed!")
		return "", false, err
	}
	if flow == nil {
		return "", false, nil
	}
	oldWfName = flow.FlowName
	// 重建流程不重试之前中断的流程，但是要重试自己
	if ignoreOtherUnCompleted && flowName != oldWfName {
		return oldWfName, false, nil
	}
	if err = m.RunUnCompletedRuntime(ctx, flow); err != nil {
		return oldWfName, false, err
	}
	return oldWfName, flow.FlowStatus == FlowStatusWaiting, nil
}

func (m *ResourceWorkflow) RetryInterruptedStep() error {
	mementoKey, flowRuntime, err := m.MementoCareTaker.GetLastWorkflowRuntime()
	if err != nil {
		return errors.Wrap(err, "get last uncompleted workflow runtime failed!")
	}
	flowMeta, ok := m.WfManager.FlowMetaMap[flowRuntime.FlowName]
	if !ok {
		return errors.New(fmt.Sprintf("get workflow runtime [%s] failed!", flowRuntime.FlowName))
	}
	if flowRuntime == nil || flowRuntime.FlowStatus != FlowStatusFailedCompleted {
		return nil
	}
	flowRuntime.FlowStatus = FlowStatusFailed
	recoverFromFirstStep := flowMeta.RecoverFromFirstStep
	for _, step := range flowRuntime.RunningSteps {
		if recoverFromFirstStep {
			// 从第一步开始恢复执行
			step.StepStatus = StepStatusPrepared
			step.RetryTimes = 0
		} else if step.StepStatus == StepStatusFailed {
			// 从中断位置开始恢复执行
			step.RetryTimes = 0
			break
		}
	}
	mementoContent, err := json.Marshal(flowRuntime)
	if err != nil {
		return err
	}
	return m.MementoCareTaker.MementoStorage.Save(mementoKey, string(mementoContent))
}

func getLastUnCompleted(flows map[string]*WfRuntime) (string, *WfRuntime, error) {
	var (
		lastId              int
		lastKey             string
		lastUnCompletedFlow *WfRuntime
	)
	for key, flow := range flows {
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

func (m *ResourceWorkflow) RunUnCompletedRuntime(ctx context.Context, wfRuntime *WfRuntime) error {
	if err := LoadUnCompletedWfRuntime(wfRuntime, m); err != nil {
		return err
	}
	return wfRuntime.Start(ctx)
}

func (m *ResourceWorkflow) getCommonContext(resource statemachine.StateResource) map[string]interface{} {
	return map[string]interface{}{
		m.WfManager.GetConfItem(define.WorkFlowResourceName):      resource.GetName(),
		m.WfManager.GetConfItem(define.WorkFlowResourceNameSpace): resource.GetNamespace(),
	}
}

func (m *ResourceWorkflow) CommonWorkFlowMainEnter(ctx context.Context, resource statemachine.StateResource, flowName string, ignoreUnCompleted bool, eventChecker statemachine.EventChecker) error {
	logger := m.Logger.WithName(flowName)
	logger.Info("begin do workflow main")

	cancelCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	go wait.Until(func() {
		if resource != nil && resource.IsCancelled() {
			cancel()
		}
	}, 5*time.Second, cancelCtx.Done())

	var err error
	var isWaiting = false

	//旧流程还要跑完，才能进入新流程
	_, isWaiting, err = m.RunLastUnCompletedRuntime(cancelCtx, flowName, ignoreUnCompleted)
	if err != nil {
		logger.Error(err, "RunLastUnCompletedRuntime failed")
		return err
	}
	if isWaiting {
		return nil
	}
	resource, err = m.Resource.Fetch()
	if err != nil {
		logger.Error(err, "reload resource failed")
		return err
	}

	event, err := eventChecker(resource)
	if err != nil {
		return err
	}
	// 如果没有了event，将状态设置为running.
	if event == nil {
		logger.Info("check event is nil, so set status to running")
		if resource, err = resource.UpdateState(statemachine.StateRunning); err != nil {
			return err
		}
		return nil
	}
	logger.Info("found event, run workflow", "event", event)
	if err = m.Run(cancelCtx, flowName, event.Params); err != nil {
		return err
	}
	resource, err = m.Resource.Fetch()
	if err != nil {
		logger.Error(err, "run workflow succeed, but reload resource failed")
		return err
	}
	return nil
}
