// Copyright 2021 The Goscript Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package parse_test

import (
	"testing"

	"github.com/golife/goscript/ast"
	"github.com/golife/goscript/parse"
)

const (
	FILE = iota
	BASIC
	BINARY
)

func TestParseFileBasic(t *testing.T) {
	f := parse.ParseFile("test.gs", "1+(3+4)*3")
	i := 0

	ast.Print(f.Root)

	var types = []int{FILE, BINARY, BASIC, BINARY, BINARY, BASIC, BASIC, BASIC}

	ast.Walk(f, 0, func(node ast.Node, level int) {
		switch node.(type) {
		case *ast.File:
			if types[i] != FILE {
				t.Fatal("Walk index:", i, "Expected:", types[i], "Got:", FILE)
			}
		case *ast.BasicLit:
			if types[i] != BASIC {
				t.Fatal("Walk index:", i, "Expected:", types[i], "Got:", BASIC)
			}
		case *ast.BinaryExpr:
			if types[i] != BINARY {
				t.Fatal("Walk index:", i, "Expected:", types[i], "Got:", BINARY)
			}
		}
		i++
	})
}

func TestParseNested(t *testing.T) {
	f := parse.ParseFile("test.gs", ";comment;\n 1+(3 + 4 + 3*(2-3) )*3")
	i := 0

	ast.Print(f.Root)

	var types = []int{FILE, BINARY, BASIC, BINARY, BINARY, BINARY, BASIC, BASIC, BINARY, BASIC, BINARY, BASIC, BASIC, BASIC}

	ast.Walk(f, 0, func(node ast.Node, level int) {
		switch node.(type) {
		case *ast.File:
			if types[i] != FILE {
				t.Fatal("Walk index:", i, "Expected:", types[i], "Got:", FILE)
			}
		case *ast.BasicLit:
			if types[i] != BASIC {
				t.Fatal("Walk index:", i, "Expected:", types[i], "Got:", BASIC)
			}
		case *ast.BinaryExpr:
			if types[i] != BINARY {
				t.Fatal("Walk index:", i, "Expected:", types[i], "Got:", BINARY)
			}
		}
		i++
	})
}

func TestInvalid(t *testing.T) {
	var tests = []string{
		"+",
		"+1",
		"(6",
		"(6+2",
		"(6+2-",
		"((6+2-",
		"6)",
		"6+2)",
		"6+2-)",
		"(6+2-))",
		"d",
		";comment",
	}
	for _, src := range tests {
		if f := parse.ParseFile("test.gs", src); f != nil {
			t.Log(src, "- not nil")
			t.Fail()
		}
	}
}
