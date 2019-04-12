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
	welcomeBulletin := NewBulletin()
	welcomeBulletin.title = "Status"
	welcomeBulletin.message = `

Programado en Go.
Patricio Whittingslow 2019.`
	welcomeBulletin.color = theme()
	var mainUI = uiElements{}
	askToRenderMain := make(chan bool)
	go RENDERER(askToRenderMain, &mainUI)
	go pollKeyboard()

	//splash(1) // TODO debug TEMP

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()
	defer close(askToRenderMain)
	defer close(askToPollKeyPress)

	speak := initStatus()
	welcomeBulletin.speaker = speak
	welcomeBulletin.Post(welcomeText)
	displayedFileNames, _, err := fileListCurrentDirectory(54) // TODO DEBUG fileNames
	if err != nil {
		log.Fatalf("Fail to read files Combinatrix: %v", err)
	}
	fileList := NewMenu()
	fileList.options = displayedFileNames
	fileList.heightStart = [3]int{0, 1, 0}
	fileList.widthStart = [3]int{0, 1, 0}
	fileList.widthEnd = [3]int{1, 3, 0}  // ancho ocupa un tercio
	fileList.heightEnd = [3]int{1, 2, 0} // alto ocupa un medio
	fileList.title = "Seleccionar Archivo de materias"
	InitMenu(&fileList)

	//mainUI.lists = append(mainUI.lists, *fileList.associatedList)

	askToRenderMain <- true
	askToPollFileList := make(chan bool)

	go fileList.Poller(askToPollFileList)
	defer close(askToPollFileList)
RESETCLASSSELECTION:
	askToPollFileList <- true

	for {
		if fileList.selection >= 0 && fileList.action == "<Enter>" {
			fileList.action = ""
			welcomeBulletin.color = ui.ColorClear
			welcomeBulletin.Post("Archivo seleccionado!\nLeyendo...")

			break
		}
		time.Sleep(20 * time.Millisecond)
		break // TODO DEBUG remove
	}
	myFile := "C:/work/gopath/src/github.com/soypat/Combinatrix/data_smol.dat"
	Classes, err := GatherClasses(myFile)
	 //TODO uncomment below
	//fileDir := fileNames[fileList.selection]
	//Classes, err := GatherClasses(fileDir[2:]) // TODO DEBUG uncomment
	if err != nil {
		welcomeBulletin.Error("Error leyendo archivo.", err)
		fileList.selection = -1
		goto RESETCLASSSELECTION
	}
	ClassNames := []string{}
	for _, v := range *Classes {
		ClassNames = append(ClassNames, v.name)
	}
	classMenu := NewMenu()
	classMenu.options = ClassNames
	classMenu.heightStart = [3]int{0, 1, 0}
	classMenu.widthStart = [3]int{1, 3, 0}
	classMenu.heightEnd = [3]int{2, 3, 0}
	classMenu.widthEnd = [3]int{2, 3, 0}
	classMenu.title = "Clases halladas"
	InitMenu(&classMenu)
	askToPollClassList := make(chan bool)

	askToPollFileList <- false
	ui.Clear()
	close(askToPollFileList)
	welcomeBulletin.Post(`Se pueden borrar clases. Apretar ENTER para continuar...`)

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
	//askToPollClassList<- true  // Probably unnecessary
	ui.Clear()
	classMenu.heightStart = [3]int{0, 1, 0}
	classMenu.widthStart = [3]int{0, 1, 0}
	classMenu.widthEnd = [3]int{1, 3, 0}  // ancho ocupa un tercio
	classMenu.heightEnd = [3]int{2, 3, 0} // alto ocupa un medio

	InitMenu(&classMenu)
	//time.Sleep(time.Millisecond*5000)
	var workingClasses []Class
	var classRemoved bool
	if len(removedClassString)>0 {
		for _, v := range *Classes {
			classRemoved = false
			for _, s := range removedClassString {
				if v.name == s {
					fmt.Printf("%s -- %s\n",v.name,s)
					classRemoved = true
				}
				if !classRemoved {
					fmt.Printf("[DEBUG] REMOVED: %s\n",v.name,s)
					workingClasses = append(workingClasses, v)
				}
			}
		}
	} else{
		workingClasses = *Classes
	}


	//for {
	//	fmt.Printf("%v\n\n",workingClasses)
	//	time.Sleep(time.Second*10)
	//}
	keepGoing = true
	for keepGoing {
		criteria := NewScheduleCriteria()
		// TODO FIX SEARCHER!
		cursadasMaster := GetSchedules(&workingClasses, &criteria) //error in searcher! gets duplicate horarios for special case when 2 classes have same hours/ number/... lots of similarities

		if cursadasMaster == nil {
			err = errors.New("No se hallaron combinaciones.")
			welcomeBulletin.Error("", err)
		}
	}

	return
}

