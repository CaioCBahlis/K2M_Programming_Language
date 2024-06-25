package lexer

import (
	"MyInterpreter/token"
	_ "fmt"
)

type Lexer struct {
	input        string //What It's gonna be reading
	position     int    // Where the lexer is in the input
	readPosition int    // pointer to the next char being read
	ch           byte   // Actual char that is being read
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0 // Stop Parsing, ch-0 means EOF in NextToken
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition += 1
}

func (l *Lexer) NextToken() token.Token{
	//Gets Char, Converts into token format, Calls readchar
	//ReadChar goes sets reader to index + 1
	// returns token
	
	var tok token.Token
	l.skipWhiteSpace()
	switch l.ch {

	case '=':
		//TO DO Abstraction that gets double chars symbols
		if l.peekChar() == '='{
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.EQ, Literal: string(ch) + string(l.ch)}
		}else{
			tok = newToken(token.ASSIGN, l.ch)
	}
	case '+':
		tok = newToken(token.PLUS, l.ch)
	case '-':
		tok = newToken(token.MINUS, l.ch)
	case '!':
		//TO DO Abstraction that gets double chars symbols
		if l.peekChar() == '='{
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.NOT_EQ, Literal: string(ch) + string(l.ch)}
		}else{
			tok = newToken(token.BANG, l.ch)
	}
	case '/':
		tok = newToken(token.SLASH, l.ch)
	case '*':
		tok = newToken(token.ASTERISK, l.ch)
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
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		if isLetter(l.ch){
			tok.Literal = l.readIdentifier() //reads all grouped letters, and returns the word they form
			tok.Type = token.LookupIdent(tok.Literal)
			return tok 
		} else if isDigit(l.ch){
			tok.Literal = l.readNumber()
			tok.Type = token.INT
			return tok
		}else{
			tok = newToken(token.ILLEGAL, l.ch)
			//if a token is not a recognized symbol or a letter, it has to be wrong
		}
	}

	l.readChar()
	return tok
}

func (l *Lexer) readIdentifier() string{
	position := l.position
	for isLetter(l.ch){ //Letters are more likely to appear together
		l.readChar()    // Reads letters together (word in college/technical terms) 
	}					//returns input for NextToken after no more letters
	
	return l.input[position:l.position]
}

func(l *Lexer) skipWhiteSpace(){
	for l.ch == ' '  || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}

}

func (l *Lexer) readNumber() string{
	position := l.position
	for isDigit(l.ch){
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) peekChar() byte{
	if l.readPosition >= len(l.input){
		return 0
	}else{
		return l.input[l.readPosition]
	}
}


func newToken(tokenType token.TokenType, ch byte) token.Token{
	return token.Token{Type: tokenType, Literal: string(ch)}
}

func NewLexer(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar() // ch is empty, if we pass the Lexer like it is
	return l	// Its going to parse null, instead of input[0]
}				// readChar() ensures that readposition aims at input[1]
				// and position and char are equal to input[0] (index and byte, respectively)


func isLetter(ch byte)bool{
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
	//every char added here is going to be interpreted as a letter, this impact var names and keywords
}

func isDigit(ch byte) bool{
	return '0' <= ch && ch <= '9'
}



