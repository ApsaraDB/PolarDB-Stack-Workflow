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

package k8sutil

import (
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ConfigMapFactoryByName(name, namespace string) (*corev1.ConfigMap, error) {
	client, err := GetCorev1Client()
	if err != nil {
		return nil, err
	}
	cm, err := client.ConfigMaps(namespace).Get(name, v1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return cm, nil
}

func ConfigMapCreateFactory(configMap *corev1.ConfigMap) error {
	client, err := GetCorev1Client()
	if err != nil {
		return err
	}
	_, err = client.ConfigMaps(configMap.Namespace).Create(configMap)
	return err
}

func ConfigMapUpdate(configMap *corev1.ConfigMap) error {
	client, err := GetCorev1Client()
	if err != nil {
		return err
	}
	_, err = client.ConfigMaps(configMap.Namespace).Update(configMap)
	return err
}
