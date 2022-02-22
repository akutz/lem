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

// lem.escape1.name=pointer outlived its call stack
// lem.escape1.alloc=1
// lem.escape1.bytes=4
func escape1(b *testing.B) {
	var sink *int32
	f1 := func() *int32 {
		return new(int32) // lem.escape1.m=new\(int32\) escapes to heap
	}
	f2 := func(p *int32) *int32 { // lem.escape1.m=leaking param: p to result ~r[0-1] level=0
		return p
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink = (f2(f1())) // lem.escape1.m=new\(int32\) escapes to heap
	}
	_ = sink
}
func init() {
	lemFuncs["escape1"] = escape1
}

// lem.escape2.name=heap cannot point to stack
// lem.escape2.alloc=1
// lem.escape2.bytes=4
func escape2(b *testing.B) {
	var sink *int32
	f := func(p *int32) *int32 { // lem.escape1.m=leaking param: p to result ~r[0-1] level=0
		return p
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink = f(new(int32)) // lem.escape1.m=new\(int32\) escapes to heap
	}
	_ = sink
}
func init() {
	lemFuncs["escape2"] = escape2
}

// lem.escape3.name=no malloc bc iface & zero value
// lem.escape3.alloc=0
// lem.escape3.bytes=0
func escape3(b *testing.B) {
	var sink interface{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var x int32
		sink = x // lem.escape3.m=x escapes to heap
	}
	_ = sink
}
func init() {
	lemFuncs["escape3"] = escape3
}

// lem.escape4.name=no malloc bc storing byte value in iface
// lem.escape4.alloc=0
// lem.escape4.bytes=0
func escape4(b *testing.B) {
	var sink interface{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var x byte
		x = 253
		sink = x // lem.escape4.m=x escapes to heap
	}
	_ = sink
}
func init() {
	lemFuncs["escape4"] = escape4
}

// lem.escape5.name=no malloc bc storing single byte-wide value in iface
// lem.escape5.alloc=0
// lem.escape5.bytes=0
func escape5(b *testing.B) {
	var sink interface{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var x int64
		x = 253
		sink = x // lem.escape5.m=x escapes to heap
	}
	_ = sink
}
func init() {
	lemFuncs["escape5"] = escape5
}

// lem.escape6.name=no malloc bc storing struct w single byte-wide value in iface
// lem.escape6.alloc=0
// lem.escape6.bytes=0
func escape6(b *testing.B) {
	var sink interface{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var s struct{ a int32 }
		s.a = 253
		sink = s // lem.escape6.m=s escapes to heap
	}
	_ = sink
}
func init() {
	lemFuncs["escape6"] = escape6
}

// lem.escape7.name=malloc bc storing struct w field value >255 in iface
// lem.escape7.alloc=1
// lem.escape7.bytes=4
func escape7(b *testing.B) {
	var sink interface{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var s struct{ a int32 }
		s.a = 256
		sink = s // lem.escape7.m=s escapes to heap
	}
	_ = sink
}
func init() {
	lemFuncs["escape7"] = escape7
}

// lem.escape8.name=malloc bc storing values >255 in iface
// lem.escape8.alloc=2
// lem.escape8.bytes=16
func escape8(b *testing.B) {
	var sink interface{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var x int16
		var y int64
		x = 256
		y = 256
		sink = x // lem.escape8.m=x escapes to heap
		sink = y // lem.escape8.m=y escapes to heap
	}
	_ = sink
}
func init() {
	lemFuncs["escape8"] = escape8
}

// lem.escape9.name=malloc bc storing values >255 in iface; 16 bytes bc alignment
// lem.escape9.alloc=2
// lem.escape9.bytes=16
func escape9(b *testing.B) {
	var sink interface{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var x int32
		var y int64
		x = 256
		y = 256
		sink = x // lem.escape9.m=x escapes to heap
		sink = y // lem.escape9.m=y escapes to heap
	}
	_ = sink
}
func init() {
	lemFuncs["escape9"] = escape9
}

