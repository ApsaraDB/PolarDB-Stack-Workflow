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

	"github.com/go-logr/logr"

	"k8s.io/klog/klogr"

	"github.com/ApsaraDB/PolarDB-Stack-Workflow/define"

	"github.com/ApsaraDB/PolarDB-Stack-Workflow/statemachine"
	"github.com/ApsaraDB/PolarDB-Stack-Workflow/test/mock"
	"github.com/ApsaraDB/PolarDB-Stack-Workflow/wfengine"
	"github.com/bouk/monkey"
	. "github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
)

func Test_CreateWfRuntime(t *testing.T) {
	Convey("create workflow runtime succeed", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		createHookFunc := func(wfRuntime *wfengine.WfRuntime) (wfengine.WfHook, error) {
			hook := mock.NewMockWfHook(ctrl)
			hook.EXPECT().OnWfInit().Return(nil)
			return hook, nil
		}

		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			createHookFunc,
			nil,
		)

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")

		resourceWf, err := wfengine.CreateResourceWorkflow(resource, wfManager)
		So(err, ShouldBeNil)

		runtime, err := wfengine.CreateWfRuntime("WF1", nil, resourceWf)
		So(runtime, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})

	Convey("createHookFunc failed", t, func() {
		errorMsg := "createHookFunc failed"
		ctrl := NewController(t)
		defer ctrl.Finish()

		createHookFunc := func(wfRuntime *wfengine.WfRuntime) (wfengine.WfHook, error) {
			return nil, errors.New(errorMsg)
		}

		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			createHookFunc,
			nil,
		)

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")

		resourceWf, err := wfengine.CreateResourceWorkflow(resource, wfManager)
		So(err, ShouldBeNil)

		runtime, err := wfengine.CreateWfRuntime("WF1", nil, resourceWf)
		So(runtime, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, errorMsg)
	})

	Convey("OnWfInit failed", t, func() {
		errorMsg := "OnWfInit failed"
		ctrl := NewController(t)
		defer ctrl.Finish()

		createHookFunc := func(wfRuntime *wfengine.WfRuntime) (wfengine.WfHook, error) {
			hook := mock.NewMockWfHook(ctrl)
			hook.EXPECT().OnWfInit().Return(errors.New(errorMsg))
			return hook, nil
		}

		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			createHookFunc,
			nil,
		)

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")

		resourceWf, err := wfengine.CreateResourceWorkflow(resource, wfManager)
		So(err, ShouldBeNil)

		runtime, err := wfengine.CreateWfRuntime("WF1", nil, resourceWf)
		So(runtime, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, errorMsg)
	})

	Convey("json.Marshal workflow runtime failed", t, func() {
		errorMsg := "json.Marshal workflow runtime failed"
		ctrl := NewController(t)
		defer ctrl.Finish()

		guard1 := monkey.Patch(j.Marshal, func(v interface{}) ([]byte, error) {
			return nil, errors.New(errorMsg)
		})
		defer guard1.Unpatch()

		createMementoStorageFunc := func(resource statemachine.StateResource) (wfengine.WfRuntimeMementoStorage, error) {
			mementoStorage := mock.NewMockWfRuntimeMementoStorage(ctrl)
			mementoStorage.EXPECT().LoadMementoMap(Any()).Return(map[string]string{}, nil)
			return mementoStorage, nil
		}

		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			nil,
			createMementoStorageFunc,
		)

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")

		resourceWf, err := wfengine.CreateResourceWorkflow(resource, wfManager)
		So(err, ShouldBeNil)

		runtime, err := wfengine.CreateWfRuntime("WF1", nil, resourceWf)
		So(runtime, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, errorMsg)
	})

	Convey("create new memento failed", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		createMementoStorageFunc := func(resource statemachine.StateResource) (wfengine.WfRuntimeMementoStorage, error) {
			mementoStorage := mock.NewMockWfRuntimeMementoStorage(ctrl)
			mementoStorage.EXPECT().LoadMementoMap(Any()).Return(map[string]string{
				"WF1": "",
			}, nil)
			return mementoStorage, nil
		}

		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			nil,
			createMementoStorageFunc,
		)

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")

		resourceWf, err := wfengine.CreateResourceWorkflow(resource, wfManager)
		So(err, ShouldBeNil)

		runtime, err := wfengine.CreateWfRuntime("WF1", nil, resourceWf)
		So(runtime, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "invalid syntax")
	})

	Convey("save new memento failed", t, func() {
		errorMsg := "save new memento failed"
		ctrl := NewController(t)
		defer ctrl.Finish()

		createMementoStorageFunc := func(resource statemachine.StateResource) (wfengine.WfRuntimeMementoStorage, error) {
			mementoStorage := mock.NewMockWfRuntimeMementoStorage(ctrl)
			mementoStorage.EXPECT().LoadMementoMap(Any()).Return(map[string]string{}, nil)
			mementoStorage.EXPECT().Save(Any(), Any()).Return(errors.New(errorMsg))
			return mementoStorage, nil
		}

		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			nil,
			createMementoStorageFunc,
		)

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")

		resourceWf, err := wfengine.CreateResourceWorkflow(resource, wfManager)
		So(err, ShouldBeNil)

		runtime, err := wfengine.CreateWfRuntime("WF1", nil, resourceWf)
		So(runtime, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, errorMsg)
	})
}

func Test_LoadUnCompletedWfRuntime(t *testing.T) {
	Convey("load uncompleted workflow runtime succeed", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()
		createHookFunc := func(wfRuntime *wfengine.WfRuntime) (wfengine.WfHook, error) {
			hook := mock.NewMockWfHook(ctrl)
			hook.EXPECT().OnWfInit().Return(nil)
			return hook, nil
		}

		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			createHookFunc,
			CreateMementoStorage(ctrl),
		)

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")

		unCompletedRuntime := &wfengine.WfRuntime{
			FlowStatus: wfengine.FlowStatusFailed,
			FlowName:   "WF1",
			InitContext: map[string]interface{}{
				"_resourceName":                 "test-resource",
				"_resourceNameSpace":            "default",
				"_flowSaveNumberKey":            "2-WF1",
				"_flowSaveResourceNameKey":      "test-resource-workflow",
				"_flowSaveResourceNameSpaceKey": "default",
			},

			RunningSteps: []*wfengine.StepRuntime{
				{
					StepName:   "Step1",
					StepStatus: wfengine.StepStatusCompleted,
				},
				{
					StepName:   "Step2",
					StepStatus: wfengine.StepStatusFailed,
					RetryTimes: 2,
				},
			},
		}
		resourceWf, err := wfengine.CreateResourceWorkflow(resource, wfManager)
		So(err, ShouldBeNil)
		err = wfengine.LoadUnCompletedWfRuntime(unCompletedRuntime, resourceWf)
		So(err, ShouldBeNil)
	})

	Convey("OnWfInit failed", t, func() {
		errorMsg := "OnWfInit failed"
		ctrl := NewController(t)
		defer ctrl.Finish()

		createHookFunc := func(wfRuntime *wfengine.WfRuntime) (wfengine.WfHook, error) {
			hook := mock.NewMockWfHook(ctrl)
			hook.EXPECT().OnWfInit().Return(errors.New(errorMsg))
			return hook, nil
		}

		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			createHookFunc,
			CreateMementoStorage(ctrl),
		)

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")

		unCompletedRuntime := &wfengine.WfRuntime{
			FlowStatus: wfengine.FlowStatusFailed,
			FlowName:   "WF1",
			InitContext: map[string]interface{}{
				"_resourceName":                 "test-resource",
				"_resourceNameSpace":            "default",
				"_flowSaveNumberKey":            "2-WF1",
				"_flowSaveResourceNameKey":      "test-resource-workflow",
				"_flowSaveResourceNameSpaceKey": "default",
			},

			RunningSteps: []*wfengine.StepRuntime{
				{
					StepName:   "Step1",
					StepStatus: wfengine.StepStatusCompleted,
				},
				{
					StepName:   "Step2",
					StepStatus: wfengine.StepStatusFailed,
					RetryTimes: 2,
				},
			},
		}
		resourceWf, err := wfengine.CreateResourceWorkflow(resource, wfManager)
		So(err, ShouldBeNil)
		err = wfengine.LoadUnCompletedWfRuntime(unCompletedRuntime, resourceWf)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, errorMsg)
	})

	Convey("createHookFunc failed", t, func() {
		errorMsg := "createHookFunc failed"
		ctrl := NewController(t)
		defer ctrl.Finish()

		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			func(wfRuntime *wfengine.WfRuntime) (wfengine.WfHook, error) {
				return nil, errors.New(errorMsg)
			},
			nil,
		)

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")

		resourceWf, err := wfengine.CreateResourceWorkflow(resource, wfManager)
		So(err, ShouldBeNil)
		err = wfengine.LoadUnCompletedWfRuntime(&wfengine.WfRuntime{}, resourceWf)

		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, errorMsg)
	})
}

