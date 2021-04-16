# Goscript v0.2.0

用Go实现Go语法子集的脚本语言

# 支持以下语法

## 变量

支持 int, float, bool

```
var a int
var b float
var c bool
a = 12
b = 12.0
c = true

a1 := 12
b2 := 12.0
c3 := true
```

# if else
```
if expression {
    //...
} else {
    //...
}

```

# function

```
func funcname (a, b int) : int {
    return a + b
}
```

# for

```
for i := 0; i < 10; i++ {
    //...
}
```

## 用法
```
go install ./ # 生成goscript二进制
goscript examples/test.gs # 运行
```