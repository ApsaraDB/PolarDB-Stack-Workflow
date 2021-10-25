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
	"context"
	j "encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/ApsaraDB/PolarDB-Stack-Workflow/statemachine"
	"github.com/ApsaraDB/PolarDB-Stack-Workflow/test/mock"
	"github.com/ApsaraDB/PolarDB-Stack-Workflow/wfengine"
	"github.com/bouk/monkey"
	"github.com/go-logr/logr"
	. "github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
)

func Test_CreateResourceWf(t *testing.T) {
	Convey("resource is nil", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		wfManager, err := wfengine.CreateWfManager(
			"",
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			CreateHook(ctrl),
			CreateMementoStorage(ctrl),
		)

		resourceWf, err := wfengine.CreateResourceWorkflow(nil, wfManager)
		So(resourceWf, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "resource is nil")
	})

	Convey("wf manager is nil", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()
		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")

		resourceWf, err := wfengine.CreateResourceWorkflow(resource, nil)
		So(resourceWf, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "WfManager is nil")
	})

	Convey("create succeed", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			CreateHook(ctrl),
			CreateMementoStorage(ctrl),
		)

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")

		resourceWf, err := wfengine.CreateResourceWorkflow(resource, wfManager)
		So(resourceWf, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})

	Convey("CreateWfRuntimeMementoCareTaker failed", t, func() {
		errorMsg := "CreateWfRuntimeMementoCareTaker failed"
		ctrl := NewController(t)
		defer ctrl.Finish()

		var createMementoStorage = func(resource statemachine.StateResource) (wfengine.WfRuntimeMementoStorage, error) {
			return nil, errors.New(errorMsg)
		}

		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			CreateHook(ctrl),
			createMementoStorage,
		)

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")

		resourceWf, err := wfengine.CreateResourceWorkflow(resource, wfManager)

		So(resourceWf, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, errorMsg)
	})

}

func Test_RunResourceWorkflow(t *testing.T) {
	Convey("flowName not found", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			CreateHook(ctrl),
			CreateMementoStorage(ctrl),
		)

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")

		resourceWf, err := wfengine.CreateResourceWorkflow(resource, wfManager)
		So(err, ShouldBeNil)
		err = resourceWf.Run(context.TODO(), "NotExistFlow", map[string]interface{}{})
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "not found")
	})

	Convey("run succeed", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			CreateHook(ctrl),
			CreateMementoStorage(ctrl),
		)

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")
		resource.EXPECT().GetState().AnyTimes().Return(statemachine.StateRunning)

		guard := monkey.PatchInstanceMethod(reflect.TypeOf(&wfengine.WfRuntime{}), "Start", func(*wfengine.WfRuntime, context.Context) (err error) {
			return nil
		})
		defer guard.Unpatch()

		resourceWf, err := wfengine.CreateResourceWorkflow(resource, wfManager)
		So(err, ShouldBeNil)
		err = resourceWf.Run(context.TODO(), "WF1", map[string]interface{}{})
		So(err, ShouldBeNil)
	})

	Convey("CreateWfRuntime failed", t, func() {
		errorMsg := "CreateWfRuntime failed"

		ctrl := NewController(t)
		defer ctrl.Finish()

		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			CreateHook(ctrl),
			CreateMementoStorage(ctrl),
		)

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")

		guard := monkey.Patch(wfengine.CreateWfRuntime, func(string, map[string]interface{}, *wfengine.ResourceWorkflow) (*wfengine.WfRuntime, error) {
			return nil, errors.New(errorMsg)
		})
		defer guard.Unpatch()

		resourceWf, err := wfengine.CreateResourceWorkflow(resource, wfManager)
		So(err, ShouldBeNil)
		err = resourceWf.Run(context.TODO(), "WF1", map[string]interface{}{})
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, errorMsg)
	})
}

