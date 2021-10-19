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
	"reflect"
	"runtime/debug"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/ApsaraDB/PolarDB-Stack-Workflow/define"
	"github.com/ApsaraDB/PolarDB-Stack-Workflow/statemachine"
	"github.com/ApsaraDB/PolarDB-Stack-Workflow/utils"
)

type FlowStatusType string

const (
	FlowStatusPrepared        FlowStatusType = "prepared"
	FlowStatusRunning         FlowStatusType = "running"
	FlowStatusFailed          FlowStatusType = "failed"
	FlowStatusWaiting         FlowStatusType = "waiting"
	FlowStatusFailedCompleted FlowStatusType = "failedAndCompleted" //失败，但不需要再次重试的状态
	FlowStatusCompleted       FlowStatusType = "completed"
)

const FlowMaxAutoRetryTimes = 10

// +k8s:deepcopy-gen:
type WfRuntime struct {
	FlowStatus       FlowStatusType         `json:"flowStatus,omitempty"`
	RetryTimes       int                    `json:"retryTimes,omitempty"`
	FlowName         string                 `json:"flowName,omitempty"`
	IgnoreError      bool                   `json:"ignoreError,omitempty"` // 是否忽略报警.
	StartTime        string                 `json:"startTime,omitempty"`
	CompleteTime     string                 `json:"completeTime,omitempty"`
	ErrorMessage     string                 `json:"errorMessage,omitempty"`
	InitContext      map[string]interface{} `json:"initContext,omitempty"` // 初始化context. 会传给第一个步骤
	RunningSteps     []*StepRuntime         `json:"runningSteps,omitempty"`
	FlowHookIns      WfHook                 `json:"-"` // hook 函数用于在步骤完成/失败后调用，在k8s场景下，在这个方法会去执行落库的动作.
	ResourceWorkflow *ResourceWorkflow      `json:"-"`
	Logger           logr.Logger            `json:"-"`
}

