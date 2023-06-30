package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

const versionNumber = "2021.v1.1.0"

var weDebugging = false

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
	if !weDebugging {
		splash(2) // TODO debug TEMP
	}

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
		if weDebugging {
			break //TODO WEDUBEGING
		}
	}
	fileDir := fileNames[fileList.selection]
	Classes, err := GatherClasses(fileDir[2:]) // UNCOMMENTO FOR NORMAL USE
	if weDebugging {                           // SKIP FILE SELECTION FOR DEBUGGING
		myFile := "C:/work/gopath/src/github.com/soypat/Combinatrix/test/data_smol.dat" // DEBUG
		Classes, err := GatherClasses(myFile)                                           // DEBUG
		fmt.Println(fileNames[0])                                                       // DEBUG
		fileDir := fileNames[0]                                                         // DEBUG
		fmt.Printf("%+v", Classes)
		fmt.Printf("%+v", fileDir)
		fmt.Printf("%+v", err)
	}

	if err != nil {
		statusBulletin.Error("Error leyendo archivo:", err)
		fileList.selection = -1
		err = nil
		goto RESETCLASSSELECTION
	} else if len(Classes) < 1 {
		err := fmt.Errorf("No se hallaron clases en %s\n\n Verificar formato del archivo", fileDir[2:])
		statusBulletin.Error("Error al leer archivo:", err)
		err = nil
		fileList.selection = -1
		goto RESETCLASSSELECTION
	}
	askToPollFileList <- false
	close(askToPollFileList)
	for {
		err = mainLoop(&statusBulletin, Classes)
		if err != nil {
			statusBulletin.Error("Se hallo un error.", err)
			time.Sleep(time.Second * 2)
		}
	}
}

func mainLoop(statusBulletin *bulletin, Classes []*Class) error {
	var err error

	ClassNames := []string{}
	for _, v := range Classes {
		ClassNames = append(ClassNames, v.name)
	}
	instrucciones := `Navegue con las flechas. 

Presione [Escape] para volver a selección de clase.`

	ui.Clear()

	var classMenu menu

	var removedClassString []string
	var amountOfClassesRemoved int
	var keepGoing bool
	statusBulletin.Post("Se pueden borrar clases y seleccionar parámetros con [M]. Aprete [ENTER] para continuar...")
	askToPollClassList := make(chan bool)
	classMenu = NewMenu()

	for _, v := range ClassNames {
		classMenu.options = append(classMenu.options, v)
	}
	classMenu.fitting = CreateFitting([3]int{1, 3, 0}, [3]int{0, 1, 0}, [3]int{2, 3, 0}, [3]int{2, 3, 0})
	classMenu.title = "Clases halladas"
	InitMenu(&classMenu)
	RenderMenu(&classMenu)
	go classMenu.Poller(askToPollClassList)
	defer close(askToPollClassList)
	askToPollClassList <- true
	removedClassString = []string{}
	amountOfClassesRemoved = 0
	criteria := NewScheduleCriteria()
	keepGoing = true
	classMenu.action = "" // Eliminar el bug del delayed respnse
	for keepGoing {       //
		askToPollClassList <- true
		switch classMenu.action {
		case "<Delete>":
			classMenu.action = ""
			if len(classMenu.options) < 1 {
				time.Sleep(time.Millisecond * 500)
				continue
			}
			removalIndex := classMenu.GetSelection()
			removedClassString = append(removedClassString, classMenu.options[removalIndex])
			classMenu.options = removeS(classMenu.options, removalIndex)
			InitMenu(&classMenu)
			amountOfClassesRemoved++
		case "<C-z>", "<C-Z>":
			classMenu.action = ""
			if amountOfClassesRemoved > 0 {
				amountOfClassesRemoved--
				classMenu.options = append(classMenu.options, removedClassString[amountOfClassesRemoved])
				removedClassString = removeS(removedClassString, amountOfClassesRemoved)

				InitMenu(&classMenu)
			}
		case "<C-d>": //debug command
			classMenu.action = ""
		case "M", "m":
			classMenu.action = ""
			classMenu.selectedColor = ui.ColorYellow
			askToPollClassList <- false

			criteriaMenu(&criteria) // CRITERIA MENU!

			askToPollClassList <- true
			classMenu.selectedColor = ui.ColorWhite
		case "<Enter>":
			classMenu.action = ""
			classMenu.selectedColor = ui.ColorYellow
			InitMenu(&classMenu)
			RenderMenu(&classMenu)
			keepGoing = false
		default:
			time.Sleep(40 * time.Millisecond)
		}
		//classMenu.action="<Enter>"
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
	combinatrix := PossibleCombinations(workingClasses)

	cursadasMaster := GetSchedules(workingClasses, &criteria) // ACA ESTA LA ESTRELLA DE TODO: GetSchedules()
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
		pressed := AskForKeyPress()
		if pressed == "<Escape>" {
			UnrenderMenuSlice(week)
			keepGoing = false
		}
		err = RenderCursada(workingClasses, &(*cursadasMaster)[currentCursada], week)

		//break  // DEBUG
		if err != nil {
			statusBulletin.Error("No se pudo mostrar horarios:", err)
			time.Sleep(time.Millisecond * 200)
		} else {
		NoPress:
			pressed := AskForKeyPress()
			switch pressed {
			default:
				time.Sleep(time.Millisecond * 20)
				goto NoPress
			case "<Right>":
				unrender = true
				currentCursada++
			case "<Left>":
				unrender = true
				currentCursada--
			case "<Escape>":
				UnrenderMenuSlice(week)
				keepGoing = false
			case "<C-p>", "<C-P>":
				statusBulletin.Post("Print no implementado!")
			}
		}

	}
	return nil
}

