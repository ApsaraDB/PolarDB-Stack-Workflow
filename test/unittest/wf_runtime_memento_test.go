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


package unittest

import (
	"testing"

	"gitlab.alibaba-inc.com/polar-as/polar-wf-engine/define"

	"github.com/pkg/errors"
	"gitlab.alibaba-inc.com/polar-as/polar-wf-engine/statemachine"

	. "github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
	"gitlab.alibaba-inc.com/polar-as/polar-wf-engine/test/mock"
	"gitlab.alibaba-inc.com/polar-as/polar-wf-engine/wfengine"
)

func Test_CreateWfRuntimeMementoCareTaker(t *testing.T) {
	Convey("CreateWfRuntimeMementoCareTaker succeed", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")
		resource.EXPECT().GetNamespace().Return("default")

		careTaker, err := wfengine.CreateWfRuntimeMementoCareTaker(
			resource,
			CreateMementoStorage(ctrl),
		)
		So(careTaker, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})

	Convey("CreateWfRuntimeMementoCareTaker failed", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")

		careTaker, err := wfengine.CreateWfRuntimeMementoCareTaker(
			resource,
			func(resource statemachine.StateResource) (wfengine.WfRuntimeMementoStorage, error) {
				mementoStorage := mock.NewMockWfRuntimeMementoStorage(ctrl)
				mementoStorage.EXPECT().LoadMementoMap(Any()).Return(nil, errors.New("LoadMementoMap failed"))
				return mementoStorage, nil
			},
		)
		So(careTaker, ShouldBeNil)
		So(err, ShouldNotBeNil)
	})
}

func Test_CreateMemento(t *testing.T) {
	Convey("CreateMemento succeed", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")
		resource.EXPECT().GetNamespace().Return("default")

		careTaker, err := wfengine.CreateWfRuntimeMementoCareTaker(
			resource,
			CreateMementoStorage(ctrl),
		)
		careTaker.MementoMap["1-WF1"] = "wf runtime context"

		memento, err := careTaker.CreateMemento("WF1")

		So(memento, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})
}

func Test_GetLastWorkflowRuntime(t *testing.T) {
	Convey("uncompleted workflow runtime not exist", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")
		resource.EXPECT().GetNamespace().Return("default")

		careTaker, err := wfengine.CreateWfRuntimeMementoCareTaker(
			resource,
			CreateMementoStorage(ctrl),
		)

		wfId, wfRuntime, err := careTaker.GetLastWorkflowRuntime()
		So(wfId, ShouldBeEmpty)
		So(wfRuntime, ShouldBeNil)
		So(err, ShouldBeNil)
	})

	Convey("GetLastWorkflowRuntime succeed", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")
		resource.EXPECT().GetNamespace().Return("default")

		careTaker, err := wfengine.CreateWfRuntimeMementoCareTaker(
			resource,
			CreateMementoStorage(ctrl),
		)
		careTaker.MementoMap["1-WF1"] = `{
			"flowStatus":"failed",
			"flowName":"WF1",
			"initContext":{
				"_flowSaveNumberKey":"2-WF1",
				"_flowSaveResourceNameKey":"test-resource-workflow",
				"_flowSaveResourceNameSpaceKey":"default",
				"_resourceName":"test-resource",
				"_resourceNameSpace":"default"
			},
			"runningSteps":[
				{"StepName":"Step1","stepStatus":"completed"},
				{"StepName":"Step2","stepStatus":"failed","retryTimes":2}
			]
		}`

		wfId, wfRuntime, err := careTaker.GetLastWorkflowRuntime()
		So(wfId, ShouldNotBeEmpty)
		So(wfRuntime, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})

	Convey("invalid JSON", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")
		resource.EXPECT().GetNamespace().Return("default")

		careTaker, err := wfengine.CreateWfRuntimeMementoCareTaker(
			resource,
			CreateMementoStorage(ctrl),
		)
		careTaker.MementoMap["1-WF1"] = `{invalid json}`

		wfId, wfRuntime, err := careTaker.GetLastWorkflowRuntime()
		So(wfId, ShouldBeEmpty)
		So(wfRuntime, ShouldBeNil)
		So(err, ShouldNotBeNil)
	})

	Convey("invalid workflow runtime id", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")
		resource.EXPECT().GetNamespace().Return("default")

		careTaker, err := wfengine.CreateWfRuntimeMementoCareTaker(
			resource,
			CreateMementoStorage(ctrl),
		)
		careTaker.MementoMap["invalidId"] = `{}`

		wfId, wfRuntime, err := careTaker.GetLastWorkflowRuntime()
		So(wfId, ShouldBeEmpty)
		So(wfRuntime, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "invalid syntax")
	})
}

