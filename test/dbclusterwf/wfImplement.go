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


package dbclusterwf

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-logr/logr"
	"gitlab.alibaba-inc.com/polar-as/polar-wf-engine/define"
	wfengineimpl "gitlab.alibaba-inc.com/polar-as/polar-wf-engine/implement/wfengine"
	"gitlab.alibaba-inc.com/polar-as/polar-wf-engine/statemachine"
	"gitlab.alibaba-inc.com/polar-as/polar-wf-engine/utils/k8sutil"
	"gitlab.alibaba-inc.com/polar-as/polar-wf-engine/wfengine"
	corev1 "k8s.io/api/core/v1"
)

var (
	dbClusterWfOnce    sync.Once
	dbClusterWfManager *wfengine.WfManager
)

func GetDbClusterWfManager() *wfengine.WfManager {
	var resourceType = "dbcluster"
	var workFlowMetaDir = "test/dbclusterwf/workflow"
	dbClusterWfOnce.Do(func() {
		if dbClusterWfManager == nil {
			var err error
			dbClusterWfManager, err = createWfManager(resourceType, workFlowMetaDir)
			dbClusterWfManager.RegisterRecover(wfengineimpl.CreateDefaultRecover())
			if err != nil {
				panic(fmt.Sprintf("create %s wf manager failed: %v", resourceType, err))
			}
			registerSteps(dbClusterWfManager)
		}
	})
	return dbClusterWfManager
}

func createWfManager(resourceType, workFlowMetaDir string) (wfManager *wfengine.WfManager, err error) {
	wfManager, err = wfengine.CreateWfManager(
		resourceType,
		workFlowMetaDir,
		wfengineimpl.CreateDefaultWfMetaLoader,
		wfengineimpl.CreateDefaultWorkflowHook,
		wfengineimpl.GetDefaultMementoStorageFactory(resourceType, false),
	)
	return
}

func CheckInstall(obj statemachine.StateResource) (*statemachine.Event, error) {
	dbCluster := obj.(*DbResource).GetDbCluster()
	clusterStatus := dbCluster.Data["clusterStatus"]
	if clusterStatus == "Init" || clusterStatus == "" || string(clusterStatus) == string(statemachine.StateCreating) {
		return statemachine.CreateEvent("Create", nil), nil
	}
	return nil, nil
}

func registerSteps(wfManager *wfengine.WfManager) {
	// 注册工作流步骤
	wfManager.RegisterStep(&CreateStep1{})
	wfManager.RegisterStep(&CreateStep2{})
}

type DbClusterStepBase struct {
	wfengine.StepAction
	Resource *corev1.ConfigMap
}

func (s *DbClusterStepBase) Init(ctx map[string]interface{}, logger logr.Logger) error {
	name := ctx[define.DefaultWfConf[define.WorkFlowResourceName]].(string)
	ns := ctx[define.DefaultWfConf[define.WorkFlowResourceNameSpace]].(string)
	res, err := k8sutil.ConfigMapFactoryByName(name, ns)
	if err != nil {
		return err
	}
	s.Resource = res
	return nil
}

func (s *DbClusterStepBase) DoStep(ctx context.Context, logger logr.Logger) error {
	panic("implement me")
}

func (s *DbClusterStepBase) Output(logger logr.Logger) map[string]interface{} {
	return map[string]interface{}{}
}

type CreateStep1 struct {
	DbClusterStepBase
}

func (step *CreateStep1) DoStep(ctx context.Context, logger logr.Logger) error {
	logger.Info("CreateStep1 resource info", "resource", step.Resource)
	return nil
}

type CreateStep2 struct {
	DbClusterStepBase
}

func (step *CreateStep2) DoStep(ctx context.Context, logger logr.Logger) error {
	logger.Info("CreateStep2 resource info", "resource", step.Resource)
	return nil
}
