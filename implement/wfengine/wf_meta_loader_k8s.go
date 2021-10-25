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
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/ApsaraDB/PolarDB-Stack-Workflow/wfengine"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"k8s.io/klog/klogr"
)

type DefaultWfMetaLoader struct {
	resourceType string
	logger       logr.Logger
}

func (l *DefaultWfMetaLoader) GetAllFlowMeta(workFlowMetaDir string) (map[string]*wfengine.FlowMeta, map[string]*wfengine.StepGroupMeta, error) {
	var err error
	var flowMap = make(map[string]*wfengine.FlowMeta)
	var stepGroupMap = make(map[string]*wfengine.StepGroupMeta)
	metaFilePath := fmt.Sprintf("%s/%s", workFlowMetaDir, l.resourceType)
	if metaFilePath == "/" {
		err = errors.New("workflow meta loader file path is required")
		l.logger.Error(err, "")
		return nil, nil, err
	}

	dirPath, err := filepath.Abs(metaFilePath)
	if err != nil {
		err = errors.Wrap(err, "workflow meta file path is invalid")
		l.logger.Error(err, "")
		return nil, nil, err
	}
	dir, err := ioutil.ReadDir(dirPath)
	if err != nil {
		err = errors.Wrap(err, "read workflow meta dir failed")
		l.logger.Error(err, "")
		return nil, nil, err
	}
	for _, fi := range dir {
		if fi.IsDir() {
			continue
		}
		ok := strings.HasSuffix(fi.Name(), ".yaml")
		if ok {
			yamlContent, err := ioutil.ReadFile(fmt.Sprintf("%s/%v", metaFilePath, fi.Name()))
			if err != nil {
				err = errors.Wrap(err, "open yaml file fail")
				l.logger.Error(err, "")
				return nil, nil, err
			}
			if isStepGroup := strings.HasSuffix(fi.Name(), "_step_group.yaml"); isStepGroup {
				stepGroupMeta := &wfengine.StepGroupMeta{}
				if err = yaml.Unmarshal(yamlContent, stepGroupMeta); err != nil {
					err = errors.Wrap(err, "unmarshal yaml file fail")
					l.logger.Error(err, "")
					return nil, nil, err
				}
				stepGroupMap[stepGroupMeta.GroupName] = stepGroupMeta
			} else {
				meta := &wfengine.FlowMeta{}
				if err = yaml.Unmarshal(yamlContent, meta); err != nil {
					err = errors.Wrap(err, "unmarshal yaml file fail")
					l.logger.Error(err, "")
					return nil, nil, err
				}
				flowMap[meta.FlowName] = meta
			}
		}
	}
	return flowMap, stepGroupMap, nil
}

func CreateDefaultWfMetaLoader(resourceType string) (wfengine.WfMetaLoader, error) {
	return &DefaultWfMetaLoader{
		resourceType: resourceType,
		logger:       klogr.New().WithName("workflow/metaloader").WithName(resourceType),
	}, nil
}
