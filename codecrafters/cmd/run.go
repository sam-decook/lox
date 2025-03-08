package main

import "fmt"

func (p *Program) Run(env *Environment) (retVal Object, ret bool) {
	for _, decl := range p.decls {
		decl.Run(env)
	}
	return nil, false
}

// This runs the function *declaration*, not the function itself, so it just
// adds it to the environment.
func (fd *FunDecl) Run(env *Environment) (retVal Object, ret bool) {
	env.Define(fd.name, &LoxFunction{name: fd.name, funDecl: fd, closure: env})
	return nil, false
}

func (b *Block) Run(env *Environment) (retVal Object, ret bool) {
	newEnv := NewEnvironment(env)

	for _, decl := range b.decls {
		retVal, ret := decl.Run(&newEnv)
		if ret {
			return retVal, true
		}
	}
	return nil, false
}

func (vd *VarDecl) Run(env *Environment) (retVal Object, ret bool) {
	if vd.expr == nil {
		env.Define(vd.name, NewNil())
	} else {
		env.Define(vd.name, vd.expr.Evaluate(env))
	}
	return nil, false
}

// Yeah, it does nothing
func (es *ExprStmt) Run(env *Environment) (retVal Object, ret bool) {
	es.expr.Evaluate(env)
	return nil, false
}

func (ps *PrintStmt) Run(env *Environment) (retVal Object, ret bool) {
	fmt.Println(ps.expr.Evaluate(env))
	return nil, false
}

func (rs *ReturnStmt) Run(env *Environment) (retVal Object, ret bool) {
	return rs.expr.Evaluate(env), true
}

func (is *IfStmt) Run(env *Environment) (retVal Object, ret bool) {
	if IsTruthy(is.condition.Evaluate(env)) {
		retVal, ret := is.thenBranch.Run(env)
		if ret {
			return retVal, true
		}
	} else if is.elseBranch != nil {
		retVal, ret := is.elseBranch.Run(env)
		if ret {
			return retVal, true
		}
	}
	return nil, false
}

func (ws *WhileStmt) Run(env *Environment) (retVal Object, ret bool) {
	for IsTruthy(ws.condition.Evaluate(env)) {
		retVal, ret := ws.body.Run(env)
		if ret {
			return retVal, true
		}
	}
	return nil, false
}
