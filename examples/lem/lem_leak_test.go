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

import "testing"

// lem.leak1.name=to sink
// lem.leak1.alloc=0
// lem.leak1.bytes=0
func leak1(b *testing.B) {
	var sink *int32
	f := func(p *int32) { // lem.leak1.m=leaking param: p
		sink = p
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f(nil)
	}
	_ = sink
}
func init() {
	lemFuncs["leak1"] = leak1
}

// lem.leak2.name=pointer to result
// lem.leak2.alloc=0
// lem.leak2.bytes=0
func leak2(b *testing.B) {
	f := func(p *int32) *int32 { // lem.leak2.m=leaking param: p to result ~r[0-1] level=0
		return p
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f(nil)
	}
}
func init() {
	lemFuncs["leak2"] = leak2
}

// lem.leak3.name=map to result
// lem.leak3.alloc=0
// lem.leak3.bytes=0
func leak3(b *testing.B) {
	var sink map[string]struct{}
	f := func(m map[string]struct{}) { // lem.leak3.m=leaking param: m
		sink = m
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f(nil)
	}
	_ = sink
}
func init() {
	lemFuncs["leak3"] = leak3
}

// lem.leak4.name=slice to result
// lem.leak4.alloc=0
// lem.leak4.bytes=0
func leak4(b *testing.B) {
	var sink []int32
	f := func(s []int32) { // lem.leak4.m=leaking param: s
		sink = s
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f(nil)
	}
	_ = sink
}
func init() {
	lemFuncs["leak4"] = leak4
}

// lem.leak5.name=chan to result
// lem.leak5.alloc=0
// lem.leak5.bytes=0
func leak5(b *testing.B) {
	var sink chan int32
	f := func(c chan int32) { // lem.leak5.m=leaking param: c
		sink = c
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f(nil)
	}
	_ = sink
}
func init() {
	lemFuncs["leak5"] = leak5
}

// lem.leak6.name=five ref params to result w zero mallocs
// lem.leak6.alloc=0
// lem.leak6.bytes=0
func leak6(b *testing.B) {
	type s struct{ a, b int32 }
	noop := func(
		a *int32, // lem.leak6.m=leaking param: a to result ~r[06] level=0
		b *int64, // lem.leak6.m=leaking param: b to result ~r[17] level=0
		c []int32, // lem.leak6.m=leaking param: c to result ~r[28] level=0
		d []*int64, // lem.leak6.m=leaking param: d to result ~r[39] level=0
		e s, // lem.leak6.m!=(leak|escape|move)
		f *s, // lem.leak6.m=leaking param: f to result ~r(5|11) level=0
	) (*int32, *int64, []int32, []*int64, s, *s) {
		return a, b, c, d, e, f
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var (
			a = new(int32)             // lem.leak6.m=new\(int32\) does not escape
			b = new(int64)             // lem.leak6.m=new\(int64\) does not escape
			c = make([]int32, 5, 5)    // lem.leak6.m=make\(\[\]int32, 5, 5\) does not escape
			d = make([]*int64, 10, 10) // lem.leak6.m=make\(\[\]\*int64, 10, 10\) does not escape
			e = s{a: 4096, b: 4096}    // lem.leak6.m!=(leak|escape|move)
			f = new(s)                 // lem.leak6.m=new\(s\) does not escape
		)
		noop(a, b, c, d, e, f)
	}
}
func init() {
	lemFuncs["leak6"] = leak6
}

// lem.leak7.name=did not escape bc return value did not outlive stack frame
// lem.leak7.alloc=0
// lem.leak7.bytes=0
func leak7(b *testing.B) {
	f := func(p *int32) *int32 { // lem.leak7.m=leaking param: p to result ~r[0-1] level=0
		return p
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var x int32 = 4096 // lem.leak7.m!=(leak|escape|move)
		var p *int32 = &x  // lem.leak7.m!=(leak|escape|move)
		var sink *int32    // lem.leak7.m!=(leak|escape|move)
		sink = f(p)
		_ = sink
	}
}
func init() {
	lemFuncs["leak7"] = leak7
}
