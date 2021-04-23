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
	topScope *ast.Scope // top-most scope;
}

func (p *exec) openScope() {
	p.topScope = ast.NewScope(p.topScope)
}

func (p *exec) closeScope() {
	p.topScope = p.topScope.Outer
}

// 入口 解析->计算
func Exec(fname, src string) int {
	c := &exec{}
	c.openScope();

	f := parse.ParseFile(fname, src)
	if f == nil {
		os.Exit(1)
	}
	c.execStmts(f.Stmts)
	return 0
	//return c.execNode(f.Root)
}

func (c *exec) execStmts(stmts []ast.Stmt) {
	for _, item := range stmts {
		switch n := item.(type) {
		case *ast.VarStmt:
			// c.topScope.Insert()
			// n.Values
			for i, name := range n.Names {
				if n.Values != nil && len(n.Values) == len(n.Names) {
					value := c.execNode(n.Values[i])
					obj := &ast.Object{Name: name.Name, Kind: ast.Var, Data: value}
					c.topScope.Insert(obj)
				}
			}
		case *ast.AssignStmt:
			for i, left := range n.Lhs { // left必须是一个ident
				ident := c.getIdent(left)
				name := ident.Name
				value := c.execNode(n.Rhs[i])
				if valueObject, ok := value.(*ast.Object); ok {
					value = valueObject.Data
				}
				fmt.Printf("assign: %v\n", value)
				obj := c.topScope.Lookup(name)
				if obj == nil { // 新的
					obj := &ast.Object{Name: name, Kind: ast.Var, Data: value}
					c.topScope.Insert(obj)
				} else { // 旧的
					obj.Data = value
				}
			}
		case *ast.ExprStmt:
			switch x := n.X.(type) {
			case *ast.CallExpr:
				ident := x.Fun.(*ast.Ident)
				if ident.Name == "print" {
					if x.Args != nil {
						for _, arg := range x.Args {
							value := c.execNode(arg)
							if valueObject, ok := value.(*ast.Object); ok {
								fmt.Printf("%v\n", valueObject.Data)
							} else {
								fmt.Printf("%v\n", value)
							}
						}
					}
				}
			}
		}
	}
}

func (c *exec) getIdent (n ast.Node) *ast.Ident {
	ident := n.(*ast.Ident)
	return ident
}

// 计算每一个Node
func (c *exec) execNode(node ast.Node) interface{} {
	switch n := node.(type) {
	case *ast.Ident:
		return c.topScope.Lookup(n.Name)
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

func (c *exec) execNodeInt(node ast.Node) int {
	t := c.execNode(node)
	if v, ok := t.(*ast.Object); ok {
		t = v.Data
	}
	tmp := t.(int)
	return tmp
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
func (c *exec) execUnaryExpr(b *ast.UnaryExpr) interface{} {
	switch b.Op {
	case token.ADD:
		return c.execNode(b.X) // b.X 肯定是 *ast.UnaryExpr, 但学是从 execNode开始吧
	case token.SUB:
		right := c.execNode(b.X)
		r := right.(int)
		return -r
	}
	return 0
}

// 计算二元操作
func (c *exec) execBinaryExpr(b *ast.BinaryExpr) int {
	var tmp int

	// 1 + 2 先用1作为tmp
	tmp = c.execNodeInt(b.X)

	// 再计算之后的
	node := b.Y
	switch b.Op {
	case token.ADD:
		tmp += c.execNodeInt(node)
	case token.SUB:
		tmp -= c.execNodeInt(node)
	case token.MUL:
		tmp *= c.execNodeInt(node)
	case token.QUO:
		tmp /= c.execNodeInt(node)
	case token.REM:
		tmp %= c.execNodeInt(node)
	}

	return tmp
}

