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
