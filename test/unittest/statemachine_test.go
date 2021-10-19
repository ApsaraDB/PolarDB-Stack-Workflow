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

	"github.com/pkg/errors"

	"k8s.io/klog/klogr"

	. "github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/ApsaraDB/PolarDB-Stack-Workflow/statemachine"
	"github.com/ApsaraDB/PolarDB-Stack-Workflow/test/mock"
)

func Test_MainEnter(t *testing.T) {
	Convey("status is not register", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("testDbCluster")
		resource.EXPECT().Fetch().AnyTimes().Return(resource, nil)
		resource.EXPECT().GetState().AnyTimes().Return(statemachine.StateNil)

		sm := statemachine.CreateStateMachineInstance("dbCluster")
		sm.RegisterStableState(statemachine.StateRunning, statemachine.StateInterrupt, statemachine.StateInit)
		err := sm.DoStateMainEnter(resource)

		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "status is not registered")
	})

	Convey("execute stable state main enter succeed", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("testDbCluster")
		resource.EXPECT().Fetch().AnyTimes().Return(resource, nil)
		resource.EXPECT().GetState().AnyTimes().Return(statemachine.StateRunning)

		sm := statemachine.CreateStateMachineInstance("dbCluster")
		sm.RegisterStableState(statemachine.StateRunning, statemachine.StateInterrupt, statemachine.StateInit)
		err := sm.DoStateMainEnter(resource)

		So(err, ShouldBeNil)
	})

	Convey("execute creating state main enter succeed", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("testDbCluster")
		resource.EXPECT().Fetch().AnyTimes().Return(resource, nil)
		resource.EXPECT().GetState().AnyTimes().Return(statemachine.StateCreating)

		sm := statemachine.CreateStateMachineInstance("dbCluster")
		sm.RegisterStateMainEnter(statemachine.StateCreating, func(obj statemachine.StateResource) error {
			return nil
		})

		err := sm.DoStateMainEnter(resource)

		So(err, ShouldBeNil)
	})

	Convey("from init state to creating state, execute creating main enter succeed", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("testDbCluster")
		resource.EXPECT().Fetch().AnyTimes().Return(resource, nil)
		resource.EXPECT().GetState().AnyTimes().Return(statemachine.StateInit)
		resource.EXPECT().UpdateState(statemachine.StateCreating).Return(resource, nil)

		sm := statemachine.CreateStateMachineInstance("dbCluster")
		sm.RegisterStableState(statemachine.StateRunning, statemachine.StateInterrupt, statemachine.StateInit)

		sm.RegisterStateTranslateMainEnter(
			statemachine.StateInit,
			func(resource statemachine.StateResource) (*statemachine.Event, error) {
				return statemachine.CreateEvent("creating", nil), nil
			},
			statemachine.StateCreating,
			func(obj statemachine.StateResource) error {
				return nil
			},
		)

		err := sm.DoStateMainEnter(resource)

		So(err, ShouldBeNil)
	})

	Convey("ignores the execution of the specified event", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("testDbCluster")
		resource.EXPECT().Fetch().AnyTimes().Return(resource, nil)
		resource.EXPECT().GetState().AnyTimes().Return(statemachine.StateRunning)

		sm := statemachine.CreateStateMachineInstance("dbCluster")
		sm.RegisterStableState(statemachine.StateRunning, statemachine.StateInterrupt, statemachine.StateInit)
		sm.RegisterStateMainEnter(statemachine.StateCreating, func(obj statemachine.StateResource) error {
			return nil
		})
		sm.RegisterIgnoreEvents([]string{"creating"})
		sm.RegisterStateTranslate(statemachine.StateRunning, func(resource statemachine.StateResource) (*statemachine.Event, error) {
			return statemachine.CreateEvent("creating", nil), nil
		}, statemachine.StateCreating)

		err := sm.DoStateMainEnter(resource)

		So(err, ShouldBeNil)
	})

	Convey("from interrupt state to rebuild state, execute rebuild main enter succeed", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("testDbCluster")
		resource.EXPECT().Fetch().AnyTimes().Return(resource, nil)
		resource.EXPECT().GetState().AnyTimes().Return(statemachine.StateInterrupt)
		resource.EXPECT().UpdateState(statemachine.StateRebuild).Times(1).Return(resource, nil)

		sm := statemachine.CreateStateMachineInstance("dbCluster")
		sm.RegisterStableState(statemachine.StateRunning, statemachine.StateInterrupt, statemachine.StateInit)
		sm.RegisterStateMainEnter(statemachine.StateRebuild, func(obj statemachine.StateResource) error {
			return nil
		})
		sm.RegisterLogger(klogr.New())
		sm.RegisterStateTranslate(statemachine.StateInterrupt, func(resource statemachine.StateResource) (*statemachine.Event, error) {
			return statemachine.CreateEvent("rebuild", nil), nil
		}, statemachine.StateRebuild)

		err := sm.DoStateMainEnter(resource)

		So(err, ShouldBeNil)
	})

	Convey("fetch resource info failed", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		errorMsg := "fetch resource info failed"
		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("testDbCluster")
		resource.EXPECT().Fetch().AnyTimes().Return(resource, errors.New(errorMsg))

		sm := statemachine.CreateStateMachineInstance("dbCluster")
		sm.RegisterStableState(statemachine.StateRunning, statemachine.StateInterrupt, statemachine.StateInit)

		err := sm.DoStateMainEnter(resource)

		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, errorMsg)
	})

	Convey("update resource state failed", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		errorMsg := "update state failed"
		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("testDbCluster")
		resource.EXPECT().Fetch().AnyTimes().Return(resource, nil)
		resource.EXPECT().GetState().AnyTimes().Return(statemachine.StateRunning)
		resource.EXPECT().UpdateState(statemachine.StateCreating).Return(nil, errors.New(errorMsg))

		sm := statemachine.CreateStateMachineInstance("dbCluster")
		sm.RegisterStableState(statemachine.StateRunning, statemachine.StateInterrupt, statemachine.StateInit)
		sm.RegisterStateMainEnter(statemachine.StateCreating, func(obj statemachine.StateResource) error {
			return nil
		})
		sm.RegisterStateTranslate(statemachine.StateRunning, func(resource statemachine.StateResource) (*statemachine.Event, error) {
			return statemachine.CreateEvent("creating", nil), nil
		}, statemachine.StateCreating)

		err := sm.DoStateMainEnter(resource)

		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, errorMsg)
	})

	Convey("check event failed", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("testDbCluster")
		resource.EXPECT().Fetch().AnyTimes().Return(resource, nil)
		resource.EXPECT().GetState().AnyTimes().Return(statemachine.StateRunning)

		sm := statemachine.CreateStateMachineInstance("dbCluster")
		sm.RegisterStableState(statemachine.StateRunning, statemachine.StateInterrupt, statemachine.StateInit)
		sm.RegisterStateMainEnter(statemachine.StateCreating, func(obj statemachine.StateResource) error {
			return nil
		})

		errorMsg := "check event failed"
		sm.RegisterStateTranslate(statemachine.StateRunning, func(resource statemachine.StateResource) (*statemachine.Event, error) {
			return nil, errors.New(errorMsg)
		}, statemachine.StateCreating)

		err := sm.DoStateMainEnter(resource)

		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, errorMsg)
	})

	Convey("no events occurred", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("testDbCluster")
		resource.EXPECT().Fetch().AnyTimes().Return(resource, nil)
		resource.EXPECT().GetState().AnyTimes().Return(statemachine.StateRunning)

		sm := statemachine.CreateStateMachineInstance("dbCluster")
		sm.RegisterStableState(statemachine.StateRunning, statemachine.StateInterrupt, statemachine.StateInit)
		sm.RegisterStateTranslate(statemachine.StateRunning, func(resource statemachine.StateResource) (*statemachine.Event, error) {
			return nil, nil
		}, statemachine.StateCreating)

		err := sm.DoStateMainEnter(resource)

		So(err, ShouldBeNil)
	})

	Convey("repeat into Running state", t, func() {
		ctrl := NewController(t)
		defer ctrl.Finish()

		resource := mock.NewMockStateResource(ctrl)
		resource.EXPECT().GetNamespace().AnyTimes().Return("default")
		resource.EXPECT().GetName().AnyTimes().Return("testDbCluster")
		resource.EXPECT().Fetch().AnyTimes().Return(resource, nil)
		resource.EXPECT().GetState().AnyTimes().Return(statemachine.StateRunning)
		resource.EXPECT().UpdateState(statemachine.StateRunning).Return(resource, nil)

		sm := statemachine.CreateStateMachineInstance("dbCluster")
		sm.RegisterStableState(statemachine.StateRunning, statemachine.StateInterrupt, statemachine.StateInit)
		sm.RegisterStateTranslate(statemachine.StateRunning, func(resource statemachine.StateResource) (*statemachine.Event, error) {
			return statemachine.CreateEvent("creating", nil), nil
		}, statemachine.StateRunning)

		err := sm.DoStateMainEnter(resource)

		So(err, ShouldBeNil)
	})
}
