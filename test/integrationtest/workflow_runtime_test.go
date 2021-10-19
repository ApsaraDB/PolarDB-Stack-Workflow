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


package integrationtest

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/ApsaraDB/PolarDB-Stack-Workflow/implement"
	"github.com/ApsaraDB/PolarDB-Stack-Workflow/statemachine"
	"github.com/ApsaraDB/PolarDB-Stack-Workflow/test"
	"github.com/ApsaraDB/PolarDB-Stack-Workflow/test/dbclusterwf"
	"github.com/ApsaraDB/PolarDB-Stack-Workflow/utils/k8sutil"
	"github.com/ApsaraDB/PolarDB-Stack-Workflow/wfengine"
	"k8s.io/klog/klogr"
)

func Test_CommonWorkFlowMainEnter(t *testing.T) {
	test.PatchClient(t, func(container *test.FakeContainer) {
		Convey("run succeed", t, func() {
			err := StartWorkFlow("db_cluster_sample", "default", "WF1", false, dbclusterwf.CheckInstall)
			So(err, ShouldBeNil)
		})

		Convey("workflow not found", t, func() {
			err := StartWorkFlow("db_cluster_sample", "default", "NotFoundWorkflow", false, dbclusterwf.CheckInstall)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "not found")
		})

		Convey("step not found", t, func() {
			err := StartWorkFlow("db_cluster_sample", "default", "WF2", false, dbclusterwf.CheckInstall)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "UNKNOWN_TASK")
		})
	})
}

func StartWorkFlow(name, namespace, flowName string, ignoreUnCompleted bool, eventChecker statemachine.EventChecker) error {
	resource, wf, err := CreateResourceWorkflow(name, namespace)
	if err != nil {
		return err
	}
	return wf.CommonWorkFlowMainEnter(context.TODO(), resource, flowName, ignoreUnCompleted, eventChecker)
}

func CreateResourceWorkflow(name, namespace string) (*dbclusterwf.DbResource, *wfengine.ResourceWorkflow, error) {
	resource, err := GetDbClusterResource(name, namespace)
	if err != nil {
		return nil, nil, err
	}
	wf, err := dbclusterwf.GetDbClusterWfManager().CreateResourceWorkflow(resource)
	if err != nil {
		return nil, nil, err
	}
	return resource, wf, nil
}

func GetDbClusterResource(name, namespace string) (*dbclusterwf.DbResource, error) {
	dbCluster, err := k8sutil.ConfigMapFactoryByName(name, namespace)
	if err != nil {
		return nil, err
	}
	resource := &dbclusterwf.DbResource{KubeResource: implement.KubeResource{Resource: dbCluster}, Logger: klogr.New()}
	return resource, nil
}
