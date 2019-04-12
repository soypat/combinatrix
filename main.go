package main

import (
	"errors"
	"fmt"
	ui "github.com/gizak/termui"
	"github.com/gizak/termui/widgets"
	"log"
	"os"
	"path/filepath"
	"time"
)

var versionNumber = "2019.10"

var isExiting = false

var colorWheel = []ui.Color{ui.ColorGreen, ui.ColorBlue, ui.ColorCyan, ui.ColorYellow, ui.ColorRed, ui.ColorMagenta}

var selectedTheme = 0
var askToPollKeyPress = make(chan string)

const welcomeText = `

Programado en Go.
Patricio Whittingslow 2019.`

func main() {
	go pollKeyboard()
	statusBulletin := NewBulletin()
	//splash(1) // TODO debug TEMP

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	defer close(askToPollKeyPress)

	InitBulletin(&statusBulletin)
	statusBulletin.Post(welcomeText)
	displayedFileNames, _, err := fileListCurrentDirectory(54) // TODO DEBUG fileNames
	if err != nil {
		log.Fatalf("Fail to read files Combinatrix: %v", err)
	}
	fileList := NewMenu()
	fileList.options = displayedFileNames
	fileList.fitting  = CreateFitting([3]int{0, 1, 0},[3]int{0, 1, 0},[3]int{1, 3, 0}, [3]int{1, 2, 0})
	fileList.title = "Seleccionar Archivo de materias"
	InitMenu(&fileList)

	askToPollFileList := make(chan bool)

	go fileList.Poller(askToPollFileList)
	defer close(askToPollFileList)
RESETCLASSSELECTION:
	askToPollFileList <- true

	for {
		if fileList.selection >= 0 && fileList.action == "<Enter>" {
			fileList.action = ""
			statusBulletin.color = ui.ColorClear
			statusBulletin.Post("Archivo seleccionado!\nLeyendo...")

			break
		}
		time.Sleep(20 * time.Millisecond)
		break // TODO DEBUG remove
	}
	myFile := "C:/work/gopath/src/github.com/soypat/Combinatrix/test/data.dat"
	Classes, err := GatherClasses(myFile) // TODO fix something in here. Returning same class twice
	//TODO uncomment below
	//fileDir := fileNames[fileList.selection]
	//Classes, err := GatherClasses(fileDir[2:]) // TODO DEBUG uncomment
	if err != nil {
		statusBulletin.Error("Error leyendo archivo.", err)
		fileList.selection = -1
		goto RESETCLASSSELECTION
	}
	ClassNames := []string{}
	for _, v := range Classes {
		ClassNames = append(ClassNames, v.name)
	}
	classMenu := NewMenu()
	classMenu.options = ClassNames
	classMenu.fitting = CreateFitting([3]int{1, 3, 0},[3]int{0, 1, 0},[3]int{2, 3, 0},[3]int{2, 3, 0})
	classMenu.title = "Clases halladas"
	InitMenu(&classMenu)
	askToPollClassList := make(chan bool)

	askToPollFileList <- false
	ui.Clear()
	close(askToPollFileList)
	statusBulletin.Post(`Se pueden borrar clases. Apretar ENTER para continuar...`)

	go classMenu.Poller(askToPollClassList)
	askToPollClassList <- true

	removedClassString := []string{}
	amountOfClassesRemoved := 0
	keepGoing := true

	for keepGoing {
		break // TODO debug remove

		switch classMenu.action {
		case "<Delete>":
			classMenu.action = ""
			//close(askToPollClassList)
			if len(classMenu.options) <= 1 {
				continue
			}

			removedClassString = append(removedClassString, classMenu.options[classMenu.selection])
			classMenu.options = removeS(classMenu.options, classMenu.selection)
			InitMenu(&classMenu)
			amountOfClassesRemoved++
		case "<C-z>", "<C-Z>":
			//classMenu.action =""
			if amountOfClassesRemoved > 0 {
				amountOfClassesRemoved--
				classMenu.options = append(classMenu.options, removedClassString[amountOfClassesRemoved])
				removedClassString = removeS(removedClassString, amountOfClassesRemoved)

				InitMenu(&classMenu)
			}
		case "d": //debug command
			fmt.Printf("%v\n", classMenu.options)
		case "<Enter>":
			keepGoing = false
		}
		classMenu.action = ""
		time.Sleep(20 * time.Millisecond)

	}
	ui.Clear()
	classMenu.fitting = CreateFitting([3]int{0, 1, 0},[3]int{0, 1, 0},[3]int{1, 3, 0},[3]int{3, 3, 0})
	InitMenu(&classMenu)

	var workingClasses []*Class
	var classRemoved bool
	if len(removedClassString) > 0 {
		for _, v := range Classes {
			classRemoved = false
			for _, s := range removedClassString {
				if v.name == s {
					classRemoved = true
				}
				if !classRemoved {
					workingClasses = append(workingClasses, v)
				}
			}
		}
	} else {
		workingClasses = Classes
	}
	// TODO add criteria menu
	combinatrix := PossibleCombinations(workingClasses)

	criteria := NewScheduleCriteria()
	cursadasMaster := GetSchedules(workingClasses, &criteria)
	if cursadasMaster == nil {
		err = errors.New("No se hallaron combinaciones.")
		statusBulletin.Error("", err)
	}
	ui.Clear()

	//var theWeek [5]menu

	keepGoing = true
	for keepGoing {
		statusBulletin.Post(fmt.Sprintf("%d combinaciones posibles. %d descartadas. %d Combinaciones viables", combinatrix, combinatrix-len(*cursadasMaster), len(*cursadasMaster)))
		time.Sleep(time.Millisecond * 500)
	}

	return
}