func Test_StartWfRuntime(t *testing.T) {
	Convey("initContext is nil", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		wfManager := &wfengine.WfManager{}
		wfManager.RegisterConf(define.DefaultWfConf)

		runtime := &wfengine.WfRuntime{
			FlowStatus:       wfengine.FlowStatusCompleted,
			FlowName:         "WF1",
			Logger:           klogr.New(),
			ResourceWorkflow: &wfengine.ResourceWorkflow{WfManager: wfManager},
		}
		err := runtime.Start(context.TODO())
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "initContext is nil")
	})

	Convey("initContext_resourceName is empty", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		wfManager := &wfengine.WfManager{}
		wfManager.RegisterConf(define.DefaultWfConf)

		runtime := &wfengine.WfRuntime{
			FlowStatus:       wfengine.FlowStatusCompleted,
			FlowName:         "WF1",
			Logger:           klogr.New(),
			ResourceWorkflow: &wfengine.ResourceWorkflow{WfManager: wfManager},
			InitContext: map[string]interface{}{
				"_resourceName":      "",
				"_resourceNameSpace": "default",
			},
		}

		err := runtime.Start(context.TODO())
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, define.DefaultWfConf[define.WorkFlowResourceName])
	})

	Convey("initContext._resourceNameSpace is empty", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		wfManager := &wfengine.WfManager{}
		wfManager.RegisterConf(define.DefaultWfConf)

		runtime := &wfengine.WfRuntime{
			FlowStatus:       wfengine.FlowStatusCompleted,
			FlowName:         "WF1",
			Logger:           klogr.New(),
			ResourceWorkflow: &wfengine.ResourceWorkflow{WfManager: wfManager},
			InitContext: map[string]interface{}{
				"_resourceName":      "test-resource",
				"_resourceNameSpace": "",
			},
		}
		err := runtime.Start(context.TODO())
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, define.DefaultWfConf[define.WorkFlowResourceNameSpace])
	})

	Convey("ignore completed workflow runtime", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		wfManager := &wfengine.WfManager{}
		wfManager.RegisterConf(define.DefaultWfConf)

		unCompletedRuntime := &wfengine.WfRuntime{
			FlowStatus:       wfengine.FlowStatusCompleted,
			FlowName:         "WF1",
			Logger:           klogr.New(),
			ResourceWorkflow: &wfengine.ResourceWorkflow{WfManager: wfManager},
			InitContext: map[string]interface{}{
				"_resourceName":      "test-resource",
				"_resourceNameSpace": "default",
			},
		}

		err := unCompletedRuntime.Start(context.TODO())
		So(err, ShouldBeNil)
	})

	Convey("ignore failedAndCompleted workflow runtime", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		wfManager := &wfengine.WfManager{}
		wfManager.RegisterConf(define.DefaultWfConf)

		unCompletedRuntime := &wfengine.WfRuntime{
			FlowStatus: wfengine.FlowStatusFailedCompleted,
			FlowName:   "WF1",
			InitContext: map[string]interface{}{
				"_resourceName":      "test-resource",
				"_resourceNameSpace": "default",
			},
			Logger:           klogr.New(),
			ResourceWorkflow: &wfengine.ResourceWorkflow{WfManager: wfManager},
		}

		err := unCompletedRuntime.Start(context.TODO())
		So(err, ShouldBeNil)
	})

	Convey("workflow meta not exist, skip it", t, func() {
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

		logger := klogr.New()
		runtime := &wfengine.WfRuntime{
			FlowStatus: wfengine.FlowStatusRunning,
			FlowName:   "NotExistWorkflow",
			InitContext: map[string]interface{}{
				"_resourceName":      "test-resource",
				"_resourceNameSpace": "default",
			},
			Logger: klogr.New(),
			ResourceWorkflow: &wfengine.ResourceWorkflow{
				Logger:    logger,
				Resource:  resource,
				WfManager: wfManager,
			},
		}

		err = runtime.Start(context.TODO())
		So(err, ShouldBeNil)
	})

	Convey("workflow set complete failed, skip it.", t, func() {

		ctrl := NewController(t)
		defer ctrl.Finish()

		errorMsg := "json.Marshal failed"
		guard := monkey.Patch(j.Marshal, func(v interface{}) ([]byte, error) {
			return nil, errors.New(errorMsg)
		})
		defer guard.Unpatch()

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

		logger := klogr.New()
		runtime := &wfengine.WfRuntime{
			FlowStatus: wfengine.FlowStatusRunning,
			FlowName:   "NotExistWorkflow",
			InitContext: map[string]interface{}{
				"_resourceName":      "test-resource",
				"_resourceNameSpace": "default",
			},
			Logger: klogr.New(),
			ResourceWorkflow: &wfengine.ResourceWorkflow{
				Logger:           logger,
				Resource:         resource,
				WfManager:        wfManager,
				MementoCareTaker: &wfengine.WfRuntimeMementoCareTaker{},
			},
		}

		err = runtime.Start(context.TODO())
		So(err, ShouldBeNil)
	})

	Convey("workflow meta step not found in runtime", t, func() {
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

		logger := klogr.New()
		runtime := &wfengine.WfRuntime{
			FlowStatus: wfengine.FlowStatusRunning,
			FlowName:   "WF1",
			RetryTimes: 1,
			InitContext: map[string]interface{}{
				"_resourceName":      "test-resource",
				"_resourceNameSpace": "default",
			},
			Logger: klogr.New(),
			ResourceWorkflow: &wfengine.ResourceWorkflow{
				Logger:    logger,
				Resource:  resource,
				WfManager: wfManager,
			},
		}

		err = runtime.Start(context.TODO())
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "not found in workflow runtime")
	})

	Convey("workflow runtime interrupt after step is retried 10 times", t, func() {
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

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetState().Return(statemachine.StateRunning)
		resource.EXPECT().UpdateState(Any()).Return(resource, nil)

		logger := klogr.New()
		runtime := &wfengine.WfRuntime{
			FlowStatus: wfengine.FlowStatusRunning,
			FlowName:   "WF1",
			RetryTimes: 1,
			InitContext: map[string]interface{}{
				"_resourceName":      "test-resource",
				"_resourceNameSpace": "default",
			},
			Logger: klogr.New(),
			ResourceWorkflow: &wfengine.ResourceWorkflow{
				Logger:    logger,
				Resource:  resource,
				WfManager: wfManager,
			},
			RunningSteps: []*wfengine.StepRuntime{
				{
					StepName:   "Step1",
					StepStatus: wfengine.StepStatusCompleted,
				},
				{
					StepName:   "Step2",
					StepStatus: wfengine.StepStatusPrepared,
					RetryTimes: 10,
				},
			},
		}

		err = runtime.Start(context.TODO())
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "too many times retry")
	})

	Convey("save running status to workflow runtime memento failed", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			CreateHook(ctrl),
			CreateMementoStorage(ctrl),
		)

		step := mock.NewMockStepAction(ctrl)
		wfManager.RegisterStep(step)

		resource := mock.NewMockStateResource(ctrl)

		logger := klogr.New()
		runtime := &wfengine.WfRuntime{
			FlowStatus: wfengine.FlowStatusRunning,
			FlowName:   "WF1",
			RetryTimes: 1,
			InitContext: map[string]interface{}{
				"_resourceName":      "test-resource",
				"_resourceNameSpace": "default",
			},
			Logger: klogr.New(),
			ResourceWorkflow: &wfengine.ResourceWorkflow{
				Logger:           logger,
				Resource:         resource,
				WfManager:        wfManager,
				MementoCareTaker: &wfengine.WfRuntimeMementoCareTaker{},
			},
			RunningSteps: []*wfengine.StepRuntime{
				{
					StepName:   "Step1",
					StepStatus: wfengine.StepStatusInited,
					RetryTimes: 1,
				},
				{
					StepName:   "Step2",
					StepStatus: wfengine.StepStatusPrepared,
					RetryTimes: 0,
				},
			},
		}

		err = runtime.Start(context.TODO())
		So(err, ShouldNotBeNil)

	})

	Convey("create step action instance failed", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			CreateHook(ctrl),
			CreateMementoStorage(ctrl),
		)

		step := mock.NewMockStepAction(ctrl)
		wfManager.RegisterStep(step)

		errorMsg := "NewStepActionIns failed"
		guard := monkey.PatchInstanceMethod(reflect.TypeOf(&wfengine.WfRuntime{}), "NewStepActionIns", func(rt *wfengine.WfRuntime, stepMeta *wfengine.StepMeta) (wfengine.StepAction, error) {
			return nil, errors.New(errorMsg)
		})
		defer guard.Unpatch()

		resource := mock.NewMockStateResource(ctrl)

		logger := klogr.New()
		runtime := &wfengine.WfRuntime{
			FlowStatus: wfengine.FlowStatusRunning,
			FlowName:   "WF1",
			RetryTimes: 1,
			InitContext: map[string]interface{}{
				"_resourceName":      "test-resource",
				"_resourceNameSpace": "default",
			},
			Logger: klogr.New(),
			ResourceWorkflow: &wfengine.ResourceWorkflow{
				Logger:    logger,
				Resource:  resource,
				WfManager: wfManager,
			},
			RunningSteps: []*wfengine.StepRuntime{
				{
					StepName:   "Step1",
					StepStatus: wfengine.StepStatusFailed,
				},
				{
					StepName:   "Step2",
					StepStatus: wfengine.StepStatusPrepared,
					RetryTimes: 1,
				},
			},
		}

		err = runtime.Start(context.TODO())
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, errorMsg)
	})

	Convey("create unknown step action instance", t, func() {
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
		resource.EXPECT().GetState().AnyTimes().Return(statemachine.StateRunning)
		resource.EXPECT().UpdateState(Any()).Return(resource, nil)

		recover := mock.NewMockRecover(ctrl)
		recover.EXPECT().SaveResourceInterruptInfo(Any(), Any(), Any(), Any(), Any()).Return(nil)
		wfManager.RegisterRecover(recover)

		logger := klogr.New()
		runtime := &wfengine.WfRuntime{
			FlowStatus: wfengine.FlowStatusRunning,
			FlowName:   "WF1",
			RetryTimes: 1,
			InitContext: map[string]interface{}{
				"_resourceName":      "test-resource",
				"_resourceNameSpace": "default",
			},
			Logger: klogr.New(),

			ResourceWorkflow: &wfengine.ResourceWorkflow{
				Logger:    logger,
				Resource:  resource,
				WfManager: wfManager,
			},
			RunningSteps: []*wfengine.StepRuntime{
				{
					StepName:   "Step1",
					StepStatus: wfengine.StepStatusFailed,
				},
				{
					StepName:   "Step2",
					StepStatus: wfengine.StepStatusPrepared,
					RetryTimes: 1,
				},
			},
		}

		err = runtime.Start(context.TODO())
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "UNKNOWN_TASK")
	})

	Convey("do step return waiting status", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			CreateHook(ctrl),
			CreateMementoStorage(ctrl),
		)

		step := mock.NewMockStepAction(ctrl)
		wfManager.RegisterStep(step)

		guard := monkey.PatchInstanceMethod(reflect.TypeOf(&wfengine.WfRuntime{}), "NewStepActionIns", func(rt *wfengine.WfRuntime, stepMeta *wfengine.StepMeta) (wfengine.StepAction, error) {
			step := mock.NewMockStepAction(ctrl)
			step.EXPECT().Init(Any(), Any()).Return(nil)
			step.EXPECT().DoStep(Any(), Any()).Return(define.NewWaitingError())
			return step, nil
		})
		defer guard.Unpatch()

		resource := mock.NewMockStateResource(ctrl)

		logger := klogr.New()
		runtime := &wfengine.WfRuntime{
			FlowStatus: wfengine.FlowStatusRunning,
			FlowName:   "WF1",
			RetryTimes: 1,
			InitContext: map[string]interface{}{
				"_resourceName":      "test-resource",
				"_resourceNameSpace": "default",
			},
			Logger: klogr.New(),
			ResourceWorkflow: &wfengine.ResourceWorkflow{
				Logger:    logger,
				Resource:  resource,
				WfManager: wfManager,
			},
			RunningSteps: []*wfengine.StepRuntime{
				{
					StepName:   "Step1",
					StepStatus: wfengine.StepStatusInited,
					RetryTimes: 1,
				},
				{
					StepName:   "Step2",
					StepStatus: wfengine.StepStatusPrepared,
					RetryTimes: 0,
				},
			},
		}

		err = runtime.Start(context.TODO())
		So(err, ShouldBeNil)
		So(runtime.FlowStatus, ShouldEqual, wfengine.FlowStatusWaiting)
		So(runtime.RunningSteps[0].RetryTimes, ShouldEqual, 2)
		So(runtime.RunningSteps[0].StepStatus, ShouldEqual, wfengine.StepStatusWaiting)
		So(runtime.RunningSteps[1].RetryTimes, ShouldEqual, 0)
		So(runtime.RunningSteps[1].StepStatus, ShouldEqual, wfengine.FlowStatusPrepared)
	})

	Convey("step is waiting status, do step return waiting status, keep waiting", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			CreateHook(ctrl),
			CreateMementoStorage(ctrl),
		)

		step := mock.NewMockStepAction(ctrl)
		wfManager.RegisterStep(step)

		guard := monkey.PatchInstanceMethod(reflect.TypeOf(&wfengine.WfRuntime{}), "NewStepActionIns", func(rt *wfengine.WfRuntime, stepMeta *wfengine.StepMeta) (wfengine.StepAction, error) {
			step := mock.NewMockStepAction(ctrl)
			step.EXPECT().Init(Any(), Any()).Return(nil)
			step.EXPECT().DoStep(Any(), Any()).Return(define.NewWaitingError())
			return step, nil
		})
		defer guard.Unpatch()

		resource := mock.NewMockStateResource(ctrl)

		logger := klogr.New()
		runtime := &wfengine.WfRuntime{
			FlowStatus: wfengine.FlowStatusWaiting,
			FlowName:   "WF1",
			RetryTimes: 1,
			InitContext: map[string]interface{}{
				"_resourceName":      "test-resource",
				"_resourceNameSpace": "default",
			},
			Logger: klogr.New(),
			ResourceWorkflow: &wfengine.ResourceWorkflow{
				Logger:    logger,
				Resource:  resource,
				WfManager: wfManager,
			},
			RunningSteps: []*wfengine.StepRuntime{
				{
					StepName:   "Step1",
					StepStatus: wfengine.StepStatusWaiting,
					RetryTimes: 1,
				},
				{
					StepName:   "Step2",
					StepStatus: wfengine.StepStatusPrepared,
					RetryTimes: 0,
				},
			},
		}

		err = runtime.Start(context.TODO())
		So(err, ShouldBeNil)
		So(runtime.FlowStatus, ShouldEqual, wfengine.FlowStatusWaiting)
		So(runtime.RunningSteps[0].RetryTimes, ShouldEqual, 1)
		So(runtime.RunningSteps[0].StepStatus, ShouldEqual, wfengine.StepStatusWaiting)
		So(runtime.RunningSteps[1].RetryTimes, ShouldEqual, 0)
		So(runtime.RunningSteps[1].StepStatus, ShouldEqual, wfengine.FlowStatusPrepared)
	})

	Convey("set waiting status failed", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			nil,
			CreateMementoStorage(ctrl),
		)

		step := mock.NewMockStepAction(ctrl)
		wfManager.RegisterStep(step)

		guard := monkey.PatchInstanceMethod(reflect.TypeOf(&wfengine.WfRuntime{}), "NewStepActionIns", func(rt *wfengine.WfRuntime, stepMeta *wfengine.StepMeta) (wfengine.StepAction, error) {
			step := mock.NewMockStepAction(ctrl)
			step.EXPECT().Init(Any(), Any()).Return(nil)
			step.EXPECT().DoStep(Any(), Any()).Return(define.NewWaitingError())
			return step, nil
		})
		defer guard.Unpatch()

		errorMsg := "OnStepWaiting failed"
		hook := mock.NewMockWfHook(ctrl)
		hook.EXPECT().OnStepInit(Any()).Return(nil)
		hook.EXPECT().OnStepWaiting(Any()).Return(errors.New(errorMsg))
		hook.EXPECT().OnStepCompleted(Any()).Return(nil)

		resource := mock.NewMockStateResource(ctrl)

		logger := klogr.New()
		runtime := &wfengine.WfRuntime{
			FlowStatus:  wfengine.FlowStatusRunning,
			FlowName:    "WF1",
			RetryTimes:  1,
			FlowHookIns: hook,
			InitContext: map[string]interface{}{
				"_resourceName":      "test-resource",
				"_resourceNameSpace": "default",
			},
			Logger: klogr.New(),
			ResourceWorkflow: &wfengine.ResourceWorkflow{
				Logger:    logger,
				Resource:  resource,
				WfManager: wfManager,
			},
			RunningSteps: []*wfengine.StepRuntime{
				{
					StepName:   "Step1",
					StepStatus: wfengine.StepStatusPrepared,
				},
				{
					StepName:   "Step2",
					StepStatus: wfengine.StepStatusPrepared,
					RetryTimes: 0,
				},
			},
		}

		err = runtime.Start(context.TODO())
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, errorMsg)
		So(runtime.FlowStatus, ShouldEqual, wfengine.StepStatusFailed)
		So(runtime.RunningSteps[0].RetryTimes, ShouldEqual, 0)
		So(runtime.RunningSteps[0].StepStatus, ShouldEqual, wfengine.StepStatusFailed)
		So(runtime.RunningSteps[1].RetryTimes, ShouldEqual, 0)
		So(runtime.RunningSteps[1].StepStatus, ShouldEqual, wfengine.FlowStatusPrepared)
	})

	Convey("set complete status failed", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			nil,
			CreateMementoStorage(ctrl),
		)

		step := mock.NewMockStepAction(ctrl)
		wfManager.RegisterStep(step)

		guard := monkey.PatchInstanceMethod(reflect.TypeOf(&wfengine.WfRuntime{}), "NewStepActionIns", func(rt *wfengine.WfRuntime, stepMeta *wfengine.StepMeta) (wfengine.StepAction, error) {
			step := mock.NewMockStepAction(ctrl)
			step.EXPECT().Init(Any(), Any()).Return(nil)
			step.EXPECT().DoStep(Any(), Any()).Return(nil)
			step.EXPECT().Output(Any()).Return(nil)
			return step, nil
		})
		defer guard.Unpatch()

		errorMsg := "OnStepCompleted failed"
		hook := mock.NewMockWfHook(ctrl)
		hook.EXPECT().OnStepInit(Any()).Return(nil)
		hook.EXPECT().OnStepCompleted(Any()).AnyTimes().Return(errors.New(errorMsg))

		resource := mock.NewMockStateResource(ctrl)

		logger := klogr.New()
		runtime := &wfengine.WfRuntime{
			FlowStatus:  wfengine.FlowStatusRunning,
			FlowName:    "WF1",
			RetryTimes:  1,
			FlowHookIns: hook,
			InitContext: map[string]interface{}{
				"_resourceName":      "test-resource",
				"_resourceNameSpace": "default",
			},
			Logger: klogr.New(),
			ResourceWorkflow: &wfengine.ResourceWorkflow{
				Logger:    logger,
				Resource:  resource,
				WfManager: wfManager,
			},
			RunningSteps: []*wfengine.StepRuntime{
				{
					StepName:   "Step1",
					StepStatus: wfengine.StepStatusPrepared,
				},
				{
					StepName:   "Step2",
					StepStatus: wfengine.StepStatusPrepared,
					RetryTimes: 0,
				},
			},
		}

		err = runtime.Start(context.TODO())
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, errorMsg)
		So(runtime.FlowStatus, ShouldEqual, wfengine.StepStatusFailed)
		So(runtime.RunningSteps[0].RetryTimes, ShouldEqual, 0)
		So(runtime.RunningSteps[0].StepStatus, ShouldEqual, wfengine.StepStatusFailed)
		So(runtime.RunningSteps[1].RetryTimes, ShouldEqual, 0)
		So(runtime.RunningSteps[1].StepStatus, ShouldEqual, wfengine.FlowStatusPrepared)
	})

	Convey("OnWfCompleted failed", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			nil,
			CreateMementoStorage(ctrl),
		)

		step := mock.NewMockStepAction(ctrl)
		wfManager.RegisterStep(step)

		guard := monkey.PatchInstanceMethod(reflect.TypeOf(&wfengine.WfRuntime{}), "NewStepActionIns", func(rt *wfengine.WfRuntime, stepMeta *wfengine.StepMeta) (wfengine.StepAction, error) {
			step := mock.NewMockStepAction(ctrl)
			step.EXPECT().Init(Any(), Any()).Return(nil)
			step.EXPECT().DoStep(Any(), Any()).Return(nil)
			step.EXPECT().Output(Any()).Return(nil)
			return step, nil
		})
		defer guard.Unpatch()

		errorMsg := "OnWfCompleted failed"
		hook := mock.NewMockWfHook(ctrl)
		hook.EXPECT().OnStepInit(Any()).AnyTimes().Return(nil)
		hook.EXPECT().OnStepCompleted(Any()).AnyTimes().Return(nil)
		hook.EXPECT().OnWfCompleted().Return(errors.New(errorMsg))

		resource := mock.NewMockStateResource(ctrl)

		logger := klogr.New()
		runtime := &wfengine.WfRuntime{
			FlowStatus:  wfengine.FlowStatusRunning,
			FlowName:    "WF1",
			RetryTimes:  1,
			FlowHookIns: hook,
			InitContext: map[string]interface{}{
				"_resourceName":      "test-resource",
				"_resourceNameSpace": "default",
			},
			Logger: klogr.New(),
			ResourceWorkflow: &wfengine.ResourceWorkflow{
				Logger:    logger,
				Resource:  resource,
				WfManager: wfManager,
			},
			RunningSteps: []*wfengine.StepRuntime{
				{
					StepName:   "Step1",
					StepStatus: wfengine.StepStatusPrepared,
				},
				{
					StepName:   "Step2",
					StepStatus: wfengine.StepStatusPrepared,
					RetryTimes: 0,
				},
			},
		}

		err = runtime.Start(context.TODO())
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, errorMsg)
		So(runtime.FlowStatus, ShouldEqual, wfengine.StepStatusFailed)
		So(runtime.RunningSteps[0].RetryTimes, ShouldEqual, 0)
		So(runtime.RunningSteps[0].StepStatus, ShouldEqual, wfengine.StepStatusCompleted)
		So(runtime.RunningSteps[1].RetryTimes, ShouldEqual, 0)
		So(runtime.RunningSteps[1].StepStatus, ShouldEqual, wfengine.StepStatusCompleted)
	})

	Convey("do step failed", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			nil,
			CreateMementoStorage(ctrl),
		)

		step := mock.NewMockStepAction(ctrl)
		wfManager.RegisterStep(step)

		errorMsg := "DoStep failed"
		guard := monkey.PatchInstanceMethod(reflect.TypeOf(&wfengine.WfRuntime{}), "NewStepActionIns", func(rt *wfengine.WfRuntime, stepMeta *wfengine.StepMeta) (wfengine.StepAction, error) {
			step := mock.NewMockStepAction(ctrl)
			step.EXPECT().Init(Any(), Any()).Return(nil)
			step.EXPECT().DoStep(Any(), Any()).Return(errors.New(errorMsg))
			return step, nil
		})
		defer guard.Unpatch()

		hook := mock.NewMockWfHook(ctrl)
		hook.EXPECT().OnStepInit(Any()).Return(nil)
		hook.EXPECT().OnStepCompleted(Any()).AnyTimes().Return(nil)

		resource := mock.NewMockStateResource(ctrl)

		logger := klogr.New()
		runtime := &wfengine.WfRuntime{
			FlowStatus:  wfengine.FlowStatusRunning,
			FlowName:    "WF1",
			RetryTimes:  1,
			FlowHookIns: hook,
			InitContext: map[string]interface{}{
				"_resourceName":      "test-resource",
				"_resourceNameSpace": "default",
			},
			Logger: klogr.New(),
			ResourceWorkflow: &wfengine.ResourceWorkflow{
				Logger:    logger,
				Resource:  resource,
				WfManager: wfManager,
			},
			RunningSteps: []*wfengine.StepRuntime{
				{
					StepName:   "Step1",
					StepStatus: wfengine.StepStatusPrepared,
				},
				{
					StepName:   "Step2",
					StepStatus: wfengine.StepStatusPrepared,
					RetryTimes: 0,
				},
			},
		}

		err = runtime.Start(context.TODO())
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, errorMsg)
		So(runtime.FlowStatus, ShouldEqual, wfengine.StepStatusFailed)
		So(runtime.RunningSteps[0].RetryTimes, ShouldEqual, 0)
		So(runtime.RunningSteps[0].StepStatus, ShouldEqual, wfengine.StepStatusFailed)
		So(runtime.RunningSteps[1].RetryTimes, ShouldEqual, 0)
		So(runtime.RunningSteps[1].StepStatus, ShouldEqual, wfengine.FlowStatusPrepared)
	})

	Convey("init step failed", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			nil,
			CreateMementoStorage(ctrl),
		)

		step := mock.NewMockStepAction(ctrl)
		wfManager.RegisterStep(step)

		errorMsg := "step init failed"
		guard := monkey.PatchInstanceMethod(reflect.TypeOf(&wfengine.WfRuntime{}), "NewStepActionIns", func(rt *wfengine.WfRuntime, stepMeta *wfengine.StepMeta) (wfengine.StepAction, error) {
			step := mock.NewMockStepAction(ctrl)
			step.EXPECT().Init(Any(), Any()).Return(errors.New(errorMsg))
			return step, nil
		})
		defer guard.Unpatch()

		hook := mock.NewMockWfHook(ctrl)
		hook.EXPECT().OnStepInit(Any()).Return(nil)
		hook.EXPECT().OnStepCompleted(Any()).AnyTimes().Return(nil)

		resource := mock.NewMockStateResource(ctrl)

		logger := klogr.New()
		runtime := &wfengine.WfRuntime{
			FlowName:    "WF1",
			RetryTimes:  1,
			FlowHookIns: hook,
			InitContext: map[string]interface{}{
				"_resourceName":      "test-resource",
				"_resourceNameSpace": "default",
			},
			Logger: klogr.New(),
			ResourceWorkflow: &wfengine.ResourceWorkflow{
				Logger:    logger,
				Resource:  resource,
				WfManager: wfManager,
			},
			RunningSteps: []*wfengine.StepRuntime{
				{
					StepName:   "Step1",
					StepStatus: wfengine.StepStatusPrepared,
				},
				{
					StepName:   "Step2",
					StepStatus: wfengine.StepStatusPrepared,
					RetryTimes: 0,
				},
			},
		}

		err = runtime.Start(context.TODO())
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, errorMsg)
		So(runtime.FlowStatus, ShouldEqual, wfengine.StepStatusFailed)
		So(runtime.RunningSteps[0].RetryTimes, ShouldEqual, 0)
		So(runtime.RunningSteps[0].StepStatus, ShouldEqual, wfengine.StepStatusFailed)
		So(runtime.RunningSteps[1].RetryTimes, ShouldEqual, 0)
		So(runtime.RunningSteps[1].StepStatus, ShouldEqual, wfengine.FlowStatusPrepared)
	})

	Convey("init step panic", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			nil,
			CreateMementoStorage(ctrl),
		)

		step := mock.NewMockStepAction(ctrl)
		wfManager.RegisterStep(step)

		errorMsg := "step init failed"
		guard := monkey.PatchInstanceMethod(reflect.TypeOf(&wfengine.WfRuntime{}), "NewStepActionIns", func(rt *wfengine.WfRuntime, stepMeta *wfengine.StepMeta) (wfengine.StepAction, error) {
			step := mock.NewMockStepAction(ctrl)
			step.EXPECT().Init(Any(), Any()).Do(func(map[string]interface{}, logr.Logger) {
				panic(errorMsg)
			}).Return(nil)
			return step, nil
		})
		defer guard.Unpatch()

		hook := mock.NewMockWfHook(ctrl)
		hook.EXPECT().OnStepInit(Any()).Return(nil)
		hook.EXPECT().OnStepCompleted(Any()).AnyTimes().Return(nil)

		resource := mock.NewMockStateResource(ctrl)

		logger := klogr.New()
		runtime := &wfengine.WfRuntime{
			FlowStatus:  wfengine.FlowStatusRunning,
			FlowName:    "WF1",
			RetryTimes:  1,
			FlowHookIns: hook,
			InitContext: map[string]interface{}{
				"_resourceName":      "test-resource",
				"_resourceNameSpace": "default",
			},
			Logger: klogr.New(),
			ResourceWorkflow: &wfengine.ResourceWorkflow{
				Logger:    logger,
				Resource:  resource,
				WfManager: wfManager,
			},
			RunningSteps: []*wfengine.StepRuntime{
				{
					StepName:   "Step1",
					StepStatus: wfengine.StepStatusPrepared,
				},
				{
					StepName:   "Step2",
					StepStatus: wfengine.StepStatusPrepared,
					RetryTimes: 0,
				},
			},
		}

		err = runtime.Start(context.TODO())
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, errorMsg)
	})

	Convey("set step status to running failed", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			nil,
			CreateMementoStorage(ctrl),
		)

		step := mock.NewMockStepAction(ctrl)
		wfManager.RegisterStep(step)

		errorMsg := "OnStepInit failed"
		hook := mock.NewMockWfHook(ctrl)
		hook.EXPECT().OnStepInit(Any()).Return(errors.New(errorMsg))
		hook.EXPECT().OnStepCompleted(Any()).Return(nil)

		resource := mock.NewMockStateResource(ctrl)

		logger := klogr.New()
		runtime := &wfengine.WfRuntime{
			FlowStatus:  wfengine.FlowStatusRunning,
			FlowName:    "WF1",
			RetryTimes:  1,
			FlowHookIns: hook,
			InitContext: map[string]interface{}{
				"_resourceName":      "test-resource",
				"_resourceNameSpace": "default",
			},
			Logger: klogr.New(),
			ResourceWorkflow: &wfengine.ResourceWorkflow{
				Logger:    logger,
				Resource:  resource,
				WfManager: wfManager,
			},
			RunningSteps: []*wfengine.StepRuntime{
				{
					StepName:   "Step1",
					StepStatus: wfengine.StepStatusPrepared,
				},
				{
					StepName:   "Step2",
					StepStatus: wfengine.StepStatusPrepared,
					RetryTimes: 0,
				},
			},
		}

		err = runtime.Start(context.TODO())
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, errorMsg)
		So(runtime.FlowStatus, ShouldEqual, wfengine.StepStatusFailed)
		So(runtime.RunningSteps[0].RetryTimes, ShouldEqual, 0)
		So(runtime.RunningSteps[0].StepStatus, ShouldEqual, wfengine.StepStatusFailed)
		So(runtime.RunningSteps[1].RetryTimes, ShouldEqual, 0)
		So(runtime.RunningSteps[1].StepStatus, ShouldEqual, wfengine.FlowStatusPrepared)
	})

	Convey("init step failed and interrupt", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			nil,
			CreateMementoStorage(ctrl),
		)

		step := mock.NewMockStepAction(ctrl)
		wfManager.RegisterStep(step)

		errorMsg := "step init failed and interrupt"
		guard := monkey.PatchInstanceMethod(reflect.TypeOf(&wfengine.WfRuntime{}), "NewStepActionIns", func(rt *wfengine.WfRuntime, stepMeta *wfengine.StepMeta) (wfengine.StepAction, error) {
			step := mock.NewMockStepAction(ctrl)
			step.EXPECT().Init(Any(), Any()).Return(define.NewInterruptError("", errorMsg))
			return step, nil
		})
		defer guard.Unpatch()

		hook := mock.NewMockWfHook(ctrl)
		hook.EXPECT().OnStepInit(Any()).Return(nil)
		hook.EXPECT().OnStepCompleted(Any()).AnyTimes().Return(nil)
		hook.EXPECT().OnWfInterrupt(Any()).AnyTimes().Return(nil)

		recover := mock.NewMockRecover(ctrl)
		recover.EXPECT().SaveResourceInterruptInfo(Any(), Any(), Any(), Any(), Any()).Return(nil)
		wfManager.RegisterRecover(recover)

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetState().Return(statemachine.StateCreating)
		resource.EXPECT().UpdateState(Any()).Return(resource, nil)

		logger := klogr.New()
		runtime := &wfengine.WfRuntime{
			FlowStatus:  wfengine.FlowStatusRunning,
			FlowName:    "WF1",
			RetryTimes:  1,
			FlowHookIns: hook,
			InitContext: map[string]interface{}{
				"_resourceName":      "test-resource",
				"_resourceNameSpace": "default",
			},
			Logger: klogr.New(),
			ResourceWorkflow: &wfengine.ResourceWorkflow{
				Logger:    logger,
				Resource:  resource,
				WfManager: wfManager,
			},
			RunningSteps: []*wfengine.StepRuntime{
				{
					StepName:   "Step1",
					StepStatus: wfengine.StepStatusPrepared,
				},
				{
					StepName:   "Step2",
					StepStatus: wfengine.StepStatusPrepared,
					RetryTimes: 0,
				},
			},
		}

		err = runtime.Start(context.TODO())
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, errorMsg)
		So(runtime.FlowStatus, ShouldEqual, wfengine.FlowStatusFailedCompleted)
		So(runtime.RunningSteps[0].RetryTimes, ShouldEqual, 0)
		So(runtime.RunningSteps[0].StepStatus, ShouldEqual, wfengine.StepStatusFailed)
		So(runtime.RunningSteps[1].RetryTimes, ShouldEqual, 0)
		So(runtime.RunningSteps[1].StepStatus, ShouldEqual, wfengine.FlowStatusPrepared)
	})

	Convey("do step failed and interrupt", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			nil,
			CreateMementoStorage(ctrl),
		)

		step := mock.NewMockStepAction(ctrl)
		wfManager.RegisterStep(step)

		errorMsg := "do step failed and interrupt"
		guard := monkey.PatchInstanceMethod(reflect.TypeOf(&wfengine.WfRuntime{}), "NewStepActionIns", func(rt *wfengine.WfRuntime, stepMeta *wfengine.StepMeta) (wfengine.StepAction, error) {
			step := mock.NewMockStepAction(ctrl)
			step.EXPECT().Init(Any(), Any()).Return(nil)
			step.EXPECT().DoStep(Any(), Any()).Return(define.NewInterruptError("", errorMsg))
			return step, nil
		})
		defer guard.Unpatch()

		hook := mock.NewMockWfHook(ctrl)
		hook.EXPECT().OnStepInit(Any()).Return(nil)
		hook.EXPECT().OnStepCompleted(Any()).AnyTimes().Return(nil)
		hook.EXPECT().OnWfInterrupt(Any()).AnyTimes().Return(nil)

		recover := mock.NewMockRecover(ctrl)
		recover.EXPECT().SaveResourceInterruptInfo(Any(), Any(), Any(), Any(), Any()).Return(nil)
		wfManager.RegisterRecover(recover)

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetState().Return(statemachine.StateCreating)
		resource.EXPECT().UpdateState(Any()).Return(resource, nil)

		logger := klogr.New()
		runtime := &wfengine.WfRuntime{
			FlowStatus:  wfengine.FlowStatusRunning,
			FlowName:    "WF1",
			RetryTimes:  1,
			FlowHookIns: hook,
			InitContext: map[string]interface{}{
				"_resourceName":      "test-resource",
				"_resourceNameSpace": "default",
			},
			Logger: klogr.New(),
			ResourceWorkflow: &wfengine.ResourceWorkflow{
				Logger:    logger,
				Resource:  resource,
				WfManager: wfManager,
			},
			RunningSteps: []*wfengine.StepRuntime{
				{
					StepName:   "Step1",
					StepStatus: wfengine.StepStatusPrepared,
				},
				{
					StepName:   "Step2",
					StepStatus: wfengine.StepStatusPrepared,
					RetryTimes: 0,
				},
			},
		}

		err = runtime.Start(context.TODO())
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, errorMsg)
		So(runtime.FlowStatus, ShouldEqual, wfengine.FlowStatusFailedCompleted)
		So(runtime.RunningSteps[0].RetryTimes, ShouldEqual, 0)
		So(runtime.RunningSteps[0].StepStatus, ShouldEqual, wfengine.StepStatusFailed)
		So(runtime.RunningSteps[1].RetryTimes, ShouldEqual, 0)
		So(runtime.RunningSteps[1].StepStatus, ShouldEqual, wfengine.FlowStatusPrepared)
	})

	Convey("do step interrupt, save interrupt info failed", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			nil,
			CreateMementoStorage(ctrl),
		)

		step := mock.NewMockStepAction(ctrl)
		wfManager.RegisterStep(step)

		errorMsg := "do step failed and interrupt"
		guard := monkey.PatchInstanceMethod(reflect.TypeOf(&wfengine.WfRuntime{}), "NewStepActionIns", func(rt *wfengine.WfRuntime, stepMeta *wfengine.StepMeta) (wfengine.StepAction, error) {
			step := mock.NewMockStepAction(ctrl)
			step.EXPECT().Init(Any(), Any()).Return(nil)
			step.EXPECT().DoStep(Any(), Any()).Return(define.NewInterruptError("", errorMsg))
			return step, nil
		})
		defer guard.Unpatch()

		hook := mock.NewMockWfHook(ctrl)
		hook.EXPECT().OnStepInit(Any()).Return(nil)
		hook.EXPECT().OnStepCompleted(Any()).AnyTimes().Return(nil)
		hook.EXPECT().OnWfInterrupt(Any()).AnyTimes().Return(nil)

		recover := mock.NewMockRecover(ctrl)
		recover.EXPECT().SaveResourceInterruptInfo(Any(), Any(), Any(), Any(), Any()).Return(errors.New("save interrupt info failed"))
		wfManager.RegisterRecover(recover)

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetState().Return(statemachine.StateCreating)

		logger := klogr.New()
		runtime := &wfengine.WfRuntime{
			FlowStatus:  wfengine.FlowStatusRunning,
			FlowName:    "WF1",
			RetryTimes:  1,
			FlowHookIns: hook,
			InitContext: map[string]interface{}{
				"_resourceName":      "test-resource",
				"_resourceNameSpace": "default",
			},
			Logger: klogr.New(),
			ResourceWorkflow: &wfengine.ResourceWorkflow{
				Logger:    logger,
				Resource:  resource,
				WfManager: wfManager,
			},
			RunningSteps: []*wfengine.StepRuntime{
				{
					StepName:   "Step1",
					StepStatus: wfengine.StepStatusPrepared,
				},
				{
					StepName:   "Step2",
					StepStatus: wfengine.StepStatusPrepared,
					RetryTimes: 0,
				},
			},
		}

		err = runtime.Start(context.TODO())
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, errorMsg)
		So(runtime.FlowStatus, ShouldEqual, wfengine.FlowStatusFailedCompleted)
		So(runtime.RunningSteps[0].RetryTimes, ShouldEqual, 0)
		So(runtime.RunningSteps[0].StepStatus, ShouldEqual, wfengine.StepStatusFailed)
		So(runtime.RunningSteps[1].RetryTimes, ShouldEqual, 0)
		So(runtime.RunningSteps[1].StepStatus, ShouldEqual, wfengine.FlowStatusPrepared)
	})

	Convey("do step interrupt, save interrupt status failed", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			nil,
			CreateMementoStorage(ctrl),
		)

		step := mock.NewMockStepAction(ctrl)
		wfManager.RegisterStep(step)

		errorMsg := "do step failed and interrupt"
		guard := monkey.PatchInstanceMethod(reflect.TypeOf(&wfengine.WfRuntime{}), "NewStepActionIns", func(rt *wfengine.WfRuntime, stepMeta *wfengine.StepMeta) (wfengine.StepAction, error) {
			step := mock.NewMockStepAction(ctrl)
			step.EXPECT().Init(Any(), Any()).Return(nil)
			step.EXPECT().DoStep(Any(), Any()).Return(define.NewInterruptError("", errorMsg))
			return step, nil
		})
		defer guard.Unpatch()

		hook := mock.NewMockWfHook(ctrl)
		hook.EXPECT().OnStepInit(Any()).Return(nil)
		hook.EXPECT().OnStepCompleted(Any()).AnyTimes().Return(nil)
		hook.EXPECT().OnWfInterrupt(Any()).AnyTimes().Return(nil)

		recover := mock.NewMockRecover(ctrl)
		recover.EXPECT().SaveResourceInterruptInfo(Any(), Any(), Any(), Any(), Any()).Return(nil)
		wfManager.RegisterRecover(recover)

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetState().Return(statemachine.StateCreating)
		resource.EXPECT().UpdateState(Any()).Return(resource, errors.New("UpdateState failed"))

		logger := klogr.New()
		runtime := &wfengine.WfRuntime{
			FlowStatus:  wfengine.FlowStatusRunning,
			FlowName:    "WF1",
			RetryTimes:  1,
			FlowHookIns: hook,
			InitContext: map[string]interface{}{
				"_resourceName":      "test-resource",
				"_resourceNameSpace": "default",
			},
			Logger: klogr.New(),
			ResourceWorkflow: &wfengine.ResourceWorkflow{
				Logger:    logger,
				Resource:  resource,
				WfManager: wfManager,
			},
			RunningSteps: []*wfengine.StepRuntime{
				{
					StepName:   "Step1",
					StepStatus: wfengine.StepStatusPrepared,
				},
				{
					StepName:   "Step2",
					StepStatus: wfengine.StepStatusPrepared,
					RetryTimes: 0,
				},
			},
		}

		err = runtime.Start(context.TODO())
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, errorMsg)
		So(runtime.FlowStatus, ShouldEqual, wfengine.FlowStatusFailedCompleted)
		So(runtime.RunningSteps[0].RetryTimes, ShouldEqual, 0)
		So(runtime.RunningSteps[0].StepStatus, ShouldEqual, wfengine.StepStatusFailed)
		So(runtime.RunningSteps[1].RetryTimes, ShouldEqual, 0)
		So(runtime.RunningSteps[1].StepStatus, ShouldEqual, wfengine.FlowStatusPrepared)
	})

	Convey("OnWfInterrupt failed", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			nil,
			CreateMementoStorage(ctrl),
		)

		step := mock.NewMockStepAction(ctrl)
		wfManager.RegisterStep(step)

		errorMsg := "do step failed and interrupt"
		guard := monkey.PatchInstanceMethod(reflect.TypeOf(&wfengine.WfRuntime{}), "NewStepActionIns", func(rt *wfengine.WfRuntime, stepMeta *wfengine.StepMeta) (wfengine.StepAction, error) {
			step := mock.NewMockStepAction(ctrl)
			step.EXPECT().Init(Any(), Any()).Return(nil)
			step.EXPECT().DoStep(Any(), Any()).Return(define.NewInterruptError("", errorMsg))
			return step, nil
		})
		defer guard.Unpatch()

		hook := mock.NewMockWfHook(ctrl)
		hook.EXPECT().OnStepInit(Any()).Return(nil)
		hook.EXPECT().OnStepCompleted(Any()).AnyTimes().Return(nil)
		hook.EXPECT().OnWfInterrupt(Any()).AnyTimes().Return(errors.New("OnWfInterrupt failed"))

		recover := mock.NewMockRecover(ctrl)
		recover.EXPECT().SaveResourceInterruptInfo(Any(), Any(), Any(), Any(), Any()).Return(nil)
		wfManager.RegisterRecover(recover)

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetState().Return(statemachine.StateCreating)
		resource.EXPECT().UpdateState(Any()).Return(resource, nil)

		logger := klogr.New()
		runtime := &wfengine.WfRuntime{
			FlowStatus:  wfengine.FlowStatusRunning,
			FlowName:    "WF1",
			RetryTimes:  1,
			FlowHookIns: hook,
			InitContext: map[string]interface{}{
				"_resourceName":      "test-resource",
				"_resourceNameSpace": "default",
			},
			Logger: klogr.New(),
			ResourceWorkflow: &wfengine.ResourceWorkflow{
				Logger:    logger,
				Resource:  resource,
				WfManager: wfManager,
			},
			RunningSteps: []*wfengine.StepRuntime{
				{
					StepName:   "Step1",
					StepStatus: wfengine.StepStatusPrepared,
				},
				{
					StepName:   "Step2",
					StepStatus: wfengine.StepStatusPrepared,
					RetryTimes: 0,
				},
			},
		}

		err = runtime.Start(context.TODO())
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, errorMsg)
		So(runtime.FlowStatus, ShouldEqual, wfengine.FlowStatusFailedCompleted)
		So(runtime.RunningSteps[0].RetryTimes, ShouldEqual, 0)
		So(runtime.RunningSteps[0].StepStatus, ShouldEqual, wfengine.StepStatusFailed)
		So(runtime.RunningSteps[1].RetryTimes, ShouldEqual, 0)
		So(runtime.RunningSteps[1].StepStatus, ShouldEqual, wfengine.FlowStatusPrepared)
	})

	Convey("interrupt after more than 10 failed", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			nil,
			CreateMementoStorage(ctrl),
		)

		step := mock.NewMockStepAction(ctrl)
		wfManager.RegisterStep(step)

		errorMsg := "do step failed"
		guard := monkey.PatchInstanceMethod(reflect.TypeOf(&wfengine.WfRuntime{}), "NewStepActionIns", func(rt *wfengine.WfRuntime, stepMeta *wfengine.StepMeta) (wfengine.StepAction, error) {
			step := mock.NewMockStepAction(ctrl)
			step.EXPECT().Init(Any(), Any()).Return(nil)
			step.EXPECT().DoStep(Any(), Any()).Return(errors.New(errorMsg))
			return step, nil
		})
		defer guard.Unpatch()

		hook := mock.NewMockWfHook(ctrl)
		hook.EXPECT().OnStepInit(Any()).Return(nil)
		hook.EXPECT().OnStepCompleted(Any()).AnyTimes().Return(nil)
		hook.EXPECT().OnWfInterrupt(Any()).AnyTimes().Return(nil)

		recover := mock.NewMockRecover(ctrl)
		recover.EXPECT().SaveResourceInterruptInfo(Any(), Any(), Any(), Any(), Any()).Return(nil)
		wfManager.RegisterRecover(recover)

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetState().Return(statemachine.StateNil)
		resource.EXPECT().UpdateState(Any()).Return(resource, nil)

		logger := klogr.New()
		runtime := &wfengine.WfRuntime{
			FlowStatus:  wfengine.FlowStatusRunning,
			FlowName:    "WF1",
			RetryTimes:  9,
			FlowHookIns: hook,
			InitContext: map[string]interface{}{
				"_resourceName":      "test-resource",
				"_resourceNameSpace": "default",
			},
			Logger: klogr.New(),
			ResourceWorkflow: &wfengine.ResourceWorkflow{
				Logger:    logger,
				Resource:  resource,
				WfManager: wfManager,
			},
			RunningSteps: []*wfengine.StepRuntime{
				{
					StepName:   "Step1",
					StepStatus: wfengine.StepStatusInited,
					RetryTimes: 9,
				},
				{
					StepName:   "Step2",
					StepStatus: wfengine.StepStatusPrepared,
					RetryTimes: 0,
				},
			},
		}

		err = runtime.Start(context.TODO())
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, errorMsg)
		So(runtime.FlowStatus, ShouldEqual, wfengine.FlowStatusFailedCompleted)
		So(runtime.RunningSteps[0].RetryTimes, ShouldEqual, 10)
		So(runtime.RunningSteps[0].StepStatus, ShouldEqual, wfengine.StepStatusFailed)
		So(runtime.RunningSteps[1].RetryTimes, ShouldEqual, 0)
		So(runtime.RunningSteps[1].StepStatus, ShouldEqual, wfengine.FlowStatusPrepared)
	})

	Convey("workflow runtime start succeed", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			nil,
			CreateMementoStorage(ctrl),
		)

		step := mock.NewMockStepAction(ctrl)
		wfManager.RegisterStep(step)

		guard := monkey.PatchInstanceMethod(reflect.TypeOf(&wfengine.WfRuntime{}), "NewStepActionIns", func(rt *wfengine.WfRuntime, stepMeta *wfengine.StepMeta) (wfengine.StepAction, error) {
			step := mock.NewMockStepAction(ctrl)
			step.EXPECT().Init(Any(), Any()).Return(nil)
			step.EXPECT().DoStep(Any(), Any()).Return(nil)
			step.EXPECT().Output(Any()).Return(nil)
			return step, nil
		})
		defer guard.Unpatch()

		resource := mock.NewMockStateResource(ctrl)

		hook := mock.NewMockWfHook(ctrl)
		hook.EXPECT().OnStepInit(Any()).AnyTimes().Return(nil)
		hook.EXPECT().OnStepCompleted(Any()).AnyTimes().Return(nil)
		hook.EXPECT().OnWfCompleted().Return(nil)

		logger := klogr.New()
		runtime := &wfengine.WfRuntime{
			FlowStatus:  wfengine.FlowStatusRunning,
			FlowName:    "WF1",
			RetryTimes:  1,
			FlowHookIns: hook,
			InitContext: map[string]interface{}{
				"_resourceName":      "test-resource",
				"_resourceNameSpace": "default",
			},
			Logger: klogr.New(),
			ResourceWorkflow: &wfengine.ResourceWorkflow{
				Logger:    logger,
				Resource:  resource,
				WfManager: wfManager,
			},
			RunningSteps: []*wfengine.StepRuntime{
				{
					StepName:   "Step1",
					StepStatus: wfengine.StepStatusInited,
					RetryTimes: 1,
				},
				{
					StepName:   "Step2",
					StepStatus: wfengine.StepStatusPrepared,
					RetryTimes: 0,
				},
			},
		}

		err = runtime.Start(context.TODO())
		So(err, ShouldBeNil)
		So(runtime.FlowStatus, ShouldEqual, wfengine.FlowStatusCompleted)
		So(runtime.RetryTimes, ShouldEqual, 2)
		So(runtime.RunningSteps[0].RetryTimes, ShouldEqual, 2)
		So(runtime.RunningSteps[0].StepStatus, ShouldEqual, wfengine.StepStatusCompleted)
		So(runtime.RunningSteps[1].RetryTimes, ShouldEqual, 0)
		So(runtime.RunningSteps[1].StepStatus, ShouldEqual, wfengine.StepStatusCompleted)

	})

	Convey("workflow cancelled", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			nil,
			CreateMementoStorage(ctrl),
		)

		step := mock.NewMockStepAction(ctrl)
		wfManager.RegisterStep(step)

		recover := mock.NewMockRecover(ctrl)
		recover.EXPECT().SaveResourceInterruptInfo(Any(), Any(), Any(), Any(), Any()).Return(nil)
		wfManager.RegisterRecover(recover)

		guard := monkey.PatchInstanceMethod(reflect.TypeOf(&wfengine.WfRuntime{}), "NewStepActionIns", func(rt *wfengine.WfRuntime, stepMeta *wfengine.StepMeta) (wfengine.StepAction, error) {
			step := mock.NewMockStepAction(ctrl)
			step.EXPECT().Init(Any(), Any()).AnyTimes().Return(nil)
			step.EXPECT().DoStep(Any(), Any()).AnyTimes().Return(nil)
			step.EXPECT().Output(Any()).AnyTimes().Return(nil)
			return step, nil
		})
		defer guard.Unpatch()

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetState().Return(statemachine.StateRunning)
		resource.EXPECT().UpdateState(Any()).Return(resource, nil)

		hook := mock.NewMockWfHook(ctrl)
		hook.EXPECT().OnStepInit(Any()).AnyTimes().Return(nil)
		hook.EXPECT().OnStepCompleted(Any()).AnyTimes().Return(nil)
		hook.EXPECT().OnWfCompleted().AnyTimes().Return(nil)
		hook.EXPECT().OnWfInterrupt(Any()).AnyTimes().Return(nil)

		logger := klogr.New()
		runtime := &wfengine.WfRuntime{
			FlowStatus:  wfengine.FlowStatusRunning,
			FlowName:    "WF1",
			RetryTimes:  1,
			FlowHookIns: hook,
			InitContext: map[string]interface{}{
				"_resourceName":      "test-resource",
				"_resourceNameSpace": "default",
			},
			Logger: klogr.New(),
			ResourceWorkflow: &wfengine.ResourceWorkflow{
				Logger:    logger,
				Resource:  resource,
				WfManager: wfManager,
			},
			RunningSteps: []*wfengine.StepRuntime{
				{
					StepName:   "Step1",
					StepStatus: wfengine.StepStatusInited,
					RetryTimes: 1,
				},
				{
					StepName:   "Step2",
					StepStatus: wfengine.StepStatusPrepared,
					RetryTimes: 0,
				},
			},
		}
		ctx, cancel := context.WithCancel(context.TODO())
		go func() {
			cancel()
		}()
		err = runtime.Start(ctx)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "CANCELLED")
		So(runtime.FlowStatus, ShouldEqual, wfengine.FlowStatusFailedCompleted)
		So(runtime.ErrorMessage, ShouldContainSubstring, "CANCELLED")
	})

}

func getCommonContext(resource statemachine.StateResource) map[string]interface{} {
	return map[string]interface{}{
		define.DefaultWfConf[define.WorkFlowResourceName]:      resource.GetName(),
		define.DefaultWfConf[define.WorkFlowResourceNameSpace]: resource.GetNamespace(),
	}
}
