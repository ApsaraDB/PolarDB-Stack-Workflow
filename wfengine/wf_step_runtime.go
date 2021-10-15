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
	"time"
)

type StepStatusType string

const (
	StepStatusPrepared  StepStatusType = "prepared"
	StepStatusInited    StepStatusType = "inited"
	StepStatusFailed    StepStatusType = "failed"
	StepStatusCompleted StepStatusType = "completed"
	StepStatusWaiting   StepStatusType = "waiting"
)

type StepRuntime struct {
	StepName             string                 `json:"StepName,omitempty"`
	StepStatus           StepStatusType         `json:"stepStatus,omitempty"`
	StepStartTime        string                 `json:"stepStartTime,omitempty"`
	LastStepCompleteTime string                 `json:"lastStepCompleteTime,omitempty"`
	ContextOutput        map[string]interface{} `json:"contextOutput,omitempty"` // 会传递给下一个step的context.
	RetryTimes           int                    `json:"retryTimes,omitempty"`
}

func (rs *StepRuntime) setFail() {
	rs.StepStatus = StepStatusFailed
	rs.LastStepCompleteTime = time.Now().Format("2006-01-02 15:04:05")
}

func (rs *StepRuntime) setInit() {
	if rs.StepStatus == StepStatusFailed || rs.StepStatus == StepStatusInited {
		rs.RetryTimes += 1
	} else if rs.StepStatus == StepStatusPrepared {
		rs.RetryTimes = 0
	}
	rs.StepStatus = StepStatusInited
	rs.StepStartTime = time.Now().Format("2006-01-02 15:04:05")
	rs.LastStepCompleteTime = time.Now().Format("2006-01-02 15:04:05")
}

func (rs *StepRuntime) setOutputContext(output map[string]interface{}) {
	rs.ContextOutput = output
}

func (rs *StepRuntime) setCompleted() {
	rs.StepStatus = StepStatusCompleted
	rs.LastStepCompleteTime = time.Now().Format("2006-01-02 15:04:05")
}

func (rs *StepRuntime) setWaiting() {
	rs.StepStatus = StepStatusWaiting
}
