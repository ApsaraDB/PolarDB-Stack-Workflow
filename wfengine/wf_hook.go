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

import "github.com/ApsaraDB/PolarDB-Stack-Workflow/define"

type WfHook interface {
	// 流程初始化钩子
	OnWfInit() error

	// 流程结束时执行
	OnWfCompleted() error

	// 流程中断时执行
	OnWfInterrupt(*define.InterruptError) error

	// 流程步骤执行前运行
	OnStepInit(step *StepRuntime) error

	// 步骤等待状态时执行
	OnStepWaiting(step *StepRuntime) error

	// 流程步骤执行后运行
	OnStepCompleted(step *StepRuntime) error
}