func Test_GetLastUnCompleted(t *testing.T) {
	Convey("no uncompleted workflow", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			CreateHook(ctrl),
			CreateMementoStorage(ctrl),
		)

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")

		guard := monkey.PatchInstanceMethod(reflect.TypeOf(&wfengine.WfRuntimeMementoCareTaker{}), "GetAllUnCompletedWorkflowRuntime", func(*wfengine.WfRuntimeMementoCareTaker) (map[string]*wfengine.WfRuntime, error) {
			return nil, nil
		})
		defer guard.Unpatch()

		resourceWf, err := wfengine.CreateResourceWorkflow(resource, wfManager)
		wfRuntime, err := resourceWf.GetLastUnCompletedRuntime()

		So(err, ShouldBeNil)
		So(wfRuntime, ShouldBeNil)
	})

	Convey("get all uncompleted workflow failed", t, func() {
		errorMsg := "GetAllUnCompletedWorkflowRuntime failed"
		ctrl := NewController(t)
		defer ctrl.Finish()

		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			CreateHook(ctrl),
			CreateMementoStorage(ctrl),
		)

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")

		guard := monkey.PatchInstanceMethod(reflect.TypeOf(&wfengine.WfRuntimeMementoCareTaker{}), "GetAllUnCompletedWorkflowRuntime", func(*wfengine.WfRuntimeMementoCareTaker) (map[string]*wfengine.WfRuntime, error) {
			return nil, errors.New(errorMsg)
		})
		defer guard.Unpatch()

		resourceWf, err := wfengine.CreateResourceWorkflow(resource, wfManager)
		wfRuntime, err := resourceWf.GetLastUnCompletedRuntime()

		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, errorMsg)
		So(wfRuntime, ShouldBeNil)
	})

	Convey("get last uncompleted workflow succeed", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			CreateHook(ctrl),
			CreateMementoStorage(ctrl),
		)

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")

		guard := monkey.PatchInstanceMethod(reflect.TypeOf(&wfengine.WfRuntimeMementoCareTaker{}), "GetAllUnCompletedWorkflowRuntime", func(*wfengine.WfRuntimeMementoCareTaker) (map[string]*wfengine.WfRuntime, error) {
			return map[string]*wfengine.WfRuntime{
				"1-WF1": {
					FlowStatus: wfengine.FlowStatusFailed,
					FlowName:   "WF1",
				},
				"2-WF1": {
					FlowStatus: wfengine.FlowStatusCompleted,
					FlowName:   "WF2",
				},
				"3-WF1": {
					FlowStatus: wfengine.FlowStatusFailed,
					FlowName:   "WF3",
				},
			}, nil
		})
		defer guard.Unpatch()

		resourceWf, err := wfengine.CreateResourceWorkflow(resource, wfManager)
		wfRuntime, err := resourceWf.GetLastUnCompletedRuntime()

		So(err, ShouldBeNil)
		So(wfRuntime, ShouldNotBeNil)
		So(wfRuntime.FlowName, ShouldContainSubstring, "WF3")
	})

	Convey("getLastUnCompleted failed", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			CreateHook(ctrl),
			CreateMementoStorage(ctrl),
		)

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")

		guard := monkey.PatchInstanceMethod(reflect.TypeOf(&wfengine.WfRuntimeMementoCareTaker{}), "GetAllUnCompletedWorkflowRuntime", func(*wfengine.WfRuntimeMementoCareTaker) (map[string]*wfengine.WfRuntime, error) {
			return map[string]*wfengine.WfRuntime{
				"1WF1": {
					FlowStatus: wfengine.FlowStatusFailed,
					FlowName:   "WF1",
				},
			}, nil
		})
		defer guard.Unpatch()

		resourceWf, err := wfengine.CreateResourceWorkflow(resource, wfManager)
		wfRuntime, err := resourceWf.GetLastUnCompletedRuntime()

		So(err, ShouldNotBeNil)
		So(wfRuntime, ShouldBeNil)
	})
}

