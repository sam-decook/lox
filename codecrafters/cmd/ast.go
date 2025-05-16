// Implements this BNF:
// program        → declaration* EOF ;
// declaration    → classDecl
//                | funDecl
//                | varDecl
//                | statement ;
// classDecl      → "class" IDENTIFIER ( "<" IDENTIFIER )? "{" function* "}" ;
// funDecl        → "fun" function ;
// function       → IDENTIFIER "(" parameters? ")" block ;
// parameters     → IDENTIFIER ( "," IDENTIFIER )* ;
// varDecl        → "var" IDENTIFIER ( "=" expression )? ";" ;
// statement      → exprStmt
//                | forStmt
//                | ifStmt
//                | printStmt
//                | returnStmt
//                | whileStmt
//                | block ;
// exprStmt       → expression ";" ;
// forStmt        → "for" "(" ( varDecl | exprStmt | ";" ) expression? ";" expression? ")" statement ;
// ifStmt         → "if" "(" expression ")" statement ( "else" statement )? ;
// printStmt      → "print" expression ";" ;
// returnStmt     → "return" expression? ";" ;
// whileStmt      → "while" "(" expression ")" statement ;
// block          → "{" declaration* "}" ;
//
// expression     → assignment ;
// assignment     → ( call "." )? IDENTIFIER "=" assignment
//                | logic_or ;
// logic_or       → logic_and ( "or" logic_and )* ;
// logic_and      → equality ( "and" equality )* ;
// equality       → comparison ( ( "!=" | "==" ) comparison )* ;
// comparison     → term ( ( ">" | ">=" | "<" | "<=" ) term )* ;
// term           → factor ( ( "-" | "+" ) factor )* ;
// factor         → unary ( ( "/" | "*" ) unary )* ;
// unary          → ( "!" | "-" ) unary | call ;
// call           → primary ( "(" arguments? ")" | "." IDENTIFIER )* ;
// arguments      → expression ( "," expression )* ;
// primary        → NUMBER | STRING | "true" | "false" | "nil" | "(" expression ")"
//                | IDENTIFIER | "super" "." IDENTIFIER ;

package main

import (
	"fmt"
	"strings"
)

type Stmt interface {
	ASTNode
	// `ret` is true if there was a return statement, and `retVal` holds the `Object`
	//
	// This is useful for distinguishing between a nil return and a LoxNil return.
	Run(lox *Interpreter) (retVal Object, ret bool)
	String() string
}

type Program struct {
	decls []Stmt
}

func (p *Program) String() string {
	sb := strings.Builder{}
	for _, stmt := range p.decls {
		sb.WriteString(stmt.String() + "\n")
	}
	return sb.String()
}

type ClassDecl struct {
	name       string
	superclass *VariableExpr
	methods    []*FunDecl
}

func (cd *ClassDecl) String() string {
	sb := strings.Builder{}
	sb.WriteString("class " + cd.name)
	if cd.superclass != nil {
		sb.WriteString("< " + cd.superclass.name.Lexeme)
	}
	sb.WriteString(" {\n")
	for _, method := range cd.methods {
		sb.WriteString("\t" + method.String() + "\n")
	}
	sb.WriteString("}")
	return sb.String()
}

type FunDecl struct {
	name   string
	params []Token
	body   []Stmt //not a block so the parameters can be more easily added
}

func (fd *FunDecl) String() string {
	sb := strings.Builder{}
	sb.WriteString("fun " + fd.name + "(")
	if len(fd.params) > 0 {
		sb.WriteString(fd.params[0].Lexeme)
		for _, arg := range fd.params[1:] {
			sb.WriteString(", " + arg.Lexeme)
		}
	}
	sb.WriteString(") ")
	for _, stmt := range fd.body {
		sb.WriteString(stmt.String() + "\n")
	}
	return sb.String()
}

type VarDecl struct {
	name string
	expr Expr
}

func (vd *VarDecl) String() string {
	sb := strings.Builder{}

	sb.WriteString("var " + vd.name)
	if vd.expr != nil {
		sb.WriteString(" = " + vd.expr.String())
	}

	return sb.String()
}

type ExprStmt struct {
	expr Expr
}

func (es *ExprStmt) String() string {
	return es.expr.String()
}

// For statements de-sugar into while statements

