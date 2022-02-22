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

package internal

import (
	"bytes"
	"fmt"
	"go/build"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// Context is an internal subset of lem.Context. Please refer to lem.Context
// for additional information.
type Context struct {
	Benchmarks    map[string]func(*testing.B)
	BuildOutput   string
	CompilerFlags []string
}

// Int64Range is an inclusive range of int64 values.
type Int64Range struct {
	Min int64
	Max int64
}

func (i Int64Range) deepEqual(b Int64Range) bool {
	return i.Min == b.Min && i.Max == b.Max
}

// Eq returns true when (Min==Max && a==Min) || (a>=Min && a<=Max).
func (i Int64Range) Eq(a int64) bool {
	if i.Min == i.Max {
		return i.Min == a
	}
	return a >= i.Min && a <= i.Max
}

// String returns the string version of this value.
func (i Int64Range) String() string {
	if i.Min == i.Max {
		return fmt.Sprintf("%d", i.Min)
	}
	return fmt.Sprintf("%d-%d", i.Min, i.Max)
}

// Build builds the specified package in order to produce the optimization
// output.
func Build(w io.Writer, pkg build.Package, ctx Context) error {

	// If there are no valid Go sources, test or otherwise, then
	// return early.
	if len(pkg.GoFiles) == 0 &&
		len(pkg.TestGoFiles) == 0 &&
		len(pkg.XTestGoFiles) == 0 {
		return nil
	}

	// Build a set of compiler flags.
	compilerFlags := []string{"-m"}
	for _, f := range ctx.CompilerFlags {
		if f != "-m" { // do not add a duplicate -m flag
			compilerFlags = append(compilerFlags, f)
		}
	}
	compilerFlagVal := strings.Join(compilerFlags, " ")

	// Build the package's test binary if there are any test files.
	var didTestBuildPackage bool
	if len(pkg.TestGoFiles) > 0 || len(pkg.XTestGoFiles) > 0 {
		tempFileName, err := getTempFileName()
		if err != nil {
			return err
		}
		defer os.RemoveAll(tempFileName)
		args := []string{
			"test",
			"-c", "-o", tempFileName,
			"-gcflags", compilerFlagVal,
			pkg.ImportPath,
		}
		if err := forkGo(w, args...); err != nil {
			return err
		}

		// Does the test import the package?
		for _, testImport := range pkg.TestImports {
			if testImport == pkg.ImportPath {
				didTestBuildPackage = true
				break
			}
		}
	}

	// Build the package if there are any sources and if the
	if len(pkg.GoFiles) > 0 && !didTestBuildPackage {
		// Build the list of arguments used to build the package.
		args := []string{
			"build",
			"-gcflags", compilerFlagVal,
			pkg.ImportPath,
		}
		if err := forkGo(w, args...); err != nil {
			return err
		}
	}

	return nil
}

func forkGo(w io.Writer, args ...string) error {
	var stderr bytes.Buffer
	cmd := exec.Command("go", args...)
	cmd.Stderr = io.MultiWriter(w, &stderr)
	if err := cmd.Run(); err != nil {
		log.Printf("failed: go %s\n", strings.Join(args, " "))
		return fmt.Errorf("%w\n%s", err, stderr.String())
	}
	return nil
}

func getTempFileName() (string, error) {
	tempFile, err := ioutil.TempFile("", "")
	if err != nil {
		return "", err
	}
	tempFileName := tempFile.Name()
	if err := tempFile.Close(); err != nil {
		return "", err
	}
	if err := os.RemoveAll(tempFileName); err != nil {
		return "", err
	}
	return tempFileName, nil
}
