package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Nigh/transliterate/pkg/transliterate"
)

var (
	help        bool
	inputPath   string
	isDry       bool
	skipComfirm bool
	isSilent    bool
)
var trans func(string) string

type FilePath string
type FileRanames struct {
	path    FilePath
	oldName string
	newName string
	isDir   bool
}
type FilePaths []FileRanames
type FilePathSlice []string

func (a FilePath) Depth() (depth int) {
	slice := strings.Split(string(a), string(filepath.Separator))
	for _, v := range slice {
		if len(v) > 0 {
			depth++
		}
	}
	return
}

func (a FilePaths) Len() int      { return len(a) }
func (a FilePaths) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a FilePaths) Less(i, j int) bool {
	di := a[i].path.Depth()
	dj := a[j].path.Depth()
	return di > dj
}

var file2Rename FilePaths

func init() {
	flag.BoolVar(&help, "help", false, "help")
	flag.BoolVar(&isDry, "dry", false, "run without actually rename files")
	flag.BoolVar(&skipComfirm, "y", false, "skip comfirm")
	flag.BoolVar(&isSilent, "s", false, "no output print")

	flag.StringVar(&inputPath, "in", "", "input path")

	file2Rename = make(FilePaths, 0)
	trans = transliterate.Sugar("-", "")
}

func main() {
	flag.Parse()
	if help || len(inputPath) == 0 {
		flag.Usage()
		return
	}
	if len(inputPath) == 0 {
		flag.Usage()
		return
	}
	inputPath = filepath.Clean(inputPath)
	_, err := os.Stat(inputPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	filepath.Walk(inputPath, walker)
	if !isSilent {
		fmt.Println("There are total", len(file2Rename), "files to rename")
	}

	if len(file2Rename) > 0 {
		sort.Sort(file2Rename)
		fmtStr := fmt.Sprintf("[%%0%dd] %%s %%s ---> %%s\n", int(math.Ceil(math.Log10(float64(len(file2Rename))))))
		for i, v := range file2Rename {
			var fType string
			if v.isDir {
				fType = "D"
			} else {
				fType = "F"
			}
			if !isDry {
				if !isSilent {
					fmt.Printf(fmtStr, i, fType, filepath.Clean(string(v.path)+string(filepath.Separator)+v.oldName), v.newName)
				}
				old := filepath.Clean(string(v.path) + string(filepath.Separator) + v.oldName)
				new := filepath.Clean(string(v.path) + string(filepath.Separator) + v.newName)
				os.Rename(old, new)
			}
		}
	}
}

func walker(realPath string, f os.FileInfo, err error) error {
	ext := filepath.Ext(f.Name())
	oldName := strings.TrimSuffix(f.Name(), filepath.Ext(f.Name()))
	newName := trans(oldName)

	if !isSilent {
		if f.IsDir() {
			fmt.Printf("[D] %s", realPath)
			if oldName != newName {
				fmt.Printf(" -> %s", newName+ext)
			}
			fmt.Printf("\n")
		} else {
			if oldName != newName {
				fmt.Printf("\t[F] %s -> %s\n", oldName+ext, newName+ext)
			}
		}
	}

	if oldName != newName {
		file2Rename = append(file2Rename, FileRanames{path: FilePath(filepath.Dir(realPath)), oldName: oldName + ext, newName: newName + ext, isDir: f.IsDir()})
	}
	return nil
}
