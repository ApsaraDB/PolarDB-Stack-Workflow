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
	"reflect"
	"strings"
	"testing"

	"github.com/bouk/monkey"

	"gitlab.alibaba-inc.com/polar-as/polar-wf-engine/test/dbclusterwf"
	"k8s.io/klog/klogr"

	. "github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
	"gitlab.alibaba-inc.com/polar-as/polar-wf-engine/statemachine"
	"gitlab.alibaba-inc.com/polar-as/polar-wf-engine/test/mock"
	"gitlab.alibaba-inc.com/polar-as/polar-wf-engine/wfengine"
)

var resourceType = "dbcluster"
var workFlowMetaDir = "test/dbclusterwf/workflow"

func CreateWfMetaLoader(ctrl *Controller) wfengine.CreateMetaLoaderFunc {
	return func(resourceType string) (wfengine.WfMetaLoader, error) {
		wfMetaLoader := mock.NewMockWfMetaLoader(ctrl)
		wfMetaLoader.EXPECT().GetAllFlowMeta(Any()).AnyTimes().Return(map[string]*wfengine.FlowMeta{
			"WF1": {
				FlowName:             "WF1",
				RecoverFromFirstStep: true,
				Steps: []*wfengine.StepMeta{
					{
						ClassName: "mock.MockStepAction",
						StepName:  "Step1",
					},
					{
						ClassName: "mock.MockStepAction",
						StepName:  "Step2",
					},
				},
			},
			"WF2": {
				FlowName:             "WF2",
				RecoverFromFirstStep: false,
				Steps: []*wfengine.StepMeta{
					{
						ClassName: "mock.MockStepAction",
						StepName:  "Step1",
					},
					{
						ClassName: "Group1",
						StepName:  "Step2",
					},
				},
			},
		}, map[string]*wfengine.StepGroupMeta{
			"Group1": {
				GroupName: "Group1",
				Steps: []*wfengine.StepMeta{
					{
						ClassName: "mock.MockStepAction",
						StepName:  "GroupStep1",
					},
					{
						ClassName: "mock.MockStepAction",
						StepName:  "GroupStep2",
					},
				},
			},
		}, nil)
		return wfMetaLoader, nil
	}
}

func CreateHook(ctrl *Controller) wfengine.CreateHookFunc {
	return func(wfRuntime *wfengine.WfRuntime) (wfengine.WfHook, error) {
		return nil, nil
	}
}

func CreateMementoStorage(ctrl *Controller) wfengine.CreateMementoStorageFunc {
	return func(resource statemachine.StateResource) (wfengine.WfRuntimeMementoStorage, error) {
		mementoStorage := mock.NewMockWfRuntimeMementoStorage(ctrl)
		mementoStorage.EXPECT().LoadMementoMap(Any()).AnyTimes().Return(map[string]string{}, nil)
		mementoStorage.EXPECT().Save(Any(), Any()).AnyTimes().Return(nil)
		return mementoStorage, nil
	}
}

func Test_CreateWfManager(t *testing.T) {
	Convey("ResourceType is empty error", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()
		wfManager, err := wfengine.CreateWfManager(
			"",
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			CreateHook(ctrl),
			CreateMementoStorage(ctrl),
		)
		So(wfManager, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "ResourceType is empty")
	})

	Convey("WorkFlowMetaDir is empty error", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()
		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			"",
			CreateWfMetaLoader(ctrl),
			CreateHook(ctrl),
			CreateMementoStorage(ctrl),
		)
		So(wfManager, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "WorkFlowMetaDir is empty")
	})

	Convey("createMetaLoaderFunc is nil error", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()
		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			nil,
			CreateHook(ctrl),
			CreateMementoStorage(ctrl),
		)
		So(wfManager, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "createMetaLoaderFunc is nil")
	})

	Convey("createMetaLoaderFunc failed", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()
		errorMsg := "createMetaLoaderFunc failed"
		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			func(resourceType string) (wfengine.WfMetaLoader, error) {
				return nil, errors.New(errorMsg)
			},
			CreateHook(ctrl),
			CreateMementoStorage(ctrl),
		)
		So(wfManager, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, errorMsg)
	})

	Convey("GetAllFlowMeta failed", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()
		errorMsg := "GetAllFlowMeta failed"
		createMetaLoaderFunc := func(resourceType string) (wfengine.WfMetaLoader, error) {
			ctrl := NewController(t)
			defer ctrl.Finish()
			wfMetaLoader := mock.NewMockWfMetaLoader(ctrl)
			wfMetaLoader.EXPECT().GetAllFlowMeta(Any()).AnyTimes().Return(nil, nil, errors.New(errorMsg))
			return wfMetaLoader, nil
		}

		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			createMetaLoaderFunc,
			CreateHook(ctrl),
			CreateMementoStorage(ctrl),
		)
		So(wfManager, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, errorMsg)
	})

	Convey("replaceGroupStep succeed", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()
		wfManager, err := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			CreateHook(ctrl),
			CreateMementoStorage(ctrl),
		)
		So(wfManager, ShouldNotBeNil)
		So(err, ShouldBeNil)
		So(wfManager.FlowMetaMap["WF2"].Steps, ShouldHaveLength, 3)
	})
}

