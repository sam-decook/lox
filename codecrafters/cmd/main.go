package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Usage: ./your_program.sh [tokenize | parse | evaluate | run] <filename>")
		os.Exit(1)
	}

	command := os.Args[1]
	filename := os.Args[2]

	lox := Interpreter{}
	lox.Scan(filename)

	switch command {
	case "tokenize":
		for _, token := range lox.tokens {
			fmt.Println(token.String())
		}

	case "parse":
		lox.Parse()
		fmt.Println(lox.ast.String())

	case "evaluate":
		// Evaluate is a special case, since it only parses an expression
		lox.parser.tokens = lox.tokens
		ast := lox.parser.expression()
		res := ast.Evaluate(&lox.env)
		if res == nil {
			fmt.Println("nil")
		} else {
			fmt.Println(res)
		}

	case "run":
		lox.Parse()
		lox.Evaluate()

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		os.Exit(1)
	}

	if lox.scanner.lexicalError {
		os.Exit(65)
	}
}
