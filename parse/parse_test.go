// Copyright 2021 The Goscript Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package parse_test

import (
	"fmt"
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

func TestStmt(t *testing.T) {
	f := parse.ParseFile("test.gs", `var a, b int = 10+2, 12+(9-3)`)
	fmt.Println(len(f.Stmts))
	for _, item := range f.Stmts {
		ast.Print(item)
	}
}

func TestStmt2(t *testing.T) {
	f := parse.ParseFile("test.gs", `var a, b int = 10+2, 12+(9-3)
a, b := 10, 12+32*3
funcName(1, 2)
func hello(name1, name2 string, age int) {
	var c = 10
	c = 100 + 10
	print("life")
}

if a > 10 && b < 100 {
	print("ok")
}

a = 103
`)
	fmt.Println(len(f.Stmts))
	for _, item := range f.Stmts {
		ast.Print(item)
	}
}