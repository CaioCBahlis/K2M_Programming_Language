package lexer

import (
	"MyInterpreter/token"
	_ "fmt"
	"unicode"
)

type Lexer struct {
	input        string // What It's gonna be reading
	position     int    // Lexer's reading position relative to the input
	readPosition int    // Pointer to the next char being read
	ch           byte   // Char being read
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0 // Stop Parsing, ch-0 means EOF in NextToken
	} else {
		l.ch = l.input[l.readPosition] //read the token in the next position
	}
	l.position = l.readPosition
	l.readPosition += 1

}

func (l *Lexer) NextToken() token.Token {
	//Gets Char, Converts into token format, Calls readchar
	//ReadChar goes sets reader to index + 1
	// returns token

	var tok token.Token
	l.skipWhiteSpace()
	switch l.ch {

	case '=':
		//TODO Abstraction that gets double chars symbols
		// Done :)
		tok = l.GetMultiCharToken(token.ASSIGN, token.EQ)
	case '+':
		tok = l.GetMultiCharToken(token.PLUS, token.PE)
	case '-':
		tok = l.GetMultiCharToken(token.MINUS, token.LE)
	case '!':
		//TODO Abstraction that gets double chars symbols
		// DONE :)
		tok = l.GetMultiCharToken(token.BANG, token.NOT_EQ)
	case '/':
		tok = l.GetMultiCharToken(token.SLASH, token.DE)
	case '*':
		tok = l.GetMultiCharToken(token.ASTERISK, token.ME, token.EXPONENT)
	case '<':
		tok = newToken(token.LT, l.ch)
	case '>':
		tok = newToken(token.GT, l.ch)
	case ';':
		tok = newToken(token.SEMICOLON, l.ch)
	case ',':
		tok = newToken(token.COMMA, l.ch)
	case '{':
		tok = newToken(token.LBRACE, l.ch)
	case '}':
		tok = newToken(token.RBRACE, l.ch)
	case '(':
		tok = newToken(token.LPAREN, l.ch)
	case ')':
		tok = newToken(token.RPAREN, l.ch)
	case '[':
		tok = newToken(token.LBRACKET, l.ch)
	case ']':
		tok = newToken(token.RBRACKET, l.ch)
	case '"':
		tok.Type = token.STRING
		tok.Literal = l.readString()
	case ':':
		tok = newToken(token.COLON, l.ch)
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier() //reads all grouped letters, and returns the word they form
			tok.Type = token.LookupIdent(tok.Literal)
			return tok
		} else if isDigit(l.ch) {
			tok.Literal = l.readNumber()
			tok.Type = token.INT
			return tok
		} else {
			tok = newToken(token.ILLEGAL, l.ch)
			//if a token is not a recognized symbol or a letter, it has to be wrong
		}
	}

	l.readChar()
	return tok
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) { //Letters are more likely to appear together
		l.readChar() // Reads letters together (word in college/technical terms)
	} //returns input for NextToken after no more letters

	return l.input[position:l.position]
}

func (l *Lexer) skipWhiteSpace() {
	for unicode.IsSpace(rune(l.ch)) {
		l.readChar()
	}

}

func (l *Lexer) readNumber() string {
	position := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	} else {
		return l.input[l.readPosition]
	}
}

func (l *Lexer) readString() string {
	position := l.position + 1

	for {
		l.readChar()
		if l.ch == '"' || l.ch == 0 {
			break
		}
	}
	return l.input[position:l.position]
}

func newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}

func NewLexer(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
	//every char added here is going to be interpreted as a letter, this impact var names and keywords
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func (l *Lexer) GetMultiCharToken(CurChar token.TokenType, ExpectChar ...token.TokenType) token.Token {
	for _, tokens := range ExpectChar {
		if l.peekChar() == tokens[1] {
			ch := l.ch
			l.readChar()
			return token.Token{Type: tokens, Literal: string(ch) + string(l.ch)}
		}
	}
	return token.Token{Type: CurChar, Literal: string(l.ch)}
}
