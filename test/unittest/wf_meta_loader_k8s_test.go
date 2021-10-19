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
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/bouk/monkey"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
	implementwfengine "github.com/ApsaraDB/PolarDB-Stack-Workflow/implement/wfengine"
	"gopkg.in/yaml.v2"
)

func Test_GetAllFlowMeta(t *testing.T) {

	Convey("recover from interrupt to creating", t, func() {
		loader, err := implementwfengine.CreateDefaultWfMetaLoader("dbcluster")
		So(err, ShouldBeNil)
		flowMeta, flowGroupMeta, err := loader.GetAllFlowMeta("test/dbclusterwf/workflow")
		So(flowMeta, ShouldNotBeNil)
		So(flowGroupMeta, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})

	Convey("meta dir not provided", t, func() {
		loader, err := implementwfengine.CreateDefaultWfMetaLoader("")
		So(err, ShouldBeNil)
		flowMeta, flowGroupMeta, err := loader.GetAllFlowMeta("")
		So(flowMeta, ShouldBeNil)
		So(flowGroupMeta, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "file path is required")
	})

	Convey("meta dir is invalid", t, func() {
		loader, err := implementwfengine.CreateDefaultWfMetaLoader("dbcluster")
		So(err, ShouldBeNil)

		guard := monkey.Patch(filepath.Abs, func(path string) (string, error) {
			return "", errors.New("dirInvalid")
		})
		defer guard.Unpatch()

		flowMeta, flowGroupMeta, err := loader.GetAllFlowMeta("test/dbclusterwf/workflow")
		So(flowMeta, ShouldBeNil)
		So(flowGroupMeta, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "dirInvalid")
	})

	Convey("meta dir is not exist", t, func() {
		loader, err := implementwfengine.CreateDefaultWfMetaLoader("dbcluster")
		So(err, ShouldBeNil)
		flowMeta, flowGroupMeta, err := loader.GetAllFlowMeta("not_exist_dir")
		So(flowMeta, ShouldBeNil)
		So(flowGroupMeta, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "read workflow meta dir failed")
	})

	Convey("read workflow meta file failed", t, func() {
		loader, err := implementwfengine.CreateDefaultWfMetaLoader("dbcluster")
		So(err, ShouldBeNil)

		guard := monkey.Patch(ioutil.ReadFile, func(filename string) ([]byte, error) {
			return nil, errors.New("fileInvalid")
		})
		defer guard.Unpatch()

		flowMeta, flowGroupMeta, err := loader.GetAllFlowMeta("test/dbclusterwf/workflow")
		So(flowMeta, ShouldBeNil)
		So(flowGroupMeta, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "fileInvalid")
	})

	Convey("workflow meta is an invalid yaml", t, func() {
		guard := monkey.Patch(yaml.Unmarshal, func(in []byte, out interface{}) (err error) {
			if !strings.Contains(reflect.TypeOf(out).String(), "Group") {
				return errors.New("invalid yaml")
			}
			return nil
		})
		defer guard.Unpatch()

		loader, err := implementwfengine.CreateDefaultWfMetaLoader("dbcluster")
		So(err, ShouldBeNil)
		flowMeta, flowGroupMeta, err := loader.GetAllFlowMeta("test/dbclusterwf/workflow")
		So(flowMeta, ShouldBeNil)
		So(flowGroupMeta, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "unmarshal yaml file fail")
	})

	Convey("group step meta is an invalid yaml", t, func() {
		guard := monkey.Patch(yaml.Unmarshal, func(in []byte, out interface{}) (err error) {
			if strings.Contains(reflect.TypeOf(out).String(), "Group") {
				return errors.New("invalid yaml")
			}
			return nil
		})
		defer guard.Unpatch()

		loader, err := implementwfengine.CreateDefaultWfMetaLoader("dbcluster")
		So(err, ShouldBeNil)
		flowMeta, flowGroupMeta, err := loader.GetAllFlowMeta("test/dbclusterwf/workflow")
		So(flowMeta, ShouldBeNil)
		So(flowGroupMeta, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "unmarshal yaml file fail")
	})
}
