package main

import (
	ui "github.com/gizak/termui"
	"github.com/gizak/termui/widgets"
	"os"
	"path/filepath"
	//"fmt"
	time2 "time"
)

func main() {
	//displayedFileNames, fileNames, err := fileListCurrentDirectory()
	//if err != nil {
	//	log.Fatalf("Failed getting Current directory: %v", err)
	//}
	//if err := ui.Init(); err != nil {
	//	log.Fatalf("failed to initialize termui: %v", err)
	//}
	defer ui.Close()
	speak := initStatus()
	time2.Sleep(time2.Millisecond*2000)
	speak("Hey fucker")
}

func initStatus() func(string) {
	speaks := widgets.NewParagraph()
	speaks.SetRect(60, 0, 100, 10)
	speaks.Title = "Status"
	speaks.Text = `Programa Iniciado.

Programado en Go.
Patricio Whittingslow 2019`
	handle := func(statusString string) { // Closure is so cool
		speaks.Text = statusString
		ui.Render(speaks)
	}
	return handle
}

// fileListCurrentDirectory
func fileListCurrentDirectory() ([]string, []string, error) {
	var files []string
	root, err := filepath.Abs("./")
	if err != nil {
		return nil, nil, err
	}

	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {

		files = append(files, path)
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	// Ahora lo que hago es excluir la parte reduntante del dir
	// C:/Go/mydir/foo/myfile.exe  ----> se convierte a ---> foo/myfile.exe
	//const numberOfFiles
	var fileLength int
	maxFileLength := 0
	minFileLength := 2047

	i := 0
	var shortFileNames, actualFileNames []string
	shortFileNames = append(files[:0:0], files...)
	actualFileNames = append(files[:0:0], files...)

	for _, file := range files {
		fileLength = len(file)
		if fileLength > maxFileLength {
			maxFileLength = fileLength
		}
		if fileLength < minFileLength {
			minFileLength = fileLength
		}
		i++
	}
	permittedStringLength := 54
	i = 0

	for _, file := range files {
		if len(file) <= minFileLength {
			files = remove(files, i)
			shortFileNames = remove(shortFileNames, i)
			actualFileNames = remove(actualFileNames, i)
			continue
		}
		if len(file) > permittedStringLength+minFileLength {

			shortFileNames[i] = `~\…` + file[len(file)-permittedStringLength:]

		} else {
			shortFileNames[i] = "~" + file[minFileLength:]

		}
		actualFileNames[i] = "~" + file[minFileLength:]
		i++
	}
	return shortFileNames, actualFileNames, nil
}

func remove(s []string, i int) []string {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}
