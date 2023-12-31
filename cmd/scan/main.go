package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/midbel/sweet"
)

func main() {
	dialect := flag.String("d", "", "dialect")
	flag.Parse()

	r, err := os.Open(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	defer r.Close()

	scan, err := sweet.Scan(r, sweet.KeywordsForDialect(*dialect))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for {
		tok := scan.Scan()
		if tok.Type == sweet.Invalid {
			fmt.Fprintf(os.Stderr, "invalid token found at %s", tok.Position)
			fmt.Fprintln(os.Stderr)
			os.Exit(1)
		}
		if tok.Type == sweet.EOF {
			break
		}
		fmt.Println(tok.Position, tok)
	}
}
