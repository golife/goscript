// Copyright 2021 The Goscript Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package exec_test

import (
	"github.com/golife/goscript/exec"
	"testing"
)

func TestInteger(t *testing.T) {
	test_handler(t, "42", 42)
}

func TestSimpleExpression(t *testing.T) {
	test_handler(t, "5 + 3", 8)
	test_handler(t, "5 + 3 + 1", 9)
	test_handler(t, "2 * 3", 6)
	test_handler(t, "2 * 3 * 4", 24)
}

func TestSimpleExpressionWithComments(t *testing.T) {
	test_handler(t, ";comment 1\n 5 * 3; comment 2", 15)
}

func TestComplexExpression(t *testing.T) {
	test_handler(t, "8 - (2+3) * 2", -2)
	test_handler(t, "8 - (2+3) * 2 - 1", -3)
	test_handler(t, "8 - (2+3) * (2 - 2)", 8)
	test_handler(t, "8 * (2+3 - 1*3*2) - (2 + 2 -1)", -11)
	test_handler(t, "8 * (2+3 - 2*3*(2+3)) - (2 + 2 -1)", -203)
}

func test_handler(t *testing.T, src string, expected int) {
	ret := exec.Exec("test", src)
	if ret != expected {
		t.Fatal(src, "Expected:", expected, "Got:", ret)
	} else {
		t.Log(src, " = ", ret)
	}
}