func Test_GetAllUnCompletedWorkflowRuntime(t *testing.T) {
	Convey("get all uncompleted workflow runtime succeed", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")
		resource.EXPECT().GetNamespace().Return("default")

		careTaker, err := wfengine.CreateWfRuntimeMementoCareTaker(
			resource,
			CreateMementoStorage(ctrl),
		)
		careTaker.MementoMap["1-WF1"] = `{
			"flowStatus":"failed",
			"flowName":"WF1",
			"initContext":{
				"_flowSaveNumberKey":"1-WF1",
				"_flowSaveResourceNameKey":"test-resource-workflow",
				"_flowSaveResourceNameSpaceKey":"default",
				"_resourceName":"test-resource",
				"_resourceNameSpace":"default"
			},
			"runningSteps":[
				{"StepName":"Step1","stepStatus":"completed"},
				{"StepName":"Step2","stepStatus":"failed","retryTimes":2}
			]
		}`
		careTaker.MementoMap["2-WF1"] = `{
			"flowStatus":"completed",
			"flowName":"WF1",
			"initContext":{
				"_flowSaveNumberKey":"2-WF1",
				"_flowSaveResourceNameKey":"test-resource-workflow",
				"_flowSaveResourceNameSpaceKey":"default",
				"_resourceName":"test-resource",
				"_resourceNameSpace":"default"
			},
			"runningSteps":[
				{"StepName":"Step1","stepStatus":"completed"},
				{"StepName":"Step2","stepStatus":"completed"}
			]
		}`
		wfRuntimes, err := careTaker.GetAllUnCompletedWorkflowRuntime()
		So(len(wfRuntimes), ShouldEqual, 1)
		So(err, ShouldBeNil)
	})

	Convey("invalid JSON", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")
		resource.EXPECT().GetNamespace().Return("default")

		careTaker, err := wfengine.CreateWfRuntimeMementoCareTaker(
			resource,
			CreateMementoStorage(ctrl),
		)
		careTaker.MementoMap["1-WF1"] = `{invalid json}`

		wfRuntimes, err := careTaker.GetAllUnCompletedWorkflowRuntime()
		So(wfRuntimes, ShouldBeNil)
		So(err, ShouldNotBeNil)
	})

}