func Test_RegisterStep(t *testing.T) {
	ctrl := NewController(t)
	defer ctrl.Finish()
	wfManager, _ := wfengine.CreateWfManager(
		resourceType,
		workFlowMetaDir,
		CreateWfMetaLoader(ctrl),
		CreateHook(ctrl),
		CreateMementoStorage(ctrl),
	)

	Convey("RegisterStep succeed", t, func() {
		step := &dbclusterwf.CreateStep1{}
		wfManager.RegisterStep(step)
		stepStructName := strings.Replace(reflect.TypeOf(step).String(), "*", "", -1)
		So(wfManager.TypeRegistry[stepStructName] == reflect.TypeOf(step), ShouldBeTrue)
	})
}

func Test_RegisterLogger(t *testing.T) {
	ctrl := NewController(t)
	defer ctrl.Finish()
	wfManager, _ := wfengine.CreateWfManager(
		resourceType,
		workFlowMetaDir,
		CreateWfMetaLoader(ctrl),
		CreateHook(ctrl),
		CreateMementoStorage(ctrl),
	)
	Convey("RegisterLogger succeed", t, func() {
		logger := klogr.New()
		wfManager.RegisterLogger(logger)
		So(wfManager.Logger, ShouldNotBeNil)
	})
}

func Test_CreateResourceWorkflow(t *testing.T) {
	ctrl := NewController(t)
	defer ctrl.Finish()
	wfManager, _ := wfengine.CreateWfManager(
		resourceType,
		workFlowMetaDir,
		CreateWfMetaLoader(ctrl),
		CreateHook(ctrl),
		CreateMementoStorage(ctrl),
	)

	Convey("CreateResourceWorkflow succeed", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()
		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")
		resourceWf, err := wfManager.CreateResourceWorkflow(resource)
		So(err, ShouldBeNil)
		So(resourceWf, ShouldNotBeNil)
	})

}

