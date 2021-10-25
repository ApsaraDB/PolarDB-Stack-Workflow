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

package statemachine

import (
	"errors"
	"fmt"

	"github.com/ApsaraDB/PolarDB-Stack-Workflow/utils"
	"github.com/go-logr/logr"
	"k8s.io/klog/klogr"
)

type State string

const (
	StateRunning   State = "Running"
	StateCreating  State = "Creating"
	StateInit      State = "Init"
	StateInterrupt State = "Interrupt"
	StateRebuild   State = "Rebuild"
	StateNil       State = ""
)

// 支持状态机的资源
type StateResource interface {
	GetName() string
	GetNamespace() string
	Fetch() (StateResource, error)
	GetState() State
	UpdateState(State) (StateResource, error)
	IsCancelled() bool
}

type StateMainEnter interface {
	MainEnter(obj StateResource) error
	GetName() string
}

type AnonymousStateMainEnter struct {
	mainEnter func(obj StateResource) error
	name      string
}

func (e *AnonymousStateMainEnter) MainEnter(obj StateResource) error {
	return e.mainEnter(obj)
}

func (e *AnonymousStateMainEnter) GetName() string {
	return e.name
}

type stateTranslate struct {
	eventChecker func(object StateResource) (*Event, error) // 事件检测方法
	targetState  State                                      // eventCheckFunc检测触发事件，资源进入targetState状态
}

type StateMachine struct {
	ignoreEvents      []string                    //忽略的事件
	stateTranslateMap map[State][]*stateTranslate // 状态转换表
	stateMainEnterMap map[State]StateMainEnter    // 进入状态会执行的动作
	Logger            logr.Logger
	ResourceType      string
}

func (sm *StateMachine) RegisterIgnoreEvents(ignoreEvents []string) *StateMachine {
	sm.ignoreEvents = ignoreEvents
	return sm
}

func (sm *StateMachine) RegisterLogger(logger logr.Logger) *StateMachine {
	sm.Logger = logger
	return sm
}

func (sm *StateMachine) DoStateMainEnter(obj StateResource) error {
	if sm.Logger == nil {
		sm.Logger = klogr.New().WithName(fmt.Sprintf("statemachine/%s/%s/%s", sm.ResourceType, obj.GetNamespace(), obj.GetName()))
	} else {
		sm.Logger = sm.Logger.WithName("statemachine")
	}
	objMu := ObjectLockIns.GetLock(obj)
	objMu.Lock()
	defer objMu.Unlock()

	reloadObj, err := obj.Fetch()
	if err != nil {
		return err
	}
	state := reloadObj.GetState()
	if state == "" {
		state = StateCreating
	}
	if stateMainEnter, ok := sm.stateMainEnterMap[state]; !ok {
		err := errors.New(fmt.Sprintf("%v status is not registered", state))
		sm.Logger.Error(err, "")
		return err
	} else {
		sm.Logger.Info(fmt.Sprintf("start do [%s]: %v main enter", stateMainEnter.GetName(), state))
		return stateMainEnter.MainEnter(reloadObj)
	}
}

func (sm *StateMachine) RegisterStateTranslateMainEnter(beforeState State, eventChecker EventChecker, afterState State, action func(obj StateResource) error) {
	sm.RegisterStateTranslate(beforeState, eventChecker, afterState)
	sm.RegisterStateMainEnter(afterState, action)
}

// 资源处于稳定状态(beforeState)，通过eventCheckFunc检查是否发生了更新，发生更新则触发事件，并将资源状态变为非稳定态(afterState)
// 非稳定态都通过RegisterStateMainEnter注册了入口方法
func (sm *StateMachine) RegisterStateTranslate(beforeState State, eventChecker EventChecker, afterState State) {
	if sm.stateTranslateMap == nil {
		sm.stateTranslateMap = map[State][]*stateTranslate{}
	}
	eventTransIns := &stateTranslate{
		eventChecker: eventChecker,
		targetState:  afterState,
	}
	sm.stateTranslateMap[beforeState] = append(sm.stateTranslateMap[beforeState], eventTransIns)
}

// 注册资源进入某非稳定状态要执行的动作
func (sm *StateMachine) RegisterStateMainEnter(state State, action func(obj StateResource) error) {
	if sm.stateMainEnterMap == nil {
		sm.stateMainEnterMap = map[State]StateMainEnter{}
	}
	sm.stateMainEnterMap[state] = &AnonymousStateMainEnter{mainEnter: action, name: string(state)}
}

// 注册稳定状态, 等待有事件触发，离开当前状态.
func (sm *StateMachine) RegisterStableState(states ...State) {
	if sm.stateMainEnterMap == nil {
		sm.stateMainEnterMap = map[State]StateMainEnter{}
	}
	for _, state := range states {
		sm.stateMainEnterMap[state] = &StableStateMainEnter{sm, state}
	}
}

// 检查当前状态下是否有需要transfer的事件发生. 如果发现 state change, 设置变更后的状态.
func (sm *StateMachine) CheckSetStateTranslate(nowState State, obj StateResource) (State, error) {
	translateMap, ok := sm.stateTranslateMap[nowState]
	if !ok {
		return StateNil, nil
	}

	var checkErr error
	for _, translate := range translateMap {
		if event, err := translate.eventChecker(obj); err != nil {
			checkErr = err
			sm.Logger.Error(checkErr, "check event translate failed", "event", event)
			continue
		} else if event != nil {
			if sm.ignoreEvents != nil && utils.ContainsString(sm.ignoreEvents, string(event.Name)) {
				sm.Logger.Info("ignore event for resource", "event", event.Name)
				return StateNil, nil
			}
			sm.Logger.Info(fmt.Sprintf("set cluster status from %v to %v", nowState, translate.targetState))

			//直接设置为目标状态，触发相应的工作流。
			if _, err := obj.UpdateState(translate.targetState); err != nil {
				return StateNil, err
			}
			return translate.targetState, nil
		}
	}
	if checkErr != nil {
		sm.Logger.Error(checkErr, "invoker all check func not found one success")
		return StateNil, checkErr
	}
	return StateNil, nil
}

func CreateStateMachineInstance(resourceType string) *StateMachine {
	return &StateMachine{ResourceType: resourceType}
}