// lem.escape10.name=malloc bc storing value >255 in iface
// lem.escape10.alloc=1
// lem.escape10.bytes=8
func escape10(b *testing.B) {
	var sink interface{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var x int32
		var y int64
		x = 255
		y = 256
		sink = x // lem.escape10.m=x escapes to heap
		sink = y // lem.escape10.m=y escapes to heap
	}
	_ = sink
}
func init() {
	lemFuncs["escape10"] = escape10
}

// lem.escape11.name=malloc bc zero value trick only works when storing value in iface
// lem.escape11.alloc=1
// lem.escape11.bytes=8
func escape11(b *testing.B) {
	var sink *int64
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x := new(int64) // lem.escape11.m=new\(int64\) escapes to heap
		*x = 0
		sink = x
	}
	_ = sink
}
func init() {
	lemFuncs["escape11"] = escape11
}

// lem.escape12.name=five ref params to result w one malloc for stored return value
// lem.escape12.alloc=1
// lem.escape12.bytes=4
func escape12(b *testing.B) {
	type s struct{ a, b int32 }
	noop := func(
		a *int32, // lem.escape12.m=leaking param: a to result ~r[06] level=0
		b *int64, // lem.escape12.m=leaking param: b to result ~r[17] level=0
		c []int32, // lem.escape12.m=leaking param: c to result ~r[28] level=0
		d []*int64, // lem.escape12.m=leaking param: d to result ~r[39] level=0
		e s, // lem.escape12.m!=(leak|escape|move)
		f *s, // lem.escape12.m=leaking param: f to result ~r(5|11) level=0
	) (*int32, *int64, []int32, []*int64, s, *s) {
		return a, b, c, d, e, f
	}
	var sink *int32
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var (
			a = new(int32)             // lem.escape12.m=new\(int32\) escapes to heap
			b = new(int64)             // lem.escape12.m=new\(int64\) does not escape
			c = make([]int32, 5, 5)    // lem.escape12.m=make\(\[\]int32, 5, 5\) does not escape
			d = make([]*int64, 10, 10) // lem.escape12.m=make\(\[\]\*int64, 10, 10\) does not escape
			e = s{a: 4096, b: 4096}    // lem.escape12.m!=(leak|escape|move)
			f = new(s)                 // lem.escape12.m=new\(s\) does not escape
		)
		sink, _, _, _, _, _ = noop(a, b, c, d, e, f)
	}
	_ = sink
}
func init() {
	lemFuncs["escape12"] = escape12
}

// lem.escape13.name=return value was captured & outlived stack frame
// lem.escape13.alloc=1
// lem.escape13.bytes=4
func escape13(b *testing.B) {
	f := func(p *int32) *int32 { // lem.escape13.m=leaking param: p to result ~r[0-1] level=0
		return p
	}
	var sink *int32
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x := new(int32) // lem.escape13.m=new\(int32\) escapes to heap
		*x = 4096
		if sink = f(x); sink == nil {
			// Do nothing
		}
		_ = x
	}
	_ = sink
}
func init() {
	lemFuncs["escape13"] = escape13
}

// lem.escape14.name=malloc w storing <256 in interface bc float32
// lem.escape14.alloc=1
// lem.escape14.bytes=4
func escape14(b *testing.B) {
	var sink interface{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var x float32
		x = 253
		sink = x // lem.escape14.m=x escapes to heap
	}
	_ = sink
}
func init() {
	lemFuncs["escape14"] = escape14
}

// lem.escape15.name=malloc w storing <256 in interface bc float64
// lem.escape15.alloc=1
// lem.escape15.bytes=8
func escape15(b *testing.B) {
	var sink interface{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var x float64
		x = 253
		sink = x // lem.escape15.m=x escapes to heap
	}
	_ = sink
}
func init() {
	lemFuncs["escape15"] = escape15
}
