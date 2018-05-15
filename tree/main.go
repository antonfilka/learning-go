package main

import (
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"strings"
)

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

func dirTree(out io.Writer, dirPath string, printFiles bool) error {

	isLastParents := make([]bool, 0)

	err := RecursiveDirRead(out, dirPath, printFiles, dirPath, isLastParents)

	if err != nil {
		return fmt.Errorf("programm error: %v", err)
	}

	return nil

}

func RecursiveDirRead(out io.Writer, dirPath string, printFiles bool, initialPath string, isLastParents []bool) error {
	nestedLevel := GetDirLevel(dirPath, initialPath)

	root, err := os.Open(dirPath)
	if err != nil {
		return fmt.Errorf("error opening dir: %v", err)
	}

	currDir, err := root.ReadDir(-1)
	if err != nil {
		return fmt.Errorf("error reading dir: %v", err)
	}

	var filteredDir []os.DirEntry

	for _, val := range currDir {
		if val.IsDir() || (!val.IsDir() && printFiles) {
			filteredDir = append(filteredDir, val)
		}
	}

	sort.Slice(filteredDir, func(i, j int) bool {
		return filteredDir[i].Name() < filteredDir[j].Name()
	})

	for i, file := range filteredDir {

		isLastElement := i == len(filteredDir)-1

		if file.IsDir() {
			printDir(out, nestedLevel, isLastElement, isLastParents, file)

			newIsLastParents := make([]bool, len(isLastParents)+1)
			newIsLastParents[nestedLevel] = isLastElement
			copy(newIsLastParents, isLastParents)

			err = RecursiveDirRead(out, path.Join(dirPath, file.Name()), printFiles, initialPath, newIsLastParents)
			if err != nil {
				return fmt.Errorf("error reading dir: %v", err)
			}
		} else if printFiles {
			if file.Name() == ".DS_Store" {
				continue
			}
			printFile(out, nestedLevel, isLastElement, isLastParents, file)
		}

	}

	return nil
}

func GetDirLevel(dirPath string, initialPath string) (level int) {
	initialPathSlice := strings.Split(path.Clean(initialPath), string(os.PathSeparator))
	pathSlice := strings.Split(path.Clean(dirPath), string(os.PathSeparator))

	if pathSlice[0] == initialPathSlice[0] && len(pathSlice) == len(initialPathSlice) {
		return 0
	}

	if initialPathSlice[0] == "." && len(initialPathSlice) == 1 {
		return len(pathSlice) - (len(initialPathSlice) - 1)
	}

	return len(pathSlice) - len(initialPathSlice)
}

func printFile(out io.Writer, level int, isLastElement bool, isLastParents []bool, file os.DirEntry) error {
	fileInfo, err := file.Info()

	if err != nil {
		return fmt.Errorf("failed getting file info %v", err)
	}

	indent := getIndent(level, isLastElement, isLastParents)

	formattedFileSize := fmt.Sprintf("(%db)", fileInfo.Size())
	if formattedFileSize == "(0b)" {
		formattedFileSize = "(empty)"
	}

	formattedString := fmt.Sprintf("%s%s %s\n", indent, fileInfo.Name(), formattedFileSize)

	_, err = fmt.Fprintf(out, "%s", formattedString)

	if err != nil {
		return fmt.Errorf("failed printing file info %v", err)
	}

	return nil
}

func printDir(out io.Writer, level int, isLastElement bool, isLastParents []bool, dir os.DirEntry) error {

	indent := getIndent(level, isLastElement, isLastParents)
	formattedString := fmt.Sprintf("%s%s\n", indent, dir.Name())

	_, err := fmt.Fprintf(out, "%s", formattedString)

	if err != nil {
		return fmt.Errorf("failed printing dir info %v", err)
	}

	return nil
}

func getIndent(level int, isLastElement bool, isLastParents []bool) string {
	spacer := "───"
	lineBeginner := "├"
	if isLastElement {
		lineBeginner = "└"
	}


	
	ownIndent := lineBeginner + spacer

	combinedSpacers := ""
	i := 0
	if level > 0 {
		combinedSpacers = ""
		for i < level {
			if !isLastParents[i] {
				combinedSpacers += "│\t"
			} else {
				combinedSpacers += "\t"
			}
			i++
		}
	}

	indent := combinedSpacers + ownIndent

	return indent
}
