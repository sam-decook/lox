package main

import (
	"fmt"
	"os"
)

// In order for variables to always evaluate to the same value (in closures?),
// we need to do a static analysis pass to resolve them.
//
// However, since this is a *static* analysis, we are not evaluating any
// expressions. We will store which scope to look for a variable in by how many
// environments back it should look it. Then instead of recursively looking in
// each parent environment, we can just skip to the correct one.
//
// We will have to match the scope behavior exactly, which won't be too hard,
// but could be a source of bugs later if changes are not made in both places.
//
// Finally, we will only do this for variables in a local scope, i.e. not
// globals, since they already obey slightly different rules.

type FunctionType int

const (
	FunctionTypeNone FunctionType = iota
	FunctionTypeFunction
	FunctionTypeInitializer
	FunctionTypeMethod
)

type ClassType int

const (
	ClassTypeNone ClassType = iota
	ClassTypeClass
	ClassTypeSubclass
)

type Resolver struct {
	locals    map[Expr]int
	scopes    []map[string]bool
	funcType  FunctionType
	classType ClassType
}

func NewResolver() *Resolver {
	return &Resolver{
		locals: make(map[Expr]int),
		scopes: []map[string]bool{},
	}
}

// Helper functions for scopes
func (r *Resolver) BeginScope() {
	r.scopes = append(r.scopes, make(map[string]bool))
}

func (r *Resolver) EndScope() {
	if len(r.scopes) == 0 {
		panic("No scope to end")
	}
	r.scopes = r.scopes[:len(r.scopes)-1]
}

// Common interface for all AST nodes to implement
type ASTNode interface {
	resolve(r *Resolver)
}

func (p *Program) resolve(r *Resolver) {
	for _, decl := range p.decls {
		decl.resolve(r)
	}
}

func (c *ClassDecl) resolve(r *Resolver) {
	enclosingClassType := r.classType
	r.classType = ClassTypeClass

	r.declare(c.name)
	r.define(c.name)

	if c.superclass != nil {
		r.classType = ClassTypeSubclass
		if c.name == c.superclass.name.Lexeme {
			fmt.Fprintf(os.Stderr, "A class can't inherit from itself.\n")
			os.Exit(65)
		}

		c.superclass.resolve(r)

		r.BeginScope()
		r.declare("super")
		r.define("super")
	}

	r.BeginScope()
	r.declare("this")
	r.define("this")

	for _, method := range c.methods {
		fnType := FunctionTypeMethod
		if method.name == "init" {
			fnType = FunctionTypeInitializer
		}
		r.resolveFunction(method, fnType)
	}

	r.EndScope()

	if c.superclass != nil {
		r.EndScope()
	}

	r.classType = enclosingClassType
}

func (fd *FunDecl) resolve(r *Resolver) {
	r.declare(fd.name)
	r.define(fd.name)

	r.resolveFunction(fd, FunctionTypeFunction)
}

func (r *Resolver) resolveFunction(fd *FunDecl, funcType FunctionType) {
	enclosingFnType := r.funcType
	r.funcType = funcType

	r.BeginScope()
	for _, param := range fd.params {
		r.declare(param.Lexeme)
		r.define(param.Lexeme)
	}
	for _, stmt := range fd.body {
		stmt.resolve(r)
	}
	r.EndScope()

	r.funcType = enclosingFnType
}

func (vd *VarDecl) resolve(r *Resolver) {
	r.declare(vd.name)
	if vd.expr != nil {
		vd.expr.resolve(r)
	}
	r.define(vd.name)
}

func (es *ExprStmt) resolve(r *Resolver) {
	es.expr.resolve(r)
}

func (is *IfStmt) resolve(r *Resolver) {
	is.condition.resolve(r)
	if is.elseBranch != nil {
		is.elseBranch.resolve(r)
	}
	is.thenBranch.resolve(r)
}

func (ps *PrintStmt) resolve(r *Resolver) {
	ps.expr.resolve(r)
}

