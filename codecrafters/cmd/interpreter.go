package main

type Interpreter struct {
	tokens  []Token
	ast     Program
	globals Environment
	env     *Environment // a pointer to the current environment
	locals  map[Expr]int // side table for how many environments up to look
}

func (lox *Interpreter) Scan(filename string) bool {
	scanner := Scanner{}
	scanner.init(filename)
	lox.tokens = scanner.scan()
	return scanner.lexicalError
}

func (lox *Interpreter) Parse() {
	parser := Parser{tokens: lox.tokens}
	lox.ast = parser.program()
}

func (lox *Interpreter) Resolve() {
	resolver := NewResolver()
	lox.ast.resolve(resolver)
	lox.locals = resolver.locals
}

func (lox *Interpreter) Evaluate() {
	lox.globals = *NewEnvironment(nil)
	lox.env = &lox.globals

	// Maybe can check for errors here
	lox.ast.Run(lox)
}

func (lox *Interpreter) NewScope() {
	lox.env = NewEnvironment(lox.env)
}

func (lox *Interpreter) EndScope() {
	lox.env = lox.env.parent
}

func (lox Interpreter) GetAt(distance int, name string) Object {
	return lox.env.Ancestor(distance).values[name]
}

func (lox *Interpreter) AssignAt(distance int, name string, obj Object) {
	lox.env.Ancestor(distance).values[name] = obj
}

func (lox *Interpreter) LookUpVariable(expr Expr, name string) Object {
	distance, isLocal := lox.locals[expr]

	if isLocal {
		return lox.GetAt(distance, name)
	} else {
		return lox.globals.Get(name)
	}
}
