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
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// LineMatcher is a regular expression used to patch an expected expression
// from build optimization output for a line in a Go source file.
type LineMatcher struct {
	// Regexp is matched against the build optimization output.
	Regexp *regexp.Regexp

	// Source is the line of source code for which this matcher was built.
	Source string
}

func (lm LineMatcher) deepEqual(b LineMatcher) bool {
	if lm.Source != b.Source {
		return false
	}
	ar, br := lm.Regexp, b.Regexp
	if ar == nil && br != nil {
		return false
	}
	if ar != nil && br == nil {
		return false
	}
	if ar != nil && br != nil && ar.String() != br.String() {
		return false
	}
	return true
}

// TestCase is a test case parsed from the lem comments in a source file.
type TestCase struct {
	// ID maps to lem.<ID>.
	ID string

	// Name maps to lem.<ID>.name=<NAME>.
	// Please see the lem package documentation for more information.
	Name string

	// AllocOp maps to lem.<ID>.alloc=\d+(-\d+)? and is the expected number
	// of allocations per operation.
	AllocOp Int64Range

	// BytesOp maps to lem.<ID>.bytes=\d+(-\d+)? and is the expected number
	// of bytes per per operation.
	BytesOp Int64Range

	// Matches maps to lem.<ID>.m= and is a list of patterns that must appear
	// in the optimization output.
	Matches []LineMatcher

	// Natches maps to lem.<ID>.m!= and is a list of patterns that must appear
	// in the optimization output.
	Natches []LineMatcher
}

func (tc TestCase) deepEqual(b TestCase) bool {
	if tc.ID != b.ID {
		return false
	}
	if tc.Name != b.Name {
		return false
	}
	if !tc.AllocOp.deepEqual(b.AllocOp) {
		return false
	}
	if !tc.BytesOp.deepEqual(b.BytesOp) {
		return false
	}
	if len(tc.Matches) != len(b.Matches) {
		return false
	}
	for i := range tc.Matches {
		if !tc.Matches[i].deepEqual(b.Matches[i]) {
			return false
		}
	}
	if len(tc.Natches) != len(b.Natches) {
		return false
	}
	for i := range tc.Natches {
		if !tc.Natches[i].deepEqual(b.Natches[i]) {
			return false
		}
	}
	return true
}

// Path returns the test case path from the provided ID and name.
// Please see the lem package documentation for more information.
func (tc TestCase) Path() []string {
	var path []string
	if len(tc.Name) == 0 || tc.Name[0] != '/' {
		path = append(path, tc.ID)
	}
	if len(tc.Name) > 0 {
		path = append(path, strings.Split(tc.Name, "/")...)
	}

	// Remove any empty elements from the slice.
	temp := path[:0]
	for _, s := range path {
		if s != "" {
			temp = append(temp, s)
		}
	}

	return temp
}

var (
	nameRx  = regexp.MustCompile(`^// lem\.([^.]+)\.name=(.+)$`)
	allocRx = regexp.MustCompile(`^// lem\.([^.]+)\.alloc=(\d+)(?:-(\d+))?$`)
	bytesRx = regexp.MustCompile(`^// lem\.([^.]+)\.bytes=(\d+)(?:-(\d+))?$`)
	matchRx = regexp.MustCompile(`^// lem\.([^.]+)\.m=(.+)$`)
	natchRx = regexp.MustCompile(`^// lem\.([^.]+)\.m!=(.+)$`)
	newlnRx = regexp.MustCompile(`\r?\n`)
)

// GetTestCases parses the provided Go source files & returns a TestCase slice.
func GetTestCases(files ...string) ([]TestCase, error) {
	var (
		testCases []TestCase
		lookupTbl = testCaseLookupTable{}
	)
	for _, filePath := range files {
		testCasesInFile, err := getTestCasesInFile(filePath, lookupTbl)
		if err != nil {
			return nil, err
		}

		// Store the length of the testCases slice and then append the
		// test cases from the file to it.
		indexOfUnmappedTestCases := len(testCases)
		testCases = append(testCases, testCasesInFile...)

		// Add the newly appended test cases to the lookup table.
		for i := indexOfUnmappedTestCases; i < len(testCases); i++ {
			lookupTbl[testCases[i].ID] = &testCases[i]
		}
	}
	return testCases, nil
}

// testCaseLookupTable provides a quick way to check if a test case already
// exists.
type testCaseLookupTable map[string]*TestCase

