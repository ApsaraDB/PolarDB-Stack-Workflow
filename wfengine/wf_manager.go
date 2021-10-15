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
	"fmt"
	"reflect"
	"strings"

	"gitlab.alibaba-inc.com/polar-as/polar-wf-engine/define"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"gitlab.alibaba-inc.com/polar-as/polar-wf-engine/statemachine"
	"k8s.io/klog/klogr"
)

type WfManager struct {
	ResourceType             string
	FlowMetaMap              map[string]*FlowMeta
	TypeRegistry             map[string]reflect.Type
	createHookFunc           CreateHookFunc
	createMementoStorageFunc CreateMementoStorageFunc
	Logger                   logr.Logger
	recover                  Recover
	config                   map[define.WFConfKey]string
}

type FlowMeta struct {
	FlowName             string      `yaml:"flowName,omitempty"`
	RecoverFromFirstStep bool        `yaml:"recoverFromFirstStep,omitempty"`
	Steps                []*StepMeta `yaml:"steps,omitempty"`
}

type StepGroupMeta struct {
	GroupName string      `yaml:"groupName,omitempty"`
	Steps     []*StepMeta `yaml:"steps,omitempty"`
}

type StepMeta struct {
	ClassName string `yaml:"className,omitempty"`
	StepName  string `yaml:"stepName,omitempty"`
}

type CreateMetaLoaderFunc func(resourceType string) (WfMetaLoader, error)
type CreateHookFunc func(wfRuntime *WfRuntime) (WfHook, error)
type CreateMementoStorageFunc func(resource statemachine.StateResource) (WfRuntimeMementoStorage, error)

func CreateWfManager(
	resourceType string,
	workFlowMetaDir string,
	createMetaLoaderFunc CreateMetaLoaderFunc,
	createHookFunc CreateHookFunc,
	createMementoStorageFunc CreateMementoStorageFunc,
) (*WfManager, error) {
	var (
		err              error
		wfMetaLoader     WfMetaLoader
		flowMetaMap      map[string]*FlowMeta
		stepGroupMetaMap map[string]*StepGroupMeta
	)
	logger := klogr.New().WithName(fmt.Sprintf("workflow/%s", resourceType))
	if resourceType == "" {
		err = errors.New("ResourceType is empty")
		logger.Error(err, "")
		return nil, err
	}
	if workFlowMetaDir == "" {
		err = errors.New("WorkFlowMetaDir is empty")
		logger.Error(err, "")
		return nil, err
	}
	if createMetaLoaderFunc == nil {
		err = errors.New("createMetaLoaderFunc is nil")
		logger.Error(err, "")
		return nil, err
	}
	wfMetaLoader, err = createMetaLoaderFunc(resourceType)
	if err != nil {
		return nil, err
	}
	flowMetaMap, stepGroupMetaMap, err = wfMetaLoader.GetAllFlowMeta(workFlowMetaDir)
	if err != nil {
		return nil, err
	}
	replaceGroupStep(flowMetaMap, stepGroupMetaMap)
	return &WfManager{
		ResourceType:             resourceType,
		FlowMetaMap:              flowMetaMap,
		TypeRegistry:             make(map[string]reflect.Type),
		createHookFunc:           createHookFunc,
		createMementoStorageFunc: createMementoStorageFunc,
		Logger:                   logger,
		config:                   define.DefaultWfConf,
	}, nil
}

func replaceGroupStep(flowMetaMap map[string]*FlowMeta, stepGroupMetaMap map[string]*StepGroupMeta) {
	if len(stepGroupMetaMap) == 0 {
		return
	}
	for _, flowMeta := range flowMetaMap {
		replaceStep(flowMeta, stepGroupMetaMap)
	}
}

func replaceStep(flowMeta *FlowMeta, stepGroupMetaMap map[string]*StepGroupMeta) {
	var (
		replace       bool
		flowStepIndex int
		flowStep      *StepMeta
		stepGroup     *StepGroupMeta
	)
	for flowStepIndex, flowStep = range flowMeta.Steps {
		if stepGroup, replace = stepGroupMetaMap[flowStep.ClassName]; replace {
			break
		}
	}
	if replace && stepGroup != nil && flowStep != nil {
		rear := append([]*StepMeta{}, flowMeta.Steps[flowStepIndex+1:]...)
		steps := append(flowMeta.Steps[:flowStepIndex], stepGroup.Steps...)
		steps = append(steps, rear...)
		flowMeta.Steps = steps
		replaceStep(flowMeta, stepGroupMetaMap)
	}
}

func (m *WfManager) RegisterSteps(steps ...StepAction) {
	for _, step := range steps {
		m.RegisterStep(step)
	}
}

func (m *WfManager) RegisterStep(step StepAction) {
	stepStructName := strings.Replace(reflect.TypeOf(step).String(), "*", "", -1)
	m.TypeRegistry[stepStructName] = reflect.TypeOf(step)
}

func (m *WfManager) CreateResourceWorkflow(resource statemachine.StateResource) (*ResourceWorkflow, error) {
	return CreateResourceWorkflow(resource, m)
}

func (m *WfManager) RegisterLogger(logger logr.Logger) *WfManager {
	m.Logger = logger.WithName(fmt.Sprintf("workflow"))
	return m
}

func (m *WfManager) RegisterRecover(recover Recover) *WfManager {
	m.recover = recover
	return m
}

func (m *WfManager) RegisterConf(conf map[define.WFConfKey]string) {
	for key, value := range conf {
		if m.config == nil {
			m.config = define.DefaultWfConf
		}
		m.config[key] = value
	}
}

func (m *WfManager) GetConfItem(confKey define.WFConfKey) string {
	return m.config[confKey]
}

func (m *WfManager) CheckRecovery(resource statemachine.StateResource) bool {
	var err error
	if resource.GetState() != statemachine.StateInterrupt {
		return false
	}

	recover, prevStatus := m.recover.GetResourceRecoverInfo(resource, m.config)
	if !recover {
		return false
	}

	var setPrevStatus = statemachine.StateRunning
	if prevStatus != statemachine.StateNil {
		if prevStatus != statemachine.StateRebuild {
			// 回到上一状态
			resourceWf, err := m.CreateResourceWorkflow(resource)
			if err != nil {
				m.Logger.Error(err, "")
				return true
			}
			err = resourceWf.RetryInterruptedStep()
			if err != nil {
				m.Logger.Error(err, "")
				return true
			}
		}
		setPrevStatus = prevStatus
	}
	if err = m.recover.CancelResourceRecover(resource, m.config); err != nil {
		m.Logger.Error(err, "recover interrupted flow, cancel resource recover info failed!")
		return true
	}
	if resource, err = resource.UpdateState(setPrevStatus); err != nil {
		m.Logger.Error(err, "recover interrupted flow, update resource status failed!")
		return true
	}
	m.Logger.Info(fmt.Sprintf("recover interrupted flow, update resource status to %s succeed!", resource.GetState()))
	return true
}
