package repl


import (
	"bufio"
	"fmt"
	"io"
	"MyInterpreter/lexer"
	"MyInterpreter/parser"
)


const PROMPT = ">>"

const MONKEY_FACE = `
						 __,__
					.--. .-" "-. .--.
					/ .. \/ .-. .-. \/ .. \
					| | '| / Y \ |' | |
					| \ \ \ 0 | 0 / / / |
					\ '- ,\.-"""""""-./, -' /
					''-' /_ ^ ^ _\ '-''
					| \._ _./ |
					\ \ '~' / /
					'._ '-=-' _.'
					'-----'
 ` 




func Start(in io.Reader, out io.Writer){
	scanner := bufio.NewScanner(in)

	for {
		fmt.Printf(PROMPT)
		scanned := scanner.Scan()
		if !scanned{
			return
		}

		line := scanner.Text()
		l := lexer.NewLexer(line)
		p := parser.NewParser(l)
		program := p.ParseProgram()

		if len(p.Errors()) > 1{
			printParserErrors(out, p.Errors())
			continue
		}
		io.WriteString(out, program.String())
		io.WriteString(out, "\n")
	}
}


func printParserErrors(out io.Writer, errors []string){
	io.WriteString(out, MONKEY_FACE)
	io.WriteString(out, "WOMP WOMP! PARSER IS NOT HAPPY")
	for _, msg := range errors{
		io.WriteString(out, "\t"+msg+"\n")
	}
}