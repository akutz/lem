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

package lem_test

import (
	"testing"
)

// lem.move1.name=pointing to stack from heap
// lem.move1.alloc=1
// lem.move1.bytes=4
func move1(b *testing.B) {
	var sink *int32
	for i := 0; i < b.N; i++ {
		var x int32 = 4096 // lem.move1.m=moved to heap: x
		sink = &x
	}
	_ = sink
}
func init() {
	lemFuncs["move1"] = move1
}

// lem.move2.name=too large for stack
// lem.move2.alloc=1
// lem.move2.bytes=15728640-15728700
func move2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var buf [15 * 1024 * 1024]byte // lem.move2.m=moved to heap: buf
		_ = buf
	}
}
func init() {
	lemFuncs["move2"] = move2
}
