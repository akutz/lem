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

package mem_test

import (
	"testing"

	"github.com/akutz/lem"
)

func TestLem(t *testing.T) {
	lem.RunWithBenchmarks(t, map[string]func(*testing.B){
		"escape1": escape1,
	})
}

// lem.escape1.alloc=2
// lem.escape1.bytes=16
func escape1(b *testing.B) {
	var sink1 interface{}
	var sink2 interface{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var x int32 = 256
		var y int64 = 256
		sink1 = x // lem.escape1.m=x escapes to heap
		sink2 = y // lem.escape1.m=y escapes to heap
	}
	_ = sink1
	_ = sink2
}
