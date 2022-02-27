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
	"sync"
	"testing"
)

type Tree struct {
	TreeNode
	testsByID map[string]*TestCase
}

// NewTree returns a new Tree with the provided test cases.
func NewTree(testCases ...TestCase) Tree {
	var tree Tree
	for _, tc := range testCases {
		tree.Insert(tc)
	}
	return tree
}

// DeepEqual returns true if the two trees are equal.
// Only exported fields are compared.
func (tr *Tree) DeepEqual(b Tree) bool {
	return tr.TreeNode.deepEqual(b.TreeNode)
}

// Run the tests for this tree.
func (tr Tree) Run(t *testing.T, ctx Context) {
	tr.run(t, ctx)
}

func (tr *Tree) Get(id string) *TestCase {
	return tr.testsByID[id]
}

func (tr *Tree) Insert(testCase TestCase) *TestCase {
	tr.Once.Do(func() {
		tr.Index = map[string]int{}
		tr.testsByID = map[string]*TestCase{}
	})
	tc := tr.Get(testCase.ID)
	if tc != nil {
		return tc
	}
	tc = tr.TreeNode.insert(testCase, testCase.Path()...)
	tr.testsByID[tc.ID] = tc
	return tc
}

// TreeNode organizes the TestCases in a tree structure.
type TreeNode struct {
	sync.Once
	Index map[string]int
	Steps []string
	Nodes []TreeNode
	Tests []TestCase
}

func (tr *TreeNode) deepEqual(b TreeNode) bool {
	if len(tr.Index) != len(b.Index) {
		return false
	}
	for ak, av := range tr.Index {
		if bv, ok := b.Index[ak]; !ok || av != bv {
			return false
		}
	}
	if len(tr.Steps) != len(b.Steps) {
		return false
	}
	for i := range tr.Steps {
		if tr.Steps[i] != b.Steps[i] {
			return false
		}
	}
	if len(tr.Tests) != len(b.Tests) {
		return false
	}
	for i := range tr.Tests {
		if !tr.Tests[i].deepEqual(b.Tests[i]) {
			return false
		}
	}
	if len(tr.Nodes) != len(b.Nodes) {
		return false
	}
	for i := range tr.Nodes {
		if !tr.Nodes[i].deepEqual(b.Nodes[i]) {
			return false
		}
	}
	return true
}

func (tr *TreeNode) insert(testCase TestCase, path ...string) *TestCase {
	tr.Once.Do(func() { tr.Index = map[string]int{} })
	if len(path) < 2 {
		if len(path) == 1 {
			testCase.Name = path[0]
		}
		tr.Tests = append(tr.Tests, testCase)
		return &tr.Tests[len(tr.Tests)-1]
	} else {
		if _, ok := tr.Index[path[0]]; !ok {
			tr.Index[path[0]] = len(tr.Nodes)
			tr.Nodes = append(tr.Nodes, TreeNode{})
			tr.Steps = append(tr.Steps, path[0])
		}
		return tr.Nodes[tr.Index[path[0]]].insert(testCase, path[1:]...)
	}
}

func (tr TreeNode) run(t *testing.T, ctx Context) {

	// Descend into any possible children.
	for i, s := range tr.Steps {
		i, s := i, s
		t.Run(s, func(t *testing.T) {
			tr.Nodes[i].run(t, ctx)
		})
	}

	// Run this node's tests.
	for i := range tr.Tests {
		tc := tr.Tests[i]
		t.Run(tc.Name, func(t *testing.T) {
			// Assert the expected leak, escape, move decisions match.
			for _, lm := range tc.Matches {
				if s := lm.Regexp.FindString(ctx.BuildOutput); s == "" {
					t.Error(getBuildOutputErr(lm, s))
				}
			}

			// Assert the expected leak, escape, move decisions do not match.
			for _, lm := range tc.Natches {
				if s := lm.Regexp.FindString(ctx.BuildOutput); s != "" {
					t.Error(getBuildOutputErr(lm, s))
				}
			}

			// Find the benchmark function.
			if benchFn, ok := ctx.Benchmarks[tc.ID]; !ok {
				if ctx.Benchmarks != nil {
					t.Logf("benchmark function not registered for %s", tc.ID)
				}
			} else {
				// Assert the expected allocs and bytes match.
				r := testing.Benchmark(benchFn)
				if ea, aa := tc.AllocOp, r.AllocsPerOp(); !ea.Eq(aa) {
					t.Errorf("exp.alloc=%d, act.alloc=%d", ea, aa)
				}
				if eb, ab := tc.BytesOp, r.AllocedBytesPerOp(); !eb.Eq(ab) {
					t.Errorf("exp.bytes=%d, act.bytes=%d", eb, ab)
				}
			}
		})
	}
}

const expectedBuildOutputNotFound = `error: build optimization
reason: not found
regexp: %s
source: %s
`

const expectedBuildOutputWasFound = `error: build optimization
reason: was found
output: %s
regexp: %s
source: %s
`

func getBuildOutputErr(lm LineMatcher, found string) string {
	if found == "" {
		return fmt.Sprintf(
			expectedBuildOutputNotFound,
			lm.Regexp.String(),
			lm.Source,
		)
	}
	return fmt.Sprintf(
		expectedBuildOutputWasFound,
		found,
		lm.Regexp.String(),
		lm.Source,
	)
}
