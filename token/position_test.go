package token_test

import (
	"fmt"
	"github.com/golife/goscript/scan"
	"github.com/golife/goscript/token"
	"testing"
)

func TestFilePosition(t *testing.T) {
	f := token.NewFile("", "(+ 1 2)")
	p := f.Position(f.Pos(0))
	if p.String() != "1:1" {
		t.Fatal("Nameless file: Expected: 1:1, Got:", p.String())
	}

	f = token.NewFile("test.gs", "(+ 1 2)")
	p = f.Position(f.Pos(1))
	if p.String() != "test.gs:1:2" {
		t.Fatal("Nameless file: Expected: test.gs:1:2, Got:", p.String())
	}

	var test_expr = "(+ 2 3)\n(- 5 4)"
	f = token.NewFile("test", test_expr)
	var tests = []struct {
		row, col int
		pos      token.Pos
	}{
		{1, 1, f.Pos(0)},
		{2, 1, f.Pos(8)},
	}

	f.AddLine(8)
	for _, v := range tests {
		p := f.Position(v.pos)
		if p.Column != v.col || p.Line != v.row {
			t.Fatal("For: pos", v.pos, "Expected:", v.row, ":", v.col, ", Got:",
				p.Line, ":", p.Column)
		}
	}
}

var tests = []struct {
	filename string
	source   []byte // may be nil
	lines    []int
}{
	{"a", []byte{}, []int{}},
	{"b", []byte("01234"), []int{0}},
	{"f", []byte("(+ 1\n3)\n"), []int{0, 5}},
	{"f", []byte("(+ 1\n3)\n\n"), []int{0, 5, 8}},
	{"f", []byte("(+ 1\n3)\n\n1"), []int{0, 5, 8, 9}},
}

func test_handler(t *testing.T, filename, src string, expectedLines []int) {
	var s scan.Scanner
	f := token.NewFile("", src)
	s.Init(f, src)
	// scan完
	_, tok, _ := s.Scan();
	for ;!tok.IsEOF(); {
		//t.Log(tok.String())
		_, tok, _ = s.Scan()
	}

	lines := f.Lines()
	if fmt.Sprintf("%v", lines) != fmt.Sprintf("%v", expectedLines) {
		t.Fatal(src, "Expect lines: ", expectedLines, "Got: ", lines)
	}
}

func TestFilePosition2(t *testing.T) {
	for _, item := range tests {
		test_handler(t, item.filename, string(item.source), item.lines)
	}
}
