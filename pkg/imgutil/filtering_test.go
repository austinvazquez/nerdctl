/*
   Copyright The containerd Authors.

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

package imgutil

import (
	"testing"

	"github.com/containerd/containerd/v2/core/images"
	"gotest.tools/v3/assert"
)

func TestFilterByReference(t *testing.T) {
	testCases := []struct {
		name              string
		referencePatterns []string
		images            []images.Image
		expectedImages    []images.Image
		expectedErr       error
	}{
		{
			name:           "EmptyList",
			images:         []images.Image{},
			expectedImages: []images.Image{},
		},
		{
			name: "MatchByReference",
			images: []images.Image{
				{
					Name: "foo:latest",
				},
				{
					Name: "docker.io/library/hello-world:latest",
				},
				{
					Name: "public.ecr.aws/docker/library/hello-world:latest",
				},
			},
			referencePatterns: []string{"hello-world"},
			expectedImages: []images.Image{
				{
					Name: "docker.io/library/hello-world:latest",
				},
				{
					Name: "public.ecr.aws/docker/library/hello-world:latest",
				},
			},
		},
		{
			name: "NoMatchExists",
			images: []images.Image{
				{
					Name: "foo:latest",
				},
				{
					Name: "docker.io/library/hello-world:latest",
				},
				{
					Name: "public.ecr.aws/docker/library/hello-world:latest",
				},
			},
			referencePatterns: []string{"foobar"},
			expectedImages:    []images.Image{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualImages, err := FilterByReference(tc.referencePatterns)(tc.images)
			if tc.expectedErr == nil {
				assert.NilError(t, err)
			} else {
				assert.ErrorIs(t, err, tc.expectedErr)
			}
			assert.Equal(t, len(actualImages), len(tc.expectedImages))
			assert.DeepEqual(t, actualImages, tc.expectedImages)
		})
	}
}

func TestFilterByDangling(t *testing.T) {
	testCases := []struct {
		name           string
		dangling       bool
		images         []images.Image
		expectedImages []images.Image
	}{
		{
			name:           "EmptyList",
			dangling:       true,
			images:         []images.Image{},
			expectedImages: []images.Image{},
		},
		{
			name:     "IsDangling",
			dangling: true,
			images: []images.Image{
				{
					Name:   "",
					Labels: map[string]string{"ref": "dangling1"},
				},
				{
					Name: "docker.io/library/hello-world:latest",
				},
				{
					Name:   "",
					Labels: map[string]string{"ref": "dangling2"},
				},
			},
			expectedImages: []images.Image{
				{
					Name:   "",
					Labels: map[string]string{"ref": "dangling1"},
				},
				{
					Name:   "",
					Labels: map[string]string{"ref": "dangling2"},
				},
			},
		},
		{
			name:     "IsNotDangling",
			dangling: false,
			images: []images.Image{
				{
					Name:   "",
					Labels: map[string]string{"ref": "dangling1"},
				},
				{
					Name: "docker.io/library/hello-world:latest",
				},
				{
					Name:   "",
					Labels: map[string]string{"ref": "dangling2"},
				},
			},
			expectedImages: []images.Image{
				{
					Name: "docker.io/library/hello-world:latest",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualImages, err := FilterByDangling(tc.dangling)(tc.images)
			assert.NilError(t, err)
			assert.Equal(t, len(actualImages), len(tc.expectedImages))
			assert.DeepEqual(t, actualImages, tc.expectedImages)
		})
	}
}