func (r *WfRuntime) Start(ctx context.Context) (err error) {
	flowStart := time.Now()
	var runningStepRuntime *StepRuntime

	defer func() {
		r.Logger.Info(fmt.Sprintf("[time cost] %f s", time.Since(flowStart).Seconds()))
		re := recover()
		if re != nil {
			err = r.getRecoverError(re)
		}
		if err != nil {
			r.setStepAndFlowFailed(err, runningStepRuntime)
		}
	}()
	var rsName string
	rsName, err = r.getResourceFullName()
	if err != nil {
		return err
	}

	if r.FlowStatus == FlowStatusCompleted {
		r.RetryTimes = 0
		r.Logger.Info("wf runtime completed, ignore!")
		return nil
	}

	if r.FlowStatus == FlowStatusFailedCompleted {
		r.Logger.Info("wf runtime failedAndComplete, ignore!", "InitContext", r.InitContext)
		return nil
	}

	flowMeta, ok := r.ResourceWorkflow.WfManager.FlowMetaMap[r.FlowName]
	if !ok {
		r.Logger.Info("workflow meta has not existed, skip it.")
		err = r.setComplete()
		if err != nil {
			r.Logger.Error(err, "workflow set complete failed, skip it.")
		}
		return nil
	}

	if r.FlowStatus == FlowStatusFailed || r.FlowStatus == FlowStatusRunning {
		r.upTaskRetryTimes()
	}

	if r.RetryTimes > 1 {
		r.Logger.Info(fmt.Sprintf("wf runtime is %d times retry !!", r.RetryTimes))
	}

	for i := 0; i < len(flowMeta.Steps); i++ {
		stepMeta := flowMeta.Steps[i]
		stepLogger := r.Logger.WithName(fmt.Sprintf("%d/%s", i+1, stepMeta.StepName))
		// 从工作流运行时中找到当前步骤运行时
		stepRuntime := findStepInRuntime(stepMeta.StepName, r)
		if stepRuntime == nil {
			errMsg := fmt.Sprintf("step %s is not found in workflow runtime, maybe workflow meta steps changed.", stepMeta.StepName)
			err = errors.New(errMsg)
			stepLogger.Error(err, "")
			return err
		}
		if stepRuntime.StepStatus == StepStatusCompleted {
			stepLogger.Info("find flowMeta already completed, continue!")
			continue
		}
		runningStepRuntime = stepRuntime
		if runningStepRuntime.RetryTimes >= FlowMaxAutoRetryTimes {
			stepLogger.Error(errors.New(r.ErrorMessage), "too many times retry", "retryTimes:", r.RetryTimes)
			return define.NewInterruptError("MANY_RETRY", fmt.Sprintf("[%s]too many times retry [%d] to do [%s], last error step [%s], err: [%s]", rsName, r.RetryTimes, r.FlowName, stepMeta.StepName, r.ErrorMessage))
		}

		stepContext := r.getStepContext(i, flowMeta)
		stepActionIns, err := r.NewStepActionIns(stepMeta)
		if err != nil {
			return err
		}
		if runningStepRuntime.StepStatus != StepStatusWaiting {
			// 初始化当前步骤运行时
			runningStepRuntime.setInit()
			if err = r.setRunning(runningStepRuntime); err != nil {
				return err
			}
			// 执行步骤逻辑
			stepLogger.Info("prepare to do step", "context", stepContext)
		}
		err, outputContext := r.doStepIns(ctx, stepLogger, stepActionIns, stepContext, stepMeta)
		if define.IsWaitingError(err) {
			stepLogger.Info(err.Error())
			if runningStepRuntime.StepStatus == StepStatusWaiting && r.FlowStatus == FlowStatusWaiting {
				return nil
			}
			if err = r.setStepWaiting(runningStepRuntime, stepLogger); err != nil {
				return err
			}
			return nil
		}
		if err != nil {
			return err
		}

		// 获取步骤执行结果
		runningStepRuntime.setOutputContext(outputContext)
		stepLogger.Info("do step set output context success", "outputContext", outputContext)

		if err = r.setStepComplete(runningStepRuntime, stepLogger, outputContext); err != nil {
			return err
		}
	}
	runningStepRuntime = nil
	err = r.setComplete()
	return err
}

func (r *WfRuntime) getResourceFullName() (string, error) {
	if r.InitContext == nil {
		return "", errors.New("initContext is nil")
	}

	rsNs, ok := r.InitContext[r.ResourceWorkflow.WfManager.GetConfItem(define.WorkFlowResourceNameSpace)]
	if !ok || rsNs == "" {
		return "", errors.New(fmt.Sprintf("initContext.%s is empty", r.ResourceWorkflow.WfManager.GetConfItem(define.WorkFlowResourceNameSpace)))
	}

	rsName, ok := r.InitContext[r.ResourceWorkflow.WfManager.GetConfItem(define.WorkFlowResourceName)]
	if !ok || rsName == "" {
		return "", errors.New(fmt.Sprintf("initContext.%s is empty", r.ResourceWorkflow.WfManager.GetConfItem(define.WorkFlowResourceName)))
	}

	return fmt.Sprintf("%s/%s", rsNs, rsName), nil
}

func (r *WfRuntime) doStepIns(ctx context.Context, stepLogger logr.Logger, stepIns StepAction, context map[string]interface{}, stepMeta *StepMeta) (err error, output map[string]interface{}) {
	if err = r.initStepIns(stepLogger, stepIns, context, stepMeta); err != nil {
		return err, nil
	}
	if err = r.doInitedStepIns(ctx, stepLogger, stepIns, context, stepMeta); err != nil {
		return err, nil
	}
	outputContext := stepIns.Output(stepLogger)
	return nil, outputContext
}

func (r *WfRuntime) initStepIns(stepLogger logr.Logger, stepIns StepAction, context map[string]interface{}, stepMeta *StepMeta) error {
	if err := stepIns.Init(context, stepLogger); err != nil {
		isInterrupt, _ := define.IsInterruptError(err)
		if isInterrupt {
			stepLogger.Error(err, "init step context failed and interrupt", "context", context)
			return err
		}
		err = errors.Wrap(err, fmt.Sprintf("init step context fail for %v", stepMeta))
		return err
	}
	stepLogger.Info("do step init complete !")
	return nil
}

