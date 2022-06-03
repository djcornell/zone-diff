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
	gz := flag.Bool("gz", false, "treat files as gzip compressed")
	v := flag.Bool("v", false, "verbose output")

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

	scanner1 := bufio.NewScanner(f1)
	scanner2 := bufio.NewScanner(f2)

	if *gz {

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

		scanner1 = bufio.NewScanner(f1gz)
		scanner2 = bufio.NewScanner(f2gz)

	}

	// read through each zone file line by line, comparring as we go, only printing out the differences
	// if the lines are the same, we don't print anything. if the left line is lexiographically greater
	// than the right line, we print the left line as a addition. If the right line is lexiographically
	// greater than the left line, we print the right line as an deletion.

	scanner1.Scan()
	scanner2.Scan()

	line1 := scanner1.Text()
	line2 := scanner2.Text()

	parts1 := []string{}
	parts2 := []string{}

	s1 := true
	s2 := true

	// as long as we have more lines in both files, we keep comparing them
	for s1 || s2 {

		// we only care about lines that are NS record entries
		if line1 != "" {
			parts1, err = parts(line1)
			if err != nil {
				if *v {
					log.Println(err)
				}
				s1 = scanner1.Scan()
				line1 = scanner1.Text()
				continue
			}
		}

		if line2 != "" {
			parts2, err = parts(line2)
			if err != nil {
				if *v {
					log.Println(err)
				}
				s2 = scanner2.Scan()
				line2 = scanner2.Text()
				continue
			}
		}

		// if the lines are the same, skip to the next line on both files
		if line1 == line2 {
			//fmt.Println("same")
			s1 = scanner1.Scan()
			s2 = scanner2.Scan()
			line1 = scanner1.Text()
			line2 = scanner2.Text()
			continue
		}

		// if the left line is lexically greater than the right line, print the the left line as a deletion
		// and do not advance the reader on the right line
		if line1 > line2 && s2 {
			//fmt.Println("left > right '%s' > '%s'", line1, line2)
			add(parts2)
			s2 = scanner2.Scan()
			line2 = scanner2.Text()
			continue
		}

		// conversely, if the right line is lexically greater than the left line, print the right line
		// as an addition and do not advance the reader on the left line
		if line1 < line2 && s1 {
			//fmt.Printf("right > left: '%s' '%s'\n", line1, line2)
			remove(parts1)
			s1 = scanner1.Scan()
			line1 = scanner1.Text()
			continue
		}

		// if we get here, we have to drain the remaining lines from the left file
		if s1 {
			fmt.Println("draining s1")
			remove(parts1)
			s1 = scanner1.Scan()
			line1 = scanner1.Text()
			continue
		}

		// or the right side file
		if s2 {
			fmt.Println("draining s2")
			add(parts2)
			s2 = scanner2.Scan()
			line2 = scanner2.Text()
			continue
		}

		log.Fatal("unhanlded case")
	}
}

func add(parts []string) {
	fmt.Printf("+ %s %s\n", parts[0], parts[4])
}

func remove(parts []string) {
	fmt.Printf("- %s %s\n", parts[0], parts[4])
}

func parts(line string) ([]string, error) {
	fields := strings.Fields(line)

	if len(fields) != 5 {
		return nil, errors.New("line has incorrect number of fields: " + line)
	}

	if fields[3] != "ns" {
		return nil, errors.New("line is not an NS record: " + line)
	}

	return fields, nil
}