// Get the test case with the specified ID, otherwise an error is returned.
func (t testCaseLookupTable) Get(id string) (*TestCase, error) {
	tc, ok := t[id]
	if !ok {
		return nil, fmt.Errorf("unknown test case ID: %s", id)
	}
	return tc, nil
}

func readLines(filePath string) ([]string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return newlnRx.Split(string(data), -1), nil
}

func getTestCasesInFile(
	filePath string,
	lookupTbl testCaseLookupTable) ([]TestCase, error) {

	var (
		testCases []TestCase
		fileName  = filepath.Base(filePath)
	)

	if lookupTbl == nil {
		lookupTbl = testCaseLookupTable{}
	}

	var fset token.FileSet
	f, err := parser.ParseFile(&fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	lines, err := readLines(filePath)
	if err != nil {
		return nil, err
	}

	// Scan each line of the file for lem comments.
	for _, cg := range f.Comments {
		for _, c := range cg.List {
			var (
				l      = c.Text
				tc     *TestCase
				lineNo = fset.Position(c.Pos()).Line
			)

			// lem.<ID>.name=<NAME>
			if m := nameRx.FindStringSubmatch(l); m != nil {
				id, name := m[1], m[2]
				if tc, _ = lookupTbl.Get(id); tc != nil {
					if tc.Name != "" {
						return nil, fmt.Errorf("duplicate lem.%s.name", id)
					}
					tc.Name = name
				} else {
					testCases = append(testCases, TestCase{ID: id, Name: name})
					lookupTbl[id] = &testCases[len(testCases)-1]
				}
			} else if m := allocRx.FindStringSubmatch(l); m != nil {
				if tc, _ = lookupTbl.Get(m[1]); tc == nil {
					testCases = append(testCases, TestCase{ID: m[1]})
					tc = &testCases[len(testCases)-1]
					lookupTbl[m[1]] = tc
				}
				min, err := strconv.ParseInt(m[2], 10, 64)
				if err != nil {
					return nil, err
				}
				tc.AllocOp.Min = min
				if len(m) < 3 || m[3] == "" {
					tc.AllocOp.Max = min
				} else {
					max, err := strconv.ParseInt(m[3], 10, 64)
					if err != nil {
						return nil, err
					}
					tc.AllocOp.Max = max
				}
			} else if m := bytesRx.FindStringSubmatch(l); m != nil {
				if tc, _ = lookupTbl.Get(m[1]); tc == nil {
					testCases = append(testCases, TestCase{ID: m[1]})
					tc = &testCases[len(testCases)-1]
					lookupTbl[m[1]] = tc
				}
				min, err := strconv.ParseInt(m[2], 10, 64)
				if err != nil {
					return nil, err
				}
				tc.BytesOp.Min = min
				if len(m) < 3 || m[3] == "" {
					tc.BytesOp.Max = min
				} else {
					max, err := strconv.ParseInt(m[3], 10, 64)
					if err != nil {
						return nil, err
					}
					tc.BytesOp.Max = max
				}
			} else if m := matchRx.FindStringSubmatch(l); m != nil {
				if tc, _ = lookupTbl.Get(m[1]); tc == nil {
					testCases = append(testCases, TestCase{ID: m[1]})
					tc = &testCases[len(testCases)-1]
					lookupTbl[m[1]] = tc
				}
				r, err := regexp.Compile(
					fmt.Sprintf(
						"(?m)^.*%s:%d:\\d+: %s$", fileName, lineNo, m[2]),
				)
				if err != nil {
					return nil, err
				}
				tc.Matches = append(tc.Matches, LineMatcher{
					Regexp: r,
					Source: lines[lineNo],
				})
			} else if m := natchRx.FindStringSubmatch(l); m != nil {
				if tc, _ = lookupTbl.Get(m[1]); tc == nil {
					testCases = append(testCases, TestCase{ID: m[1]})
					tc = &testCases[len(testCases)-1]
					lookupTbl[m[1]] = tc
				}
				r, err := regexp.Compile(
					fmt.Sprintf(
						"(?m)^.*%s:%d:\\d+:.*%s.*$", fileName, lineNo, m[2]),
				)
				if err != nil {
					return nil, err
				}
				tc.Natches = append(tc.Natches, LineMatcher{
					Regexp: r,
					Source: lines[lineNo],
				})
			}
		}
	}

	return testCases, nil
}