func Test_RunLastUnCompletedRuntime(t *testing.T) {
	Convey("run uncompleted workflow before starting the new workflow", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			CreateHook(ctrl),
			CreateMementoStorage(ctrl),
		)

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")

		guard := monkey.PatchInstanceMethod(reflect.TypeOf(&wfengine.WfRuntimeMementoCareTaker{}), "GetAllUnCompletedWorkflowRuntime", func(*wfengine.WfRuntimeMementoCareTaker) (map[string]*wfengine.WfRuntime, error) {
			return map[string]*wfengine.WfRuntime{
				"1-WF1": {
					FlowStatus: wfengine.FlowStatusFailed,
					FlowName:   "WF1",
				},
			}, nil
		})
		defer guard.Unpatch()

		guard1 := monkey.PatchInstanceMethod(reflect.TypeOf(&wfengine.ResourceWorkflow{}), "RunUnCompletedRuntime", func(*wfengine.ResourceWorkflow, context.Context, *wfengine.WfRuntime) error {
			return nil
		})
		defer guard1.Unpatch()

		resourceWf, err := wfengine.CreateResourceWorkflow(resource, wfManager)
		oldWfName, isWaiting, err := resourceWf.RunLastUnCompletedRuntime(context.TODO(), "WF2", false)

		So(err, ShouldBeNil)
		So(isWaiting, ShouldBeFalse)
		So(oldWfName, ShouldContainSubstring, "WF1")
	})

	Convey("retry uncompleted workflow", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			CreateHook(ctrl),
			CreateMementoStorage(ctrl),
		)

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")

		guard := monkey.PatchInstanceMethod(reflect.TypeOf(&wfengine.WfRuntimeMementoCareTaker{}), "GetAllUnCompletedWorkflowRuntime", func(*wfengine.WfRuntimeMementoCareTaker) (map[string]*wfengine.WfRuntime, error) {
			return map[string]*wfengine.WfRuntime{
				"1-WF1": {
					FlowStatus: wfengine.FlowStatusFailed,
					FlowName:   "WF1",
				},
			}, nil
		})
		defer guard.Unpatch()

		guard1 := monkey.PatchInstanceMethod(reflect.TypeOf(&wfengine.ResourceWorkflow{}), "RunUnCompletedRuntime", func(*wfengine.ResourceWorkflow, context.Context, *wfengine.WfRuntime) error {
			return nil
		})
		defer guard1.Unpatch()

		resourceWf, err := wfengine.CreateResourceWorkflow(resource, wfManager)
		oldWfName, isWaiting, err := resourceWf.RunLastUnCompletedRuntime(context.TODO(), "WF1", false)

		So(err, ShouldBeNil)
		So(isWaiting, ShouldBeFalse)
		So(oldWfName, ShouldContainSubstring, "WF1")
	})

	Convey("no uncompleted workflow", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			CreateHook(ctrl),
			CreateMementoStorage(ctrl),
		)

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")

		guard := monkey.PatchInstanceMethod(reflect.TypeOf(&wfengine.WfRuntimeMementoCareTaker{}), "GetAllUnCompletedWorkflowRuntime", func(*wfengine.WfRuntimeMementoCareTaker) (map[string]*wfengine.WfRuntime, error) {
			return map[string]*wfengine.WfRuntime{}, nil
		})
		defer guard.Unpatch()

		resourceWf, err := wfengine.CreateResourceWorkflow(resource, wfManager)
		oldWfName, isWaiting, err := resourceWf.RunLastUnCompletedRuntime(context.TODO(), "WF1", false)

		So(err, ShouldBeNil)
		So(isWaiting, ShouldBeFalse)
		So(oldWfName, ShouldBeEmpty)
	})

	Convey("get last uncompleted workflow failed", t, func() {
		errorMsg := "GetLastUnCompletedRuntime failed"
		ctrl := NewController(t)
		defer ctrl.Finish()

		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			CreateHook(ctrl),
			CreateMementoStorage(ctrl),
		)

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")

		guard := monkey.PatchInstanceMethod(reflect.TypeOf(&wfengine.WfRuntimeMementoCareTaker{}), "GetAllUnCompletedWorkflowRuntime", func(*wfengine.WfRuntimeMementoCareTaker) (map[string]*wfengine.WfRuntime, error) {
			return nil, errors.New(errorMsg)
		})
		defer guard.Unpatch()

		resourceWf, err := wfengine.CreateResourceWorkflow(resource, wfManager)
		oldWfName, isWaiting, err := resourceWf.RunLastUnCompletedRuntime(context.TODO(), "WF1", false)

		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, errorMsg)
		So(isWaiting, ShouldBeFalse)
		So(oldWfName, ShouldBeEmpty)
	})

	Convey("run rebuild workflow, ignore the last uncompleted workflow", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			CreateHook(ctrl),
			CreateMementoStorage(ctrl),
		)

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")

		guard := monkey.PatchInstanceMethod(reflect.TypeOf(&wfengine.WfRuntimeMementoCareTaker{}), "GetAllUnCompletedWorkflowRuntime", func(*wfengine.WfRuntimeMementoCareTaker) (map[string]*wfengine.WfRuntime, error) {
			return map[string]*wfengine.WfRuntime{
				"1-WF1": {
					FlowStatus: wfengine.FlowStatusFailed,
					FlowName:   "WF1",
				},
			}, nil
		})
		defer guard.Unpatch()

		resourceWf, err := wfengine.CreateResourceWorkflow(resource, wfManager)
		oldWfName, isWaiting, err := resourceWf.RunLastUnCompletedRuntime(context.TODO(), "Rebuild", true)

		So(err, ShouldBeNil)
		So(isWaiting, ShouldBeFalse)
		So(oldWfName, ShouldContainSubstring, "WF1")
	})

	Convey("run rebuild workflow, not ignore the last uncompleted rebuild workflow", t, func() {

		ctrl := NewController(t)
		defer ctrl.Finish()

		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			CreateHook(ctrl),
			CreateMementoStorage(ctrl),
		)

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")

		guard := monkey.PatchInstanceMethod(reflect.TypeOf(&wfengine.WfRuntimeMementoCareTaker{}), "GetAllUnCompletedWorkflowRuntime", func(*wfengine.WfRuntimeMementoCareTaker) (map[string]*wfengine.WfRuntime, error) {
			return map[string]*wfengine.WfRuntime{
				"1-WF1": {
					FlowStatus: wfengine.FlowStatusFailed,
					FlowName:   "Rebuild",
				},
			}, nil
		})
		defer guard.Unpatch()

		resourceWf, err := wfengine.CreateResourceWorkflow(resource, wfManager)
		oldWfName, isWaiting, err := resourceWf.RunLastUnCompletedRuntime(context.TODO(), "Rebuild", true)

		So(err, ShouldNotBeNil)
		So(isWaiting, ShouldBeFalse)
		So(oldWfName, ShouldContainSubstring, "Rebuild")
	})
}

