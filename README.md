# Goscript v0.2.0

用Go实现Go语法子集的脚本语言

# 语法规范

## 变量声明

支持 int, float, bool

```
VarDecl     = "var" VarSpec .
VarSpec     = IdentifierList ( Type [ "=" ExpressionList ] | "=" ExpressionList ) .
IdentifierList = identifier { "," identifier } .
ExpressionList = Expression { "," Expression } .

var a int
var b float
var c bool
a = 12
b = 12.0
c = true

var U, V, W float
var x, y float = -1, -2
```

## 变量声明+赋值缩写

```
ShortVarDecl = IdentifierList ":=" ExpressionList .
a1 := 12
b2 := 12.0
c3 := true

a, b, c := 1, 2, 3
```

## 赋值
```
Assignment = ExpressionList assign_op ExpressionList .
assign_op = "=" .

x = 1
x, y, z = 1, 2, 3
```

## if else
```
IfStmt = "if" [ SimpleStmt ";" ] Expression Block [ "else" ( IfStmt | Block ) ] .

if expression {
    //...
} else {
    //...
}

```

## for

```
ForStmt = "for" [ Condition | ForClause ] Block .
Condition = Expression .

ForClause = [ InitStmt ] ";" [ Condition ] ";" [ PostStmt ] .
InitStmt = SimpleStmt .
PostStmt = SimpleStmt .

for i := 0; i < 10; i++ {
    //...
}
```

## function

```
FunctionDecl = "func" FunctionName Signature [ FunctionBody ] .
FunctionName = identifier .
FunctionBody = Block .
Signature    = Parameters [ Result ] .
Result       = Parameters | Type .
Parameters   = "(" [ ParameterList [ "," ] ] ")" .
ParameterList  = ParameterDecl { "," ParameterDecl } .
ParameterDecl  = [ IdentifierList ] [ "..." ] Type .

Block = "{" StatementList "}" .
StatementList = { Statement ";" } .

Statement = VarDecl | SimpleStmt | ReturnStmt | Block | IfStmt | ForStmt .
SimpleStmt = EmptyStmt | ExpressionStmt | Assignment | ShortVarDecl .
ReturnStmt = "return" [ ExpressionList ] .
ExpressionStmt = Expression .
EmptyStmt = . # The empty statement does nothing.

func()
func(x int) int
func(a, _ int, z float32) bool
func(a, b int, z float32) (bool)
func(prefix string, values ...int)
func(a, b int, z float64, opt ...interface{}) (success bool)
func(int, int, float64) (float64, *[]int)
func(n int) func(p *T)

func funcname (a, b int) : int {
    return a + b
}
```

## Expression
```
Expression = Term { addOp Term }
Term = UnaryExpr { mulop UnaryExpr }
UnaryExpr  = PrimaryExpr | unary_op UnaryExpr .
PrimaryExpr -> BasicLit | "(" Expression ")" | MethodExpr
BasicLit    = int_lit | float_lit | imaginary_lit | rune_lit | string_lit .

mulop = "*" | "/" | "%"
addOp = "+" | "-"
unary_op   = "+" | "-"

```

## File
```
File = { (Statement ";") | (FunctionDecl ";") }
```

## 用法
```
go install ./ # 生成goscript二进制
goscript examples/test.gs # 运行
```