package main

//
// a file grep application "plugin" for MapReduce.
//
// go build -buildmode=plugin grep.go
//

import "../mr"
import "strings"
import "bufio"
import "fmt"
import "regexp"

//
// The map function is called once for each file of input. The first
// argument is the name of the input file, and the second is the
// file's complete contents. The return value is a slice containing
// the filename and the line contents
//
func Map(filename string, contents string, pattern string) []mr.KeyValue {
	regexPattern, _ := regexp.Compile(pattern)
	kva := []mr.KeyValue{}

	// scan string line by line checking for regex match
    sc := bufio.NewScanner(strings.NewReader(contents))
	lineNumber := 1
    for sc.Scan() {
		cur_line := sc.Text()
		if regexPattern.MatchString(cur_line) {
			kv := mr.KeyValue{filename, cur_line}
			kva = append(kva, kv)
		}

		lineNumber++
    }

	return kva
}

//
// The reduce function is called once for each key generated by the
// any map task.
//
func Reduce(key string, values string) string {
	return fmt.Sprintf("%v %v\n", key, values)
}
