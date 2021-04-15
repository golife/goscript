// Copyright (c) 2014, Rob Thornton
// All rights reserved.
// This source code is governed by a Simplied BSD-License. Please see the
// LICENSE included in this distribution for a copy of the full license
// or, if one is not included, you may also find a copy at
// http://opensource.org/licenses/BSD-2-Clause

// Copyright 2021 The Goscript Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package token_test

import (
	"testing"

	"github.com/golife/goscript/token"
)

func TestLookup(t *testing.T) {
	var tests = []struct {
		str string
		tok token.Token
	}{
		{"+", token.ADD},
		{"%", token.REM},
		{"EOF", token.EOF},
		{"Integer", token.INTEGER},
		{"Comment", token.COMMENT},
		{"", token.ILLEGAL},
	}

	for i, v := range tests {
		if res := token.Lookup(v.str); res != v.tok {
			t.Fatal(i, "- Expected:", v.tok, "Got:", res)
		}
	}
}

func TestIsLiteral(t *testing.T) {
	var tests = []struct {
		tok token.Token
		exp bool
	}{
		{token.ADD, false},
		{token.REM, false},
		{token.EOF, false},
		{token.INTEGER, true},
		{token.COMMENT, false},
	}

	for _, v := range tests {
		if res := v.tok.IsLiteral(); res != v.exp {
			t.Fatal(v.tok, "- Expected:", v.exp, "Got:", res)
		}
	}
}

func TestIsOperator(t *testing.T) {
	var tests = []struct {
		tok token.Token
		exp bool
	}{
		{token.ADD, true},
		{token.REM, true},
		{token.EOF, false},
		{token.INTEGER, false},
		{token.COMMENT, false},
	}

	for _, v := range tests {
		if res := v.tok.IsOperator(); res != v.exp {
			t.Fatal(v.tok, "- Expected:", v.exp, "Got:", res)
		}
	}
}

func TestString(t *testing.T) {
	var tests = []struct {
		str string
		tok token.Token
	}{
		{"+", token.ADD},
		{"%", token.REM},
		{"EOF", token.EOF},
		{"Integer", token.INTEGER},
		{"Comment", token.COMMENT},
	}

	for i, v := range tests {
		if res := v.tok.String(); res != v.str {
			t.Fatal(i, "- Expected:", v.str, "Got:", res)
		}
	}
}

func TestValid(t *testing.T) {
	var tests = []struct {
		tok token.Token
		exp bool
	}{
		{token.ADD, true},
		{token.REM, true},
		{token.EOF, true},
		{token.INTEGER, true},
		{token.COMMENT, true},
		{token.Token(-1), false},
		{token.Token(999999), false},
	}

	for _, v := range tests {
		if res := v.tok.Valid(); res != v.exp {
			t.Fatal(v.tok, "- Expected:", v.exp, "Got:", res)
		}
	}
}
