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
	"time"

	"github.com/containerd/containerd/v2/core/images"
	"gotest.tools/v3/assert"
)

func TestApplyFilters(t *testing.T) {
	tests := []struct {
		name           string
		images         []images.Image
		filters        []FilterFunc
		expectedImages []images.Image
		expectedErr    error
	}{
		{
			name:   "EmptyList",
			images: []images.Image{},
			filters: []FilterFunc{
				FilterDanglingImages(),
			},
			expectedImages: []images.Image{},
		},
		{
			name: "ApplyNoFilters",
			images: []images.Image{
				{
					Name: "<none>",
				},
				{
					Name: "docker.io/library/hello-world:latest",
				},
			},
			filters: []FilterFunc{},
			expectedImages: []images.Image{
				{
					Name: "<none>",
				},
				{
					Name: "docker.io/library/hello-world:latest",
				},
			},
		},
		{
			name: "ApplySingleFilter",
			images: []images.Image{
				{
					Name: "<none>",
				},
				{
					Name: "docker.io/library/hello-world:latest",
				},
			},
			filters: []FilterFunc{
				FilterDanglingImages(),
			},
			expectedImages: []images.Image{
				{
					Name: "<none>",
				},
			},
		},
		{
			name: "ApplyMultipleFilters",
			images: []images.Image{
				{
					Name: "<none>",
				},
				{
					Name: "alpine:3.19",
				},
				{
					Name: "docker.io/library/hello-world:latest",
				},
				{
					Name: "public.ecr.aws/docker/library/hello-world:latest",
				},
			},
			filters: []FilterFunc{
				FilterTaggedImages(),
				FilterByReference([]string{"hello-world"}),
			},
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
			name: "ReturnErrorAndEmptyListOnFilterError",
			images: []images.Image{
				{
					Name: "<none>:<none>",
				},
				{
					Name: "docker.io/library/hello-world:latest",
				},
			},
			filters: []FilterFunc{
				FilterDanglingImages(),
				FilterUntil(""),
			},
			expectedImages: []images.Image{},
			expectedErr:    errNoUntilTimestamp,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actualImages, err := ApplyFilters(test.images, test.filters...)
			if test.expectedErr == nil {
				assert.NilError(t, err)
			} else {
				assert.ErrorIs(t, err, test.expectedErr)
			}
			assert.Equal(t, len(actualImages), len(test.expectedImages))
			assert.DeepEqual(t, actualImages, test.expectedImages)
		})
	}
}

func TestFilterUntil(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		name           string
		until          string
		images         []images.Image
		expectedImages []images.Image
		expectedErr    error
	}{
		{
			name:           "EmptyTimestampReturnsError",
			until:          "",
			images:         []images.Image{},
			expectedImages: []images.Image{},
			expectedErr:    errNoUntilTimestamp,
		},
		{
			name:           "UnparseableTimestampReturnsError",
			until:          "-2006-01-02T15:04:05Z07:00",
			images:         []images.Image{},
			expectedImages: []images.Image{},
			expectedErr:    errUnparsableUntilTimestamp,
		},
		{
			name:  "ImagesOlderThan3Hours(Go duration)",
			until: "3h",
			images: []images.Image{
				{
					Name:      "image:yesterday",
					CreatedAt: now.Add(-24 * time.Hour),
				},
				{
					Name:      "image:today",
					CreatedAt: now.Add(-12 * time.Hour),
				},
				{
					Name:      "image:latest",
					CreatedAt: now,
				},
			},
			expectedImages: []images.Image{
				{
					Name:      "image:yesterday",
					CreatedAt: now.Add(-24 * time.Hour),
				},
				{
					Name:      "image:today",
					CreatedAt: now.Add(-12 * time.Hour),
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actualImages, err := FilterUntil(test.until)(test.images)
			if test.expectedErr == nil {
				assert.NilError(t, err)
			} else {
				assert.ErrorIs(t, err, test.expectedErr)
			}
			assert.Equal(t, len(actualImages), len(test.expectedImages))
			assert.DeepEqual(t, actualImages, test.expectedImages)
		})
	}
}

