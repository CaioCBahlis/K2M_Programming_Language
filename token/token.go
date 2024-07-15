package token

type TokenType string

// Not Perfomant but easy to use
// Using integer or byte would be more performant
// Future Optimizations

type Token struct {
	Type    TokenType
	Literal string
}

var keywords = map[string]TokenType{
	"fn":     FUNCTION,
	"let":    LET,
	"True":   TRUE,
	"False":  FALSE,
	"if":     IF,
	"else":   ELSE,
	"return": RETURN,
	"while":  WHILE,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
	//there's no way of knowing if a word is the name of a var or a keyword
	// here, we just if the word is on the keyword list, if not, it must be the ident
	// of a var or function
}

const (
	ILLEGAL = "ILLEGAL" // Not known Token
	EOF     = "EOF"     //End of Life, stop parsing

	IDENT  = "IDENT" //TokenType for a Variable
	INT    = "INT"
	STRING = "STRING"

	ASSIGN   = "="
	PLUS     = "+"
	MINUS    = "-"
	BANG     = "!"
	ASTERISK = "*"
	SLASH    = "/"
	LT       = "<"
	GT       = ">"
	EQ       = "=="
	NOT_EQ   = "!="
	PE       = "+="
	ME       = "*="
	DE       = "/="
	LE       = "-="
	EXPONENT = "**"

	COMMA     = ","
	SEMICOLON = ";"
	COLON     = ":"

	LPAREN   = "("
	RPAREN   = ")"
	LBRACE   = "{"
	RBRACE   = "}"
	LBRACKET = "["
	RBRACKET = "]"

	FUNCTION = "FUNCTION"
	RETURN   = "RETURN"
	LET      = "LET"

	IF   = "IF"
	ELSE = "ELSE"

	TRUE  = "TRUE"
	FALSE = "FALSE"

	WHILE = "WHILE"
)
