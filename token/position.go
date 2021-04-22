// Copyright (c) 2014, Rob Thornton
// All rights reserved.
// This source code is governed by a Simplied BSD-License. Please see the
// LICENSE included in this distribution for a copy of the full license
// or, if one is not included, you may also find a copy at
// http://opensource.org/licenses/BSD-2-Clause

// Copyright 2021 The Goscript Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
在扫描的时候, 从流中获得第一个字符开始, 从上往下, 从左往右的的进行.
第一个字符的偏移量是零.

相比而言, 当用户希望知道汇报出来的错误是发生在哪行哪列的时候, 第一个字符应当在第一行, 第一列.
因此, 需要将字符的位置信息翻译为对最终用户有意义的信息.

位置(Pos)是字符的偏移量加上文件的基数. 如果基数是1, 字符串的偏移是0, 这个字符串的 Pos 是1.

位置(Pos)为0是非法的,因为这意味着文件之外的地方.
同样的, 如果一个位置大于文件的基数加上文件的长度,那么也是非法的.

为什么要考虑这么复杂的事情呢? 当你需要解析多个文件的时候,要确定错误信息是从哪个文件中产生的是一件很麻烦的事情.
Pos 使事情变得简单.在后面的文章中会有更多关于此的介绍.

Position 类型严格用于错误报告.它允许我们输出清晰的关于哪行, 哪列, 以及哪个文件发生了错误的信息.

在这个阶段,我们只需要处理单独的一个文件, 但是将来我们会对这段代码会很有用.
*/

package token

import "fmt"

// Pos是某个代码在所有文件的位置, 通过Pos可以很快速的找到代码在某个文件的某行某列
// 如果p1, p2在同一个文件里, p1 < p2 表示p1在p2之前
// 如果p1, p2在不同的文件里, p1 < p2 表示p1在p2之前被解析
// 当前只有一个文件, 所以不会这么复杂
type Pos uint

var NoPos = Pos(0)

func (p Pos) IsValid() bool {
	return p != NoPos
}

type Position struct {
	Filename string // filename, if any
	Offset   int    // offset, starting at 0
	Line     int    // line number, starting at 1
	Column   int    // column number, starting at 1 (byte count)
}

func (p Position) String() string {
	if p.Filename == "" {
		return fmt.Sprintf("%d:%d", p.Line, p.Column)
	}
	return fmt.Sprintf("%s:%d:%d", p.Filename, p.Line, p.Column)
}

// File represents a single source file. It is used to track the number of
// newlines in the file, it's size, name and position within a fileset.
type File struct {
	base int    // 当前阶段只有一个文件, 所以base总为1 Pos的范围: [base...base+size]
	name string // 文件名
	src  string // 源码
	size int    // == len(src)
	// lines 存储每一行的第一个字符的offset, 所以lines[0] == 0 (注意, 这和base没关系)
	// a\nbc\nd -> lines = {0, 2, 5} \n算一个字符
	// "ab\nc\n" -> {0, 3}.
	lines []int
}

func NewFile(name, src string) *File {
	f := &File{
		base:  1,
		name:  name,
		src:   src,
		size:  len(src),
		lines: []int{},
	}
	if f.size > 0 {
		f.lines = []int{0}
	}
	return f
}

// Name returns the file name of file f as registered with AddFile.
func (f *File) Name() string {
	return f.name
}

// Base returns the base offset of file f as registered with AddFile.
func (f *File) Base() int {
	return f.base
}

// Size returns the size of file f as registered with AddFile.
func (f *File) Size() int {
	return f.size
}

func (f *File) Lines() []int {
	return f.lines
}

// AddLine adds the position of the start of a line in the source file at
// the given offset. Every file consists of at least one line at offset zero.
// 扫描器检测到一个新行(\n), 就需要将这个行首添加到文件的行列表中
//
//func (f *File) AddLine(offset int) {
//	if offset >= f.base-1 && offset < f.base+len(f.src) {
//		f.lines = append(f.lines, offset)
//	}
//}

// AddLine adds the line offset for a new line.
// The line offset must be larger than the offset for the previous line
// and smaller than the file size; otherwise the line offset is ignored.
// 扫描器检测到一个新行(\n), 就需要将这个行首添加到文件的行列表中 offset > line[上一行] < size
//
func (f *File) AddLine(offset int) {
	if i := len(f.lines); (i == 0 || f.lines[i-1] < offset) && offset < f.size {
		f.lines = append(f.lines, offset)
	} else {
		panic("Invalid AddLine offset")
	}
}

// Pos returns the Pos value for the given file offset;
// the offset must be <= f.Size().
// f.Pos(f.Offset(p)) == p.
//
func (f *File) Pos(offset int) Pos {
	if offset > f.size {
		panic("illegal file offset")
	}
	return Pos(f.base + offset)
}

// Offset returns the offset for the given file position p;
// p must be a valid Pos value in that file.
// f.Offset(f.Pos(offset)) == offset.
//
func (f *File) Offset(p Pos) int {
	if int(p) < f.base || int(p) > f.base+f.size {
		panic("illegal Pos value")
	}
	return int(p) - f.base
}

// pos -> position
// 注意, 这里的Pos已经 = base + offset了
func (f *File) Position(p Pos) Position {
	row, col := 0, 0

	for _row, offset := range f.lines {
		lineStartPos := f.Pos(offset)
		if p >= lineStartPos {
			row = _row + 1
			col = int(p - lineStartPos) + 1
		}
	}

	return Position{Filename: f.name, Column: col, Line: row}
}

// Line returns the line number for the given file position p;
// p must be a Pos value in that file or NoPos.
//
func (f *File) Line(p Pos) int {
	return f.Position(p).Line
}
