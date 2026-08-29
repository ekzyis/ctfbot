// Command leet rewrites text in leetspeak.
//
// It reads text from the command-line arguments, or from stdin if none are
// given, substitutes lookalike digits for letters, and prints the result.
// Everything else — case of unmapped letters, spaces, punctuation — is left
// untouched, so the output can be padded/wrapped into a flag by hand.
//
//	$ leet you found the secret
//	y0u f0und 7h3 53cr37
//	$ echo "Capture The Flag" | leet
//	C4p7ur3 Th3 F149
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"
)

// leetMap substitutes letters with lookalike digits. Lookup is
// case-insensitive; every replacement stays within [0-9].
var leetMap = map[rune]rune{
	'a': '4', 'e': '3', 'i': '1', 'o': '0',
	's': '5', 't': '7', 'l': '1', 'g': '9', 'b': '8',
}

// leet rewrites s in leetspeak: each letter with a mapping becomes its digit
// (matched case-insensitively); all other runes are passed through unchanged.
func leet(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if sub, ok := leetMap[unicode.ToLower(r)]; ok {
			b.WriteRune(sub)
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: leet [text...]\n\n")
		fmt.Fprintf(os.Stderr, "Rewrites text in leetspeak. Reads the arguments, or stdin if none.\n")
	}
	flag.Parse()

	// With arguments, leetspeak them and print a line.
	if args := flag.Args(); len(args) > 0 {
		fmt.Println(leet(strings.Join(args, " ")))
		return
	}

	// Otherwise act as a stdin->stdout filter, preserving input verbatim
	// apart from the substitutions.
	in, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "leet: reading stdin: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(leet(string(in)))
}
