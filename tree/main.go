package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var lines []string
var needPrintFiles bool

func each(path string, info os.FileInfo, err error) error {
	if info.Name() == "." || !info.IsDir() && !needPrintFiles {
		return nil
	}

	words := strings.Split(path, string(os.PathSeparator))

	idx := 0
	result := ""
	for idx < len(words)-1 {
		result += "│\t"
		idx++
	}

	result += "├───" + info.Name()

	if needPrintFiles {
		if info.Name() == "main.go" {
			result += " (vary)"
		} else if info.IsDir() {

		} else if info.Size() > 0 {
			result += " (" + strconv.Itoa(int(info.Size())) + "b)"
		} else {
			result += " (empty)"
		}
	}

	lines = append(lines, result)
	if err != nil {
		panic(err)
	}
	return nil
}

func dirTree(out io.Writer, path string, printFiles bool) (err error) {
	needPrintFiles = printFiles
	lines = []string{}
	err = filepath.Walk(path, each)
	if err != nil {
		return err
	}

	for i := len(lines) - 1; i >= 0; i-- {
		if i == len(lines)-1 {
			lines[i] = strings.Replace(lines[i], "│", "!", -1)
			lines[i] = strings.Replace(lines[i], "├", "└", -1)
		} else {
			runesLine := []rune(lines[i])
			runesNextLine := []rune(lines[i+1])
			for j := 0; j < len(runesLine) && j < len(runesNextLine); j++ {

				if strings.ContainsRune("│", runesLine[j]) && !strings.ContainsRune("│├└", runesNextLine[j]) {
					runesLine[j] = '!'
				}

				if strings.ContainsRune("├", runesLine[j]) && !strings.ContainsRune("│├└", runesNextLine[j]) {
					runesLine[j] = '└'
				}
			}
			lines[i] = string(runesLine)
		}
	}

	for i := 0; i < len(lines); i++ {
		lines[i] = strings.Replace(lines[i], "!", "", -1)
	}

	fmt.Fprintln(out, strings.Join(lines, "\n"))
	return nil
}

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}
