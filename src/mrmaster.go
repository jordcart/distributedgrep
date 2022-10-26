package main

import "./mr"
import "time"
import "os"
import "fmt"
import "regexp"

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: mrmaster regex/searchstring inputfiles...\n")
		os.Exit(1)
	}

	pattern := os.Args[1]
	filesToGrep := os.Args[2:]

	_, error := regexp.Compile(pattern)

	if error != nil {
		fmt.Fprintf(os.Stderr, "distributedgrep: invalid regular expression\n")
		os.Exit(1)
	}

	m := mr.MakeMaster(pattern, filesToGrep, 10)
	for m.Done() == false {
		time.Sleep(time.Second)
	}

	time.Sleep(time.Second)
}
