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
	"testing"

	"gitlab.alibaba-inc.com/polar-as/polar-wf-engine/define"

	"gitlab.alibaba-inc.com/polar-as/polar-wf-engine/statemachine"

	"gitlab.alibaba-inc.com/polar-as/polar-wf-engine/implement"
	implement_wfengine "gitlab.alibaba-inc.com/polar-as/polar-wf-engine/implement/wfengine"
	"gitlab.alibaba-inc.com/polar-as/polar-wf-engine/test"
	"gitlab.alibaba-inc.com/polar-as/polar-wf-engine/test/dbclusterwf"
	"gitlab.alibaba-inc.com/polar-as/polar-wf-engine/utils/k8sutil"
	"k8s.io/klog/klogr"

	. "github.com/smartystreets/goconvey/convey"
)

func GetDbClusterResource(name, namespace string) (*dbclusterwf.DbResource, error) {
	dbCluster, err := k8sutil.ConfigMapFactoryByName(name, namespace)
	if err != nil {
		return nil, err
	}
	resource := &dbclusterwf.DbResource{KubeResource: implement.KubeResource{Resource: dbCluster}, Logger: klogr.New()}
	return resource, nil
}

func Test_GetResourceRecoverInfo(t *testing.T) {
	test.PatchClient(t, func(container *test.FakeContainer) {
		Convey("recover from interrupt to creating", t, func() {
			recover := implement_wfengine.CreateDefaultRecover()
			resource, err := GetDbClusterResource("db_cluster_sample_1", "default")
			So(err, ShouldBeNil)
			r, prevState := recover.GetResourceRecoverInfo(resource, define.DefaultWfConf)
			So(r, ShouldBeTrue)
			So(prevState, ShouldEqual, statemachine.StateCreating)
		})

		Convey("non interrupt state dose not need to be recovered", t, func() {
			recover := implement_wfengine.CreateDefaultRecover()
			resource, err := GetDbClusterResource("db_cluster_sample", "default")
			So(err, ShouldBeNil)
			r, prevState := recover.GetResourceRecoverInfo(resource, define.DefaultWfConf)
			So(r, ShouldBeFalse)
			So(prevState, ShouldBeEmpty)
		})

		Convey("interrupt state dose not need to be recovered", t, func() {
			recover := implement_wfengine.CreateDefaultRecover()
			resource, err := GetDbClusterResource("db_cluster_sample_2", "default")
			So(err, ShouldBeNil)
			r, prevState := recover.GetResourceRecoverInfo(resource, define.DefaultWfConf)
			So(r, ShouldBeFalse)
			So(prevState, ShouldBeEmpty)
		})

		Convey("interrupt state without previous state info dose not need to be recovered", t, func() {
			recover := implement_wfengine.CreateDefaultRecover()
			resource, err := GetDbClusterResource("db_cluster_sample_3", "default")
			So(err, ShouldBeNil)
			r, prevState := recover.GetResourceRecoverInfo(resource, define.DefaultWfConf)
			So(r, ShouldBeFalse)
			So(prevState, ShouldBeEmpty)
		})
	})
}

func Test_CancelResourceRecover(t *testing.T) {
	test.PatchClient(t, func(container *test.FakeContainer) {
		Convey("remove recover info from resource", t, func() {
			recover := implement_wfengine.CreateDefaultRecover()
			resource, err := GetDbClusterResource("db_cluster_sample_2", "default")
			So(err, ShouldBeNil)
			annotations := resource.Resource.GetAnnotations()
			_, ok := annotations[define.DefaultWfConf[define.WFInterruptToRecover]]
			So(ok, ShouldBeTrue)
			_, ok = annotations[define.DefaultWfConf[define.WFInterruptPrevious]]
			So(ok, ShouldBeTrue)
			err = recover.CancelResourceRecover(resource, define.DefaultWfConf)
			So(err, ShouldBeNil)
			res, err := GetDbClusterResource("db_cluster_sample_2", "default")
			So(err, ShouldBeNil)
			annotations = res.Resource.GetAnnotations()
			_, ok = annotations[define.DefaultWfConf[define.WFInterruptToRecover]]
			So(ok, ShouldBeFalse)
			_, ok = annotations[define.DefaultWfConf[define.WFInterruptPrevious]]
			So(ok, ShouldBeFalse)
		})

		Convey("no recover info in resource", t, func() {
			recover := implement_wfengine.CreateDefaultRecover()
			resource, err := GetDbClusterResource("db_cluster_sample", "default")
			err = recover.CancelResourceRecover(resource, define.DefaultWfConf)
			So(err, ShouldBeNil)
		})
	})
}

func Test_SaveResourceInterruptInfo(t *testing.T) {
	test.PatchClient(t, func(container *test.FakeContainer) {
		Convey("save recover info", t, func() {
			recover := implement_wfengine.CreateDefaultRecover()
			resource, err := GetDbClusterResource("db_cluster_sample", "default")
			So(err, ShouldBeNil)

			reason := "interrupt reason"
			message := "interrupt message"
			prev := "Creating"
			err = recover.SaveResourceInterruptInfo(resource, define.DefaultWfConf, reason, message, prev)
			So(err, ShouldBeNil)

			res, err := GetDbClusterResource("db_cluster_sample", "default")
			So(err, ShouldBeNil)
			annotations := res.Resource.GetAnnotations()
			r, ok := annotations[define.DefaultWfConf[define.WFInterruptToRecover]]
			So(ok, ShouldBeTrue)
			So(r, ShouldEqual, "F")
			p, ok := annotations[define.DefaultWfConf[define.WFInterruptPrevious]]
			So(ok, ShouldBeTrue)
			So(p, ShouldEqual, prev)
			re, ok := annotations[define.DefaultWfConf[define.WFInterruptReason]]
			So(ok, ShouldBeTrue)
			So(re, ShouldEqual, reason)
			m, ok := annotations[define.DefaultWfConf[define.WFInterruptMessage]]
			So(ok, ShouldBeTrue)
			So(m, ShouldEqual, message)
		})
	})
}
