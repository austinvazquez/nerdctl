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
	"gotest.tools/v3/assert"
)

func TestBuildContextWithOCILayout(t *testing.T) {
	testutil.DockerIncompatible(t)
	testutil.RequiresBuild(t)
	testutil.RegisterBuildCacheCleanup(t)

	base := testutil.NewBase(t)
	imageName := testutil.Identifier(t)
	t.Cleanup(func() { base.Cmd("rmi", imageName) })

	dockerfile := fmt.Sprintf(`FROM %s
LABEL layer=oci-layout
CMD ["echo", "nerdctl-build-oci-layout-build-context"]`, testutil.CommonImage)
	buildCtx := createBuildContext(t, dockerfile)

	tarPath := fmt.Sprintf("%s/%s", buildCtx, "test.tar")
	base.Cmd("build", buildCtx, fmt.Sprintf("--output=type=oci,dest=%s", tarPath)).Run()

	ociLayoutDir := t.TempDir()

	err := extractTarFile(ociLayoutDir, tarPath)
	assert.NilError(t, err)

	ociLayout := "test"
	dockerfile = fmt.Sprintf(`FROM %s
CMD ["echo", "nerdctl-test-build-context-oci-layout"]`, ociLayout)
	buildCtx = createBuildContext(t, dockerfile)

	base.Cmd("build", buildCtx, fmt.Sprintf("--build-context=%s=oci-layout://%s", ociLayout, ociLayoutDir), "-t", imageName).AssertOK()
}
