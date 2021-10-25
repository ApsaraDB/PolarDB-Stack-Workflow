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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	implementwfengine "github.com/ApsaraDB/PolarDB-Stack-Workflow/implement/wfengine"
	"github.com/ApsaraDB/PolarDB-Stack-Workflow/test"
	"github.com/ApsaraDB/PolarDB-Stack-Workflow/utils/k8sutil"
	"github.com/bouk/monkey"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
	corev1 "k8s.io/api/core/v1"
)

func Test_LoadMementoMap(t *testing.T) {
	test.PatchClient(t, func(container *test.FakeContainer) {
		Convey("load resource workflow runtime memento", t, func() {
			resource, err := GetDbClusterResource("db_cluster_sample", "default")
			So(err, ShouldBeNil)

			mementoStorageFactory := implementwfengine.GetDefaultMementoStorageFactory("dbcluster", false)
			storage, _ := mementoStorageFactory(resource)
			memento, err := storage.LoadMementoMap("db_cluster_sample-workflow")
			So(memento, ShouldNotBeNil)
			So(err, ShouldBeNil)
		})

		Convey("workflow runtime memento is not found", t, func() {
			resource, err := GetDbClusterResource("db_cluster_sample", "default")
			So(err, ShouldBeNil)

			mementoStorageFactory := implementwfengine.GetDefaultMementoStorageFactory("dbcluster", false)
			storage, _ := mementoStorageFactory(resource)
			memento, err := storage.LoadMementoMap("test-resource-workflow")
			So(memento, ShouldBeNil)
			So(err, ShouldBeNil)
		})

		Convey("get workflow runtime memento configmap failed", t, func() {
			resource, err := GetDbClusterResource("db_cluster_sample", "default")
			So(err, ShouldBeNil)

			guard := monkey.Patch(k8sutil.ConfigMapFactoryByName, func(name, namespace string) (*corev1.ConfigMap, error) {
				return nil, errors.New("getConfigmapErr")
			})
			defer guard.Unpatch()

			mementoStorageFactory := implementwfengine.GetDefaultMementoStorageFactory("dbcluster", false)
			storage, _ := mementoStorageFactory(resource)
			memento, err := storage.LoadMementoMap("test-resource-workflow")
			So(memento, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})
	})
}

func Test_SaveMemento(t *testing.T) {
	test.PatchClient(t, func(container *test.FakeContainer) {
		Convey("save workflow runtime memento succeed", t, func() {
			resource, err := GetDbClusterResource("db_cluster_sample", "default")
			So(err, ShouldBeNil)

			mementoStorageFactory := implementwfengine.GetDefaultMementoStorageFactory("dbcluster", false)
			storage, _ := mementoStorageFactory(resource)
			memento, err := storage.LoadMementoMap("db_cluster_sample-workflow")
			err = storage.Save("2-WF1", "")
			So(err, ShouldBeNil)
			memento, err = storage.LoadMementoMap("db_cluster_sample-workflow")
			So(err, ShouldBeNil)
			So(memento, ShouldHaveLength, 2)
		})

		Convey("set controller reference failed", t, func() {
			resource, err := GetDbClusterResource("db_cluster_sample", "default")
			So(err, ShouldBeNil)

			guard := monkey.Patch(controllerutil.SetControllerReference, func(owner, controlled metav1.Object, scheme *runtime.Scheme) error {
				return errors.New("SetControllerReferenceErr")
			})
			defer guard.Unpatch()

			mementoStorageFactory := implementwfengine.GetDefaultMementoStorageFactory("dbcluster", false)
			storage, _ := mementoStorageFactory(resource)
			_, err = storage.LoadMementoMap("db_cluster_sample-1-workflow")
			err = storage.Save("1-WF1", "")
			So(err, ShouldNotBeNil)
		})

		Convey("create new configmap before save resource workflow runtime memento", t, func() {
			resource, err := GetDbClusterResource("db_cluster_sample", "default")
			So(err, ShouldBeNil)

			mementoStorageFactory := implementwfengine.GetDefaultMementoStorageFactory("dbcluster", false)
			storage, _ := mementoStorageFactory(resource)
			memento, err := storage.LoadMementoMap("db_cluster_sample-1-workflow")
			err = storage.Save("1-WF1", "")
			So(err, ShouldBeNil)
			memento, err = storage.LoadMementoMap("db_cluster_sample-1-workflow")
			So(err, ShouldBeNil)
			So(memento, ShouldHaveLength, 1)
		})

		Convey("get workflow runtime memento configmap failed", t, func() {
			resource, err := GetDbClusterResource("db_cluster_sample", "default")
			So(err, ShouldBeNil)

			guard := monkey.Patch(k8sutil.ConfigMapFactoryByName, func(name, namespace string) (*corev1.ConfigMap, error) {
				return nil, errors.New("getConfigmapErr")
			})
			defer guard.Unpatch()

			mementoStorageFactory := implementwfengine.GetDefaultMementoStorageFactory("dbcluster", false)
			storage, _ := mementoStorageFactory(resource)
			err = storage.Save("1-WF1", "")
			So(err, ShouldNotBeNil)
		})
	})
}
