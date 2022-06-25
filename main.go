package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {

	//Retrieve file name parameters from command line
	var (
		source_file string
		target_file string
	)
	flag.StringVar(&source_file, "source", "", "The Source file to be searched")
	flag.StringVar(&target_file, "target", "", "The Target File to be founded")
	flag.Parse()
	seen := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) { seen[f.Name] = true })
	if !(seen["source"] && seen["target"]) {
		flag.PrintDefaults()
		os.Exit(2)
	}

	chSource := openFileChannel(source_file)
	chTarget := openFileChannel(target_file)

	//Main loop
	sourceLine, isSourceActive := <-chSource
	targetLine, isTargetActive := <-chTarget
	for isSourceActive || isTargetActive {
		sourceLineFields := strings.Split(sourceLine, ";")
		targetLineFields := strings.Split(targetLine, ";")
		var action, value string
		var nextSource, nextTarget bool
		if !isTargetActive {
			action, value, nextSource, nextTarget = "outOfTarget", sourceLine, true, false
		} else if !isSourceActive {
			action, value, nextSource, nextTarget = "outOfSource", targetLine, false, true
		} else {
			if targetLineFields[0] == sourceLineFields[0] {
				value_ := strings.Join(([]string{sourceLineFields[0], sourceLineFields[1], targetLineFields[1]}), ";")
				action, value, nextSource, nextTarget = "match", value_, false, true //false & true to find 1:N next target for same souce
			} else if targetLine > sourceLine {
				action, value, nextSource, nextTarget = "notFound", sourceLine, true, false
			} else if targetLine < sourceLine {
				action, value, nextSource, nextTarget = "unUsed", targetLine, false, true
			}
		}
		if nextSource {
			sourceLine, isSourceActive = <-chSource
		}
		if nextTarget {
			targetLine, isTargetActive = <-chTarget
		}
		fmt.Println(value + "," + action)
	}
}

func openFileChannel(file string) <-chan string {
	ch := make(chan string)
	go func(ch chan string) {
		defer close(ch)
		f, err := os.Open(file)
		if err != nil {
			log.Printf("Could not open file: %v. %v", file, err)
			return
		}
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			ch <- scanner.Text()
		}
		if err := scanner.Err(); err != nil {
			log.Printf("Error while scanning file: %v. %v", file, err)
		}
	}(ch)
	return ch
}
