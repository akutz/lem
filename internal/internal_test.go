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

package internal_test

import (
	"encoding/json"
	"reflect"
	"regexp"
	"testing"

	"github.com/akutz/lem/internal"
)

func TestTestCasePath(t *testing.T) {
	testCases := []struct {
		name string
		data internal.TestCase
		path []string
	}{
		{
			name: "id",
			data: internal.TestCase{
				ID:   "leak",
				Name: "to sink",
			},
			path: []string{"leak", "to sink"},
		},
		{
			name: "id & name w multiple parts",
			data: internal.TestCase{
				ID:   "leak",
				Name: "to result/malloc",
			},
			path: []string{"leak", "to result", "malloc"},
		},
		{
			name: "id & no ID in path",
			data: internal.TestCase{
				ID:   "leak",
				Name: "/to result",
			},
			path: []string{"to result"},
		},
		{
			name: "id & no ID in path & name w multiple parts",
			data: internal.TestCase{
				ID:   "escape",
				Name: "/no malloc/storing single byte-wide value",
			},
			path: []string{"no malloc", "storing single byte-wide value"},
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			if e, a := tc.path, tc.data.Path(); !reflect.DeepEqual(a, a) {
				t.Errorf("expPath=%v, actPath=%v", e, a)
			}
		})
	}
}

