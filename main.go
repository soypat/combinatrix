package main

import (
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

	splash(1)

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()
	defer close(askToRenderMain)
	defer close(askToPollKeyPress)

	speak := initStatus()
	welcomeBulletin.speaker = speak
	welcomeBulletin.Post()
	//go displayMessage(welcomeBulletin, askToPollKeyPress, 1)
	displayedFileNames, fileNames, err := fileListCurrentDirectory(54)
	if err != nil {
		log.Fatalf("Fail to read files Combinatrix: %v", err)
	}
	fileList := NewMenu()
	fileList.options = displayedFileNames
	fileList.heightStart = [2]int{0, 0}
	fileList.widthStart = [2]int{0, 0}
	fileList.widthEnd = [2]int{3, 0}  // ancho ocupa un tercio
	fileList.heightEnd = [2]int{2, 0} // alto ocupa un medio
	fileList.title = "Seleccionar Archivo de materias"
	InitMenu(&fileList)

	//mainUI.lists = append(mainUI.lists, *fileList.associatedList)

	askToRenderMain <- true
	askToPollFileList := make(chan bool)

	go fileList.Poller(askToPollFileList)
	askToPollFileList <- true
	defer close(askToPollFileList)

	for {
		if fileList.selection>=0 {
			welcomeBulletin.message = "Archivo seleccionado!\nLeyendo..."
			welcomeBulletin.color = ui.ColorClear
			welcomeBulletin.Post()

			break
		}
		time.Sleep(20 * time.Millisecond)

	}
	Classes, err :=GatherClasses(fileNames[fileList.selection])
	if err != nil {
		welcomeBulletin.Error("Error leyendo archivo.",err)
	}
	ClassNames := []string{}
	for i,v := range *Classes {
		ClassNames[i] = v.name
	}
	//classList := NewMenu()
	//classList.options = ClassNames
	//fileList.heightStart = [2]int{3, 0}
	//fileList.widthStart = [2]int{0, 0}
	//fileList.widthEnd = [2]int{3, 0}  // ancho ocupa un tercio
	//fileList.heightEnd = [2]int{2, 0} // alto ocupa un medio

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

func (myBulletin *bulletin) Post() {
	myBulletin.speaker(*myBulletin) // More brainfuckery
}
func (myBulletin *bulletin) Error(message string,err error) {
	message = fmt.Sprintf(message+"\n" ,err)
	myBulletin.message = message
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
		input.Post()
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
	ui.Close()
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
