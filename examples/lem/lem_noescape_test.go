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

// lem.noescape1.name=pointer does not outlive its call stack
// lem.noescape1.alloc=0
// lem.noescape1.bytes=0
func noescape1(b *testing.B) {
	f := func(p *int32) *int32 { // lem.noescape1.m=leaking param: p to result ~r[0-1] level=0
		return p
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f(new(int32)) // lem.noescape1.m=new\(int32\) does not escape
	}
}
func init() {
	lemFuncs["noescape1"] = noescape1
}

// lem.noescape2.name=value types do not escape
// lem.noescape2.alloc=0
// lem.noescape2.bytes=0
func noescape2(b *testing.B) {
	var sink int32
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var x int32
		x = 256
		sink = x
	}
	_ = sink
}
func init() {
	lemFuncs["noescape2"] = noescape2
}

// lem.noescape3.name=value types do not leak
// lem.noescape3.alloc=0
// lem.noescape3.bytes=0
func noescape3(b *testing.B) {
	var sink int32
	f := func(x int32) {
		sink = x
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f(0)
	}
	_ = sink
}
func init() {
	lemFuncs["noescape3"] = noescape3
}
