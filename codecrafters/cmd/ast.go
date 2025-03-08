// Implements this BNF:
// program        → declaration* EOF ;
// declaration    → funDecl | varDecl | statement ;
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
// forStmt        → "for" "(" ( varDecl | exprStmt | ";" )
//                  expression? ";"
//                  expression? ")" statement ;
// ifStmt         → "if" "(" expression ")" statement ( "else" statement )? ;
// printStmt      → "print" expression ";" ;
// returnStmt     → "return" expression? ";" ;
// whileStmt      → "while" "(" expression ")" statement ;
// block          → "{" declaration* "}" ;
//
// expression     → assignment ;
// assignment     → IDENTIFIER "=" assignment | logic_or ;
// logic_or       → logic_and ( "or" logic_and )* ;
// logic_and      → equality ( "and" equality )* ;
// equality       → comparison ( ( "!=" | "==" ) comparison )* ;
// comparison     → term ( ( ">" | ">=" | "<" | "<=" ) term )* ;
// term           → factor ( ( "-" | "+" ) factor )* ;
// factor         → unary ( ( "/" | "*" ) unary )* ;
// unary          → ( "!" | "-" ) unary | call ;
// call           → primary ( "(" arguments? ")" )* ;
// arguments      → expression ( "," expression )* ;
// primary        → NUMBER | STRING | "true" | "false" | "nil"
//                | "(" expression ")"
//                | IDENTIFIER;

package main

import (
	"fmt"
	"strings"
)

type Stmt interface {
	Run(env *Environment) (retVal Object, ret bool)
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

// Technically this is a function, we are combining it with assigning it to a variable
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

type WhileStmt struct {
	condition Expr
	body      Stmt
}

func (ws *WhileStmt) String() string {
	return fmt.Sprintf("while (%s) %s", ws.condition, ws.body)
}

type ForStmt struct {
	initializer Stmt
	condition   Expr
	increment   Expr
	body        Stmt
}

func (fs *ForStmt) String() string {
	sb := strings.Builder{}

	sb.WriteString("for (")

	if fs.initializer != nil {
		sb.WriteString(fs.initializer.String())
	}
	sb.WriteString(";")

	if fs.condition != nil {
		sb.WriteString(fs.condition.String())
	}
	sb.WriteString(";")

	if fs.increment != nil {
		sb.WriteString(fs.increment.String())
	}
	sb.WriteString(") ")

	sb.WriteString(fs.body.String())

	return sb.String()
}

type ExprStmt struct {
	expr Expr
}

func (es *ExprStmt) String() string {
	return es.expr.String()
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

type Expr interface {
	Evaluate(env *Environment) Object
	String() string
}

type AssignmentExpr struct {
	name string
	expr Expr
}

func (ae *AssignmentExpr) String() string {
	return fmt.Sprintf("%s = %s", ae.name, ae.expr)
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

type BinaryExpr struct {
	left  Expr
	op    Token
	right Expr
}

func (be *BinaryExpr) String() string {
	return fmt.Sprintf("(%s %s %s)", be.op.Lexeme, be.left, be.right)
}

type GroupExpr struct {
	group Expr
}

func (ge *GroupExpr) String() string {
	return fmt.Sprintf("(group %s)", ge.group)
}

type LiteralExpr struct {
	token Token
	value string
}

func (le *LiteralExpr) String() string {
	return le.value
}

type VariableExpr struct {
	name Token
}

func (ve *VariableExpr) String() string {
	return ve.name.Lexeme
}
