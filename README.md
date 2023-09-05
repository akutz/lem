# Leak, escape, move (lem)

Lem is a bespoke, Golang test framework for asserting expected [escape analysis](#escape-analysis) results and heap allocations.

* [**Overview**](#overview): an overview of lem
* [**Directives**](#directives): the comments used to configure lem
* [**Benchmarks**](#benchmarks): using benchmark functions to assert heap behavior
* [**Examples**](#examples): common use cases in action
* [**Appendix**](#appendix): helpful information germane to lem


## Overview

This project allows developers to statically assert leak, escape, move and heap memory characteristics about values throughout a project's source code. For instance, please consider the following example ([./examples/hello/world.go](./examples/hello/world.go)):

```go
package hello

func World() *string {
	s := "Hello, world."
	return &s
}
```

If compiled with build optimization output enabled (the compiler flag `-m`), a developer would see something similar to the following:

```bash
$ go tool compile -m -l ./examples/hello/world.go
./examples/hello/world.go:20:2: moved to heap: s
```

Go's escape analysis determined that the value in variable `s` should be moved to the heap. The lem framework provides a simple means to assert the expected escape analysis result by adding a trailing comment like so:

```go
	s := "Hello, world." // lem.World.m=moved to heap: s
```

The comment `lem.World.m=moved to heap: s` informs the lem framework that escape analysis should emit the message `moved to heap: s` for the line where the comment is defined. Having lem act on this assertion is as simple as introducing a test file ([./examples/hello/world_test.go](./examples/hello/world_test.go)):

```go
package hello

import (
	"testing"

	"github.com/akutz/lem"
)

func TestHello(t *testing.T) {
	lem.Run(t)
}
```

Now let's run the test and see what happens:

```bash
$ go test -v ./examples/hello
=== RUN   TestHello
=== RUN   TestHello/World
--- PASS: TestHello (0.64s)
    --- PASS: TestHello/World (0.00s)
PASS
ok  	github.com/akutz/lem/examples/hello	0.889s
```

Okay, maybe it did not actually work and the lack of an error is just evidence that the framework is buggy? To verify it _did_ work, let's make this change in [./examples/hello/world.go](./examples/hello/world.go):

```go
	s := "Hello, world." // lem.World.m=escapes to heap: s
```

Run the test again:

```bash
$ go test -v ./examples/hello
=== RUN   TestHello
=== RUN   TestHello/World
    tree.go:147: exp.m=(?m)^.*world.go:20:\d+: escapes to heap: s$
--- FAIL: TestHello (0.58s)
    --- FAIL: TestHello/World (0.00s)
FAIL
FAIL	github.com/akutz/lem/examples/hello	0.694s
FAIL
```

So clearly the assertions _are_ working. Pretty cool, right? Keep reading, there is quite a bit more :smile:


## Directives

The term _directive_ refers to the comments used to configure lem. Please note the following about the table below:

* All directives are optional.
* The _Positional_ column indicates the location of a directive in the source code matters:
  * Non-positional directives may be placed anywhere in source code
  * Positional directives are line-number specific
* Directives with the same `<ID>` value are considered part of the same test case.
* The _Multiple_ column indicates whether a given directive may occur multiple times for the same `<ID>`.
* The directives for expected allocs and bytes are ignored unless lem is provided a benchmark function for a given `<ID>`.


| Name | Pattern | Positional | Multiple | Description |
|---|---------|:---:|:---:|-------------|
| [Name](#name) | `^// lem\.(?P<ID>[^.]+)\.name=(?P<NAME>.+)$` |  |  | The test case name. If omitted the `<ID>` is used as the name. |
| [Expected allocs](#expected-allocs) | `^// lem\.(?P<ID>[^.]+)\.alloc=(?P<MIN>\d+)(?:-(?P<MAX>\d+))?$` |  |  | Number of expected allocations. |
| [Expected bytes](#expected-bytes) | `^// lem\.(?P<ID>[^.]+)\.bytes=(?P<MIN>\d+)(?:-(?P<MAX>\d+))?$` |  |  | Number of expected, allocated bytes. |
| [Match](#match) | `^// lem\.(?P<ID>[^.]+)\.m=(?P<MATCH>.+)$` | ✓ | ✓ | A regex pattern that must appear in the build optimization output. |
| [Natch](#natch) | `^// lem\.(?P<ID>[^.]+)\.m!=(?P<NATCH>.+)$` | ✓ | ✓ | A regex pattern that must _**not**_ appear in the build optimization output. |


### Name

The name directive is used to provide a more verbose description for the test than just the `<ID>`. Please consider the following example ([./examples/name/name_test.go](./examples/name/name_test.go)):

```go
package name_test

import (
	"testing"

	"github.com/akutz/lem"
)

func TestLem(t *testing.T) {
	lem.Run(t)
}

var sink *int32

func leak1(p *int32) *int32 { // lem.leak1.m=leaking param: p to result ~r[0-1] level=0
	return p
}

// lem.leak2.name=to sink
func leak2(p *int32) *int32 { // lem.leak2.m=leaking param: p
	sink = p
	return p
}
```

The `leak1` test case does _not_ have a name directive, but `leak2` _does_. Lem uses the name for `leak2` when running the tests:

```bash
$ go test -v ./examples/name
=== RUN   TestLem
=== RUN   TestLem/leak2
=== RUN   TestLem/leak2/to_sink
=== RUN   TestLem/leak1
--- PASS: TestLem (0.43s)
    --- PASS: TestLem/leak2 (0.00s)
        --- PASS: TestLem/leak2/to_sink (0.00s)
    --- PASS: TestLem/leak1 (0.00s)
PASS
ok  	github.com/akutz/lem/examples/name	0.672s
```

A name directive can make it easier to find a test in the lem output.


### Expected allocs

This directive asserts the number of allocations expected to occur for a specific test case. The directive may be specified as an exact value:

```go
// lem.move1.alloc=1
```

or as an inclusive range:

```go
// lem.move2.alloc=1-2
```

Please note this directive has no effect unless a [benchmark](#benchmarks) function is provided for the test case.


### Expected bytes

This directive asserts the number of allocated bytes expected to occur for a specific test case. The directive may be specified as an exact value:

```go
// lem.move1.bytes=8
```

or as an inclusive range:

```go
// lem.move2.bytes=8-12
```

Please note this directive has no effect unless a [benchmark](#benchmarks) function is provided for the test case.


### Match

The match directive may occur multiple times for a single test case and is used to assert that a specific pattern must be present in the build optimization output for the line on which the directive is defined. For example ([./examples/match/match_test.go](./examples/match/match_test.go)):

```go
package match_test

import (
	"testing"

	"github.com/akutz/lem"
)

func TestLem(t *testing.T) {
	lem.Run(t)
}

var sink interface{}

func put(x, y int32) {
	sink = x // lem.put.m=x escapes to heap
	sink = y // lem.put.m=y escapes to heap
}
```

Not only is there no issue with multiple match directives for a single test case, it is likely there _will be_ multiple match directives for a single test case.


### Natch

The inverse of the match directive -- the specified patterns _**cannot**_ occur in the build optimization output. For example ([./examples/natch/natch_test.go](./examples/natch/natch_test.go)):

```go
package natch_test

import (
	"testing"

	"github.com/akutz/lem"
)

func TestLem(t *testing.T) {
	lem.Run(t)
}

var sink int32

func put(x, y int32) {
	sink = x // lem.put.m!=(leak|escape|move)
	sink = y // lem.put.m!=(leak|escape|move)
}
```

And just like the match directive, multiple natch directives are allowed.


## Benchmarks

In order to assert an expected number of allocations or bytes, a benchmark must be provided to lem ([./examples/mem/mem_test.go](./examples/mem/mem_test.go)):

```golang
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
```

Running the above test does not produce anything too spectacular output-wise:

```bash
$ go test -v ./examples/mem
=== RUN   TestLem
=== RUN   TestLem/escape1
--- PASS: TestLem (1.72s)
    --- PASS: TestLem/escape1 (1.06s)
PASS
ok  	github.com/akutz/lem/examples/mem	1.908s
```

However, internally lem runs the provided benchmark in order to compare the result to the expected number of allocations and bytes allocated.

---

:wave: _**16 bytes?!**_

Some of you may have noticed that the example asserts 16 bytes are allocated. Except the size of an `int32` is 4 bytes, and the size of an `int64` is 8? Why is the number of allocated bytes not 12? In fact, on a 32-bit system it _would_ be 12 bytes.

Go aligns memory based on the platform -- 4 bytes on 32-bit platforms, 8 bytes on 64-bit systems. That means an `int32` (4 bytes) + an `int64` (8 bytes) will reserve 16 bytes on the stack as it is the next alignment after the 12 bytes the two types actually require.

Go's own documentation does not delve too deeply into [alignment guarantees](https://go.dev/ref/spec#Size_and_alignment_guarantees), but there _is_ a great article avaialble on [Go memory layouts](https://go101.org/article/memory-layout.html).

---


## Examples

There are several examples in this repository to help you get started:

* [**gcflags**](./examples/gcflags/): how to specify custom compiler flags when running lem
* [**hello**](./examples/hello): the "Hello, world." example
* [**lem**](./examples/lem): wide coverage for escape analysis and heap behavior
* [**match**](./examples/match): the example for the [match](#match) directive
* [**mem**](./examples/mem): the example for the [benchmarks](#benchmarks) section
* [**name**](./examples/name): the example for the [name](#name) directive
* [**natch**](./examples/natch): the example for the [natch](#natch) directive


## Appendix

* [**Escape analysis**](#escape-analysis): a brief overview of escape analysis


### Escape analysis

Escape analysis is a process the Go compiler uses to determine which values can be placed on the stack versus the heap. A few key points include:

* static analysis that runs against the source's abstract syntax tree (AST)
* does not always tell the whole story with respect to heap allocations
* may be enabled with the compiler flag `-m`, ex. `go build -gcflags "-m"`

For a full, in-depth look at escape analysis, please refer to [this documentation](https://github.com/akutz/go-interface-values/tree/main/docs/03-escape-analysis).