func Test_RetryInterruptedStep(t *testing.T) {
	Convey("get last workflow runtime failed", t, func() {
		errorMsg := "GetLastWorkflowRuntime failed"
		ctrl := NewController(t)
		defer ctrl.Finish()

		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			CreateHook(ctrl),
			CreateMementoStorage(ctrl),
		)

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")

		guard := monkey.PatchInstanceMethod(reflect.TypeOf(&wfengine.WfRuntimeMementoCareTaker{}), "GetLastWorkflowRuntime", func(*wfengine.WfRuntimeMementoCareTaker) (string, *wfengine.WfRuntime, error) {
			return "", nil, errors.New(errorMsg)
		})
		defer guard.Unpatch()

		resourceWf, err := wfengine.CreateResourceWorkflow(resource, wfManager)
		err = resourceWf.RetryInterruptedStep()

		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, errorMsg)
	})

	Convey("get flow meta failed", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			CreateHook(ctrl),
			CreateMementoStorage(ctrl),
		)

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")

		guard := monkey.PatchInstanceMethod(reflect.TypeOf(&wfengine.WfRuntimeMementoCareTaker{}), "GetLastWorkflowRuntime", func(*wfengine.WfRuntimeMementoCareTaker) (string, *wfengine.WfRuntime, error) {
			return "1-WF1", &wfengine.WfRuntime{
				FlowStatus: wfengine.FlowStatusFailed,
				FlowName:   "NotExistFlow",
			}, nil
		})
		defer guard.Unpatch()

		resourceWf, err := wfengine.CreateResourceWorkflow(resource, wfManager)
		err = resourceWf.RetryInterruptedStep()

		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "get workflow runtime")
	})

	Convey("only failedAndCompleted workflow runtime can be retried", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			CreateHook(ctrl),
			CreateMementoStorage(ctrl),
		)

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")

		guard := monkey.PatchInstanceMethod(reflect.TypeOf(&wfengine.WfRuntimeMementoCareTaker{}), "GetLastWorkflowRuntime", func(*wfengine.WfRuntimeMementoCareTaker) (string, *wfengine.WfRuntime, error) {
			return "1-WF1", &wfengine.WfRuntime{
				FlowStatus: wfengine.FlowStatusFailed,
				FlowName:   "WF1",
			}, nil
		})
		defer guard.Unpatch()

		resourceWf, err := wfengine.CreateResourceWorkflow(resource, wfManager)
		err = resourceWf.RetryInterruptedStep()

		So(err, ShouldBeNil)
	})

	Convey("recover from first step", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			CreateHook(ctrl),
			CreateMementoStorage(ctrl),
		)

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")

		guard := monkey.PatchInstanceMethod(reflect.TypeOf(&wfengine.WfRuntimeMementoCareTaker{}), "GetLastWorkflowRuntime", func(*wfengine.WfRuntimeMementoCareTaker) (string, *wfengine.WfRuntime, error) {
			return "1-WF1", &wfengine.WfRuntime{
				FlowStatus: wfengine.FlowStatusFailedCompleted,
				FlowName:   "WF1",
				RunningSteps: []*wfengine.StepRuntime{
					{
						StepName:   "Step1",
						StepStatus: wfengine.StepStatusCompleted,
					},
					{
						StepName:   "Step2",
						StepStatus: wfengine.StepStatusFailed,
					},
				},
			}, nil
		})
		defer guard.Unpatch()

		resourceWf, err := wfengine.CreateResourceWorkflow(resource, wfManager)
		err = resourceWf.RetryInterruptedStep()

		So(err, ShouldBeNil)
	})

	Convey("recover from failed step", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			CreateHook(ctrl),
			CreateMementoStorage(ctrl),
		)

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")

		guard := monkey.PatchInstanceMethod(reflect.TypeOf(&wfengine.WfRuntimeMementoCareTaker{}), "GetLastWorkflowRuntime", func(*wfengine.WfRuntimeMementoCareTaker) (string, *wfengine.WfRuntime, error) {
			return "1-WF2", &wfengine.WfRuntime{
				FlowStatus: wfengine.FlowStatusFailedCompleted,
				FlowName:   "WF2",
				RunningSteps: []*wfengine.StepRuntime{
					{
						StepName:   "Step1",
						StepStatus: wfengine.StepStatusCompleted,
					},
					{
						StepName:   "GroupStep1",
						StepStatus: wfengine.StepStatusFailed,
						RetryTimes: 10,
					},
				},
			}, nil
		})
		defer guard.Unpatch()

		resourceWf, err := wfengine.CreateResourceWorkflow(resource, wfManager)
		err = resourceWf.RetryInterruptedStep()

		So(err, ShouldBeNil)
	})

	Convey("json.Marshal failed", t, func() {
		errorMsg := "json.Marshal failed"
		ctrl := NewController(t)
		defer ctrl.Finish()

		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			CreateHook(ctrl),
			CreateMementoStorage(ctrl),
		)

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")

		guard := monkey.PatchInstanceMethod(reflect.TypeOf(&wfengine.WfRuntimeMementoCareTaker{}), "GetLastWorkflowRuntime", func(*wfengine.WfRuntimeMementoCareTaker) (string, *wfengine.WfRuntime, error) {
			return "1-WF2", &wfengine.WfRuntime{
				FlowStatus: wfengine.FlowStatusFailedCompleted,
				FlowName:   "WF2",
				RunningSteps: []*wfengine.StepRuntime{
					{
						StepName:   "Step1",
						StepStatus: wfengine.StepStatusCompleted,
					},
					{
						StepName:   "GroupStep1",
						StepStatus: wfengine.StepStatusFailed,
						RetryTimes: 10,
					},
				},
			}, nil
		})
		defer guard.Unpatch()

		guard1 := monkey.Patch(j.Marshal, func(v interface{}) ([]byte, error) {
			return nil, errors.New(errorMsg)
		})
		defer guard1.Unpatch()

		resourceWf, err := wfengine.CreateResourceWorkflow(resource, wfManager)
		err = resourceWf.RetryInterruptedStep()

		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, errorMsg)
	})

}