func (r *WfRuntime) doInitedStepIns(ctx context.Context, stepLogger logr.Logger, stepIns StepAction, context map[string]interface{}, stepMeta *StepMeta) error {
	stepStart := time.Now()
	var err error
	done := make(chan error, 1)
	go func() {
		defer close(done)
		done <- stepIns.DoStep(ctx, stepLogger)
	}()
	select {
	case <-ctx.Done():
		err = define.NewInterruptError("CANCELLED", "workflow cancelled!")
	case result := <-done:
		err = result
	}
	stepTimeCost := time.Since(stepStart)
	stepLogger.Info(fmt.Sprintf("[time cost] %f s", stepTimeCost.Seconds()))
	if err != nil {
		isInterrupt, _ := define.IsInterruptError(err)
		if isInterrupt {
			stepLogger.Error(err, "do step fail and interrupt !", "context", context)
			return err
		}
		if define.IsWaitingError(err) {
			return err
		}

		err = errors.Wrap(err, fmt.Sprintf("step %s fail", stepMeta.ClassName))

		if r.RetryTimes >= FlowMaxAutoRetryTimes {
			stepLogger.Error(err, "do step fail, too many times retry", "retryTimes", r.RetryTimes, "context", context)
			return define.NewInterruptError("MANY_RETRY", fmt.Sprintf("too many times retry [%d] to do [%s], last err step[%s], err:[%v]", r.RetryTimes, r.FlowName, stepMeta.StepName, err))
		}
		return err
	}
	stepLogger.Info("do step succeed", "context", context)
	return nil
}

func (r *WfRuntime) setStepComplete(runningStepRuntime *StepRuntime, stepLogger logr.Logger, outputContext map[string]interface{}) error {
	runningStepRuntime.setCompleted()
	if err := r.setStepCompleted(runningStepRuntime); err != nil {
		return err
	}
	stepLogger.Info("do step set completed succeed, ", "outputContext", outputContext)
	return nil
}

func (r *WfRuntime) setStepWaiting(runningStepRuntime *StepRuntime, stepLogger logr.Logger) error {
	runningStepRuntime.setWaiting()
	if err := r.setWaiting(runningStepRuntime); err != nil {
		return err
	}
	stepLogger.Info("do step set waiting succeed, ")
	return nil
}

func (r *WfRuntime) NewStepActionIns(stepMeta *StepMeta) (StepAction, error) {
	stepActionClass, ok := r.ResourceWorkflow.WfManager.TypeRegistry[stepMeta.ClassName]
	if !ok {
		return nil, define.NewInterruptError("UNKNOWN_TASK", fmt.Sprintf("[%s %s]can not find step meta in type resistry!", stepMeta.StepName, stepMeta.ClassName))
	}

	stepActionInsTmp := reflect.New(stepActionClass.Elem()).Interface()
	stepIns := stepActionInsTmp.(StepAction)
	return stepIns, nil
}

func (r *WfRuntime) getStepContext(i int, flowMetaIns *FlowMeta) map[string]interface{} {
	var stepContext map[string]interface{}
	if i == 0 {
		stepContext = r.InitContext
	} else {
		contextTmp := findStepInRuntime(flowMetaIns.Steps[i-1].StepName, r).ContextOutput
		stepContext = utils.MapMerge(r.InitContext, contextTmp)
	}
	return stepContext
}

func (r *WfRuntime) setStepAndFlowFailed(err error, runningStep *StepRuntime) {
	r.Logger.Error(err, "flow run failed")
	if runningStep != nil {
		runningStep.setFail()
	}
	if hErr := r.setFail(err, runningStep); hErr != nil {
		r.Logger.Error(err, "flow setFail error")
	}
}

