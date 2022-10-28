package main

import "./mr"
import "time"
import "os"
import "io/ioutil"
import "fmt"
import "regexp"
import "strconv"

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: mrmaster regex/searchstring inputfiles...\n")
		os.Exit(1)
	}

	pattern := os.Args[1]
	filesToGrep := os.Args[2:]
	numReduce := 10

	_, error := regexp.Compile(pattern)

	if error != nil {
		fmt.Fprintf(os.Stderr, "distributedgrep: invalid regular expression\n")
		os.Exit(1)
	}

	m := mr.MakeMaster(pattern, filesToGrep, numReduce)
	for m.Done() == false {
		time.Sleep(time.Second)
	}

	for i := 0; i < numReduce; i++ {
		filename := "mr-out-" + strconv.Itoa(i)
		dat, err := ioutil.ReadFile(filename)
		if err != nil {
			continue
		}
		fmt.Print(string(dat))
	}

	time.Sleep(time.Second)
}
