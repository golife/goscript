package ast

import "github.com/golife/goscript/token"

// ----------------------------------------------------------------------------
// Statements

// A statement is represented by a tree consisting of one
// or more of the following concrete statement nodes.
//
type (
	// A BadStmt node is a placeholder for statements containing
	// syntax errors for which no correct statement nodes can be
	// created.
	//
	BadStmt struct {
		From, To token.Pos // position range of bad statement
	}

	// An EmptyStmt node represents an empty statement.
	// The "position" of the empty statement is the position
	// of the immediately following (explicit or implicit) semicolon.
	//
	EmptyStmt struct {
		Semicolon token.Pos // position of following ";"
		Implicit  bool      // if set, ";" was omitted in the source
	}

	// var
	VarStmt struct {
		TokPos token.Pos   // position of Tok
		Tok    token.Token // IMPORT, CONST, TYPE, or VAR

		Names  []*Ident // value names (len(Names) > 0)
		Type   *Ident   // value type; or nil
		Values []Expr   // initial values; or nil
	}

	// func
	FuncStmt struct {
		TokPos token.Pos   // position of Tok
		Tok    token.Token // func

		Name *Ident     // function/method name
		Type *FuncType  // function signature: type and value parameters, results, and position of "func" keyword
		Body *BlockStmt // function body; or nil for external (non-Go) function
	}

	// A DeclStmt node represents a declaration in a statement list.
	DeclStmt struct {
		Decl Decl // *GenDecl with CONST, TYPE, or VAR token
	}

	// An ExprStmt node represents a (stand-alone) expression
	// in a statement list.
	//
	ExprStmt struct {
		X Expr // expression
	}

	// An AssignStmt node represents an assignment or
	// a short variable declaration.
	// a = 10
	// a, b = 10, 12
	//
	AssignStmt struct {
		Lhs    []Expr      // left
		TokPos token.Pos   // position of Tok
		Tok    token.Token // assignment token, DEFINE
		Rhs    []Expr      // right
	}

	// A ReturnStmt node represents a return statement.
	ReturnStmt struct {
		Return  token.Pos // position of "return" keyword
		Results []Expr    // result expressions; or nil
	}

	// A BlockStmt node represents a braced statement list.
	BlockStmt struct {
		Lbrace token.Pos // position of "{"
		List   []Stmt
		Rbrace token.Pos // position of "}", if any (may be absent due to syntax error)
	}

	// An IfStmt node represents an if statement.
	IfStmt struct {
		If   token.Pos // position of "if" keyword
		Init Stmt      // initialization statement; or nil
		Cond Expr      // condition
		Body *BlockStmt
		Else Stmt // else branch; or nil
	}

	// A ForStmt represents a for statement.
	ForStmt struct {
		For  token.Pos // position of "for" keyword
		Init Stmt      // initialization statement; or nil
		Cond Expr      // condition; or nil
		Post Stmt      // post iteration statement; or nil
		Body *BlockStmt
	}
)

func (s *BadStmt) Pos() token.Pos    { return s.From }
func (s *DeclStmt) Pos() token.Pos   { return s.Decl.Pos() }
func (s *EmptyStmt) Pos() token.Pos  { return s.Semicolon }
func (s *ExprStmt) Pos() token.Pos   { return s.X.Pos() }
func (s *AssignStmt) Pos() token.Pos { return s.Lhs[0].Pos() }
func (s *ReturnStmt) Pos() token.Pos { return s.Return }
func (s *BlockStmt) Pos() token.Pos  { return s.Lbrace }
func (s *IfStmt) Pos() token.Pos     { return s.If }
func (s *ForStmt) Pos() token.Pos    { return s.For }
func (s *VarStmt) Pos() token.Pos    { return s.TokPos }
func (s *FuncStmt) Pos() token.Pos   { return s.TokPos }

func (s *BadStmt) End() token.Pos  { return s.To }
func (s *DeclStmt) End() token.Pos { return s.Decl.End() }
func (s *EmptyStmt) End() token.Pos {
	if s.Implicit {
		return s.Semicolon
	}
	return s.Semicolon + 1 /* len(";") */
}
func (s *ExprStmt) End() token.Pos   { return s.X.End() }
func (s *AssignStmt) End() token.Pos { return s.Rhs[len(s.Rhs)-1].End() }
func (s *ReturnStmt) End() token.Pos {
	if n := len(s.Results); n > 0 {
		return s.Results[n-1].End()
	}
	return s.Return + 6 // len("return")
}
func (s *BlockStmt) End() token.Pos {
	if s.Rbrace.IsValid() {
		return s.Rbrace + 1
	}
	if n := len(s.List); n > 0 {
		return s.List[n-1].End()
	}
	return s.Lbrace + 1
}
func (s *IfStmt) End() token.Pos {
	if s.Else != nil {
		return s.Else.End()
	}
	return s.Body.End()
}
func (s *ForStmt) End() token.Pos  { return s.Body.End() }
func (s *VarStmt) End() token.Pos  { return s.Values[len(s.Values)-1].End() }
func (s *FuncStmt) End() token.Pos { return s.Body.Rbrace }

// stmtNode() ensures that only statement nodes can be
// assigned to a Stmt.
//
func (*BadStmt) stmtNode()    {}
func (*DeclStmt) stmtNode()   {}
func (*EmptyStmt) stmtNode()  {}
func (*ExprStmt) stmtNode()   {}
func (*AssignStmt) stmtNode() {}
func (*ReturnStmt) stmtNode() {}
func (*BlockStmt) stmtNode()  {}
func (*IfStmt) stmtNode()     {}
func (*ForStmt) stmtNode()    {}
func (*VarStmt) stmtNode()    {}
func (*FuncStmt) stmtNode()    {}
