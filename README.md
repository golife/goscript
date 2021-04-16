# Goscript v0.1.1

用Go实现一个简单的计算器

# 规范

```
expression -> term { addOp term }*
addOp -> "+" | "-"
term -> factor { mulop factor }*
mulop -> "*" | "/"
factor -> NUM | "(" expression ")"  
```

## 示例

```
2 + 3
(3+1) * (4/2)
1 * 2 * 3 * 4 * 5
```

## 用法
```
go install ./ # 生成goscript二进制
goscript examples/test.gs # 运行
```