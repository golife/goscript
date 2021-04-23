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
	test_handler(t, "-42", -42)
}

func TestSimpleExpression(t *testing.T) {
	test_handler(t, "5 + 3", 8)
	test_handler(t, "-5 + 3", -2)
	test_handler(t, "5 + 3 + 1", 9)
	test_handler(t, "-5 + 3 + 1", -1)
	test_handler(t, "-5 - 3 + 1", -7)
	test_handler(t, "-5 -- 5", 0) // 暂时没有 -- 所以可以
	test_handler(t, "2 * 3", 6)
	test_handler(t, "-2 * 3", -6)
	test_handler(t, "2 * -3", -6)
	test_handler(t, "2 * -3 * -2", 12)
	test_handler(t, "2 * (-3)", -6)
	test_handler(t, "2 * 3 * 4", 24)
	test_handler(t, "1 * 2 * 3 * 4 * 5", 120)
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
	test_handler(t, "8 * (-2+3 - 2*-3*(2+3)) - (-2 - 2 -1)", 253)
	test_handler(t, "-8 * (-2-3 - (-2)*-3*(2+3)) - (-2 - 2 -1)", 285)
}

func test_handler(t *testing.T, src string, expected int) {
	ret := exec.Exec("test", src)
	if ret != expected {
		t.Fatal(src, "Expected:", expected, "Got:", ret)
	} else {
		t.Log(src, " = ", ret)
	}
}

func TestStmts(t *testing.T) {
	src := `var a = 10
a = a + 2 + a * 3
print(a, a * 2)
print(a - 2)
`
	exec.Exec("test.gs", src)
}