func Test_CommonWorkFlowMainEnter(t *testing.T) {
	Convey("workflow run succeed", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			CreateHook(ctrl),
			CreateMementoStorage(ctrl),
		)

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")
		resource.EXPECT().Fetch().AnyTimes().Return(resource, nil)
		resource.EXPECT().IsCancelled().AnyTimes().Return(false)

		guard2 := monkey.PatchInstanceMethod(reflect.TypeOf(&wfengine.ResourceWorkflow{}), "Run", func(m *wfengine.ResourceWorkflow, ctx context.Context, flowName string, initContext map[string]interface{}) (err error) {
			return nil
		})
		defer guard2.Unpatch()

		resourceWf, err := wfengine.CreateResourceWorkflow(resource, wfManager)
		err = resourceWf.CommonWorkFlowMainEnter(context.TODO(), resource, "WF1", false, func(resource statemachine.StateResource) (*statemachine.Event, error) {
			return statemachine.CreateEvent("WF1", nil), nil
		})

		So(err, ShouldBeNil)
	})

	Convey("workflow run failed", t, func() {
		errorMsg := "workflow run failed"
		ctrl := NewController(t)
		defer ctrl.Finish()

		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			CreateHook(ctrl),
			CreateMementoStorage(ctrl),
		)

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")
		resource.EXPECT().Fetch().AnyTimes().Return(resource, nil)
		resource.EXPECT().IsCancelled().AnyTimes().Return(false)

		guard2 := monkey.PatchInstanceMethod(reflect.TypeOf(&wfengine.ResourceWorkflow{}), "Run", func(m *wfengine.ResourceWorkflow, ctx context.Context, flowName string, initContext map[string]interface{}) (err error) {
			return errors.New(errorMsg)
		})
		defer guard2.Unpatch()

		resourceWf, err := wfengine.CreateResourceWorkflow(resource, wfManager)
		err = resourceWf.CommonWorkFlowMainEnter(context.TODO(), resource, "WF1", false, func(resource statemachine.StateResource) (*statemachine.Event, error) {
			return statemachine.CreateEvent("WF1", nil), nil
		})

		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, errorMsg)
	})

	Convey("cancel running workflow", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			CreateHook(ctrl),
			CreateMementoStorage(ctrl),
		)
		recover := mock.NewMockRecover(ctrl)
		recover.EXPECT().SaveResourceInterruptInfo(Any(), Any(), Any(), Any(), Any()).Return(nil)
		wfManager.RegisterRecover(recover)

		step := mock.NewMockStepAction(ctrl)
		step.EXPECT().Init(Any(), Any()).AnyTimes().Return(nil)
		step.EXPECT().DoStep(Any(), Any()).Do(func(context.Context, logr.Logger) error {
			time.Sleep(10 * time.Second)
			return nil
		})
		wfManager.RegisterStep(step)

		guard := monkey.PatchInstanceMethod(reflect.TypeOf(&wfengine.WfRuntime{}), "NewStepActionIns", func(rt *wfengine.WfRuntime, stepMeta *wfengine.StepMeta) (wfengine.StepAction, error) {
			return step, nil
		})
		defer guard.Unpatch()

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")
		resource.EXPECT().Fetch().AnyTimes().Return(resource, nil)
		resource.EXPECT().IsCancelled().AnyTimes().Return(true)
		resource.EXPECT().GetState().Return(statemachine.StateRunning)
		resource.EXPECT().UpdateState(Any()).Return(resource, nil)

		resourceWf, err := wfengine.CreateResourceWorkflow(resource, wfManager)
		err = resourceWf.CommonWorkFlowMainEnter(context.TODO(), resource, "WF1", false, func(resource statemachine.StateResource) (*statemachine.Event, error) {
			return statemachine.CreateEvent("WF1", nil), nil
		})

		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "CANCELLED")
	})

	Convey("set status to running when event is nil", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			CreateHook(ctrl),
			CreateMementoStorage(ctrl),
		)

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")
		resource.EXPECT().Fetch().AnyTimes().Return(resource, nil)
		resource.EXPECT().IsCancelled().AnyTimes().Return(false)
		resource.EXPECT().UpdateState(statemachine.StateRunning).Times(1).Return(resource, nil)

		resourceWf, err := wfengine.CreateResourceWorkflow(resource, wfManager)
		err = resourceWf.CommonWorkFlowMainEnter(context.TODO(), resource, "WF1", false, func(resource statemachine.StateResource) (*statemachine.Event, error) {
			return nil, nil
		})

		So(err, ShouldBeNil)
	})

	Convey("set status to running failed when event is nil", t, func() {
		errorMsg := "update status failed"
		ctrl := NewController(t)
		defer ctrl.Finish()

		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			CreateHook(ctrl),
			CreateMementoStorage(ctrl),
		)

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")
		resource.EXPECT().Fetch().AnyTimes().Return(resource, nil)
		resource.EXPECT().IsCancelled().AnyTimes().Return(false)
		resource.EXPECT().UpdateState(statemachine.StateRunning).Times(1).Return(nil, errors.New(errorMsg))

		resourceWf, err := wfengine.CreateResourceWorkflow(resource, wfManager)
		err = resourceWf.CommonWorkFlowMainEnter(context.TODO(), resource, "WF1", false, func(resource statemachine.StateResource) (*statemachine.Event, error) {
			return nil, nil
		})

		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, errorMsg)
	})

	Convey("run last uncompleted workflow runtime failed", t, func() {
		errorMsg := "run last uncompleted workflow runtime failed"
		ctrl := NewController(t)
		defer ctrl.Finish()

		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			CreateHook(ctrl),
			CreateMementoStorage(ctrl),
		)

		guard2 := monkey.PatchInstanceMethod(reflect.TypeOf(&wfengine.ResourceWorkflow{}), "RunLastUnCompletedRuntime", func(m *wfengine.ResourceWorkflow, ctx context.Context, flowName string, ignoreOtherUnCompleted bool) (oldWfName string, isWaiting bool, err error) {
			return "", false, errors.New(errorMsg)
		})
		defer guard2.Unpatch()

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")
		resource.EXPECT().IsCancelled().AnyTimes().Return(false)

		resourceWf, err := wfengine.CreateResourceWorkflow(resource, wfManager)
		err = resourceWf.CommonWorkFlowMainEnter(context.TODO(), resource, "WF1", false, func(resource statemachine.StateResource) (*statemachine.Event, error) {
			return nil, nil
		})

		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, errorMsg)
	})

	Convey("last uncompleted workflow is waiting status", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			CreateHook(ctrl),
			CreateMementoStorage(ctrl),
		)

		guard2 := monkey.PatchInstanceMethod(reflect.TypeOf(&wfengine.ResourceWorkflow{}), "RunLastUnCompletedRuntime", func(m *wfengine.ResourceWorkflow, ctx context.Context, flowName string, ignoreOtherUnCompleted bool) (oldWfName string, isWaiting bool, err error) {
			return "", true, nil
		})
		defer guard2.Unpatch()

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")
		resource.EXPECT().IsCancelled().AnyTimes().Return(false)

		resourceWf, err := wfengine.CreateResourceWorkflow(resource, wfManager)
		err = resourceWf.CommonWorkFlowMainEnter(context.TODO(), resource, "WF1", false, func(resource statemachine.StateResource) (*statemachine.Event, error) {
			return nil, nil
		})

		So(err, ShouldBeNil)
	})

	Convey("fetch resource failed", t, func() {
		errorMsg := "fetch resource failed"
		ctrl := NewController(t)
		defer ctrl.Finish()

		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			CreateHook(ctrl),
			CreateMementoStorage(ctrl),
		)

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")
		resource.EXPECT().IsCancelled().AnyTimes().Return(false)
		resource.EXPECT().Fetch().Times(1).Return(nil, errors.New(errorMsg))

		resourceWf, err := wfengine.CreateResourceWorkflow(resource, wfManager)
		err = resourceWf.CommonWorkFlowMainEnter(context.TODO(), resource, "WF1", false, func(resource statemachine.StateResource) (*statemachine.Event, error) {
			return nil, nil
		})

		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, errorMsg)
	})

	Convey("event check failed", t, func() {
		errorMsg := "event check failed"
		ctrl := NewController(t)
		defer ctrl.Finish()

		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			CreateHook(ctrl),
			CreateMementoStorage(ctrl),
		)

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")
		resource.EXPECT().IsCancelled().AnyTimes().Return(false)
		resource.EXPECT().Fetch().Times(1).Return(resource, nil)

		resourceWf, err := wfengine.CreateResourceWorkflow(resource, wfManager)
		err = resourceWf.CommonWorkFlowMainEnter(context.TODO(), resource, "WF1", false, func(resource statemachine.StateResource) (*statemachine.Event, error) {
			return nil, errors.New(errorMsg)
		})

		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, errorMsg)
	})

	Convey("run workflow succeed, but reload resource failed", t, func() {
		errorMsg := "fetch resource failed"
		ctrl := NewController(t)
		defer ctrl.Finish()

		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			CreateHook(ctrl),
			CreateMementoStorage(ctrl),
		)

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")
		resource.EXPECT().IsCancelled().AnyTimes().Return(false)
		first := resource.EXPECT().Fetch().Return(resource, nil)
		resource.EXPECT().Fetch().Return(nil, errors.New(errorMsg)).After(first)

		guard2 := monkey.PatchInstanceMethod(reflect.TypeOf(&wfengine.ResourceWorkflow{}), "Run", func(m *wfengine.ResourceWorkflow, ctx context.Context, flowName string, initContext map[string]interface{}) (err error) {
			return nil
		})
		defer guard2.Unpatch()

		resourceWf, err := wfengine.CreateResourceWorkflow(resource, wfManager)
		err = resourceWf.CommonWorkFlowMainEnter(context.TODO(), resource, "WF1", false, func(resource statemachine.StateResource) (*statemachine.Event, error) {
			return statemachine.CreateEvent("WF1", nil), nil
		})

		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, errorMsg)
	})

}
