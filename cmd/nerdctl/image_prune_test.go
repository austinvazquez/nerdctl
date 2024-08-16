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

package main

import (
	"fmt"
	"testing"

	"github.com/containerd/nerdctl/v2/pkg/testutil"
)

func TestImagePrune(t *testing.T) {
	testutil.RequiresBuild(t)
	testutil.RegisterBuildCacheCleanup(t)

	base := testutil.NewBase(t)
	imageName := testutil.Identifier(t)
	defer base.Cmd("rmi", imageName).AssertOK()

	dockerfile := fmt.Sprintf(`FROM %s
	CMD ["echo", "nerdctl-test-image-prune"]`, testutil.CommonImage)

	buildCtx := createBuildContext(t, dockerfile)

	base.Cmd("build", buildCtx).AssertOK()
	base.Cmd("build", "-t", imageName, buildCtx).AssertOK()
	base.Cmd("images").AssertOutContainsAll(imageName, "<none>")

	base.Cmd("image", "prune", "--force").AssertOutNotContains(imageName)
	base.Cmd("images").AssertOutNotContains("<none>")
	base.Cmd("images").AssertOutContains(imageName)
}

func TestImagePruneAll(t *testing.T) {
	testutil.RequiresBuild(t)
	testutil.RegisterBuildCacheCleanup(t)

	base := testutil.NewBase(t)
	imageName := testutil.Identifier(t)

	dockerfile := fmt.Sprintf(`FROM %s
	CMD ["echo", "nerdctl-test-image-prune"]`, testutil.CommonImage)

	buildCtx := createBuildContext(t, dockerfile)

	base.Cmd("build", "-t", imageName, buildCtx).AssertOK()
	// The following commands will clean up all images, so it should fail at this point.
	defer base.Cmd("rmi", imageName).AssertFail()
	base.Cmd("images").AssertOutContains(imageName)

	tID := testutil.Identifier(t)
	base.Cmd("run", "--name", tID, imageName).AssertOK()
	base.Cmd("image", "prune", "--force", "--all").AssertOutNotContains(imageName)
	base.Cmd("images").AssertOutContains(imageName)

	base.Cmd("rm", "-f", tID).AssertOK()
	base.Cmd("image", "prune", "--force", "--all").AssertOutContains(imageName)
	base.Cmd("images").AssertOutNotContains(imageName)
}

func TestImagePruneFilterBefore(t *testing.T) {
	testutil.RequiresBuild(t)
	testutil.RegisterBuildCacheCleanup(t)

	base := testutil.NewBase(t)

	imagePrefix := testutil.Identifier(t)

	images := []string{}
	for _, i := range []int{1, 2, 3} {
		images = append(images, fmt.Sprintf("%s-%d", imagePrefix, i))
	}

	for _, imageName := range images {
		dockerfile := fmt.Sprintf(`FROM %s
		CMD ["echo", "nerdctl-test-image-prune-filter-before"]`, testutil.CommonImage)

		buildCtx := createBuildContext(t, dockerfile)
		base.Cmd("build", "-t", imageName, buildCtx).AssertOK()

		base.Cmd("images").AssertOutContains(imageName)
	}

	beforeFilter := fmt.Sprintf("--filter=before=%s", images[1])
	base.Cmd("image", "prune", "--force", beforeFilter).AssertOK()

	base.Cmd("images").AssertOutNotContains(images[0])
	base.Cmd("images").AssertOutNotContains(images[1])
	base.Cmd("images").AssertOutContains(images[2])
}
