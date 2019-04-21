package main

import (
	"errors"
	"fmt"
	ui "github.com/gizak/termui"
	"github.com/gizak/termui/widgets"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

const versionNumber = "2019.BETA"

var isExiting = false

var colorWheel = []ui.Color{ui.ColorGreen, ui.ColorBlue, ui.ColorCyan, ui.ColorYellow, ui.ColorRed, ui.ColorMagenta}

var selectedTheme = 0
var askToPollKeyPress = make(chan string)
var version = widgets.NewParagraph()

const welcomeText = `

Programado en Go.
Patricio Whittingslow 2019.`

func main() {
	go pollKeyboard()
	statusBulletin := NewBulletin()

	//splash(2) // TODO debug TEMP

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	defer close(askToPollKeyPress)

	InitBulletin(&statusBulletin)
	statusBulletin.Post(welcomeText)
	fileList := NewMenu()

	fileList.fitting = CreateFitting([3]int{0, 1, 0}, [3]int{0, 1, 0}, [3]int{1, 3, 0}, [3]int{1, 2, 0})
	fileListWidth, _ := fileList.GetDims()

	displayedFileNames, fileNames, err := fileListCurrentDirectory(fileListWidth - 6)
	if err != nil {
		log.Fatalf("Fail to read files Combinatrix: %v", err)
	}
	if len(displayedFileNames) < 1 {
		statusBulletin.Error("No se hallaron archivos en la carpeta de trabajo o las subyacentes. \n\nEl programa cerrará...", nil)
		time.Sleep(time.Second * 3)
		return
	}
	fileList.options = displayedFileNames

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
		//break //TODO  DEBUG
	}
	//myFile := "C:/work/gopath/src/github.com/soypat/Combinatrix/test/lottamat.dat" // DEBUG
	//Classes, err := GatherClasses(myFile)
	//fmt.Sprintf(fileNames[1])// DEBUG COMMENT
	//fileDir:=fileNames[0] // DEBUG
	//TODO uncomment below
	fileDir := fileNames[fileList.selection]
	Classes, err := GatherClasses(fileDir[2:]) // UNCOMMENTO FOR NORMAL USE
	if err != nil {
		statusBulletin.Error("Error leyendo archivo:", err)
		fileList.selection = -1
		err = nil
		goto RESETCLASSSELECTION
	} else if len(Classes) < 1 {
		err := fmt.Errorf("No se hallaron clases en %s\n\n Verificar formato del archivo", fileDir[2:])
		statusBulletin.Error("Error al leer archivo:",err)
		err = nil
		fileList.selection = -1
		goto RESETCLASSSELECTION
	}
	askToPollFileList <- false
	close(askToPollFileList)
	for {
		err = mainLoop(&statusBulletin, Classes)
		if err != nil {
			statusBulletin.Error("Se hallo un error.",err)
			time.Sleep(time.Second*3)
	}
}

	return
}

