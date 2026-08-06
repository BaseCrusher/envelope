package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/BaseCrusher/envelope"
)

func main() {
	prefix := flag.String("prefix", "", "only convert variables whose name starts with this prefix; the prefix is stripped from the key")
	flag.Parse()

	out, err := envelope.Marshal(os.Environ(), *prefix)
	if err != nil {
		fmt.Fprintln(os.Stderr, "envelope:", err)
		os.Exit(1)
	}
	os.Stdout.Write(out)
}
