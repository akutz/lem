/*
Copyright 2022

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

package lem

import (
	"bytes"
	"flag"
	"fmt"
	"go/build"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/akutz/lem/internal"
)

// Context provides a means to configure the test execution.
type Context struct {
	// Benchmarks is an optional map of functions to benchmark.
	//
	// Keys in this map should correspond go the <ID> from "lem.<ID>" comments.
	//
	// Please note this is required to assert allocations and/or bytes.
	Benchmarks map[string]func(*testing.B)

	// BuildContext is the support context for building the specified
	// packages and discovering their source files.
	//
	// Please see https://pkg.go.dev/go/build#Context for more information.
	BuildContext *build.Context

	// BuildOutput may be used in place of building any of the specified
	// packages.
	// If this field is specified then there will be no calls to "go build"
	// or "go test."
	BuildOutput string

	// CompilerFlags is a list of flags to pass to the compiler.
	//
	// Please note the "-m" flag will always be used, whether it is included
	// in this list or not.
	CompilerFlags []string

	// ImportedPackages is a list of imported packages to include in the
	// testing.
	//
	// Please note if this field has a non-zero number of elements then
	// the Packages field is ignored.
	ImportedPackages []build.Package

	// Packages is a list of packages to include in the testing.
	//
	// Please note this field is ignored if the ImportedPackages field has a
	// non-zero number of elements.
	Packages []string
}

// Copy returns a copy of this context.
func (src Context) Copy() Context {
	return Context{
		Benchmarks:       copyNillableBenchmarksMap(src.Benchmarks),
		BuildContext:     copyNillableGoBuildContext(src.BuildContext),
		BuildOutput:      src.BuildOutput,
		CompilerFlags:    copyNillableStringSlice(src.CompilerFlags),
		ImportedPackages: copyNillableImportedPackageSlice(src.ImportedPackages),
		Packages:         copyNillableStringSlice(src.Packages),
	}
}

func (src Context) toInternal() internal.Context {
	return internal.Context{
		Benchmarks:    copyNillableBenchmarksMap(src.Benchmarks),
		BuildOutput:   src.BuildOutput,
		CompilerFlags: copyNillableStringSlice(src.CompilerFlags),
	}
}

// copyGoBuildContext returns a copy of the provided, Go build context.
func copyGoBuildContext(src build.Context) build.Context {
	return build.Context{
		GOARCH:        src.GOARCH,
		GOOS:          src.GOOS,
		GOROOT:        src.GOROOT,
		GOPATH:        src.GOROOT,
		Dir:           src.Dir,
		CgoEnabled:    src.CgoEnabled,
		UseAllFiles:   src.UseAllFiles,
		Compiler:      src.Compiler,
		BuildTags:     copyNillableStringSlice(src.BuildTags),
		ToolTags:      copyNillableStringSlice(src.ToolTags),
		ReleaseTags:   copyNillableStringSlice(src.ReleaseTags),
		InstallSuffix: src.InstallSuffix,
		JoinPath:      src.JoinPath,
		SplitPathList: src.SplitPathList,
		IsAbsPath:     src.IsAbsPath,
		IsDir:         src.IsDir,
		HasSubdir:     src.HasSubdir,
		ReadDir:       src.ReadDir,
		OpenFile:      src.OpenFile,
	}
}

// NewBuildContext returns a copy of Go's default build context.
func NewBuildContext() build.Context {
	return copyGoBuildContext(build.Default)
}

func copyNillableGoBuildContext(src *build.Context) *build.Context {
	if src == nil {
		return nil
	}
	dst := copyGoBuildContext(*src)
	return &dst
}

func copyNillableBenchmarksMap(
	src map[string]func(*testing.B)) map[string]func(*testing.B) {
	if src == nil {
		return nil
	}
	dst := map[string]func(*testing.B){}
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func copyNillableStringSlice(src []string) []string {
	if src == nil {
		return nil
	}
	if len(src) == 0 {
		return []string{}
	}
	dst := make([]string, len(src))
	for i := range src {
		dst[i] = src[i]
	}
	return dst
}

func copyNillableImportedPackageSlice(
	src []build.Package) []build.Package {
	if src == nil {
		return nil
	}
	if len(src) == 0 {
		return []build.Package{}
	}
	dst := make([]build.Package, len(src))
	for i := range src {
		dst[i] = src[i]
	}
	return dst
}

func theirDirectory() (string, error) {
	_, callersFilePath, _, ok := runtime.Caller(2)
	if !ok {
		return "", fmt.Errorf("failed to obtain caller's directory")
	}
	return filepath.Dir(callersFilePath), nil
}

// Sets the value of the -test.benchtime flag and returns the original
// value if one was present, otherwise an empty string is returned.
//
// Please note this function is a no-op if the flag is not already
// defined.
func SetBenchtime(s string) string {
	f := flag.Lookup("test.benchtime")
	if f == nil {
		return ""
	}
	og := f.Value.String()
	f.Value.Set(s)
	return og
}

// Sets the value of the -test.benchmem flag and returns the original
// value if one was present, otherwise an empty string is returned.
//
// Please note this function is a no-op if the flag is not already
// defined.
func SetBenchmem(s string) string {
	f := flag.Lookup("test.benchmem")
	if f == nil {
		return ""
	}
	og := f.Value.String()
	f.Value.Set(s)
	return og
}

// Tags returns a slice of the value of the tags flag.
func Tags() []string {
	f := flag.Lookup("tags")
	if f == nil {
		return nil
	}
	var tags []string
	for _, t := range strings.Split(f.Value.String(), ",") {
		t := strings.TrimSpace(t)
		if t != "" {
			tags = append(tags, t)
		}
	}
	return tags
}

// Run validates the leak, escape, and move assertions for the caller's
// package and test package (if different).
func Run(t *testing.T) {
	dir, err := theirDirectory()
	if err != nil {
		t.Fatal(err)
	}
	run(t, dir, Context{})
}

// RunWithBenchmarks validates the leak, escape, move assertions, and
// heap allocation assertions for the caller's package and test package
// (if different).
func RunWithBenchmarks(t *testing.T, benchmarks map[string]func(*testing.B)) {
	dir, err := theirDirectory()
	if err != nil {
		t.Fatal(err)
	}
	run(t, dir, Context{Benchmarks: benchmarks})
}

// RunWithContext validates the leak, escape, and move assertions for the
// packages specified in the provided options. Heap allocation assertions
// may also occur if the provided context includes the benchmarks map.
func RunWithContext(t *testing.T, ctx Context) {
	dir, err := theirDirectory()
	if err != nil {
		t.Fatal(err)
	}
	run(t, dir, ctx)
}

func run(t *testing.T, srcDir string, ctx Context) {
	ctx = ctx.Copy()

	// Create a new build context if one does not exist.
	if ctx.BuildContext == nil {
		buildContext := NewBuildContext()
		ctx.BuildContext = &buildContext
		ctx.BuildContext.BuildTags = Tags()
	}

	// If no package was specified then default to the package relative to
	// the provided source directory.
	if len(ctx.Packages) == 0 {
		ctx.Packages = []string{"."}
	}

	// If ctx.ImportedPackages is empty then create it from the
	// packages specified in ctx.Packages.
	if len(ctx.ImportedPackages) == 0 {
		ctx.ImportedPackages = make([]build.Package, len(ctx.Packages))
		for i, pkg := range ctx.Packages {
			ipkg, err := ctx.BuildContext.Import(
				pkg,
				srcDir,
				build.IgnoreVendor)
			if err != nil {
				t.Fatalf("failed to import pkg %s: %v", pkg, err)
			}
			ctx.ImportedPackages[i] = *ipkg
		}
	}

	var (
		allSrcFiles []string
		buildOutput bytes.Buffer
	)

	for _, pkg := range ctx.ImportedPackages {

		// Build the package if build output has not already been supplied.
		if ctx.BuildOutput == "" {
			if err := internal.Build(
				&buildOutput,
				pkg,
				ctx.toInternal()); err != nil {

				t.Fatalf("failed to build pkg %s: %v", pkg.ImportPath, err)
			}
		}

		// Get the package's sources and sort them so they maintain
		// lexographical order between all different types of sources.
		pkgSrcs := append([]string{}, pkg.GoFiles...)
		pkgSrcs = append(pkgSrcs, pkg.TestGoFiles...)
		pkgSrcs = append(pkgSrcs, pkg.XTestGoFiles...)
		sort.Strings(pkgSrcs)

		// Append the package sources to the overall number of sources.
		allSrcFiles = append(allSrcFiles, pkgSrcs...)
	}

	if ctx.BuildOutput == "" {
		ctx.BuildOutput = buildOutput.String()
	}

	testCases, err := internal.GetTestCases(allSrcFiles...)
	if err != nil {
		t.Fatalf("failed to get test cases: %v", err)
	}

	// Build a test case tree and run the tests.
	internal.NewTree(testCases...).Run(t, ctx.toInternal())
}
