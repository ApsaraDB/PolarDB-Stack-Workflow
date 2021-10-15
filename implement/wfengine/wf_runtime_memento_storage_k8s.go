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
	"github.com/pkg/errors"
	"gitlab.alibaba-inc.com/polar-as/polar-wf-engine/implement"
	"gitlab.alibaba-inc.com/polar-as/polar-wf-engine/statemachine"
	"gitlab.alibaba-inc.com/polar-as/polar-wf-engine/utils/k8sutil"
	"gitlab.alibaba-inc.com/polar-as/polar-wf-engine/wfengine"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const WorkFlowJobConfigMapTag = "flowJobsDomain"
const EventsReasonFormat = "WF_%s"

type DefaultMementoStorage struct {
	careTakerName string
	resource      implement.IKubeResource
	resourceType  string
	enableEvent   bool
}

func GetDefaultMementoStorageFactory(
	resourceType string,
	enableEvent bool,
) wfengine.CreateMementoStorageFunc {
	return func(resource statemachine.StateResource) (wfengine.WfRuntimeMementoStorage, error) {
		return &DefaultMementoStorage{
			resource:     resource.(implement.IKubeResource),
			resourceType: resourceType,
			enableEvent:  enableEvent,
		}, nil
	}
}

// 工作流运行过程中保存运行时信息
func (s *DefaultMementoStorage) Save(mementoKey, mementoContent string) error {
	cm, err := s.getConfigMap()
	if err != nil {
		if apierrors.IsNotFound(err) {
			ncm := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      s.careTakerName,
					Namespace: s.resource.GetResource().GetNamespace(),
					Labels: map[string]string{
						WorkFlowJobConfigMapTag: s.resource.GetResource().GetName(),
					},
				},
				Data: map[string]string{
					mementoKey: mementoContent,
				},
			}
			if err := controllerutil.SetControllerReference(s.resource.GetResource(), ncm, s.resource.GetScheme()); err != nil {
				err := errors.Wrap(err, "set workflow runtime meta reference fail!")
				return err
			}
			return k8sutil.ConfigMapCreateFactory(ncm)
		} else {
			return err
		}
	}
	cm.Data[mementoKey] = mementoContent
	return k8sutil.ConfigMapUpdate(cm)
}

func (s *DefaultMementoStorage) LoadMementoMap(careTakerName string) (map[string]string, error) {
	s.careTakerName = careTakerName
	cm, err := s.getConfigMap()
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil
		} else {
			return nil, err
		}
	}
	return cm.Data, nil
}

func (s *DefaultMementoStorage) getConfigMap() (*corev1.ConfigMap, error) {
	return k8sutil.ConfigMapFactoryByName(s.careTakerName, s.resource.GetResource().GetNamespace())
}
