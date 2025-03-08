package main

type Interpreter struct {
	filename string
	scanner  Scanner
	tokens   []Token
	parser   Parser
	ast      Program
	env      Environment
}

func (i *Interpreter) Scan(filename string) {
	i.filename = filename

	i.scanner.init(filename)

	i.tokens = i.scanner.scan()
}

func (i *Interpreter) Parse() {
	i.parser.tokens = i.tokens
	i.ast = i.parser.program()
}

func (i *Interpreter) Evaluate() {
	i.env = NewEnvironment(nil)

	// Maybe can check for errors here
	i.ast.Run(&i.env)
}
