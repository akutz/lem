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

// Package lem provides a comment-driven approach for writing benchmarks
// that assert when values leak, escape, and move to the heap, as well as
// the number of allocations and memory as a result.
//
// Callers may define benchmark functions with the following signature:
//
//     func(*testing.Benchmark)
//
// The functions can then be annotated with comments that are parsed by the
// lem test harness, β. These comments all begin the same way:
//
//     // lem.<ID>
//
// The <ID> is a unique ID that can be any value a developer wishes it to
// be, and the value is used for two purposes:
//
// 1. The <ID> is a key in the map β.Benchmarks and points to the actual
// benchmark function for that <ID>.
//
// 2. The <ID> is used to group all associated "lem.<ID>"" comments.
//
// There are two types of comments:
//
// 1. Those placed above the function signature
//
// 2. Those placed alongside lines inside the function
//
// The first comment occurs above the function signature and defines the
// test's name, the value used as the argument for the "name" parameter in
// the "Run" function for the types "testing.T"  and "testing.B". The
// comment takes the form "lem.<ID>.name=<NAME>".
//
// The <ID> and <NAME> values are used to build the test's path string:
//
// 1. If <NAME> does not start with "/" then <ID> is added to the path,
//    followed by a "/" character.
// 2. The <NAME> value is added to the path.
//
// The path string is then split with "/" as the separator. The following
// are examples of valid forms of "lem.<ID>.name=<NAME>":
//
//     // lem.leak1.name=to sink
//     // lem.leak2.name=/to result
//     // lem.move.name=too large
//
// The above comments translate to the following strings:
//
//     leak1/to sink
//     to result
//     move/too large
//
// The next comment also occurs above the function's signature and takes
// the form "lem.<ID>.alloc=<VALUE>" or "lem.<ID>.alloc=<MIN>-<MAX>".
// This comment asserts the number of allocations expected to occur during
// the execution of the benchmark. The number may be exact or an inclusive
// range. Examples include:
//
//     // lem.leak1.alloc=1
//     // lem.leak1.alloc=2-4
//
// The first example asserts a single allocation should occur while the
// second example asserts either two, three, or four allocations are
// expected to have occurred.
//
// The next comment also occurs above the function's signature and takes
// the form "lem.<ID>.bytes=<VALUE>" or "lem.<ID>.bytes=<MIN>-<MAX>".
// This comment asserts the number of bytes expected to be allocated during
// the execution of the benchmark. For more documentation please refer
// to "lem.<ID>.alloc" as both comments have the same format rules.
//
// The next comment occurs alongside a line inside of a function, and it is
// "lem.<ID>.m=<REGEX>". This comment asserts that the Go compiler's
// optimization flag "-m" should emit some type of message for the line of
// code where the comment appears, ex.
//
//     /* line 70 */ sink = x // lem.escape3.m=x escapes to heap
//
// The above code is on line 70 of a source file named escape.go, and the
// comment asserts the output from "go build -gcflags -m" should match
// the regex "escape.go:70:\d+: x escapes to heap". Please note that
// special characters must be escaped, such as "new\(int32\) escapes to heap".
//
// The last comment is a variant of the previous and takes the form
// "lem.<ID>.m!=<REGEX>". This comment asserts a provided pattern should
// not match the compiler optimization output. This is useful when you want
// to assert a variable did not escape, leak, or move. For example:
//
//     /* line 80 */ x = new(int32) // lem.escape12.m!=(escape|leak|move)
//
// The above comment asserts none of the words "escape", "leak", or "move"
// appeared in the compiler optimization output for line 80 for the source
// file in which the comment exists.
package lem
