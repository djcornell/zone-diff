package main

import (
	"bufio"
	"compress/gzip"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {

	zoneFile1 := flag.String("1", "", "first zone file to compare")
	zoneFile2 := flag.String("2", "", "second zone file to compare")

	flag.Parse()

	// open both zonefiles. They are gzipped so we need to open them with a gzip reader
	f1, err := os.Open(*zoneFile1)
	if err != nil {
		log.Fatal(err)
	}
	defer f1.Close()

	f2, err := os.Open(*zoneFile2)
	if err != nil {
		log.Fatal(err)
	}
	defer f2.Close()

	f1gz, err := gzip.NewReader(f1)

	if err != nil {
		log.Fatal(err)
	}
	defer f1gz.Close()

	f2gz, err := gzip.NewReader(f2)

	if err != nil {
		log.Fatal(err)
	}
	defer f2gz.Close()

	// read through each zone file line by line, comparring as we go, only printing out the differences
	// if the lines are the same, we don't print anything. if the left line is lexiographically greater
	// than the right line, we print the left line as a deletion. If the right line is lexiographically
	// greater than the left line, we print the right line as an addition.

	scanner1 := bufio.NewScanner(f1gz)
	scanner2 := bufio.NewScanner(f2gz)

	scanner1.Scan()
	scanner2.Scan()

	line1 := scanner1.Text()
	line2 := scanner2.Text()

	s1 := true
	s2 := true

	// as long as we have more lines in both files, we keep comparing them
	for s1 || s2 {

		// if the lines are the same, skip to the next line on both files
		if line1 == line2 {
			s1 = scanner1.Scan()
			s2 = scanner2.Scan()
			line1 = scanner1.Text()
			line2 = scanner2.Text()
			continue
		}

		// if the left line is lexically greater than the right line, print the the left line as a deletion
		// and do not advance the reader on the right line
		if line1 > line2 && s2 {
			add(line2)
			s2 = scanner2.Scan()
			line2 = scanner2.Text()
			continue
		}

		// conversely, if the right line is lexically greater than the left line, print the right line
		// as an addition and do not advance the reader on the left line
		if line1 < line2 && s1 {
			remove(line1)
			s1 = scanner1.Scan()
			line1 = scanner1.Text()
			continue
		}

		// if we get here, we have to drain the remaining lines from the left file
		if s1 {
			remove(line1)
			s1 = scanner1.Scan()
			line1 = scanner1.Text()
			continue
		}

		if s2 {
			add(line2)
			s2 = scanner2.Scan()
			line2 = scanner2.Text()
			continue
		}

	}
}

func add(line string) {
	parts, err := parts(line)

	if err != nil {
		//log.Println(err)
		return
	}

	// print parts
	fmt.Printf("+ %s %s\n", parts[0], parts[4])

}

func remove(line string) {
	parts, err := parts(line)

	if err != nil {
		//log.Println(err)
		return
	}

	// print parts
	fmt.Printf("- %s %s\n", parts[0], parts[4])
}

func parts(line string) ([]string, error) {
	fields := strings.Fields(line)

	if len(fields) != 5 {
		return nil, errors.New("line has incorrect number of fields: " + line)
	}

	return fields, nil
}