func TestFilterByReference(t *testing.T) {
	tests := []struct {
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

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actualImages, err := FilterByReference(test.referencePatterns)(test.images)
			if test.expectedErr == nil {
				assert.NilError(t, err)
			} else {
				assert.ErrorIs(t, err, test.expectedErr)
			}
			assert.Equal(t, len(actualImages), len(test.expectedImages))
			assert.DeepEqual(t, actualImages, test.expectedImages)
		})
	}
}

func TestFilterDanglingImages(t *testing.T) {
	tests := []struct {
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
					Name:   "<none>",
					Labels: map[string]string{"ref": "dangling2"},
				},
			},
			expectedImages: []images.Image{
				{
					Name:   "",
					Labels: map[string]string{"ref": "dangling1"},
				},
				{
					Name:   "<none>",
					Labels: map[string]string{"ref": "dangling2"},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actualImages, err := FilterDanglingImages()(test.images)
			assert.NilError(t, err)
			assert.Equal(t, len(actualImages), len(test.expectedImages))
			assert.DeepEqual(t, actualImages, test.expectedImages)
		})
	}
}

func TestFilterTaggedImages(t *testing.T) {
	tests := []struct {
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
			name:     "IsTagged",
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
					Name:   "<none>",
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

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actualImages, err := FilterTaggedImages()(test.images)
			assert.NilError(t, err)
			assert.Equal(t, len(actualImages), len(test.expectedImages))
			assert.DeepEqual(t, actualImages, test.expectedImages)
		})
	}
}

func TestImageCreatedBetween(t *testing.T) {
	tests := []struct {
		name         string
		image        images.Image
		lhs          time.Time
		rhs          time.Time
		fallsBetween bool
	}{
		{
			name: "BetweenImage",
			image: images.Image{
				CreatedAt: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			lhs:          time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
			rhs:          time.Now(),
			fallsBetween: true,
		},
		{
			name: "ExclusiveLeft",
			image: images.Image{
				CreatedAt: time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			lhs:          time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
			rhs:          time.Now(),
			fallsBetween: false,
		},
		{
			name: "ExclusiveRight",
			image: images.Image{
				CreatedAt: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			lhs:          time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
			rhs:          time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			fallsBetween: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, imageCreatedBetween(test.image, test.lhs, test.rhs), test.fallsBetween)
		})
	}
}

func TestMatchesAnyLabel(t *testing.T) {
	tests := []struct {
		name          string
		imageLabels   map[string]string
		labelsToMatch map[string]string
		matches       bool
	}{
		{
			name:          "DefaultFalse",
			imageLabels:   map[string]string{},
			labelsToMatch: map[string]string{},
			matches:       false,
		},
		{
			name:          "ImageHasNoLabels",
			imageLabels:   map[string]string{},
			labelsToMatch: map[string]string{"foo": "bar"},
			matches:       false,
		},
		{
			name:          "SingleMatchingLabel",
			imageLabels:   map[string]string{"org": "com.containerd.nerdctl"},
			labelsToMatch: map[string]string{"org": "com.containerd.nerdctl"},
			matches:       true,
		},
		{
			name:          "KeyOnlyMatchingLabel",
			imageLabels:   map[string]string{"org": "com.containerd.nerdctl"},
			labelsToMatch: map[string]string{"org": ""},
			matches:       true,
		},
		{
			name:          "KeyValueDoesNotMatch",
			imageLabels:   map[string]string{"org": "com.containerd.nerdctl"},
			labelsToMatch: map[string]string{"org": "com.containerd.containerd"},
			matches:       false,
		},
		{
			name:          "AnyMatchingLabel",
			imageLabels:   map[string]string{"org": "com.containerd.nerdctl", "foo": "bar"},
			labelsToMatch: map[string]string{"org": "com.containerd.containerd", "foo": "bar"},
			matches:       true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, matchesAnyLabel(test.imageLabels, test.labelsToMatch), test.matches)
		})
	}
}
