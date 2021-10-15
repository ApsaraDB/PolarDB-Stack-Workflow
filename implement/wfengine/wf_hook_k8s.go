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


package implement_wfengine

import (
	"github.com/go-logr/logr"

	"gitlab.alibaba-inc.com/polar-as/polar-wf-engine/define"
	"gitlab.alibaba-inc.com/polar-as/polar-wf-engine/wfengine"
)

type DefaultWorkflowHook struct {
	wfRuntime       *wfengine.WfRuntime
	historyRecorder *DefaultHistoryRecorder
	logger          logr.Logger
}

func CreateDefaultWorkflowHook(wfRuntime *wfengine.WfRuntime) (wfengine.WfHook, error) {
	logger := wfRuntime.Logger.WithName("hook")
	resourceType := wfRuntime.ResourceWorkflow.WfManager.ResourceType
	historyRecorder := CreateDefaultHistoryRecorder(resourceType, logger, "")
	return &DefaultWorkflowHook{wfRuntime: wfRuntime, historyRecorder: historyRecorder, logger: logger}, nil
}

// 流程初始化钩子
func (wfh DefaultWorkflowHook) OnWfInit() error {
	if wfh.wfRuntime.FlowStatus != wfengine.FlowStatusWaiting {
		wfh.historyRecorder.WriteWorkflowStepLog(wfh.wfRuntime, nil, WorkFlowStart)
	}
	return nil
}

// 流程结束时执行
func (wfh DefaultWorkflowHook) OnWfCompleted() error {
	wfh.historyRecorder.WriteWorkflowStepLog(wfh.wfRuntime, nil, WorkFlowCompleted)
	return nil
}

// 流程中断时执行
func (wfh DefaultWorkflowHook) OnWfInterrupt(err *define.InterruptError) error {
	wfh.historyRecorder.WriteWorkflowStepLog(wfh.wfRuntime, nil, WorkFlowInterrupt)
	return nil
}

// 流程步骤执行前运行
func (wfh DefaultWorkflowHook) OnStepInit(step *wfengine.StepRuntime) error {
	wfh.historyRecorder.WriteWorkflowStepLog(wfh.wfRuntime, step, "")
	return nil
}

// 流程步骤执行前运行
func (wfh DefaultWorkflowHook) OnStepWaiting(step *wfengine.StepRuntime) error {
	wfh.historyRecorder.WriteWorkflowStepLog(wfh.wfRuntime, step, "")
	return nil
}

// 流程步骤执行后运行
func (wfh DefaultWorkflowHook) OnStepCompleted(step *wfengine.StepRuntime) error {
	wfh.historyRecorder.WriteWorkflowStepLog(wfh.wfRuntime, step, "")
	return nil
}