type uiElements struct {
	pars  []widgets.Paragraph
	lists []widgets.List
}

// LEAVE ALONE:

func RENDERER(askedToPleaseRender <-chan bool, uiBlocks *uiElements) {
	rendering := false
	for {
		rendering = getRequest(askedToPleaseRender, rendering)
		if rendering {
			if len(uiBlocks.pars) != 0 {
				for _, v := range uiBlocks.pars {
					ui.Render(&v)
				}
			}
			if len(uiBlocks.lists) != 0 {
				for _, v := range uiBlocks.lists {
					ui.Render(&v)
				}
			}
		}
		time.Sleep(40 * time.Millisecond)
	}
}

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

type bulletin struct {
	message string
	title   string
	color   ui.Color
	speaker func(bulletin) // tipo, wat the fuck... no me esperaba que iba a funcionar. Brain implosion
}

func (myBulletin *bulletin) Post(str string) {
	myBulletin.message = str
	myBulletin.speaker(*myBulletin) // More brainfuckery
}
func (myBulletin *bulletin) Error(message string, err error) {
	message = fmt.Sprintf(message+"\n", err)
	myBulletin.message = message
	myBulletin.color = ui.ColorRed
	myBulletin.speaker(*myBulletin) // More brainfuckery
}

func NewBulletin() bulletin {
	return bulletin{title: "Status", color: theme()}
}

//func InitMenu() func
func displayMessage(input bulletin, keyPress <-chan string, timeToDisplay time.Duration) {
	for i := 0; i < 1000; i++ {
		pressed := AskForKeyPress()
		if pressed == "<Space>" || pressed == "<Enter>" {
			break
		}
		input.Post(pressed)
		time.Sleep(timeToDisplay * time.Millisecond)
	}
}

func initStatus() func(bulletin) {
	width, height := ui.TerminalDimensions()
	speaks := widgets.NewParagraph()
	speaks.SetRect(2*width/3, 0, width+1, height/2+1)
	speaks.Border = false
	speaks.Title = "Status"
	speaks.TextStyle = ui.NewStyle(colorWheel[selectedTheme])
	speaks.Text = welcomeText
	ui.Render(speaks)
	handle := func(input bulletin) { // Closure is so cool
		widthRefresh, heightRefresh := ui.TerminalDimensions()
		if widthRefresh != width || heightRefresh != height {
			width = widthRefresh
			height = heightRefresh
			speaks.SetRect(2*width/3, -1, width+1, height/2+1)
		}
		speaks.Text = input.message
		speaks.TextStyle = ui.NewStyle(theme())
		if input.title != "" {
			speaks.Title = input.title
		}
		ui.Render(speaks)
	}
	return handle
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
	//s[i] = s[len(s)-1]
	//return s[:len(s)-1]
}
func removeI(v []int, i int) []int {
	v = append(v[:i], v[i+1:]...)
	return v
	//v[i] = v[len(v)-1]
	//return v[:len(v)-1]

}