// ░▒▓█ FOUR VALUE CHARACTER
func getRequest(response <-chan bool, currentUnderstanding bool) bool { // Basic concurrency function.
	select {
	case whatIHeard := <-response:
		if whatIHeard == true {
			return true
		} else {
			return false
		}
	default:
		return currentUnderstanding // If we do not recieve answer, continue doing what you were doing
	}
}

func splash(seconds time.Duration) {
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	width, height := ui.TerminalDimensions()
	xPos := (width-58)/2 - 2
	yPos := (height-7)/2 - 2
	if xPos < 0 {
		xPos = 0
	}
	welcome := widgets.NewParagraph()
	welcome.Border = false
	welcome.SetRect(xPos, yPos-2, xPos+61, yPos+1)
	welcome.Text = `        Bienvenidos a`
	welcome.TextStyle = ui.NewStyle(ui.ColorWhite)
	exitMessage := widgets.NewParagraph()
	exitMessage.Border = false
	exitMessage.Text = `                   Presione (q) para salir`
	exitMessage.SetRect(xPos, (height+yPos+8)/2-1, xPos+61, height+2)
	version := widgets.NewParagraph()
	version.Border = false
	version.SetRect(width-len(versionNumber)-2, height-2, width+2, height+2)
	version.Text = versionNumber
	splash := widgets.NewParagraph()
	splash.SetRect(xPos, yPos, xPos+61, yPos+8)

	splash.Border = false
	splashText := ` _____                 _     _             _        _      
/  __ \               | |   (_)           | |      (_)     
| /  \/ ___  _ __ ___ | |__  _ _ __   __ _| |_ _ __ ___  __
| |    / _ \| '_'_ \| '_ \| | '_ \ / _' | __| '__| \ \/ /
| \__/\ (_) | | | | | | |_) | | | | | (_| | |_| |  | |>  < 
 \____/\___/|_| |_| |_|_.__/|_|_| |_|\__,_|\__|_|  |_/_/\_\` //58 lines wide, 8 lines high
	splash.TextStyle = ui.NewStyle(colorWheel[selectedTheme])
	splash.Text = splashText
	ui.Render(version)
	ui.Render(welcome)
	ui.Render(exitMessage)
	ui.Render(splash)
	for i := 0; i < 1000; i++ {
		pressed := AskForKeyPress()
		if pressed == "<Space>" || pressed == "<Enter>" {
			break
		}
		splash.TextStyle = ui.NewStyle(colorWheel[selectedTheme])
		ui.Render(splash)
		time.Sleep(seconds * time.Millisecond)
	}
	ui.Clear()
}

func wait(seconds time.Duration) {
	time.Sleep(time.Second * seconds)
}
func AskForKeyPress() string {
	select {
	case pressed := <-askToPollKeyPress:
		//if pressed == "w" {
		//	selectedTheme++
		//	if selectedTheme > len(colorWheel)-1 {
		//		selectedTheme = 0
		//	}
		//}
		return pressed
	default:
		break
	}
	return ""
}

func pollKeyboard() { // Writes pressed key to channel
	uiEvents := ui.PollEvents()
	for {
		e := <-uiEvents
		switch e.ID {
		case "q", "<C-q>", "Q":
			isExiting = true
			close(askToPollKeyPress)
			ui.Clear()
			ui.Close()
			//close(askToRenderMain)

			panic("User Exited. (q) Press.")
		case "<C-w>", "<C-W>":
			selectedTheme++
			if selectedTheme > len(colorWheel)-1 {
				selectedTheme = 0
			}

		default:
			askToPollKeyPress <- e.ID
		}
	}
}

func theme() ui.Color {
	return colorWheel[selectedTheme]
}

// fileListCurrentDirectory
func fileListCurrentDirectory(permittedStringLength int) ([]string, []string, error) {
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
	//permittedStringLength := 54
	i = 0

	for _, file := range files {
		if len(file) <= minFileLength {
			files = removeS(files, i)
			shortFileNames = removeS(shortFileNames, i)
			actualFileNames = removeS(actualFileNames, i)
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

func removeS(s []string, i int) []string {
	s = append(s[:i], s[i+1:]...)
	return s
}
func removeI(v []int, i int) []int {
	v = append(v[:i], v[i+1:]...)
	return v
}

func factorial(i int) int {
	if i > 1 {
		return i * factorial(i-1)
	} else if i < 0 {
		return -1

	} else {
		return 1
	}
}

func NCR(n int, r int) int {
	if n < 1 || r < 1 {
		return -1
	} else if n >= r {
		return factorial(n) / (factorial(r) * factorial(n-r))
	} else {
		return -1000
	}
}