func Test_CheckRecovery(t *testing.T) {
	ctrl := NewController(t)
	defer ctrl.Finish()

	createWfManager := func() *wfengine.WfManager {
		wfManager, _ := wfengine.CreateWfManager(
			resourceType,
			workFlowMetaDir,
			CreateWfMetaLoader(ctrl),
			CreateHook(ctrl),
			CreateMementoStorage(ctrl),
		)
		return wfManager
	}
	createResource := func() *mock.MockStateResource {
		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")
		return resource
	}

	guard := monkey.Patch(
		wfengine.CreateWfRuntimeMementoCareTaker,
		func(resource statemachine.StateResource, createStorageFunc wfengine.CreateMementoStorageFunc) (*wfengine.WfRuntimeMementoCareTaker, error) {
			return &wfengine.WfRuntimeMementoCareTaker{
				Name:              resource.GetName() + "-" + "workflow",
				Namespace:         resource.GetNamespace(),
				FlowApplyResource: resource.GetName(),
				MementoMap:        map[string]string{},
				MementoStorage:    nil,
			}, nil
		},
	)
	defer guard.Unpatch()

	Convey("not recover when status is not interrupted", t, func() {
		recover := mock.NewMockRecover(ctrl)
		wfManager := createWfManager()
		wfManager.RegisterRecover(recover)

		resource := createResource()
		resource.EXPECT().GetState().AnyTimes().Return(statemachine.StateRunning)
		result := wfManager.CheckRecovery(resource)
		So(result, ShouldBeFalse)
	})

	Convey("not recover when the conditions for recovery are not met", t, func() {
		recover := mock.NewMockRecover(ctrl)
		recover.EXPECT().GetResourceRecoverInfo(Any(), Any()).Return(false, statemachine.StateNil)
		wfManager := createWfManager()
		wfManager.RegisterRecover(recover)

		resource := createResource()
		resource.EXPECT().GetState().AnyTimes().Return(statemachine.StateInterrupt)

		result := wfManager.CheckRecovery(resource)
		So(result, ShouldBeFalse)
	})

	Convey("retry when failed to recover to the previous state", t, func() {
		errorMsg := "CreateResourceWorkflow failed"
		recover := mock.NewMockRecover(ctrl)
		recover.EXPECT().GetResourceRecoverInfo(Any(), Any()).Return(true, statemachine.StateCreating)
		wfManager := createWfManager()
		wfManager.RegisterRecover(recover)

		resource := createResource()
		resource.EXPECT().GetState().AnyTimes().Return(statemachine.StateInterrupt)
		resource.EXPECT().UpdateState(Any()).AnyTimes().Return(resource, nil)

		guard := monkey.Patch(wfengine.CreateResourceWorkflow, func(resource statemachine.StateResource, wfManager *wfengine.WfManager) (*wfengine.ResourceWorkflow, error) {
			return nil, errors.New(errorMsg)
		})
		defer guard.Unpatch()

		result := wfManager.CheckRecovery(resource)
		So(result, ShouldBeTrue)
	})

	Convey("retry again when failed to retry interrupted step", t, func() {
		errorMsg := "RetryInterruptedStep failed"
		recover := mock.NewMockRecover(ctrl)
		recover.EXPECT().GetResourceRecoverInfo(Any(), Any()).Return(true, statemachine.StateCreating)
		wfManager := createWfManager()
		wfManager.RegisterRecover(recover)

		resource := createResource()
		resource.EXPECT().GetState().AnyTimes().Return(statemachine.StateInterrupt)
		resource.EXPECT().UpdateState(Any()).AnyTimes().Return(resource, nil)

		guard1 := monkey.PatchInstanceMethod(reflect.TypeOf(&wfengine.ResourceWorkflow{}), "RetryInterruptedStep", func(m *wfengine.ResourceWorkflow) error {
			return errors.New(errorMsg)
		})
		defer guard1.Unpatch()

		guard2 := monkey.Patch(wfengine.CreateResourceWorkflow, func(resource statemachine.StateResource, wfManager *wfengine.WfManager) (*wfengine.ResourceWorkflow, error) {
			return &wfengine.ResourceWorkflow{}, nil
		})
		defer guard2.Unpatch()

		result := wfManager.CheckRecovery(resource)
		So(result, ShouldBeTrue)
	})

	Convey("recover from creating state", t, func() {
		recover := mock.NewMockRecover(ctrl)
		recover.EXPECT().GetResourceRecoverInfo(Any(), Any()).Return(true, statemachine.StateCreating)
		recover.EXPECT().CancelResourceRecover(Any(), Any()).Return(nil)
		wfManager := createWfManager()
		wfManager.RegisterRecover(recover)

		resource := createResource()
		resource.EXPECT().GetState().AnyTimes().Return(statemachine.StateInterrupt)
		resource.EXPECT().UpdateState(Any()).AnyTimes().Return(resource, nil)

		guardCoreV1 := monkey.PatchInstanceMethod(reflect.TypeOf(&wfengine.ResourceWorkflow{}), "RetryInterruptedStep", func(m *wfengine.ResourceWorkflow) error {
			return nil
		})
		defer guardCoreV1.Unpatch()

		result := wfManager.CheckRecovery(resource)
		So(result, ShouldBeTrue)
	})

	Convey("recover from rebuild state", t, func() {
		recover := mock.NewMockRecover(ctrl)
		recover.EXPECT().GetResourceRecoverInfo(Any(), Any()).Return(true, statemachine.StateRebuild)
		recover.EXPECT().CancelResourceRecover(Any(), Any()).Return(nil)
		wfManager := createWfManager()
		wfManager.RegisterRecover(recover)

		resource := createResource()
		resource.EXPECT().GetState().AnyTimes().Return(statemachine.StateInterrupt)
		resource.EXPECT().UpdateState(Any()).AnyTimes().Return(resource, nil)

		guardCoreV1 := monkey.PatchInstanceMethod(reflect.TypeOf(&wfengine.ResourceWorkflow{}), "RetryInterruptedStep", func(m *wfengine.ResourceWorkflow) error {
			return nil
		})
		defer guardCoreV1.Unpatch()

		result := wfManager.CheckRecovery(resource)
		So(result, ShouldBeTrue)
	})

	Convey("recover interrupted flow, cancel resource recover info failed", t, func() {
		errorMsg := "CancelResourceRecover failed"
		recover := mock.NewMockRecover(ctrl)
		recover.EXPECT().GetResourceRecoverInfo(Any(), Any()).Return(true, statemachine.StateRebuild)
		recover.EXPECT().CancelResourceRecover(Any(), Any()).Return(errors.New(errorMsg))
		wfManager := createWfManager()
		wfManager.RegisterRecover(recover)

		resource := createResource()
		resource.EXPECT().GetState().AnyTimes().Return(statemachine.StateInterrupt)

		result := wfManager.CheckRecovery(resource)
		So(result, ShouldBeTrue)
	})

	Convey("recover interrupted flow, update resource status failed", t, func() {
		errorMsg := "UpdateState failed"
		recover := mock.NewMockRecover(ctrl)
		recover.EXPECT().GetResourceRecoverInfo(Any(), Any()).Return(true, statemachine.StateRebuild)
		recover.EXPECT().CancelResourceRecover(Any(), Any()).Return(nil)
		wfManager := createWfManager()
		wfManager.RegisterRecover(recover)

		resource := createResource()
		resource.EXPECT().GetState().AnyTimes().Return(statemachine.StateInterrupt)
		resource.EXPECT().UpdateState(Any()).AnyTimes().Return(nil, errors.New(errorMsg))

		result := wfManager.CheckRecovery(resource)
		So(result, ShouldBeTrue)
	})
}
