/*
 *
 * Copyright 2019 The Tekton Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package resources

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"

	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	tb "github.com/tektoncd/pipeline/test/builder"
)

var c = tb.Condition("conditionname", "foo")

var notStartedState = TaskConditionCheckState{{
	ConditionCheckName: "foo",
	Condition:          c,
}}

var runningState = TaskConditionCheckState{{
	ConditionCheckName: "foo",
	Condition:          c,
	ConditionCheck: &v1alpha1.ConditionCheck{
		ObjectMeta: metav1.ObjectMeta{
			Name: "running-condition-check",
		},
	},
}}

var successState = TaskConditionCheckState{{
	ConditionCheckName: "foo",
	Condition:          c,
	ConditionCheck: &v1alpha1.ConditionCheck{
		ObjectMeta: metav1.ObjectMeta{
			Name: "successful-condition-check",
		},
		Spec: v1alpha1.TaskRunSpec{},
		Status: v1alpha1.TaskRunStatus{
			Status: duckv1beta1.Status{
				Conditions: []apis.Condition{{
					Type:   apis.ConditionSucceeded,
					Status: corev1.ConditionTrue,
				}},
			},
		},
	},
}}

var failedState = TaskConditionCheckState{{
	ConditionCheckName: "foo",
	Condition:          c,
	ConditionCheck: &v1alpha1.ConditionCheck{
		ObjectMeta: metav1.ObjectMeta{
			Name: "failed-condition-check",
		},
		Spec: v1alpha1.TaskRunSpec{},
		Status: v1alpha1.TaskRunStatus{
			Status: duckv1beta1.Status{
				Conditions: []apis.Condition{{
					Type:   apis.ConditionSucceeded,
					Status: corev1.ConditionFalse,
				}},
			},
		},
	},
}}

func TestTaskConditionCheckState_HasStarted(t *testing.T) {
	tcs := []struct {
		name  string
		state TaskConditionCheckState
		want  bool
	}{{
		name:  "no-condition-checks",
		state: notStartedState,
		want:  false,
	}, {
		name:  "running-condition-check",
		state: runningState,
		want:  true,
	}, {
		name:  "successful-condition-check",
		state: successState,
		want:  true,
	}, {
		name:  "failed-condition-check",
		state: failedState,
		want:  true,
	}}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.state.HasStarted()
			if got != tc.want {
				t.Errorf("Expected HasStarted to be %v but got %v for %s", tc.want, got, tc.name)
			}
		})
	}
}

func TestTaskConditionCheckState_IsComplete(t *testing.T) {
	tcs := []struct {
		name  string
		state TaskConditionCheckState
		want  bool
	}{{
		name:  "no-condition-checks",
		state: notStartedState,
		want:  false,
	}, {
		name:  "running-condition-check",
		state: runningState,
		want:  false,
	}, {
		name:  "successful-condition-check",
		state: successState,
		want:  true,
	}, {
		name:  "failed-condition-check",
		state: failedState,
		want:  true,
	}}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.state.IsDone()
			if got != tc.want {
				t.Errorf("Expected IsComplete to be %v but got %v for %s", tc.want, got, tc.name)
			}
		})
	}
}

func TestTaskConditionCheckState_IsSuccess(t *testing.T) {
	tcs := []struct {
		name  string
		state TaskConditionCheckState
		want  bool
	}{{
		name:  "no-condition-checks",
		state: notStartedState,
		want:  false,
	}, {
		name:  "running-condition-check",
		state: runningState,
		want:  false,
	}, {
		name:  "successful-condition-check",
		state: successState,
		want:  true,
	}, {
		name:  "failed-condition-check",
		state: failedState,
		want:  false,
	}}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.state.IsSuccess()
			if got != tc.want {
				t.Errorf("Expected IsSuccess to be %v but got %v for %s", tc.want, got, tc.name)
			}
		})
	}
}

func TestResolvedConditionCheck_ConditionToTaskSpec(t *testing.T) {
	tcs := []struct {
		name string
		cond *v1alpha1.Condition
		want v1alpha1.TaskSpec
	}{{
		name: "user-provided-container-name",
		cond: tb.Condition("name", "foo", tb.ConditionSpec(
			tb.ConditionSpecCheck("foo", "ubuntu"),
		)),
		want: v1alpha1.TaskSpec{
			Steps: []corev1.Container{{
				Name:  "foo",
				Image: "ubuntu",
			}},
		},
	}, {
		name: "default-container-name",
		cond: tb.Condition("bar", "foo", tb.ConditionSpec(
			tb.ConditionSpecCheck("", "ubuntu"),
		)),
		want: v1alpha1.TaskSpec{
			Steps: []corev1.Container{{
				Name:  "condition-check-bar",
				Image: "ubuntu",
			}},
		},
	}, {
		name: "with-input-params",
		cond: tb.Condition("bar", "foo", tb.ConditionSpec(
			tb.ConditionSpecCheck("${params.name}", "${params.img}",
				tb.WorkingDir("${params.not.replaced}")),
			tb.ConditionParamSpec("name", v1alpha1.ParamTypeString),
			tb.ConditionParamSpec("img", v1alpha1.ParamTypeString),
		)),
		want: v1alpha1.TaskSpec{
			Inputs: &v1alpha1.Inputs{
				Params: []v1alpha1.ParamSpec{{
					Name: "name",
					Type: "string",
				}, {
					Name: "img",
					Type: "string",
				}},
			},
			Steps: []corev1.Container{{
				Name:       "${inputs.params.name}",
				Image:      "${inputs.params.img}",
				WorkingDir: "${params.not.replaced}",
			}},
		},
	}}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			rcc := &ResolvedConditionCheck{Condition: tc.cond}
			if d := cmp.Diff(tc.want, *rcc.ConditionToTaskSpec()); d != "" {
				t.Errorf("TaskSpec generated from Condition is unexpected -want, +got: %v", d)
			}
		})
	}
}