func TestTreeInsert(t *testing.T) {
	testCases := []struct {
		name string
		data []internal.TestCase
		tree internal.Tree
		noeq bool
	}{
		{
			name: "test case w no name",
			data: []internal.TestCase{
				{
					ID:   "a1",
					Name: "",
				},
			},
			tree: internal.Tree{
				TreeNode: internal.TreeNode{
					Index: map[string]int{},
					Tests: []internal.TestCase{
						{
							ID:   "a1",
							Name: "a1",
						},
					},
				},
			},
		},
		{
			name: "expected test case name does not match",
			noeq: true,
			data: []internal.TestCase{
				{
					ID:   "a1",
					Name: "leak",
				},
			},
			tree: internal.Tree{
				TreeNode: internal.TreeNode{
					Index: map[string]int{},
					Tests: []internal.TestCase{
						{
							ID:   "a1",
							Name: "a1",
						},
					},
				},
			},
		},
		{
			name: "test case w no name & alpha id",
			data: []internal.TestCase{
				{
					ID:   "a",
					Name: "",
				},
			},
			tree: internal.Tree{
				TreeNode: internal.TreeNode{
					Index: map[string]int{},
					Tests: []internal.TestCase{
						{
							ID:   "a",
							Name: "a",
						},
					},
				},
			},
		},
		{
			name: "tree w multiple tests in nested path",
			data: []internal.TestCase{
				{
					ID:   "a1",
					Name: "/a/1/hello",
				},
				{
					ID:   "a2",
					Name: "/a/1/world",
				},
			},
			tree: internal.Tree{
				TreeNode: internal.TreeNode{
					Index: map[string]int{"a": 0},
					Steps: []string{"a"},
					Nodes: []internal.TreeNode{
						{
							Index: map[string]int{"1": 0},
							Steps: []string{"1"},
							Nodes: []internal.TreeNode{
								{
									Tests: []internal.TestCase{
										{
											ID:   "a1",
											Name: "hello",
										},
										{
											ID:   "a2",
											Name: "world",
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "tree w multiple paths",
			data: []internal.TestCase{
				{
					ID:   "a1",
					Name: "/a/1/hello",
				},
				{
					ID:   "a2",
					Name: "/a/1/world",
				},
				{
					ID:   "a3",
					Name: "/a/2/hi",
				},
			},
			tree: internal.Tree{
				TreeNode: internal.TreeNode{
					Index: map[string]int{"a": 0},
					Steps: []string{"a"},
					Nodes: []internal.TreeNode{
						{
							Index: map[string]int{"1": 0, "2": 1},
							Steps: []string{"1", "2"},
							Nodes: []internal.TreeNode{
								{
									Tests: []internal.TestCase{
										{
											ID:   "a1",
											Name: "hello",
										},
										{
											ID:   "a2",
											Name: "world",
										},
									},
								},
								{
									Tests: []internal.TestCase{
										{
											ID:   "a3",
											Name: "hi",
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "test case w one match",
			data: []internal.TestCase{
				{
					ID:   "a",
					Name: "1",
					Matches: []internal.LineMatcher{
						{
							Regexp: regexp.MustCompile("(?s)hello"),
							Source: "hello",
						},
					},
				},
			},
			tree: internal.Tree{
				TreeNode: internal.TreeNode{
					Index: map[string]int{"a": 0},
					Steps: []string{"a"},
					Nodes: []internal.TreeNode{
						{
							Tests: []internal.TestCase{
								{
									ID:   "a",
									Name: "1",
									Matches: []internal.LineMatcher{
										{
											Regexp: regexp.MustCompile("(?s)hello"),
											Source: "hello",
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "test case w matches",
			data: []internal.TestCase{
				{
					ID:   "a",
					Name: "1",
					Matches: []internal.LineMatcher{
						{
							Regexp: regexp.MustCompile("(?s)hello"),
							Source: "hello",
						},
						{
							Regexp: regexp.MustCompile("(?s)world"),
							Source: "world",
						},
					},
				},
			},
			tree: internal.Tree{
				TreeNode: internal.TreeNode{
					Index: map[string]int{"a": 0},
					Steps: []string{"a"},
					Nodes: []internal.TreeNode{
						{
							Tests: []internal.TestCase{
								{
									ID:   "a",
									Name: "1",
									Matches: []internal.LineMatcher{
										{
											Regexp: regexp.MustCompile("(?s)hello"),
											Source: "hello",
										},
										{
											Regexp: regexp.MustCompile("(?s)world"),
											Source: "world",
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "test case w unexpected matches",
			noeq: true,
			data: []internal.TestCase{
				{
					ID:   "a",
					Name: "1",
					Matches: []internal.LineMatcher{
						{
							Regexp: regexp.MustCompile("(?m)hello"),
							Source: "hello",
						},
						{
							Regexp: regexp.MustCompile("(?m)world"),
							Source: "world",
						},
					},
				},
			},
			tree: internal.Tree{
				TreeNode: internal.TreeNode{
					Index: map[string]int{"a": 0},
					Steps: []string{"a"},
					Nodes: []internal.TreeNode{
						{
							Tests: []internal.TestCase{
								{
									ID:   "a",
									Name: "1",
									Matches: []internal.LineMatcher{
										{
											Regexp: regexp.MustCompile("(?m)hello"),
											Source: "hello",
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "test case w unexpected match options",
			noeq: true,
			data: []internal.TestCase{
				{
					ID:   "a",
					Name: "1",
					Matches: []internal.LineMatcher{
						{
							Regexp: regexp.MustCompile("(?m)hello"),
							Source: "hello",
						},
					},
				},
			},
			tree: internal.Tree{
				TreeNode: internal.TreeNode{
					Index: map[string]int{"a": 0},
					Steps: []string{"a"},
					Nodes: []internal.TreeNode{
						{
							Tests: []internal.TestCase{
								{
									ID:   "a",
									Name: "1",
									Matches: []internal.LineMatcher{
										{
											Regexp: regexp.MustCompile("(?s)hello"),
											Source: "hello",
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "test case w one natch",
			data: []internal.TestCase{
				{
					ID:   "a",
					Name: "1",
					Natches: []internal.LineMatcher{
						{
							Regexp: regexp.MustCompile("(?s)hello"),
							Source: "hello",
						},
					},
				},
			},
			tree: internal.Tree{
				TreeNode: internal.TreeNode{
					Index: map[string]int{"a": 0},
					Steps: []string{"a"},
					Nodes: []internal.TreeNode{
						{
							Tests: []internal.TestCase{
								{
									ID:   "a",
									Name: "1",
									Natches: []internal.LineMatcher{
										{
											Regexp: regexp.MustCompile("(?s)hello"),
											Source: "hello",
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "test case w natches",
			data: []internal.TestCase{
				{
					ID:   "a",
					Name: "1",
					Natches: []internal.LineMatcher{
						{
							Regexp: regexp.MustCompile("(?s)hello"),
							Source: "hello",
						},
						{
							Regexp: regexp.MustCompile("(?s)world"),
							Source: "world",
						},
					},
				},
			},
			tree: internal.Tree{
				TreeNode: internal.TreeNode{
					Index: map[string]int{"a": 0},
					Steps: []string{"a"},
					Nodes: []internal.TreeNode{
						{
							Tests: []internal.TestCase{
								{
									ID:   "a",
									Name: "1",
									Natches: []internal.LineMatcher{
										{
											Regexp: regexp.MustCompile("(?s)hello"),
											Source: "hello",
										},
										{
											Regexp: regexp.MustCompile("(?s)world"),
											Source: "world",
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "test case w unexpected natches",
			noeq: true,
			data: []internal.TestCase{
				{
					ID:   "a",
					Name: "1",
					Natches: []internal.LineMatcher{
						{
							Regexp: regexp.MustCompile("(?m)hello"),
							Source: "hello",
						},
						{
							Regexp: regexp.MustCompile("(?m)world"),
							Source: "world",
						},
					},
				},
			},
			tree: internal.Tree{
				TreeNode: internal.TreeNode{
					Index: map[string]int{"a": 0},
					Steps: []string{"a"},
					Nodes: []internal.TreeNode{
						{
							Tests: []internal.TestCase{
								{
									ID:   "a",
									Name: "1",
									Natches: []internal.LineMatcher{
										{
											Regexp: regexp.MustCompile("(?m)hello"),
											Source: "hello",
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "test case w unexpected natch options",
			noeq: true,
			data: []internal.TestCase{
				{
					ID:   "a",
					Name: "1",
					Natches: []internal.LineMatcher{
						{
							Regexp: regexp.MustCompile("(?m)hello"),
							Source: "hello",
						},
					},
				},
			},
			tree: internal.Tree{
				TreeNode: internal.TreeNode{
					Index: map[string]int{"a": 0},
					Steps: []string{"a"},
					Nodes: []internal.TreeNode{
						{
							Tests: []internal.TestCase{
								{
									ID:   "a",
									Name: "1",
									Natches: []internal.LineMatcher{
										{
											Regexp: regexp.MustCompile("(?s)hello"),
											Source: "hello",
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "tree w heap assertions",
			data: []internal.TestCase{
				{
					ID:      "a1",
					Name:    "/a/1/hello",
					AllocOp: internal.Int64Range{Min: 1, Max: 2},
					BytesOp: internal.Int64Range{Min: 3, Max: 4},
				},
				{
					ID:      "a2",
					Name:    "/a/1/world",
					AllocOp: internal.Int64Range{Min: 5, Max: 6},
					BytesOp: internal.Int64Range{Min: 7, Max: 8},
				},
				{
					ID:      "a3",
					Name:    "/a/2/hi",
					AllocOp: internal.Int64Range{Min: 9, Max: 10},
					BytesOp: internal.Int64Range{Min: 11, Max: 12},
				},
			},
			tree: internal.Tree{
				TreeNode: internal.TreeNode{
					Index: map[string]int{"a": 0},
					Steps: []string{"a"},
					Nodes: []internal.TreeNode{
						{
							Index: map[string]int{"1": 0, "2": 1},
							Steps: []string{"1", "2"},
							Nodes: []internal.TreeNode{
								{
									Tests: []internal.TestCase{
										{
											ID:      "a1",
											Name:    "hello",
											AllocOp: internal.Int64Range{Min: 1, Max: 2},
											BytesOp: internal.Int64Range{Min: 3, Max: 4},
										},
										{
											ID:      "a2",
											Name:    "world",
											AllocOp: internal.Int64Range{Min: 5, Max: 6},
											BytesOp: internal.Int64Range{Min: 7, Max: 8},
										},
									},
								},
								{
									Tests: []internal.TestCase{
										{
											ID:      "a3",
											Name:    "hi",
											AllocOp: internal.Int64Range{Min: 9, Max: 10},
											BytesOp: internal.Int64Range{Min: 11, Max: 12},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "tree w unexpected heap assertions",
			noeq: true,
			data: []internal.TestCase{
				{
					ID:      "a1",
					Name:    "/a/1/hello",
					AllocOp: internal.Int64Range{Min: 0, Max: 2},
					BytesOp: internal.Int64Range{Min: 3, Max: 4},
				},
				{
					ID:      "a2",
					Name:    "/a/1/world",
					AllocOp: internal.Int64Range{Min: 5, Max: 6},
					BytesOp: internal.Int64Range{Min: 7, Max: 8},
				},
				{
					ID:      "a3",
					Name:    "/a/2/hi",
					AllocOp: internal.Int64Range{Min: 9, Max: 10},
					BytesOp: internal.Int64Range{Min: 11, Max: 12},
				},
			},
			tree: internal.Tree{
				TreeNode: internal.TreeNode{
					Index: map[string]int{"a": 0},
					Steps: []string{"a"},
					Nodes: []internal.TreeNode{
						{
							Index: map[string]int{"1": 0, "2": 1},
							Steps: []string{"1", "2"},
							Nodes: []internal.TreeNode{
								{
									Tests: []internal.TestCase{
										{
											ID:      "a1",
											Name:    "hello",
											AllocOp: internal.Int64Range{Min: 1, Max: 2},
											BytesOp: internal.Int64Range{Min: 3, Max: 4},
										},
										{
											ID:      "a2",
											Name:    "world",
											AllocOp: internal.Int64Range{Min: 5, Max: 6},
											BytesOp: internal.Int64Range{Min: 7, Max: 8},
										},
									},
								},
								{
									Tests: []internal.TestCase{
										{
											ID:      "a3",
											Name:    "hi",
											AllocOp: internal.Int64Range{Min: 9, Max: 10},
											BytesOp: internal.Int64Range{Min: 11, Max: 12},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			e, a := tc.tree, internal.NewTree(tc.data...)
			printTrees := func(msg string) {
				eb, err := json.MarshalIndent(e.TreeNode, "", "  ")
				if err != nil {
					t.Fatal(err)
				}
				ab, err := json.MarshalIndent(a.TreeNode, "", "  ")
				if err != nil {
					t.Fatal(err)
				}
				t.Errorf("%s\nexpTree=%s\nactTree=%s",
					msg, string(eb), string(ab))
			}
			if eq := e.DeepEqual(a); eq && tc.noeq {
				printTrees("trees are unexpectedly equal")
			} else if !eq && !tc.noeq {
				printTrees("trees are not equal")
			}
		})
	}
}
