package main

type Callable interface {
	Call(lox *Interpreter, args []Object) (ret Object)
	Arity() int
}

func (f *LoxFunction) Call(lox *Interpreter, args []Object) (ret Object) {
	oldScope := lox.env
	lox.env = NewEnvironment(f.closure)
	defer func() {
		lox.env = oldScope
	}()

	for i, arg := range args {
		lox.env.Define(f.funDecl.params[i].Lexeme, arg)
	}

	for _, stmt := range f.funDecl.body {
		if retVal, ret := stmt.Run(lox); ret {
			if f.isInit {
				return lox.env.Get("this")
			}
			return retVal
		}
	}

	if f.isInit {
		return f.closure.Get("this")
	}
	return &LoxNil{}
}

func (f *LoxFunction) Arity() int {
	return len(f.funDecl.params)
}

// Adds a new environment where "this" is a variable holding the instance
func (f *LoxFunction) bind(loxInstance *LoxInstance) *LoxFunction {
	env := NewEnvironment(f.closure)
	env.Define("this", loxInstance)
	return &LoxFunction{funDecl: f.funDecl, closure: env, isInit: f.isInit}
}

func (c *LoxClass) Call(lox *Interpreter, args []Object) (ret Object) {
	instance := &LoxInstance{loxClass: *c, fields: make(map[string]Object)}

	// If there is an initializer, call it before returning the instance
	if initializer := c.FindMethod("init"); initializer != nil {
		initializer.bind(instance).Call(lox, args)
	}
	return instance
}

func (c *LoxClass) Arity() int {
	if initializer := c.FindMethod("init"); initializer != nil {
		return initializer.Arity()
	}
	return 0
}

func (c *LoxClass) FindMethod(name string) *LoxFunction {
	if m, ok := c.methods[name]; ok {
		return m
	}
	if c.superclass != nil {
		if m := c.superclass.FindMethod(name); m != nil {
			return m
		}
	}
	return nil
}

func (i *LoxInstance) Get(name string) Object {
	if field, ok := i.fields[name]; ok {
		return field
	}
	method := i.loxClass.FindMethod(name)
	if method == nil {
		runtimeError("Undefined property: " + name)
	}
	return method.bind(i)
}

func (i *LoxInstance) Set(name string, value Object) {
	i.fields[name] = value
}
