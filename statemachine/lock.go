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

import (
	"sync"
)

var ObjectLockIns = KubeObjectLock{}

type KubeObjectLock struct {
	// 假设持有sync.Mutex代价比较低，内存占用较少，暂时不考虑内存泄露问题.
	objectLock map[string]*sync.Mutex

	// 内部使用锁.
	mu sync.Mutex
}

func (dl *KubeObjectLock) GetLock(obj StateResource) *sync.Mutex {
	dl.mu.Lock()
	defer dl.mu.Unlock()
	name := obj.GetName()
	mu, ok := dl.objectLock[name]
	if ok {
		return mu
	}
	if dl.objectLock == nil {
		dl.objectLock = map[string]*sync.Mutex{}
	}
	dl.objectLock[name] = &sync.Mutex{}
	return dl.objectLock[name]
}