func (r *WfRuntime) getRecoverError(re interface{}) (err error) {
	r.Logger.Info("recovered wf runtime")
	debug.PrintStack()
	switch x := re.(type) {
	case error:
		err = x
	default:
		err = errors.New(fmt.Sprintf("%v", x))
	}
	return err
}

func (r *WfRuntime) upTaskRetryTimes() {
	r.RetryTimes += 1
}

func (r *WfRuntime) setWaiting(runningStep *StepRuntime) error {
	r.FlowStatus = FlowStatusWaiting
	r.ErrorMessage = ""
	if err := r.saveMemento(false); err != nil {
		return err
	}
	if r.FlowHookIns != nil {
		if err := r.FlowHookIns.OnStepWaiting(runningStep); err != nil {
			return errors.New(fmt.Sprintf("workflow runtime waiting, but do hook fail %v", err))
		}
	}
	return nil
}

func (r *WfRuntime) setComplete() error {
	r.FlowStatus = FlowStatusCompleted
	r.CompleteTime = time.Now().Format("2006-01-02 15:04:05")
	r.ErrorMessage = ""
	if err := r.saveMemento(false); err != nil {
		return err
	}
	if r.FlowHookIns != nil {
		if err := r.FlowHookIns.OnWfCompleted(); err != nil {
			return errors.New(fmt.Sprintf("workflow runtime completed, but do hook fail %v", err))
		}
	}
	return nil
}

func (r *WfRuntime) setRunning(runningStep *StepRuntime) error {
	r.FlowStatus = FlowStatusRunning
	r.ErrorMessage = ""
	if err := r.saveMemento(false); err != nil {
		return err
	}
	if r.FlowHookIns != nil {
		if err := r.FlowHookIns.OnStepInit(runningStep); err != nil {
			return errors.New(fmt.Sprintf("do step prepare done, but do hook fail %v", err))
		}
	}
	return nil
}

func (r *WfRuntime) setStepCompleted(runningStep *StepRuntime) error {
	r.FlowStatus = FlowStatusRunning
	r.ErrorMessage = ""
	if err := r.saveMemento(false); err != nil {
		return err
	}
	if r.FlowHookIns != nil {
		if err := r.FlowHookIns.OnStepCompleted(runningStep); err != nil {
			return errors.New(fmt.Sprintf("do step prepare done, but do hook fail %v", err))
		}
	}
	return nil
}

func (r *WfRuntime) setFail(err error, runningStep *StepRuntime) error {
	r.ErrorMessage = fmt.Sprintf("%v", err)

	//增加对流程中断状态的标注
	isInterrupted, brkErr := define.IsInterruptError(err)
	if isInterrupted {
		r.FlowStatus = FlowStatusFailedCompleted
	} else {
		r.FlowStatus = FlowStatusFailed
	}
	if isInterrupted && brkErr != nil {
		err = r.updateInterruptInfoToResource(brkErr)
		if err != nil {
			return err
		}
	}
	if err := r.saveMemento(false); err != nil {
		return err
	}
	if r.FlowHookIns != nil {
		if hErr := r.FlowHookIns.OnStepCompleted(runningStep); hErr != nil {
			r.Logger.Error(err, "flow step run failed, and do hook function fail", "stepName", runningStep.StepName)
			return hErr
		}
	}
	if isInterrupted && brkErr != nil && r.FlowHookIns != nil {
		iErr := r.FlowHookIns.OnWfInterrupt(brkErr)
		if iErr != nil {
			r.Logger.Error(err, "flow step run failed, and do interrupted hook function fail", "stepName", runningStep.StepName, "error", iErr)
			return iErr
		}
	}

	return nil
}

func (r *WfRuntime) updateInterruptInfoToResource(brkErr *define.InterruptError) error {
	// set Resource interrupt
	prevState := r.ResourceWorkflow.Resource.GetState()
	if prevState == "" {
		prevState = statemachine.StateCreating
	}
	if err := r.ResourceWorkflow.WfManager.recover.SaveResourceInterruptInfo(
		r.ResourceWorkflow.Resource,
		r.ResourceWorkflow.WfManager.config,
		brkErr.Reason,
		brkErr.ErrorMsg,
		string(prevState),
	); err != nil {
		return err
	}
	if _, err := r.ResourceWorkflow.Resource.UpdateState(statemachine.StateInterrupt); err != nil {
		return err
	}
	return nil
}

