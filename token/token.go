package token



type TokenType string

// Not Perfomant but easy to use
// Using integer or byte would be more performant



type Token struct{
	Type TokenType
	Literal string
}

var keywords = map[string]TokenType{
	"fn": FUNCTION,
	"let": LET,
	"True": TRUE,
	"False": FALSE,
	"if": IF,
	"else": ELSE,
	"return": RETURN,

}

func LookupIdent(ident string) TokenType{
	if tok, ok := keywords[ident]; ok{
		return tok
	}
	return IDENT
}


const (
	ILLEGAL = "ILLEGAL"// Not known Token
	EOF = "EOF" //End of Life, stop parsing

	IDENT = "IDENT" //TokenType for a Variable
	INT = "INT"

	ASSIGN = "="
	PLUS = "+"
	MINUS = "-"
	BANG = "!"
	ASTERISK = "*"
	SLASH = "/"
	LT = "<"
	GT = ">"
	EQ = "=="
	NOT_EQ = "!="


	COMMA = ","
	SEMICOLON = ";"

	LPAREN = "("
	RPAREN = ")"
	LBRACE = "{"
	RBRACE = "}"

	FUNCTION = "FUNCTION"
	RETURN = "RETURN"
	LET = "LET"

	IF = "IF"
	ELSE = "ELSE"

	TRUE = "TRUE"
	FALSE = "FALSE"


)