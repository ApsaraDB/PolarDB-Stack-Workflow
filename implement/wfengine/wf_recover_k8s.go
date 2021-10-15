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
	"gitlab.alibaba-inc.com/polar-as/polar-wf-engine/define"
	"gitlab.alibaba-inc.com/polar-as/polar-wf-engine/implement"
	"gitlab.alibaba-inc.com/polar-as/polar-wf-engine/statemachine"
)

type DefaultRecover struct {
}

func CreateDefaultRecover() *DefaultRecover {
	return &DefaultRecover{}
}

func (*DefaultRecover) GetResourceRecoverInfo(resource statemachine.StateResource, conf map[define.WFConfKey]string) (recover bool, preStatus statemachine.State) {
	kubeResource := resource.(implement.IKubeResource)
	if kubeResource.GetState() != statemachine.StateInterrupt {
		return false, statemachine.StateNil
	}
	recovery, ok := kubeResource.GetResource().GetAnnotations()[conf[define.WFInterruptToRecover]]
	if !ok || recovery == "F" {
		return false, statemachine.StateNil
	}
	prevStatus, ok := kubeResource.GetResource().GetAnnotations()[conf[define.WFInterruptPrevious]]
	if ok && prevStatus != "" {
		return true, statemachine.State(prevStatus)
	}
	return false, statemachine.StateNil
}

func (*DefaultRecover) CancelResourceRecover(resource statemachine.StateResource, conf map[define.WFConfKey]string) error {
	kubeResource := resource.(implement.IKubeResource)
	annos := kubeResource.GetResource().GetAnnotations()
	if annos == nil {
		return nil
	}
	delete(annos, conf[define.WFInterruptToRecover])
	delete(annos, conf[define.WFInterruptReason])
	delete(annos, conf[define.WFInterruptMessage])
	delete(annos, conf[define.WFInterruptPrevious])
	kubeResource.GetResource().SetAnnotations(annos)
	return kubeResource.Update()
}

func (*DefaultRecover) SaveResourceInterruptInfo(resource statemachine.StateResource, conf map[define.WFConfKey]string, reason, message, prevWf string) error {
	kubeResource := resource.(implement.IKubeResource)
	annos := kubeResource.GetResource().GetAnnotations()
	if annos == nil {
		annos = make(map[string]string)
	}
	annos[conf[define.WFInterruptToRecover]] = "F"
	annos[conf[define.WFInterruptReason]] = reason
	annos[conf[define.WFInterruptMessage]] = message
	annos[conf[define.WFInterruptPrevious]] = prevWf
	kubeResource.GetResource().SetAnnotations(annos)
	return kubeResource.Update()
}
