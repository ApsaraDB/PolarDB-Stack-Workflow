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
	"encoding/json"
	"flag"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/sirupsen/logrus"
	"gitlab.alibaba-inc.com/polar-as/polar-wf-engine/wfengine"
)

type StepType string

type WorkflowStep struct {
	ResourceName string            `json:"resource_name"`
	Namespace    string            `json:"namespace"`
	WorkflowID   string            `json:"workflow_id"`
	StepID       string            `json:"step_id"`
	FlowName     string            `json:"flow_name"`
	StepName     string            `json:"step_name"`
	Type         StepType          `json:"type"`
	StartTime    string            `json:"start_time"`
	EndTime      string            `json:"end_time"`
	Step         int               `json:"step"`
	StepCount    int               `json:"step_count"`
	Retry        int               `json:"retry"`
	Error        string            `json:"error"`
	ExtraCtx     map[string]string `json:"extra_ctx"`
	resourceType string
}

const (
	WorkFlowStart     StepType = "WorkFlowStart"
	WorkFlowInterrupt StepType = "WorkFlowInterrupt"
	WorkFlowCompleted StepType = "WorkFlowCompleted"
	StepInited        StepType = "StepInited"
	StepWaiting       StepType = "StepWaiting"
	StepError         StepType = "StepError"
	StepCompleted     StepType = "StepCompleted"
)

var (
	wfLogger *logrus.Logger
	wfOnce   sync.Once
)

type DefaultHistoryRecorder struct {
	resourceType string
	logDir       string
	logger       logr.Logger
}

func CreateDefaultHistoryRecorder(resourceType string, logger logr.Logger, logDir string) *DefaultHistoryRecorder {
	if logDir == "" {
		flag := flag.CommandLine.Lookup("log_dir")
		if flag != nil {
			logDir = flag.Value.String()
		}
	}
	return &DefaultHistoryRecorder{
		resourceType: resourceType,
		logger:       logger.WithName("history"),
		logDir:       logDir,
	}
}

func (r *DefaultHistoryRecorder) WriteWorkflowStepLog(flow *wfengine.WfRuntime, step *wfengine.StepRuntime, stepType StepType) {
	wfStep := r.getWorkflowStepLogObj(flow, step, stepType)
	if wfStep == nil {
		return
	}
	r.logWorkflowStep(wfStep)
}

func (r *DefaultHistoryRecorder) getWorkflowStepLogObj(flow *wfengine.WfRuntime, step *wfengine.StepRuntime, stepType StepType) *WorkflowStep {
	startTime, err := time.Parse("2006-01-02 15:04:05", flow.StartTime)
	if err != nil {
		r.logger.Error(err, "failed to parse workflow startTime!")
		return nil
	}
	workflowStep := &WorkflowStep{
		Namespace:    flow.ResourceWorkflow.Resource.GetNamespace(),
		ResourceName: flow.ResourceWorkflow.Resource.GetName(),
		StepCount:    len(flow.RunningSteps),
		WorkflowID:   strconv.FormatInt(startTime.UnixNano()/int64(time.Millisecond), 10),
		StepID:       strconv.FormatInt(time.Now().UnixNano(), 10),
		FlowName:     flow.FlowName,
		EndTime:      time.Now().Format("2006-01-02 15:04:05"),
		Retry:        flow.RetryTimes,
	}
	if step != nil {
		for i, s := range flow.RunningSteps {
			if s.StepName == step.StepName {
				workflowStep.Step = i + 1
				break
			}
		}
	}

	switch {
	// if step is nil => workflow is complete || start || interrupt
	case step == nil:
		workflowStep.Type = stepType
		if stepType == "" {
			workflowStep.Type = WorkFlowCompleted
			workflowStep.StartTime = workflowStep.EndTime
		} else if stepType == WorkFlowStart {
			workflowStep.StartTime = flow.StartTime
			workflowStep.Retry = 0
		} else if stepType == WorkFlowInterrupt {
			workflowStep.StartTime = workflowStep.EndTime
			workflowStep.Error = flow.ErrorMessage
		}
	case step.StepStatus == wfengine.StepStatusInited:
		workflowStep.StepName = step.StepName
		workflowStep.Type = StepInited
		workflowStep.StartTime = step.StepStartTime
		workflowStep.Error = ""
	case step.StepStatus == wfengine.StepStatusWaiting:
		workflowStep.StepName = step.StepName
		workflowStep.Type = StepWaiting
		workflowStep.StartTime = step.StepStartTime
		workflowStep.Error = ""
	case step.StepStatus == wfengine.StepStatusFailed:
		workflowStep.StepName = step.StepName
		workflowStep.Type = StepError
		workflowStep.StartTime = step.StepStartTime
		workflowStep.Error = flow.ErrorMessage
	case step.StepStatus == wfengine.StepStatusCompleted:
		workflowStep.StepName = step.StepName
		workflowStep.Type = StepCompleted
		workflowStep.StartTime = step.StepStartTime
	default:
		return nil
	}

	return workflowStep
}

func (r *DefaultHistoryRecorder) logWorkflowStep(stepInfo interface{}) {
	var fields map[string]interface{}
	stepInfoJson, err := json.Marshal(stepInfo)
	if err != nil {
		r.logger.Error(err, "failed to convert json!")
		return
	}
	err = json.Unmarshal(stepInfoJson, &fields)
	if err != nil {
		r.logger.Error(err, "failed to convert json to map!", "stepInfoJson", stepInfoJson)
		return
	}
	convertDateTimeFormat(fields, "start_time", "end_time")

	wfOnce.Do(func() {
		logFile, err := os.OpenFile(filepath.Join(r.logDir, r.resourceType+"_workflow.txt"), os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			r.logger.Error(err, "workflow log file init failed!")
			return
		}
		// 重置日志文件的光标.
		_, _ = logFile.Seek(0, io.SeekEnd)
		wfLogger = &logrus.Logger{
			Out:       logFile,
			Formatter: new(logrus.JSONFormatter),
			Level:     logrus.InfoLevel,
		}
	})
	if wfLogger != nil {
		wfLogger.WithFields(fields).Info()
	} else {
		r.logger.Info("", "stepInfo", stepInfo)
	}
}

func convertDateTimeFormat(obj map[string]interface{}, fields ...string) {
	for _, field := range fields {
		if value, ok := obj[field]; ok {
			if v, ok := value.(string); ok {
				timeObj, _ := time.Parse("2006-01-02 15:04:05", v)
				obj[field] = timeObj.Format(time.RFC3339)
			}
		}
	}
}
