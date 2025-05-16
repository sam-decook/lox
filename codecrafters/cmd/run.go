package main

import "fmt"

func (p *Program) Run(lox *Interpreter) (retVal Object, ret bool) {
	for _, decl := range p.decls {
		decl.Run(lox)
	}
	return nil, false
}

func (c *ClassDecl) Run(lox *Interpreter) (retVal Object, ret bool) {
	lox.env.Define(c.name, &LoxNil{})

	var superclass *LoxClass
	if c.superclass != nil {
		if sc, ok := c.superclass.Evaluate(lox).(*LoxClass); ok {
			superclass = sc
		} else {
			runtimeError("Superclass must be a class.")
		}

		lox.env = NewEnvironment(lox.env)
		lox.env.Define("super", superclass)
	}

	loxClass := LoxClass{c.name, superclass, make(map[string]*LoxFunction, len(c.methods))}

	for _, method := range c.methods {
		loxClass.methods[method.name] = &LoxFunction{
			funDecl: method,
			closure: lox.env,
			isInit:  method.name == "init",
		}
	}

	if c.superclass != nil {
		lox.env = lox.env.parent
	}

	lox.env.Assign(c.name, &loxClass)

	return nil, false
}

// This runs the function *declaration*, not the function itself, so it just
// adds it to the environment.
func (fd *FunDecl) Run(lox *Interpreter) (retVal Object, ret bool) {
	lox.env.Define(fd.name, &LoxFunction{funDecl: fd, closure: lox.env})
	return nil, false
}

func (b *Block) Run(lox *Interpreter) (retVal Object, ret bool) {
	lox.NewScope()
	defer lox.EndScope()

	for _, decl := range b.decls {
		retVal, ret := decl.Run(lox)
		if ret {
			return retVal, true
		}
	}
	return nil, false
}

func (vd *VarDecl) Run(lox *Interpreter) (retVal Object, ret bool) {
	if vd.expr == nil {
		lox.env.Define(vd.name, &LoxNil{})
	} else {
		lox.env.Define(vd.name, vd.expr.Evaluate(lox))
	}
	return nil, false
}

// Yeah, it does nothing
func (es *ExprStmt) Run(lox *Interpreter) (retVal Object, ret bool) {
	es.expr.Evaluate(lox)
	return nil, false
}

func (ps *PrintStmt) Run(lox *Interpreter) (retVal Object, ret bool) {
	fmt.Println(ps.expr.Evaluate(lox))
	return nil, false
}

func (rs *ReturnStmt) Run(lox *Interpreter) (retVal Object, ret bool) {
	retVal = &LoxNil{}
	if rs.expr != nil {
		retVal = rs.expr.Evaluate(lox)
	}
	return retVal, true
}

func (is *IfStmt) Run(lox *Interpreter) (retVal Object, ret bool) {
	if IsTruthy(is.condition.Evaluate(lox)) {
		retVal, ret := is.thenBranch.Run(lox)
		if ret {
			return retVal, true
		}
	} else if is.elseBranch != nil {
		retVal, ret := is.elseBranch.Run(lox)
		if ret {
			return retVal, true
		}
	}
	return nil, false
}

func (ws *WhileStmt) Run(lox *Interpreter) (retVal Object, ret bool) {
	for IsTruthy(ws.condition.Evaluate(lox)) {
		retVal, ret := ws.body.Run(lox)
		if ret {
			return retVal, true
		}
	}
	return nil, false
}