type IfStmt struct {
	condition  Expr
	thenBranch Stmt
	elseBranch Stmt
}

func (is *IfStmt) String() string {
	sb := strings.Builder{}
	sb.WriteString("if (" + is.condition.String() + ") ") // extra space in case a block is next
	sb.WriteString(is.thenBranch.String())
	if is.elseBranch != nil {
		sb.WriteString("else " + is.elseBranch.String())
	}
	return sb.String()
}

type PrintStmt struct {
	expr Expr
}

func (ps *PrintStmt) String() string {
	return "print " + ps.expr.String()
}

type ReturnStmt struct {
	keyword Token //for locating & error reporting
	expr    Expr
}

func (rs *ReturnStmt) String() string {
	str := "return"
	if rs.expr != nil {
		str += " " + rs.expr.String()
	}
	return str
}

type WhileStmt struct {
	condition Expr
	body      Stmt
}

func (ws *WhileStmt) String() string {
	return fmt.Sprintf("while (%s) %s", ws.condition, ws.body)
}

type Block struct {
	decls []Stmt
}

// TODO: add indentation based on depth using a variable
func (b *Block) String() string {
	sb := strings.Builder{}
	sb.WriteString("{\n")
	for _, decl := range b.decls {
		sb.WriteString("    " + decl.String() + "\n")
	}
	sb.WriteByte('}')
	return sb.String()
}

type Expr interface {
	ASTNode
	Evaluate(lox *Interpreter) Object
	String() string
}

type AssignmentExpr struct {
	name string
	expr Expr
}

func (ae *AssignmentExpr) String() string {
	return fmt.Sprintf("%s = %s", ae.name, ae.expr)
}

type SetExpr struct {
	object Expr
	name   string
	value  Expr
}

func (se *SetExpr) String() string {
	return fmt.Sprintf("%s.%s = %s", se.object, se.name, se.value)
}

type ThisExpr struct {
	keyword Token
}

func (te *ThisExpr) String() string {
	return fmt.Sprintf("this")
}

type LogicOrExpr struct {
	left  Expr
	right Expr
	op    Token
}

func (loe *LogicOrExpr) String() string {
	return fmt.Sprintf("(%s %s %s)", loe.op.Lexeme, loe.left, loe.right)
}

type LogicAndExpr struct {
	left  Expr
	right Expr
	op    Token
}

func (lae *LogicAndExpr) String() string {
	return fmt.Sprintf("(%s %s %s)", lae.op.Lexeme, lae.left, lae.right)
}

type BinaryExpr struct {
	left  Expr
	op    Token
	right Expr
}

func (be *BinaryExpr) String() string {
	return fmt.Sprintf("(%s %s %s)", be.op.Lexeme, be.left, be.right)
}

type UnaryExpr struct {
	op    Token
	right Expr
}

func (ue *UnaryExpr) String() string {
	return fmt.Sprintf("(%s %s)", ue.op.Lexeme, ue.right)
}

type CallExpr struct {
	callee Expr
	// paren	Token // the book has this, I'm not sure why atm
	args []Expr
}

func (ce *CallExpr) String() string {
	sb := strings.Builder{}
	sb.WriteString(ce.callee.String())
	sb.WriteByte('(')
	if len(ce.args) > 0 {
		sb.WriteString(ce.args[0].String())
		for _, arg := range ce.args[1:] {
			sb.WriteString(", " + arg.String())
		}
	}
	sb.WriteByte(')')
	return sb.String()
}

type GetExpr struct {
	object Expr
	name   Token
}

func (ge *GetExpr) String() string {
	return fmt.Sprintf("%s.%s", ge.object, ge.name.Lexeme)
}

type LiteralExpr struct {
	token Token
	value string
}

func (le *LiteralExpr) String() string {
	return le.value
}

type GroupExpr struct {
	group Expr
}

func (ge *GroupExpr) String() string {
	return fmt.Sprintf("(group %s)", ge.group)
}

type VariableExpr struct {
	name Token
}

func (ve *VariableExpr) String() string {
	return ve.name.Lexeme
}

type SuperExpr struct {
	keyword,
	method Token
}

func (se *SuperExpr) String() string {
	return fmt.Sprintf("%s.%s", se.keyword, se.method)
}
