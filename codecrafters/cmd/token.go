package main

import "fmt"

type TokenType int

const (
	EOF TokenType = iota
	LEFT_PAREN
	RIGHT_PAREN
	LEFT_BRACE
	RIGHT_BRACE
	COMMA
	DOT
	MINUS
	PLUS
	SEMICOLON
	STAR
	SLASH
	EQUAL
	EQUAL_EQUAL
	BANG
	BANG_EQUAL
	LESS
	LESS_EQUAL
	GREATER
	GREATER_EQUAL
	STRING
	NUMBER
	IDENTIFIER
	AND
	CLASS
	ELSE
	FALSE
	FOR
	FUN
	IF
	NIL
	OR
	PRINT
	RETURN
	SUPER
	THIS
	TRUE
	VAR
	WHILE
)

var tokens = [...]string{
	EOF:           "EOF",
	LEFT_PAREN:    "LEFT_PAREN",
	RIGHT_PAREN:   "RIGHT_PAREN",
	LEFT_BRACE:    "LEFT_BRACE",
	RIGHT_BRACE:   "RIGHT_BRACE",
	COMMA:         "COMMA",
	DOT:           "DOT",
	MINUS:         "MINUS",
	PLUS:          "PLUS",
	SEMICOLON:     "SEMICOLON",
	STAR:          "STAR",
	SLASH:         "SLASH",
	EQUAL:         "EQUAL",
	EQUAL_EQUAL:   "EQUAL_EQUAL",
	BANG:          "BANG",
	BANG_EQUAL:    "BANG_EQUAL",
	LESS:          "LESS",
	LESS_EQUAL:    "LESS_EQUAL",
	GREATER:       "GREATER",
	GREATER_EQUAL: "GREATER_EQUAL",
	STRING:        "STRING",
	NUMBER:        "NUMBER",
	IDENTIFIER:    "IDENTIFIER",
	AND:           "AND",
	CLASS:         "CLASS",
	ELSE:          "ELSE",
	FALSE:         "FALSE",
	FOR:           "FOR",
	FUN:           "FUN",
	IF:            "IF",
	NIL:           "NIL",
	OR:            "OR",
	PRINT:         "PRINT",
	RETURN:        "RETURN",
	SUPER:         "SUPER",
	THIS:          "THIS",
	TRUE:          "TRUE",
	VAR:           "VAR",
	WHILE:         "WHILE",
}

var reserved = map[string]TokenType{
	"and":    AND,
	"class":  CLASS,
	"else":   ELSE,
	"false":  FALSE,
	"for":    FOR,
	"fun":    FUN,
	"if":     IF,
	"nil":    NIL,
	"or":     OR,
	"print":  PRINT,
	"return": RETURN,
	"super":  SUPER,
	"this":   THIS,
	"true":   TRUE,
	"var":    VAR,
	"while":  WHILE,
}

type Token struct {
	Type TokenType
	// The characters matched from the input
	Lexeme string
	// The value which will be used, e.g. 42.0 -> Type: NUMBER, Lexeme: 42.0, Literal: 42
	Literal string
	Line    int
}

func (t Token) String() string {
	lit := t.Literal
	if lit == "" {
		lit = "null"
	}
	return fmt.Sprintf("%s %s %s", tokens[t.Type], t.Lexeme, lit)
}
