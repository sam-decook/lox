package main

import (
	"fmt"
	"os"
)

type Parser struct {
	tokens []Token
	idx    int
}

func (p *Parser) program() Program {
	program := Program{}
	for !p.atEnd() {
		program.decls = append(program.decls, p.declaration())
	}
	return program
}

func (p *Parser) declaration() Stmt {
	switch {
	case p.match(FUN):
		return p.funDecl()
	case p.match(VAR):
		return p.varDecl()
	default:
		return p.statement()
	}
}

func (p *Parser) funDecl() Stmt {
	name := p.consume(IDENTIFIER, "Expect an identifier after 'fun'")
	p.consume(LEFT_PAREN, "Expect '(' after function name")

	params := []Token{}
	if !p.check(RIGHT_PAREN) {
		params = append(params, p.consume(IDENTIFIER, "Expect an identifier"))
		for p.match(COMMA) {
			params = append(params, p.consume(IDENTIFIER, "Expect an identifier"))
		}
	}

	p.consume(RIGHT_PAREN, "Expect ')' after parameters")

	p.consume(LEFT_BRACE, "Expect '{' before function body")
	body := p.block().(*Block)
	// block consumes the trailing '}'

	return &FunDecl{name: name.Lexeme, params: params, body: body.decls}
}

func (p *Parser) varDecl() Stmt {
	p.consume(IDENTIFIER, "An variable declaration must have an identifier")

	vd := VarDecl{}
	vd.name = p.previous().Lexeme

	if p.match(EQUAL) {
		vd.expr = p.expression()
	}
	p.match(SEMICOLON)

	return &vd
}

func (p *Parser) statement() Stmt {
	switch {
	case p.match(FOR):
		return p.forStmt()
	case p.match(IF):
		return p.ifStmt()
	case p.match(PRINT):
		return p.printStmt()
	case p.match(RETURN):
		return p.returnStmt()
	case p.match(WHILE):
		return p.whileStmt()
	case p.match(LEFT_BRACE):
		return p.block()
	default:
		return p.exprStmt()
	}
}

func (p *Parser) exprStmt() Stmt {
	expr := p.expression()
	p.match(SEMICOLON)
	return &ExprStmt{expr}
}

func (p *Parser) printStmt() Stmt {
	expr := p.expression()
	p.match(SEMICOLON)
	return &PrintStmt{expr}
}

func (p *Parser) returnStmt() Stmt {
	key := p.previous()
	if p.match(SEMICOLON) {
		// It's a ghost nil! 👻
		return &ReturnStmt{key, &LiteralExpr{
			token: Token{Type: NIL, Line: key.Line},
			value: "nil"},
		}
	} else {
		expr := p.expression()
		p.consume(SEMICOLON, "Expected ';' after return value")
		return &ReturnStmt{key, expr}
	}
}

func (p *Parser) ifStmt() Stmt {
	p.consume(LEFT_PAREN, "Expected '(' after 'if'")
	condition := p.expression()
	p.consume(RIGHT_PAREN, "Expected ')' after if condition")
	thenBranch := p.statement()
	var elseBranch Stmt
	if p.match(ELSE) {
		elseBranch = p.statement()
	}
	return &IfStmt{condition, thenBranch, elseBranch}
}

func (p *Parser) whileStmt() Stmt {
	p.consume(LEFT_PAREN, "Expected '(' after 'while'")
	condition := p.expression()
	p.consume(RIGHT_PAREN, "Expected ')' after while condition")
	body := p.statement()
	return &WhileStmt{condition, body}
}

func (p *Parser) forStmt() Stmt {
	p.consume(LEFT_PAREN, "Expected '(' after 'for'")

	// Initializer
	var initializer Stmt
	switch {
	case p.match(SEMICOLON):
		initializer = nil
	case p.match(VAR):
		// The varDecl function expects a VAR token to have already been consumed
		initializer = p.varDecl()
	default:
		initializer = p.exprStmt()
	}

	// Condition
	var condition Expr = nil
	if !p.check(SEMICOLON) {
		condition = p.expression()
	}
	p.consume(SEMICOLON, "Expected ';' after loop condition")

	// Increment
	var increment Expr = nil
	if !p.check(RIGHT_PAREN) {
		increment = p.expression()
	}
	p.consume(RIGHT_PAREN, "Expected ')' to end for loop clauses")

	body := p.statement()

	return forToWhile(initializer, condition, increment, body)
}

// Desugars a for loop into a while loop.
func forToWhile(initializer Stmt, condition Expr, increment Expr, body Stmt) Stmt {
	// Add the increment first, since it is in the inner block
	whileBody := body
	if increment != nil {
		whileBody = &Block{decls: []Stmt{body, &ExprStmt{increment}}}
	}

	// Now, turn the body into a while loop
	if condition == nil {
		condition = &LiteralExpr{token: Token{Type: TRUE, Lexeme: "true"}}
	}
	while := &WhileStmt{condition, whileBody}

	// The only thing left is to add the initializer
	whileComplex := Stmt(while)
	if initializer != nil {
		whileComplex = &Block{decls: []Stmt{initializer, while}}
	}

	return whileComplex
}

func (p *Parser) block() Stmt {
	stmts := []Stmt{}

	for !p.check(RIGHT_BRACE) && !p.atEnd() {
		stmts = append(stmts, p.declaration())
	}

	p.consume(RIGHT_BRACE, "Expected '}' after block")

	return &Block{decls: stmts}
}

