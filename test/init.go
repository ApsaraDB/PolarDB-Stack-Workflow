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

package test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	r "runtime"
	"testing"

	"github.com/ApsaraDB/PolarDB-Stack-Workflow/test/mock"
	"github.com/ApsaraDB/PolarDB-Stack-Workflow/utils/k8sutil"
	. "github.com/bouk/monkey"
	. "github.com/golang/mock/gomock"
	apicorev1 "k8s.io/api/core/v1"
	k8sFake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"sigs.k8s.io/yaml"
)

var configmaps []*apicorev1.ConfigMap
var dbClusters []*apicorev1.ConfigMap // 不自定义CRD，使用configmap保存dbcluster元数据

// FakeContainer ...
type FakeContainer struct {
	FakeClientset               *k8sFake.Clientset
	CoreV1Client                corev1.CoreV1Interface
	MockStateResource           *mock.MockStateResource
	MockRecover                 *mock.MockRecover
	MockWfHook                  *mock.MockWfHook
	MockWfMetaLoader            *mock.MockWfMetaLoader
	MockWfRuntimeMementoStorage *mock.MockWfRuntimeMementoStorage
}

func init() {
	// chdir project dir to base
	_, filename, _, _ := r.Caller(0)
	rootDir := path.Join(path.Dir(filename), "..")
	err := os.Chdir(rootDir)
	if err != nil {
		panic(err)
	}
	testing.Init()
	configmaps = loadConfigMapResources("configmap")
	dbClusters = loadConfigMapResources("dbcluster")
	_ = scheme.AddToScheme(scheme.Scheme)
}

// InitFakeContainer ...
func InitFakeContainer(t *testing.T) (container *FakeContainer) {
	fakeClientset := k8sFake.NewSimpleClientset()
	container = &FakeContainer{
		FakeClientset: fakeClientset,
		CoreV1Client:  fakeClientset.CoreV1(),
	}

	for _, obj := range configmaps {
		_, err := container.FakeClientset.CoreV1().ConfigMaps(obj.Namespace).Create(obj)
		if err != nil {
			fmt.Println(err)
		}
	}

	for _, obj := range dbClusters {
		_, err := container.FakeClientset.CoreV1().ConfigMaps(obj.Namespace).Create(obj)
		if err != nil {
			fmt.Println(err)
		}
	}

	ctrl := NewController(t)
	defer ctrl.Finish()

	container.MockWfRuntimeMementoStorage = mock.NewMockWfRuntimeMementoStorage(ctrl)
	container.MockWfMetaLoader = mock.NewMockWfMetaLoader(ctrl)
	container.MockWfHook = mock.NewMockWfHook(ctrl)
	container.MockRecover = mock.NewMockRecover(ctrl)
	container.MockStateResource = mock.NewMockStateResource(ctrl)

	return
}

// PatchClient ...
func PatchClient(t *testing.T, action func(container *FakeContainer)) {
	container := InitFakeContainer(t)
	guardGetCoreV1 := Patch(k8sutil.GetCorev1Client, func() (corev1.CoreV1Interface, error) {
		return container.CoreV1Client, nil
	})
	defer guardGetCoreV1.Unpatch()

	action(container)
}

func loadConfigMapResources(subDir string) (objs []*apicorev1.ConfigMap) {
	items := loadTypedResources(subDir, func() interface{} { return &apicorev1.ConfigMap{} })
	for _, item := range items {
		objs = append(objs, item.(*apicorev1.ConfigMap))
	}
	return
}

func loadTypedResources(subDir string, createObject func() interface{}) (objs []interface{}) {
	dir := "test/data/k8s"
	if subDir != "" {
		dir = dir + "/" + subDir
	}
	result := loadFromDir(dir)
	for _, bytes := range result {
		obj := createObject()
		err := yaml.Unmarshal(bytes, obj)
		if err != nil {
			fmt.Println(err)
			return
		}
		objs = append(objs, obj)
	}
	return
}

func loadFromDir(dir string) (objs [][]byte) {
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			dirInfo, _ := ioutil.ReadDir(path)
			for _, fileInfo := range dirInfo {
				objs = append(objs, loadFromDir(path+"/"+fileInfo.Name())...)
			}
		}
		bytes, err := ioutil.ReadFile(path)
		if err != nil {
			fmt.Println(err)
			return err
		}

		objs = append(objs, bytes)

		return nil
	})
	return
}
