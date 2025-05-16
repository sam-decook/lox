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
	lexicalError := lox.Scan(filename)

	switch command {
	case "tokenize":
		for _, token := range lox.tokens {
			fmt.Println(token.String())
		}

	case "parse":
		lox.Parse()
		fmt.Println(lox.ast.String())

	case "evaluate":
		// Evaluate is a special case, since it only parses expressions
		parser := Parser{}
		parser.tokens = lox.tokens
		ast := parser.expression()
		res := ast.Evaluate(&lox)
		// This check might be old, now that I'm using Objects
		if res == nil {
			fmt.Println("nil")
		} else {
			fmt.Println(res)
		}

	case "run":
		lox.Parse()
		lox.Resolve()
		lox.Evaluate()

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		os.Exit(1)
	}

	if lexicalError {
		os.Exit(65)
	}
}