func (p *Parser) expression() Expr {
	return p.assignment()
}

// This function is a little weird. Go read the book: 8.4.1
func (p *Parser) assignment() Expr {
	expr := p.logicOr()

	if p.match(EQUAL) {
		// equals := p.previous() // I think for reporting an error
		value := p.assignment() // ugh it's recursive

		ve, ok := expr.(*VariableExpr)
		if !ok {
			p.error("Invalid assignment target")
		}

		return &AssignmentExpr{name: ve.name.Lexeme, expr: value}
	}

	return expr
}

func (p *Parser) logicOr() Expr {
	// This acts as the left side while there is "or"s left
	expr := p.logicAnd()

	for p.match(OR) {
		op := p.previous()
		right := p.logicAnd()
		expr = &LogicOrExpr{left: expr, right: right, op: op}
	}

	return expr
}

func (p *Parser) logicAnd() Expr {
	expr := p.equality()

	for p.match(AND) {
		op := p.previous()
		right := p.equality()
		expr = &LogicAndExpr{left: expr, right: right, op: op}
	}

	return expr
}

func (p *Parser) equality() Expr {
	expr := p.comparison()

	for p.match(EQUAL_EQUAL, BANG_EQUAL) {
		op := p.previous()
		right := p.comparison()
		expr = &BinaryExpr{
			left:  expr,
			op:    op,
			right: right,
		}
	}

	return expr
}

func (p *Parser) comparison() Expr {
	expr := p.term()

	for p.match(LESS, LESS_EQUAL, GREATER, GREATER_EQUAL) {
		op := p.previous()
		right := p.term()
		expr = &BinaryExpr{
			left:  expr,
			op:    op,
			right: right,
		}
	}

	return expr
}

func (p *Parser) term() Expr {
	expr := p.factor()

	for p.match(PLUS, MINUS) {
		op := p.previous()
		right := p.factor()
		expr = &BinaryExpr{
			left:  expr,
			op:    op,
			right: right,
		}
	}

	return expr
}

func (p *Parser) factor() Expr {
	expr := p.unary()

	for p.match(STAR, SLASH) {
		op := p.previous()
		right := p.unary()
		expr = &BinaryExpr{
			left:  expr,
			op:    op,
			right: right,
		}
	}

	return expr
}

func (p *Parser) unary() Expr {
	if p.match(BANG, MINUS) {
		op := p.previous()
		right := p.unary()
		return &UnaryExpr{
			op:    op,
			right: right,
		}
	}

	return p.call()
}

func (p *Parser) call() Expr {
	expr := p.primary()

	for {
		if p.match(LEFT_PAREN) {
			expr = p.arguments(expr)
		} else {
			break
		}
	}

	return expr
}

func (p *Parser) arguments(callee Expr) Expr {
	args := []Expr{}

	if !p.check(RIGHT_PAREN) {
		args = append(args, p.expression())
		for p.match(COMMA) {
			args = append(args, p.expression())
		}
	}

	p.consume(RIGHT_PAREN, "Expected ')' after arguments")

	return &CallExpr{callee: callee, args: args}
}

func (p *Parser) primary() Expr {
	expr := &LiteralExpr{}

	switch {
	case p.match(TRUE):
		expr.value = "true"
	case p.match(FALSE):
		expr.value = "false"
	case p.match(NIL):
		expr.value = "nil"
	case p.match(NUMBER):
		expr.value = p.previous().Literal
	case p.match(STRING):
		expr.value = p.previous().Literal
	case p.match(LEFT_PAREN):
		group := p.expression()
		p.consume(RIGHT_PAREN, "Expected ')' after expression")
		return &GroupExpr{group: group}
	case p.match(IDENTIFIER):
		// TODO: maybe VariableExpr should be renamed to IdentifierExpr
		return &VariableExpr{name: p.previous()}
	default:
		p.error("Expected an expression")
	}

	expr.token = p.previous()
	return expr
}

// --------------- Helper Functions --------------- //

// Check if any of the types match the current token type, advances if true.
func (p *Parser) match(types ...TokenType) bool {
	for _, typ := range types {
		if p.check(typ) {
			p.advance()
			return true
		}
	}
	return false
}

func (p *Parser) consume(typ TokenType, msg string) Token {
	if p.current().Type != typ {
		p.error(msg)
	}
	tok := p.current()
	p.advance()
	return tok
}

// Checks the current token, does not advance.
func (p *Parser) check(typ TokenType) bool {
	return !p.atEnd() && p.current().Type == typ
}

func (p *Parser) advance() Token {
	tok := p.current()
	if !p.atEnd() {
		p.idx++
	}
	return tok
}

func (p *Parser) atEnd() bool {
	return p.current().Type == EOF
}

func (p *Parser) current() Token {
	return p.tokens[p.idx]
}

func (p *Parser) previous() Token {
	if p.idx > 0 {
		return p.tokens[p.idx-1]
	} else {
		return p.current()
	}
}

func (p *Parser) error(msg string) {
	tok := p.tokens[p.idx]
	fmt.Fprintf(os.Stderr, "[line %d] Error at '%s': %s\n", tok.Line, tok.Lexeme, msg)
	os.Exit(65)
}