func (rs *ReturnStmt) resolve(r *Resolver) {
	if r.funcType == FunctionTypeNone {
		fmt.Fprintf(os.Stderr, "Cannot return from top-level code.")
		os.Exit(65)
	}
	if rs.expr != nil {
		if r.funcType == FunctionTypeInitializer {
			fmt.Fprintf(os.Stderr, "Cannot return from initializer.")
			os.Exit(65)
		}
		rs.expr.resolve(r)
	}
}

func (ws *WhileStmt) resolve(r *Resolver) {
	ws.condition.resolve(r)
	ws.body.resolve(r)
}

func (b *Block) resolve(r *Resolver) {
	r.BeginScope()
	for _, decl := range b.decls {
		decl.resolve(r)
	}
	r.EndScope()
}

func (ae *AssignmentExpr) resolve(r *Resolver) {
	ae.expr.resolve(r)
	r.resolveLocal(ae, ae.name)
}

func (se *SetExpr) resolve(r *Resolver) {
	se.value.resolve(r)
	// The name is dynamically evaluated
	se.object.resolve(r)
}

func (te *ThisExpr) resolve(r *Resolver) {
	if r.classType == ClassTypeNone {
		fmt.Fprintf(os.Stderr, "Cannot use 'this' outside of a class.")
		os.Exit(65)
	}
	r.resolveLocal(te, te.keyword.Lexeme)
}

func (loe *LogicOrExpr) resolve(r *Resolver) {
	loe.left.resolve(r)
	loe.right.resolve(r)
}

func (lae *LogicAndExpr) resolve(r *Resolver) {
	lae.left.resolve(r)
	lae.right.resolve(r)
}

func (be *BinaryExpr) resolve(r *Resolver) {
	be.left.resolve(r)
	be.right.resolve(r)
}

func (ue *UnaryExpr) resolve(r *Resolver) {
	ue.right.resolve(r)
}

func (ce *CallExpr) resolve(r *Resolver) {
	ce.callee.resolve(r)
	for _, arg := range ce.args {
		arg.resolve(r)
	}
}

func (ge *GetExpr) resolve(r *Resolver) {
	ge.object.resolve(r)
	// The name is dynamically evaluated
}

func (le *LiteralExpr) resolve(r *Resolver) {
	// Nothing to resolve
}

func (ge *GroupExpr) resolve(r *Resolver) {
	ge.group.resolve(r)
}

func (ve *VariableExpr) resolve(r *Resolver) {
	last := len(r.scopes) - 1
	if last >= 0 {
		defined, declared := r.scopes[last][ve.name.Lexeme]
		if declared && !defined {
			msg := "Can't read local variable in its own initializer."
			fmt.Fprintf(os.Stderr, "[line %d] Error at '%s': %s\n", ve.name.Line, ve.name.Lexeme, msg)
			os.Exit(65)
		}
	}

	r.resolveLocal(ve, ve.name.Lexeme)
}

func (se *SuperExpr) resolve(r *Resolver) {
	if r.classType == ClassTypeNone {
		fmt.Fprintf(os.Stderr, "Can't use 'super' outside of a class.")
		os.Exit(65)
	} else if r.classType != ClassTypeSubclass {
		fmt.Fprintf(os.Stderr, "Can't use 'super' without a superclass.")
		os.Exit(65)
	}
	r.resolveLocal(se, se.keyword.Lexeme)
}

// Helper functions for resolving
func (r *Resolver) declare(name string) {
	if len(r.scopes) == 0 {
		return
	}

	scope := r.scopes[len(r.scopes)-1]
	if _, ok := scope[name]; ok {
		fmt.Fprintf(os.Stderr, "Already a variable named %s in this scope.", name)
		os.Exit(65)
	}

	scope[name] = false
}

func (r *Resolver) define(name string) {
	if len(r.scopes) == 0 {
		return
	}

	scope := r.scopes[len(r.scopes)-1]
	scope[name] = true
}

// The expr *MUST* be a pointer to something that implements the Expr interface
func (r *Resolver) resolveLocal(expr Expr, name string) {
	last := len(r.scopes) - 1
	for i := last; i >= 0; i-- {
		if _, ok := r.scopes[i][name]; ok {
			// Store how many scopes back to look
			r.locals[expr] = last - i
			return
		}
	}
}
