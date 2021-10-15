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


package utils

func MapMerge(m1 map[string]interface{}, m2 map[string]interface{}) map[string]interface{} {
	nm := map[string]interface{}{}
	if m1 != nil {
		for k, v := range m1 {
			nm[k] = v
		}
	}
	if m2 != nil {
		for k, v := range m2 {
			nm[k] = v
		}
	}
	return nm
}

func MaxIntSlice(v []int) int {
	var m int
	for i, e := range v {
		if i == 0 || e > m {
			m = e
		}
	}
	return m
}

func ContainsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