func findStepInRuntime(stepName string, flow *WfRuntime) *StepRuntime {
	for _, step := range flow.RunningSteps {
		if step.StepName == stepName {
			return step
		}
	}
	return nil
}

func (r *WfRuntime) saveMemento(isNew bool) error {
	if r.ResourceWorkflow.MementoCareTaker != nil {
		var memento *WfRuntimeMemento
		flowName := r.FlowName
		runtimeJson, err := json.Marshal(r)
		if err != nil {
			return err
		}
		if isNew {
			memento, err = r.ResourceWorkflow.MementoCareTaker.CreateMemento(flowName)
			if err != nil {
				r.Logger.Error(err, "failed to crete workflow runtime memento", "initContext", r.InitContext)
				return err
			}
			r.InitContext[flowSaveResourceNameKey] = memento.name
			r.InitContext[flowSaveResourceNameSpaceKey] = memento.namespace
			r.InitContext[flowSaveNumberKey] = memento.mementoKey
		} else {
			memento, err = r.ResourceWorkflow.MementoCareTaker.LoadMemento(r)
			if err != nil {
				r.Logger.Error(err, "failed to load workflow runtime memento", "initContext", r.InitContext)
				return err
			}
		}
		memento.content = string(runtimeJson)
		if err := r.ResourceWorkflow.MementoCareTaker.SaveMemento(memento); err != nil {
			r.Logger.Error(err, "failed to save workflow runtime meta data", "initContext", r.InitContext)
			return err
		}
	}
	return nil
}

func CreateWfRuntime(
	flowName string,
	flowContext map[string]interface{},
	wf *ResourceWorkflow,
) (*WfRuntime, error) {
	logger := wf.Logger.WithName(flowName)
	if flowContext == nil {
		flowContext = map[string]interface{}{}
	}
	wfRuntime := &WfRuntime{
		FlowStatus:       FlowStatusPrepared,
		FlowName:         flowName,
		StartTime:        time.Now().Format("2006-01-02 15:04:05.000"),
		RunningSteps:     make([]*StepRuntime, 0),
		InitContext:      flowContext,
		RetryTimes:       0,
		IgnoreError:      false,
		ResourceWorkflow: wf,
		Logger:           logger,
	}

	for _, step := range wf.WfManager.FlowMetaMap[flowName].Steps {
		wfRuntime.RunningSteps = append(wfRuntime.RunningSteps, &StepRuntime{StepName: step.StepName, StepStatus: StepStatusPrepared})
	}

	var err error
	if wf.WfManager.createHookFunc != nil {
		if wfRuntime.FlowHookIns, err = wf.WfManager.createHookFunc(wfRuntime); err != nil {
			return nil, err
		}
	}
	if err := wfRuntime.saveMemento(true); err != nil {
		return nil, err
	}
	if wfRuntime.FlowHookIns != nil {
		if err := wfRuntime.FlowHookIns.OnWfInit(); err != nil {
			return nil, err
		}
	}
	return wfRuntime, nil
}

func LoadUnCompletedWfRuntime(wfRuntime *WfRuntime, wf *ResourceWorkflow) error {
	var err error
	logger := wf.Logger.WithName(wfRuntime.FlowName)
	wfRuntime.Logger = logger
	wfRuntime.ResourceWorkflow = wf
	if wf.WfManager.createHookFunc != nil {
		if wfRuntime.FlowHookIns, err = wf.WfManager.createHookFunc(wfRuntime); err != nil {
			return err
		}
	}
	if err := wfRuntime.saveMemento(false); err != nil {
		return err
	}
	if err := wfRuntime.FlowHookIns.OnWfInit(); err != nil {
		return err
	}
	return nil
}
