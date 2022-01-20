package main

import (
	"fmt"
	"io"
	"os"
	"sort"
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

func dirTree(out io.Writer, path string, printFiles bool) error {
	err := getTreeItems(out, path, printFiles, "")
	if err != nil {
		return err
	}

	return nil
}

func getTreeItems(out io.Writer, path string, printFiles bool, prefix string) error {

	var dirs []os.FileInfo
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	fileInfo, err := f.Readdir(-1)
	if err != nil {
		return err
	}
	err = f.Close()

	if !printFiles {
		for _, file := range fileInfo {
			if file.IsDir() {
				dirs = append(dirs, file)
			}
		}
		fileInfo = dirs
	}

	sort.Slice(fileInfo, func(i, j int) bool {
		return fileInfo[i].Name() < fileInfo[j].Name()
	})
	lastIndex := len(fileInfo) - 1
	// length := len(fileInfo)
	var isLast bool
	for i, file := range fileInfo {
		if i == lastIndex {
			isLast = true
		}

		var pref string
		if file.IsDir() {

			if !isLast {

				fmt.Fprintf(out, prefix+"├───"+"%s\n", file.Name())
				pref = prefix + "│\t"

			} else {
				fmt.Fprintf(out, prefix+"└───"+"%s\n", file.Name())
				pref = prefix + "\t"
			}

		} else {
			if !isLast {
				if file.Size() > 0 {
					fmt.Fprintf(out, prefix+"├───%s (%vb)\n", file.Name(), file.Size())
				} else {
					fmt.Fprintf(out, prefix+"├───%s (empty)\n", file.Name())
				}
			} else {
				if file.Size() > 0 {
					fmt.Fprintf(out, prefix+"└───%s (%vb)\n", file.Name(), file.Size())
				} else {
					fmt.Fprintf(out, prefix+"└───%s (empty)\n", file.Name())
				}
			}

		}
		getTreeItems(out, path+"/"+file.Name(), printFiles, pref)

	}
	return err
}
