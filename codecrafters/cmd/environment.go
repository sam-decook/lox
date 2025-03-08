package main

type Environment struct {
	parent *Environment
	values map[string]Object
}

func NewEnvironment(parent *Environment) Environment {
	return Environment{
		parent: parent,
		values: make(map[string]Object, 11),
	}
}

func (e *Environment) Define(name string, value Object) {
	// Overwrite if it already exists
	// Nice for a REPL (you don't want to mentally track every declaration)
	// Might hide accidental redeclarations, and be better to make users
	// assign the variable a new value instead
	e.values[name] = value
}

func (e *Environment) Assign(name string, value Object) {
	for env := e; env != nil; env = env.parent {
		if _, found := env.values[name]; found {
			env.values[name] = value
			return
		}
	}
	runtimeError("Undefined variable: " + name)
}

func (e Environment) Get(name string) Object {
	value, found := e.values[name]
	if !found && e.parent != nil {
		return e.parent.Get(name)
	}
	if !found {
		runtimeError("Undefined variable: " + name)
	}
	return value
}
