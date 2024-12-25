package main

import (
	"fmt"
	"sort"

	//"fmt"
	"io"
	"os"
	//"path/filepath"
	//"strings"
)

func dirTree(out io.Writer, path string, printFiles bool) error {
	return printDir(out, path, "", printFiles)
}

func printDir(out io.Writer, path, prefix string, printFiles bool) error {
	dir, err := os.Open(path)

	if err != nil {
		return fmt.Errorf("open %s: %v", path, err)
	}

	defer dir.Close()

	files, err := dir.Readdir(-1)

	if err != nil {
		return fmt.Errorf("readdir %s: %v", path, err)
	}

	sort.Slice(files, func(i, j int) bool { return files[i].Name() < files[j].Name() })

	if !printFiles {
		var dirsOnly []os.FileInfo
		for _, file := range files {
			if file.IsDir() {
				dirsOnly = append(dirsOnly, file)
			}
		}
		files = dirsOnly
	}

	for i, file := range files {
		connector := "├───"
		if i == len(files)-1 {
			connector = "└───"
		}

		if file.IsDir() {
			newPrefix := prefix + "\t"
			if i != len(files)-1 {
				newPrefix = prefix + "│\t"
			}

			fmt.Fprintf(out, "%s%s%s\n", prefix, connector, file.Name())

			err := printDir(out, path+string(os.PathSeparator)+file.Name(), newPrefix, printFiles)

			if err != nil {
				return err
			}
		} else if printFiles {
			size := "empty"

			if file.Size() > 0 {
				size = fmt.Sprint(file.Size())
				size = size + "b"
			}

			fmt.Fprintf(out, "%s%s%s (%s)\n", prefix, connector, file.Name(), size)
		}
	}

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
