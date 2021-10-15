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

import "fmt"

// 稳定状态，一直在等待event触发自身状态发生变化，除此以外，不做任何动作
type StableStateMainEnter struct {
	sm       *StateMachine
	nowState State
}

func (ss *StableStateMainEnter) GetName() string {
	return "StableStateMainEnter"
}

func (ss *StableStateMainEnter) MainEnter(obj StateResource) error {
	newState, err := ss.sm.CheckSetStateTranslate(ss.nowState, obj)
	if err != nil {
		return err
	}
	// 如果发生状态变更, 则进入变更后的状态的 mainEnter .
	if newState != StateNil {
		// 变更后的状态等于变更前的状态，为了防止重复进入当前同样状态，打印日志 check !
		if newState == ss.nowState {
			ss.sm.Logger.Info(fmt.Sprintf("repeat into state: %v, please check.", ss.nowState))
		}
		// 发现变更到了新的状态，不在这里进入新的状态的main. 因为已经修改了status, 会触发新的event.
	} else {
		ss.sm.Logger.Info("not find any state, so exit!!")
	}
	return nil
}
