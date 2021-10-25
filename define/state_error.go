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

package define

import (
	"fmt"

	"github.com/pkg/errors"
)

type InterruptError struct {
	Reason   string
	ErrorMsg string
}

func (e *InterruptError) Error() string {
	return fmt.Sprintf("reson: %s, message: %s", e.Reason, e.ErrorMsg)
}

func NewInterruptError(reason, msg string) error {
	err := InterruptError{
		Reason:   reason,
		ErrorMsg: msg,
	}
	var rErr error
	rErr = &err
	return rErr
}

func IsInterruptError(err error) (bool, *InterruptError) {
	pErr, ok := errors.Cause(err).(*InterruptError)
	if ok {
		return true, pErr
	}
	return false, nil
}

type WaitingError struct {
}

func (e *WaitingError) Error() string {
	return fmt.Sprint("waiting for resources to be ready")
}

func NewWaitingError() error {
	return &WaitingError{}
}

func IsWaitingError(err error) bool {
	_, ok := errors.Cause(err).(*WaitingError)
	return ok
}
