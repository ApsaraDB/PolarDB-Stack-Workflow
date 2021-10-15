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
	j "encoding/json"
	"testing"

	"github.com/bouk/monkey"
	"github.com/pkg/errors"

	. "github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
	implement_wfengine "gitlab.alibaba-inc.com/polar-as/polar-wf-engine/implement/wfengine"
	"gitlab.alibaba-inc.com/polar-as/polar-wf-engine/test/mock"
	"gitlab.alibaba-inc.com/polar-as/polar-wf-engine/wfengine"
	"k8s.io/klog/klogr"
)

func Test_WriteWorkflowStepLog(t *testing.T) {

	Convey("write workflow start log", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		logger := klogr.New()
		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")

		runtime := &wfengine.WfRuntime{
			FlowStatus: wfengine.FlowStatusRunning,
			FlowName:   "WF1",
			Logger:     logger,
			StartTime:  "2021-08-02 10:03:05",
			ResourceWorkflow: &wfengine.ResourceWorkflow{
				Logger:   logger,
				Resource: resource,
				WfManager: &wfengine.WfManager{
					ResourceType: "dbcluster",
				},
			},
			RunningSteps: []*wfengine.StepRuntime{
				{
					StepName: "Step1",
				},
				{
					StepName: "Step2",
				},
			},
		}

		recorder := implement_wfengine.CreateDefaultHistoryRecorder("dbcluster", klogr.New(), "")
		So(recorder, ShouldNotBeNil)
		recorder.WriteWorkflowStepLog(runtime, nil, implement_wfengine.WorkFlowStart)
	})

	Convey("write workflow complete log", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		logger := klogr.New()
		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")

		runtime := &wfengine.WfRuntime{
			FlowStatus: wfengine.FlowStatusRunning,
			FlowName:   "WF1",
			Logger:     logger,
			StartTime:  "2021-08-02 10:03:05",
			ResourceWorkflow: &wfengine.ResourceWorkflow{
				Resource: resource,
			},
		}

		recorder := implement_wfengine.CreateDefaultHistoryRecorder("dbcluster", klogr.New(), "")
		So(recorder, ShouldNotBeNil)
		recorder.WriteWorkflowStepLog(runtime, nil, "")
	})

	Convey("write workflow interrupt log", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		logger := klogr.New()
		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")

		runtime := &wfengine.WfRuntime{
			FlowStatus: wfengine.FlowStatusRunning,
			FlowName:   "WF1",
			Logger:     logger,
			StartTime:  "2021-08-02 10:03:05",
			ResourceWorkflow: &wfengine.ResourceWorkflow{
				Resource: resource,
			},
		}

		recorder := implement_wfengine.CreateDefaultHistoryRecorder("dbcluster", klogr.New(), "")
		So(recorder, ShouldNotBeNil)
		recorder.WriteWorkflowStepLog(runtime, nil, implement_wfengine.WorkFlowInterrupt)
	})

	Convey("failed to parse workflow startTime", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		logger := klogr.New()
		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")

		runtime := &wfengine.WfRuntime{
			FlowStatus: wfengine.FlowStatusRunning,
			FlowName:   "WF1",
			Logger:     logger,
			StartTime:  "2021-08-02_10:03:05",
			ResourceWorkflow: &wfengine.ResourceWorkflow{
				Resource: resource,
			},
		}

		recorder := implement_wfengine.CreateDefaultHistoryRecorder("dbcluster", klogr.New(), "")
		So(recorder, ShouldNotBeNil)
		recorder.WriteWorkflowStepLog(runtime, nil, implement_wfengine.WorkFlowInterrupt)
	})

	Convey("write step init log", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		logger := klogr.New()
		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")

		runtime := &wfengine.WfRuntime{
			FlowStatus: wfengine.FlowStatusRunning,
			FlowName:   "WF1",
			Logger:     logger,
			StartTime:  "2021-08-02 10:03:05",
			ResourceWorkflow: &wfengine.ResourceWorkflow{
				Logger:   logger,
				Resource: resource,
				WfManager: &wfengine.WfManager{
					ResourceType: "dbcluster",
				},
			},
			RunningSteps: []*wfengine.StepRuntime{
				{
					StepName: "Step1",
				},
				{
					StepName: "Step2",
				},
			},
		}

		stepRuntime := &wfengine.StepRuntime{
			StepName:      "Step1",
			StepStatus:    wfengine.StepStatusInited,
			StepStartTime: "2021-08-02 10:04:05",
		}

		recorder := implement_wfengine.CreateDefaultHistoryRecorder("dbcluster", klogr.New(), "")
		So(recorder, ShouldNotBeNil)
		recorder.WriteWorkflowStepLog(runtime, stepRuntime, "")
	})

	Convey("write step waiting log", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		logger := klogr.New()
		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")

		runtime := &wfengine.WfRuntime{
			FlowStatus: wfengine.FlowStatusRunning,
			FlowName:   "WF1",
			Logger:     logger,
			StartTime:  "2021-08-02 10:03:05",
			ResourceWorkflow: &wfengine.ResourceWorkflow{
				Logger:   logger,
				Resource: resource,
				WfManager: &wfengine.WfManager{
					ResourceType: "dbcluster",
				},
			},
			RunningSteps: []*wfengine.StepRuntime{
				{
					StepName: "Step1",
				},
				{
					StepName: "Step2",
				},
			},
		}

		stepRuntime := &wfengine.StepRuntime{
			StepName:      "Step1",
			StepStatus:    wfengine.StepStatusWaiting,
			StepStartTime: "2021-08-02 10:04:05",
		}

		recorder := implement_wfengine.CreateDefaultHistoryRecorder("dbcluster", klogr.New(), "")
		So(recorder, ShouldNotBeNil)
		recorder.WriteWorkflowStepLog(runtime, stepRuntime, "")
	})

	Convey("write step failed log", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		logger := klogr.New()
		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")

		runtime := &wfengine.WfRuntime{
			FlowStatus: wfengine.FlowStatusRunning,
			FlowName:   "WF1",
			Logger:     logger,
			StartTime:  "2021-08-02 10:03:05",
			ResourceWorkflow: &wfengine.ResourceWorkflow{
				Logger:   logger,
				Resource: resource,
				WfManager: &wfengine.WfManager{
					ResourceType: "dbcluster",
				},
			},
			RunningSteps: []*wfengine.StepRuntime{
				{
					StepName: "Step1",
				},
				{
					StepName: "Step2",
				},
			},
		}

		stepRuntime := &wfengine.StepRuntime{
			StepName:      "Step1",
			StepStatus:    wfengine.StepStatusFailed,
			StepStartTime: "2021-08-02 10:04:05",
		}

		recorder := implement_wfengine.CreateDefaultHistoryRecorder("dbcluster", klogr.New(), "")
		So(recorder, ShouldNotBeNil)
		recorder.WriteWorkflowStepLog(runtime, stepRuntime, "")
	})

	Convey("write step complete log", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		logger := klogr.New()
		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")

		runtime := &wfengine.WfRuntime{
			FlowStatus: wfengine.FlowStatusRunning,
			FlowName:   "WF1",
			Logger:     logger,
			StartTime:  "2021-08-02 10:03:05",
			ResourceWorkflow: &wfengine.ResourceWorkflow{
				Logger:   logger,
				Resource: resource,
				WfManager: &wfengine.WfManager{
					ResourceType: "dbcluster",
				},
			},
			RunningSteps: []*wfengine.StepRuntime{
				{
					StepName: "Step1",
				},
				{
					StepName: "Step2",
				},
			},
		}

		stepRuntime := &wfengine.StepRuntime{
			StepName:      "Step1",
			StepStatus:    wfengine.StepStatusCompleted,
			StepStartTime: "2021-08-02 10:04:05",
		}

		recorder := implement_wfengine.CreateDefaultHistoryRecorder("dbcluster", klogr.New(), "")
		So(recorder, ShouldNotBeNil)
		recorder.WriteWorkflowStepLog(runtime, stepRuntime, "")
	})

	Convey("not write step log", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		logger := klogr.New()
		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")

		runtime := &wfengine.WfRuntime{
			FlowStatus: wfengine.FlowStatusRunning,
			FlowName:   "WF1",
			Logger:     logger,
			StartTime:  "2021-08-02 10:03:05",
			ResourceWorkflow: &wfengine.ResourceWorkflow{
				Logger:   logger,
				Resource: resource,
				WfManager: &wfengine.WfManager{
					ResourceType: "dbcluster",
				},
			},
			RunningSteps: []*wfengine.StepRuntime{
				{
					StepName: "Step1",
				},
				{
					StepName: "Step2",
				},
			},
		}

		stepRuntime := &wfengine.StepRuntime{
			StepName:      "Step1",
			StepStatus:    wfengine.StepStatusPrepared,
			StepStartTime: "2021-08-02 10:04:05",
		}

		recorder := implement_wfengine.CreateDefaultHistoryRecorder("dbcluster", klogr.New(), "")
		So(recorder, ShouldNotBeNil)
		recorder.WriteWorkflowStepLog(runtime, stepRuntime, "")
	})

	Convey("json.Marshal failed", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		errorMsg := "json.Marshal failed"
		guard := monkey.Patch(j.Marshal, func(v interface{}) ([]byte, error) {
			return nil, errors.New(errorMsg)
		})
		defer guard.Unpatch()

		logger := klogr.New()
		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")

		runtime := &wfengine.WfRuntime{
			FlowStatus: wfengine.FlowStatusRunning,
			FlowName:   "WF1",
			Logger:     logger,
			StartTime:  "2021-08-02 10:03:05",
			ResourceWorkflow: &wfengine.ResourceWorkflow{
				Logger:   logger,
				Resource: resource,
				WfManager: &wfengine.WfManager{
					ResourceType: "dbcluster",
				},
			},
			RunningSteps: []*wfengine.StepRuntime{
				{
					StepName: "Step1",
				},
				{
					StepName: "Step2",
				},
			},
		}

		stepRuntime := &wfengine.StepRuntime{
			StepName:      "Step1",
			StepStatus:    wfengine.StepStatusCompleted,
			StepStartTime: "2021-08-02 10:04:05",
		}

		recorder := implement_wfengine.CreateDefaultHistoryRecorder("dbcluster", klogr.New(), "")
		So(recorder, ShouldNotBeNil)
		recorder.WriteWorkflowStepLog(runtime, stepRuntime, "")
	})

	Convey("json.Unmarshal failed", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		errorMsg := "json.Unmarshal failed"
		guard := monkey.Patch(j.Unmarshal, func(data []byte, v interface{}) error {
			return errors.New(errorMsg)
		})
		defer guard.Unpatch()

		logger := klogr.New()
		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("test-resource")

		runtime := &wfengine.WfRuntime{
			FlowStatus: wfengine.FlowStatusRunning,
			FlowName:   "WF1",
			Logger:     logger,
			StartTime:  "2021-08-02 10:03:05",
			ResourceWorkflow: &wfengine.ResourceWorkflow{
				Logger:   logger,
				Resource: resource,
				WfManager: &wfengine.WfManager{
					ResourceType: "dbcluster",
				},
			},
			RunningSteps: []*wfengine.StepRuntime{
				{
					StepName: "Step1",
				},
				{
					StepName: "Step2",
				},
			},
		}

		stepRuntime := &wfengine.StepRuntime{
			StepName:      "Step1",
			StepStatus:    wfengine.StepStatusCompleted,
			StepStartTime: "2021-08-02 10:04:05",
		}

		recorder := implement_wfengine.CreateDefaultHistoryRecorder("dbcluster", klogr.New(), "")
		So(recorder, ShouldNotBeNil)
		recorder.WriteWorkflowStepLog(runtime, stepRuntime, "")
	})
}
