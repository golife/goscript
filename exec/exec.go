// Copyright 2021 The Goscript Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package exec

import (
	"fmt"
	"os"
	"strconv"

	"github.com/golife/goscript/ast"
	"github.com/golife/goscript/parse"
	"github.com/golife/goscript/token"
)

type exec struct {
}

// 入口 解析->计算
func Exec(fname, src string) int {
	var c exec
	f := parse.ParseFile(fname, src)
	if f == nil {
		os.Exit(1)
	}
	return c.execNode(f.Root)
}

// 计算每一个Node
func (c *exec) execNode(node ast.Node) int {
	switch n := node.(type) {
	case *ast.BasicLit:
		return c.execBasicLit(n)
	case *ast.UnaryExpr:
		return c.execUnaryExpr(n)
	case *ast.BinaryExpr:
		return c.execBinaryExpr(n)
	default:
		return 0 /* can't be reached */
	}
}

// 计算字符面值
func (c *exec) execBasicLit(n *ast.BasicLit) int {
	i, err := strconv.Atoi(n.Lit)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return i
}

// 计算一元
func (c *exec) execUnaryExpr(b *ast.UnaryExpr) int {
	switch b.Op {
	case token.ADD:
		return c.execNode(b.X)
	case token.SUB:
		return -c.execNode(b.X)
	}
	return 0
}

// 计算二元操作
func (c *exec) execBinaryExpr(b *ast.BinaryExpr) int {
	var tmp int

	// 1 + 2 先用1作为tmp
	tmp = c.execNode(b.X)

	// 再计算之后的
	node := b.Y
	switch b.Op {
	case token.ADD:
		tmp += c.execNode(node)
	case token.SUB:
		tmp -= c.execNode(node)
	case token.MUL:
		tmp *= c.execNode(node)
	case token.QUO:
		tmp /= c.execNode(node)
	case token.REM:
		tmp %= c.execNode(node)
	}

	return tmp
}

