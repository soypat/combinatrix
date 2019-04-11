// Copyright 2017 Zack Guo <zack.y.guo@gmail.com>. All rights reserved.
// Use of this source code is governed by a MIT license that can
// be found in the LICENSE file.
// +build termuiver
package main

import (
	ui "github.com/gizak/termui"
	"github.com/gizak/termui/widgets"
	"log"
	"os"
	"path/filepath"
	"time"
)

var isExiting = false

var colorWheel = []ui.Color{ui.ColorGreen, ui.ColorBlue, ui.ColorCyan, ui.ColorYellow, ui.ColorRed, ui.ColorMagenta}

var selectedTheme = 0

const welcomeText = `

Programado en Go.
Patricio Whittingslow 2019.`
func main() {
	welcomeBulletin := NewBulletin()
	welcomeBulletin.title = "Status"
	welcomeBulletin.message =`

Programado en Go.
Patricio Whittingslow 2019.`
	welcomeBulletin.color = theme()

	keyPress := make(chan string)
	go pollEvent(keyPress)

	splash(3, keyPress)

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()
	speak := initStatus()
	welcomeBulletin.speaker = speak
	welcomeBulletin.Post()
	go displayMessage(welcomeBulletin, keyPress, 3)
	//go displayMessage(welcomeBulletin, 3)
	wait(4)

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

// LEAVE ALONE:

type bulletin struct {
	message string
	title string
	color ui.Color
	speaker func(bulletin)   // tipo, wat the fuck... no me esperaba que iba a funcionar. Brain implosion
}

func (myBulletin *bulletin) Post() {
	myBulletin.speaker(*myBulletin)   // More brainfuckery
}

func NewBulletin() bulletin {
	return bulletin{title:"Status", color: theme()}
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
			width = widthRefresh;
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

func splash(seconds time.Duration, keyPress <-chan string) {
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
		pressed := askForKeyPress(keyPress)
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
func askForKeyPress(keyPress <-chan string) string {
	select {
	case pressed := <-keyPress:
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

func pollEvent(keyPress chan<- string) {
	//previousKey := ""
	uiEvents := ui.PollEvents()
	eventTicker := time.NewTicker(50 * time.Millisecond)
	for {
		e := <-uiEvents
		<-eventTicker.C
		switch e.ID {
		case "q", "<C-q>", "Q":
			isExiting = true
			panic("User Exited. (q) Press.")
		case "<C-w>","<C-W>":
			selectedTheme++
			if selectedTheme > len(colorWheel)-1 {
				selectedTheme = 0
			}

		default:
			keyPress <- e.ID
			//previousKey = e.ID
		}
	}
}

func theme() ui.Color {
	return colorWheel[selectedTheme]
}

func displayMessage(input bulletin, keyPress <-chan string, timeToDisplay time.Duration) {
	for i := 0; i < 1000; i++ {
		pressed := askForKeyPress(keyPress)
		if pressed == "<Space>" || pressed == "<Enter>" {
			break
		}
		input.Post()
		time.Sleep(timeToDisplay * time.Millisecond)
	}
}