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
	"testing"

	"github.com/bouk/monkey"

	implement_wfengine "github.com/ApsaraDB/PolarDB-Stack-Workflow/implement/wfengine"
	"github.com/ApsaraDB/PolarDB-Stack-Workflow/test/mock"
	"github.com/ApsaraDB/PolarDB-Stack-Workflow/wfengine"
	. "github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
	"k8s.io/klog/klogr"
)

func Test_CreateDefaultWorkflowHook(t *testing.T) {
	Convey("create default workflow hook succeed", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		logger := klogr.New()
		resource := mock.NewMockStateResource(ctrl)
		runtime := &wfengine.WfRuntime{
			FlowStatus: wfengine.FlowStatusRunning,
			FlowName:   "WF1",
			Logger:     logger,
			ResourceWorkflow: &wfengine.ResourceWorkflow{
				Logger:   logger,
				Resource: resource,
				WfManager: &wfengine.WfManager{
					ResourceType: "dbcluster",
				},
			},
		}
		hook, err := implement_wfengine.CreateDefaultWorkflowHook(runtime)
		So(err, ShouldBeNil)
		So(hook, ShouldNotBeNil)
	})
}

func Test_OnWfInit(t *testing.T) {
	Convey("OnWfInit succeed", t, func() {
		mockAndPatch(t, func(hook wfengine.WfHook) {
			err := hook.OnWfInit()
			So(err, ShouldBeNil)
		})
	})
}

func Test_OnStepCompleted(t *testing.T) {
	Convey("OnStepCompleted succeed", t, func() {
		mockAndPatch(t, func(hook wfengine.WfHook) {
			err := hook.OnStepCompleted(nil)
			So(err, ShouldBeNil)
		})
	})
}

func Test_OnStepInit(t *testing.T) {
	Convey("OnStepInit succeed", t, func() {
		mockAndPatch(t, func(hook wfengine.WfHook) {
			err := hook.OnStepInit(nil)
			So(err, ShouldBeNil)
		})
	})
}

func Test_OnWfInterrupt(t *testing.T) {
	Convey("OnWfInterrupt succeed", t, func() {
		mockAndPatch(t, func(hook wfengine.WfHook) {
			err := hook.OnWfInterrupt(nil)
			So(err, ShouldBeNil)
		})
	})
}

func Test_OnWfCompleted(t *testing.T) {
	Convey("OnWfCompleted succeed", t, func() {
		mockAndPatch(t, func(hook wfengine.WfHook) {
			err := hook.OnWfCompleted()
			So(err, ShouldBeNil)
		})
	})
}

func Test_OnStepWaiting(t *testing.T) {
	Convey("OnStepWaiting succeed", t, func() {
		mockAndPatch(t, func(hook wfengine.WfHook) {
			err := hook.OnStepWaiting(nil)
			So(err, ShouldBeNil)
		})
	})
}

func mockAndPatch(t *testing.T, action func(wfengine.WfHook)) {
	ctrl := NewController(t)
	defer ctrl.Finish()

	logger := klogr.New()
	resource := mock.NewMockStateResource(ctrl)
	runtime := &wfengine.WfRuntime{
		FlowStatus: wfengine.FlowStatusRunning,
		FlowName:   "WF1",
		Logger:     logger,
		ResourceWorkflow: &wfengine.ResourceWorkflow{
			Logger:   logger,
			Resource: resource,
			WfManager: &wfengine.WfManager{
				ResourceType: "dbcluster",
			},
		},
	}
	hook, err := implement_wfengine.CreateDefaultWorkflowHook(runtime)
	So(err, ShouldBeNil)
	So(hook, ShouldNotBeNil)

	guard := monkey.PatchInstanceMethod(
		reflect.TypeOf(&implement_wfengine.DefaultHistoryRecorder{}), "WriteWorkflowStepLog", func(*implement_wfengine.DefaultHistoryRecorder, *wfengine.WfRuntime, *wfengine.StepRuntime, implement_wfengine.StepType) {
			return
		})
	defer guard.Unpatch()

	action(hook)
}
