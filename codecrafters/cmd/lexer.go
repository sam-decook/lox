package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Scanner struct {
	line         int //line number in file
	contents     []byte
	idx          int  //current spot in the source
	ch           byte //current character in the source
	lexicalError bool
}

func (s *Scanner) init(filename string) {
	contents, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	s.line = 1
	s.contents = contents
	s.idx = -1
	s.ch = 0
	s.lexicalError = false
}

// Returns false if at EOF
func (s *Scanner) next() bool {
	if s.idx == len(s.contents)-1 {
		return false
	}

	s.idx += 1
	s.ch = s.contents[s.idx]
	return true
}

// Returns the next byte (if there is one), but does not advance
func (s *Scanner) peek() byte {
	if s.idx == len(s.contents)-1 {
		return 0
	}

	return s.contents[s.idx+1]
}

func (s *Scanner) peekTwo() byte {
	if s.idx == len(s.contents)-2 {
		return 0
	}

	return s.contents[s.idx+2]
}

func (s *Scanner) comment() {
	for {
		if !s.next() || s.ch == '\n' {
			break
		}
	}
	s.line += 1
}

func (s *Scanner) stringLiteral() (string, bool) {
	start := s.idx

	for {
		if !s.next() {
			fmt.Fprintf(os.Stderr, "[line %d] Error: Unterminated string.", s.line)
			s.lexicalError = true
			return "", false
		} else if s.ch == '"' {
			break
		}
	}

	return string(s.contents[start : s.idx+1]), true
}

func (s *Scanner) numberLiteral() (string, string) {
	start := s.idx

	for isDigit(s.peek()) {
		s.next()
	}
	if s.peek() == '.' && isDigit(s.peekTwo()) {
		s.next()
	}
	for isDigit(s.peek()) {
		s.next()
	}

	lexeme := string(s.contents[start : s.idx+1])
	f, _ := strconv.ParseFloat(lexeme, 64)
	literal := fmt.Sprintf("%g", f)
	if !strings.Contains(literal, ".") {
		literal += ".0"
	}

	return lexeme, literal
}

func (s *Scanner) identifier() string {
	start := s.idx

	for isAlphaNumeric(s.peek()) {
		s.next()
	}

	return string(s.contents[start : s.idx+1])
}

func (s *Scanner) scan() []Token {
	toks := make([]Token, 0, len(s.contents)+1)

	for s.next() {
		switch s.ch {
		case ' ', '\t', '\r':
			//nothing
		case '\n':
			s.line += 1
		case '(':
			toks = append(toks, Token{Type: LEFT_PAREN, Lexeme: string(s.ch), Line: s.line})
		case ')':
			toks = append(toks, Token{Type: RIGHT_PAREN, Lexeme: string(s.ch), Line: s.line})
		case '{':
			toks = append(toks, Token{Type: LEFT_BRACE, Lexeme: string(s.ch), Line: s.line})
		case '}':
			toks = append(toks, Token{Type: RIGHT_BRACE, Lexeme: string(s.ch), Line: s.line})
		case ',':
			toks = append(toks, Token{Type: COMMA, Lexeme: string(s.ch), Line: s.line})
		case '.':
			toks = append(toks, Token{Type: DOT, Lexeme: string(s.ch), Line: s.line})
		case '-':
			toks = append(toks, Token{Type: MINUS, Lexeme: string(s.ch), Line: s.line})
		case '+':
			toks = append(toks, Token{Type: PLUS, Lexeme: string(s.ch), Line: s.line})
		case ';':
			toks = append(toks, Token{Type: SEMICOLON, Lexeme: string(s.ch), Line: s.line})
		case '*':
			toks = append(toks, Token{Type: STAR, Lexeme: string(s.ch), Line: s.line})
		case '/':
			if s.peek() == '/' {
				s.comment()
			} else {
				toks = append(toks, Token{Type: SLASH, Lexeme: string(s.ch), Line: s.line})
			}
		case '=':
			if s.peek() == '=' {
				s.next()
				toks = append(toks, Token{Type: EQUAL_EQUAL, Lexeme: "==", Line: s.line})
			} else {
				toks = append(toks, Token{Type: EQUAL, Lexeme: string(s.ch), Line: s.line})
			}
		case '!':
			if s.peek() == '=' {
				s.next()
				toks = append(toks, Token{Type: BANG_EQUAL, Lexeme: "!=", Line: s.line})
			} else {
				toks = append(toks, Token{Type: BANG, Lexeme: string(s.ch), Line: s.line})
			}
		case '<':
			if s.peek() == '=' {
				s.next()
				toks = append(toks, Token{Type: LESS_EQUAL, Lexeme: "<=", Line: s.line})
			} else {
				toks = append(toks, Token{Type: LESS, Lexeme: string(s.ch), Line: s.line})
			}
		case '>':
			if s.peek() == '=' {
				s.next()
				toks = append(toks, Token{Type: GREATER_EQUAL, Lexeme: ">=", Line: s.line})
			} else {
				toks = append(toks, Token{Type: GREATER, Lexeme: string(s.ch), Line: s.line})
			}
		case '"':
			str, found := s.stringLiteral()
			if found {
				toks = append(toks, Token{Type: STRING, Lexeme: str, Literal: strings.Trim(str, "\""), Line: s.line})
			}
		default:
			if isDigit(s.ch) {
				lexeme, literal := s.numberLiteral()
				toks = append(toks, Token{Type: NUMBER, Lexeme: lexeme, Literal: literal, Line: s.line})
			} else if isAlpha(s.ch) {
				ident := s.identifier()
				if r, found := reserved[ident]; found {
					toks = append(toks, Token{Type: r, Lexeme: ident, Line: s.line})
				} else {
					toks = append(toks, Token{Type: IDENTIFIER, Lexeme: ident, Line: s.line})
				}
			} else {
				fmt.Fprintf(os.Stderr, "[line %d] Error: Unexpected character: %s\n", s.line, string(s.ch))
				s.lexicalError = true
			}
		}
	}

	toks = append(toks, Token{Type: EOF, Line: s.line})
	return toks
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'z') ||
		c == '_'
}

func isAlphaNumeric(c byte) bool {
	return isAlpha(c) || isDigit(c)
}
