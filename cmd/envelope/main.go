package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/BaseCrusher/envelope"
)

func main() {
	prefix := flag.String("prefix", "", "only convert variables whose name starts with this prefix; the prefix is stripped from the key")
	out := flag.String("out", "", "write YAML to this file instead of stdout")
	flag.Parse()

	doc, err := envelope.Marshal(os.Environ(), *prefix)
	if err == nil {
		if *out == "" {
			_, err = os.Stdout.Write(doc)
		} else {
			err = os.WriteFile(*out, doc, 0o644)
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "envelope:", err)
		os.Exit(1)
	}
}