func mainLoop(statusBulletin *bulletin, Classes []*Class) error {
	var err error

	ClassNames := []string{}
	for _, v := range Classes {
		ClassNames = append(ClassNames, v.name)
	}
	instrucciones := `Navegue con las flechas. 

Presione Escape para volver a selección de clase.`

	ui.Clear()
	//var askToPollClassList chan bool

	//if askToPollClassList != nil {
	//	close(askToPollClassList)
	//}
	var classMenu menu

	var removedClassString []string
	var amountOfClassesRemoved int
	var keepGoing bool
	statusBulletin.Post("Se pueden borrar clases. Apretar ENTER para continuar...")
	askToPollClassList := make(chan bool)
	classMenu = NewMenu()

	for _, v := range ClassNames {
		classMenu.options = append(classMenu.options,v )
	}

	classMenu.fitting = CreateFitting([3]int{1, 3, 0}, [3]int{0, 1, 0}, [3]int{2, 3, 0}, [3]int{2, 3, 0})
	classMenu.title = "Clases halladas"
	InitMenu(&classMenu)
	RenderMenu(&classMenu)

	//askToPollClassList <- true
	go classMenu.Poller(askToPollClassList)
	defer close(askToPollClassList)
	askToPollClassList <- true

	removedClassString = []string{}
	amountOfClassesRemoved = 0
	keepGoing = true

	for keepGoing { //

		switch classMenu.action {
		case  "<Delete>":
			classMenu.action = ""
			if len(classMenu.options) <= 1 {
				time.Sleep(time.Millisecond * 500)
				continue
			}
			removalIndex := classMenu.GetSelection()
			removedClassString = append(removedClassString, classMenu.options[removalIndex])
			classMenu.options = removeS(classMenu.options, removalIndex)
			InitMenu(&classMenu)
			amountOfClassesRemoved++
			//keepGoing = false
		case "<C-z>", "<C-Z>":
			classMenu.action = ""
			if amountOfClassesRemoved > 0 {
				amountOfClassesRemoved--
				classMenu.options = append(classMenu.options, removedClassString[amountOfClassesRemoved])
				removedClassString = removeS(removedClassString, amountOfClassesRemoved)

				InitMenu(&classMenu)
			}
		case "d": //debug command
			classMenu.action = ""
			fmt.Printf("%v\n", classMenu.options)
		case "<Enter>":
			classMenu.action = ""
			classMenu.associatedList.SelectedRowStyle = ui.NewStyle(ui.ColorYellow)
			InitMenu(&classMenu)
			RenderMenu(&classMenu)
			keepGoing = false
		default:
			time.Sleep(40 * time.Millisecond)
		}
	}
	askToPollClassList <- false

	var workingClasses []*Class
	var isRemoved bool
	if len(removedClassString) > 0 {
		for _, v := range Classes {
			isRemoved = false
			for _, s := range removedClassString {
				if v.name == s {
					isRemoved = true
				}
			}
			if !isRemoved {
				workingClasses = append(workingClasses, v)
			}

		}
	} else {
		workingClasses = Classes
	}

	ui.Clear()
	classMenu.fitting = CreateFitting([3]int{0, 1, 0}, [3]int{0, 1, 0}, [3]int{1, 3, 0}, [3]int{3, 3, 0})

	// TODO add criteria menu
	combinatrix := PossibleCombinations(workingClasses)

	criteria := NewScheduleCriteria()
	cursadasMaster := GetSchedules(workingClasses, &criteria)
	if cursadasMaster == nil {
		err = errors.New("No se hallaron combinaciones.")
		statusBulletin.Error("", err)
		return err
	}
	var week []*menu
	for i := 0; i < 5; i++ {
		emptyMenu := NewMenu()
		week = append(week, &emptyMenu)
	}

	keepGoing = true
	currentCursada := 0

	var unrender = true
	for keepGoing {
		if currentCursada > len(*cursadasMaster)-1 {
			currentCursada = 0
		} else if currentCursada < 0 {
			currentCursada = len(*cursadasMaster) - 1
		}
		if unrender {
			UnrenderMenuSlice(week)
			time.Sleep(time.Millisecond * 5)
			unrender = false
			indexString := fmt.Sprintf("Cursada %d/%d. ", currentCursada+1, len(*cursadasMaster))
			if currentCursada == 0 {
				posiblesComb := fmt.Sprintf("\n%d combinaciones posibles.\n%d descartadas.\n%d Combinaciones viables\n\n", combinatrix, combinatrix-len(*cursadasMaster), len(*cursadasMaster))
				statusBulletin.Post(indexString + posiblesComb + instrucciones)
				PrintVersion()
			} else {
				statusBulletin.Post(indexString + instrucciones)
				PrintVersion()
			}
		}

		err = RenderCursada(workingClasses, &(*cursadasMaster)[currentCursada], week)
		//break  // DEBUG

		if err != nil {
			statusBulletin.Error("No se pudo mostrar horarios:", err)
		} else {
			pressed := AskForKeyPress()
			switch pressed {
			case "":
				time.Sleep(time.Millisecond * 20)
				continue
			case "<Right>":
				unrender = true
				currentCursada++
			case "<Left>":
				unrender = true
				currentCursada--
			case "<Escape>":
				//statusBulletin.Post("Escape not implemented")
				UnrenderMenuSlice(week)
				keepGoing = false
			case "<C-p>", "<C-P>":
				statusBulletin.Post("Print!")
			}
		}

	}
	//statusBulletin.Post("Volvined")
	//time.Sleep(time.Second*2)

	return nil
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
	PrintVersion()
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

func PrintVersion() {
	width, height := ui.TerminalDimensions()

	version.Border = false
	version.SetRect(width-len(versionNumber)-2, height-2, width+2, height+2)
	version.Text = versionNumber
	ui.Render(version)
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
		case "q", "Q":
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
			time.Sleep(time.Millisecond * 10)
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
	//shortFileNames = append(files[:0:0], files...)
	//actualFileNames = append(files[:0:0], files...)

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
	//i = 0
	reDir := regexp.MustCompile(`[\\]{1}[\w]{0,99}$`)
	reExecutable := regexp.MustCompile(`\.exe$`) // TODO No hace falta estos checks si voy por el otro camino
	reReadable := regexp.MustCompile(`\.txt$|\.dat$`)
	// FOR to remove base folder entry y ejecutables
	//i=0
	for i = 0; i < len(files); {
		currentString := files[i]
		if reDir.MatchString(currentString) || reExecutable.MatchString(currentString) {
			files = removeS(files, i)
			continue
		}
		if !reReadable.MatchString(currentString) {
			files = removeS(files, i)
			continue
		}
		i++
	}

	for _, file := range files {
		//if len(file) <= minFileLength {
		//	files = removeS(files, i)
		//	shortFileNames = removeS(shortFileNames, i)
		//	actualFileNames = removeS(actualFileNames, i)
		//	continue
		//}
		if len(file) > permittedStringLength+minFileLength {

			shortFileNames = append(shortFileNames, `~\…`+file[len(file)-permittedStringLength:])

		} else {
			shortFileNames = append(shortFileNames, "~"+file[minFileLength:])

		}
		actualFileNames = append(actualFileNames, "~"+file[minFileLength:])
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