func Test_LoadMemento(t *testing.T) {
	Convey("load memento succeed", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")
		resource.EXPECT().GetNamespace().Return("default")

		careTaker, err := wfengine.CreateWfRuntimeMementoCareTaker(
			resource,
			CreateMementoStorage(ctrl),
		)
		wfManager := &wfengine.WfManager{}
		wfManager.RegisterConf(define.DefaultWfConf)
		resourWf := &wfengine.ResourceWorkflow{
			WfManager: wfManager,
		}
		wfRuntime := &wfengine.WfRuntime{
			FlowName: "WF1",
			InitContext: map[string]interface{}{
				"_resourceName":                 "test-resource",
				"_resourceNameSpace":            "default",
				"_flowSaveNumberKey":            "2-WF1",
				"_flowSaveResourceNameKey":      "test-resource-workflow",
				"_flowSaveResourceNameSpaceKey": "default",
			},
			ResourceWorkflow: resourWf,
		}
		memento, err := careTaker.LoadMemento(wfRuntime)
		So(memento, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})

	Convey("failed to get flowSaveResourceNameSpaceKey by context", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")
		resource.EXPECT().GetNamespace().Return("default")

		careTaker, err := wfengine.CreateWfRuntimeMementoCareTaker(
			resource,
			CreateMementoStorage(ctrl),
		)
		wfManager := &wfengine.WfManager{}
		wfManager.RegisterConf(define.DefaultWfConf)
		resourWf := &wfengine.ResourceWorkflow{
			WfManager: wfManager,
		}
		wfRuntime := &wfengine.WfRuntime{
			FlowName: "WF1",
			InitContext: map[string]interface{}{
				"_resourceName":            "test-resource",
				"_resourceNameSpace":       "default",
				"_flowSaveNumberKey":       "2-WF1",
				"_flowSaveResourceNameKey": "test-resource-workflow",
			},
			ResourceWorkflow: resourWf,
		}
		memento, err := careTaker.LoadMemento(wfRuntime)
		So(memento, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "failed to get flowSaveResourceNameSpaceKey")
	})

	Convey("failed to get flowSaveResourceNameKey by context", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")
		resource.EXPECT().GetNamespace().Return("default")
		wfManager := &wfengine.WfManager{}
		wfManager.RegisterConf(define.DefaultWfConf)
		resourWf := &wfengine.ResourceWorkflow{
			WfManager: wfManager,
		}
		careTaker, err := wfengine.CreateWfRuntimeMementoCareTaker(
			resource,
			CreateMementoStorage(ctrl),
		)
		wfRuntime := &wfengine.WfRuntime{
			FlowName: "WF1",
			InitContext: map[string]interface{}{
				"_resourceName":                 "test-resource",
				"_resourceNameSpace":            "default",
				"_flowSaveNumberKey":            "2-WF1",
				"_flowSaveResourceNameSpaceKey": "default",
			},
			ResourceWorkflow: resourWf,
		}
		memento, err := careTaker.LoadMemento(wfRuntime)
		So(memento, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "failed to get flowSaveResourceNameKey")
	})

	Convey("failed to get flowSaveNumberKey by context", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")
		resource.EXPECT().GetNamespace().Return("default")

		wfManager := &wfengine.WfManager{}
		wfManager.RegisterConf(define.DefaultWfConf)
		resourWf := &wfengine.ResourceWorkflow{
			WfManager: wfManager,
		}

		careTaker, err := wfengine.CreateWfRuntimeMementoCareTaker(
			resource,
			CreateMementoStorage(ctrl),
		)
		wfRuntime := &wfengine.WfRuntime{
			FlowName: "WF1",
			InitContext: map[string]interface{}{
				"_resourceName":                 "test-resource",
				"_resourceNameSpace":            "default",
				"_flowSaveResourceNameKey":      "test-resource-workflow",
				"_flowSaveResourceNameSpaceKey": "default",
			},
			ResourceWorkflow: resourWf,
		}
		memento, err := careTaker.LoadMemento(wfRuntime)
		So(memento, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "failed to get flowSaveNumberKey")
	})

	Convey("failed to get resourceName by context", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")
		resource.EXPECT().GetNamespace().Return("default")

		wfManager := &wfengine.WfManager{}
		wfManager.RegisterConf(define.DefaultWfConf)
		resourWf := &wfengine.ResourceWorkflow{
			WfManager: wfManager,
		}

		careTaker, err := wfengine.CreateWfRuntimeMementoCareTaker(
			resource,
			CreateMementoStorage(ctrl),
		)
		wfRuntime := &wfengine.WfRuntime{
			FlowName: "WF1",
			InitContext: map[string]interface{}{
				"_resourceNameSpace":            "default",
				"_flowSaveNumberKey":            "2-WF1",
				"_flowSaveResourceNameKey":      "test-resource-workflow",
				"_flowSaveResourceNameSpaceKey": "default",
			},
			ResourceWorkflow: resourWf,
		}
		memento, err := careTaker.LoadMemento(wfRuntime)
		So(memento, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "failed to get resourceName")
	})
}