func criteriaMenu(criteria *scheduleCriteria) {
	askToPollCriteria := make(chan bool)

	critMenu := NewMenu()
	critMenu.fitting = CreateFitting([3]int{0, 3, 0}, [3]int{0, 1, 0}, [3]int{1, 3, -4}, [3]int{2, 3, 0})
	critValues := NewMenu()

	critValues.fitting = CreateFitting([3]int{1, 3, -4}, [3]int{0, 1, 0}, [3]int{1, 3, 1}, [3]int{2, 3, 0})
	critMenu.title = "Superposición de horarios y días libres"
	lasOpciones := []string{"Max horas seguidas de SP", "Max SP total", "Max nro de SP", "Dias Libres"}
	lasOpcionesSelection := []float32{criteria.maxSuperposition, criteria.maxTotalSuperposition, float32(criteria.maxNumberOfSuperpositions), float32(criteria.minFreeDays)}
	//	maxSuperposition          float32
	//	maxTotalSuperposition     float32
	//	maxNumberOfSuperpositions int
	//	freeDays                  [len(Days)]bool
	//	minFreeDays               int
	critMenu.options = lasOpciones
	keepGoing := true
	var keyPressed = true
	critValues.options = float32SlicetoString(lasOpcionesSelection)
	InitMenu(&critMenu)
	RenderMenu(&critMenu)
	go critMenu.Poller(askToPollCriteria)
	defer close(askToPollCriteria)
	askToPollCriteria <- true
	critValues.color = ui.ColorWhite

	InitMenu(&critValues)
	RenderMenu(&critValues)

	var selectedIndex int
	var pressed string
	for keepGoing {
		if keyPressed {
			keyPressed = false
			critValues.options = float32SlicetoString(lasOpcionesSelection)
			InitMenu(&critValues)
			RenderMenu(&critValues)
			critMenu.action = ""
			askToPollCriteria <- true

			time.Sleep(time.Millisecond * 20)
		}
		pressed = critMenu.action
		switch pressed {
		case "":
			time.Sleep(time.Millisecond * 20)
			continue
		case "<Down>":
			keyPressed = true
			//critMenu.action = pressed

		case "<Up>":
			keyPressed = true
			//critMenu.action = pressed
		case "<Left>":
			selectedIndex = critMenu.GetSelection()
			keyPressed = true
			lasOpcionesSelection[selectedIndex] = lasOpcionesSelection[selectedIndex] - 1

		case "<Right>":
			selectedIndex = critMenu.GetSelection()
			keyPressed = true
			lasOpcionesSelection[selectedIndex] = lasOpcionesSelection[selectedIndex] + 1
		case "d", "D":
			selectedIndex = critMenu.GetSelection()
		case "<Escape>", "<Enter>":
			askToPollCriteria <- false
			keepGoing = false
			critMenu.selectedColor = ui.ColorYellow
			criteria.maxSuperposition = lasOpcionesSelection[0]
			criteria.maxTotalSuperposition = lasOpcionesSelection[1]
			criteria.maxNumberOfSuperpositions = int(lasOpcionesSelection[2])
			criteria.minFreeDays = int(lasOpcionesSelection[3])
			return
		default:
			time.Sleep(time.Millisecond * 20)
			continue
		}

		// Filtro los numeros
		if lasOpcionesSelection[selectedIndex] < 0 || lasOpcionesSelection[selectedIndex] > 6 {
			lasOpcionesSelection[selectedIndex] = 0
		}
		if lasOpcionesSelection[0] > lasOpcionesSelection[1] {
			lasOpcionesSelection[1] = lasOpcionesSelection[0]
		} // UPDATE numeros:
		//critValues.options = float32SlicetoString(lasOpcionesSelection)
		//InitMenu(&critValues)
		//RenderMenu(&critValues)
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
	exitMessage.Text = `                   Presione [q] para salir`
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

			panic("User Exited. [q] Press.")
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

func float32SlicetoString(is []float32) []string {
	ss := []string{}
	for v := range is {
		ss = append(ss, fmt.Sprintf("%.0f", is[v]))
	}
	return ss
}
