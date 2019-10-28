/*
Copyright 2019 The Tekton Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1_test

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"

	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	tb "github.com/tektoncd/pipeline/test/builder"
	"github.com/tektoncd/pipeline/test/names"
)

const digestImage = "override-with-imagedigest-exporter-image:latest"

func Test_Invalid_NewImageResource(t *testing.T) {
	r := tb.PipelineResource("git-resource", "default", tb.PipelineResourceSpec(v1alpha1.PipelineResourceTypeGit))

	_, err := v1alpha1.NewImageResource(digestImage, r)
	if err == nil {
		t.Error("Expected error creating Image resource")
	}
}

func Test_Valid_NewImageResource(t *testing.T) {
	want := &v1alpha1.ImageResource{
		Name:        "image-resource",
		Type:        v1alpha1.PipelineResourceTypeImage,
		URL:         "https://test.com/test/test",
		Digest:      "test",
		DigestImage: digestImage,
	}

	r := tb.PipelineResource(
		"image-resource",
		"default",
		tb.PipelineResourceSpec(
			v1alpha1.PipelineResourceTypeImage,
			tb.PipelineResourceSpecParam("URL", "https://test.com/test/test"),
			tb.PipelineResourceSpecParam("Digest", "test"),
		),
	)

	got, err := v1alpha1.NewImageResource(digestImage, r)
	if err != nil {
		t.Fatalf("Unexpected error creating Image resource: %s", err)
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Mismatch of Image resource: %s", diff)
	}
}

func Test_ImageResource_Replacements(t *testing.T) {
	ir := &v1alpha1.ImageResource{
		Name:   "image-resource",
		Type:   v1alpha1.PipelineResourceTypeImage,
		URL:    "https://test.com/test/test",
		Digest: "test",
	}

	want := map[string]string{
		"name":   "image-resource",
		"type":   string(v1alpha1.PipelineResourceTypeImage),
		"url":    "https://test.com/test/test",
		"digest": "test",
	}

	got := ir.Replacements()

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Mismatch of ImageResource Replacements: %s", diff)
	}
}

func Test_ImgResource_GetOutputTaskModifier(t *testing.T) {
	names.TestingSeed()

	r := &v1alpha1.ImageResource{
		Name:        "image-resource",
		Type:        v1alpha1.PipelineResourceTypeImage,
		URL:         "docker.io/test/test:v0.1",
		DigestImage: digestImage,
	}

	ts := v1alpha1.TaskSpec{}
	modifier, err := r.GetOutputTaskModifier(&ts, "/test/test")
	if err != nil {
		t.Fatalf("Unexpected error getting GetOutputTaskModifier: %s", err)
	}

	output := []*v1alpha1.ImageResource{r}
	imagesJSON, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("Unexpected error converting to json: %s", err)
	}

	want := []v1alpha1.Step{{Container: corev1.Container{
		Name:    "image-digest-exporter-image-resource-9l9zj",
		Image:   "override-with-imagedigest-exporter-image:latest",
		Command: []string{"/ko-app/imagedigestexporter"},
		Args: []string{
			"-images", string(imagesJSON),
		},
		TerminationMessagePolicy: corev1.TerminationMessageFallbackToLogsOnError,
	}}}

	if diff := cmp.Diff(want, modifier.GetStepsToAppend()); diff != "" {
		t.Errorf("Mismatch of ImageResource OutputContainerSpec: %s", diff)
	}
}
