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
	"fmt"

	"github.com/pkg/errors"

	"k8s.io/client-go/kubernetes"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func GetKubeClient() (*kubernetes.Clientset, *rest.Config, error) {
	var cfg *rest.Config
	var err error
	cfg, err = config.GetConfig()
	if err != nil {
		klog.Error(err, "unable to set up client config")
		return nil, nil, fmt.Errorf("could not get kubernetes config from kubeconfig: %v", err)
	}
	clientSet, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, nil, errors.Wrap(err, "could not get kubernetes client: %s")
	}
	return clientSet, cfg, nil
}

func GetCorev1Client() (corev1client.CoreV1Interface, error) {
	_, config, err := GetKubeClient()
	if err != nil {
		return nil, errors.Wrap(err, "GetCorev1Client error")
	}

	coreV1Client, err := corev1client.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "corev1client.NewForConfig error")
	}

	return coreV1Client, nil
}
