package main

import (
	"fmt"
	ui "github.com/gizak/termui"
	"github.com/gizak/termui/widgets"
	"time"
)

type bulletin struct {
	message string
	title   string
	color   ui.Color
	speaker func(*bulletin) // tipo, wat the fuck... no me esperaba que iba a funcionar. Brain implosion
	border  bool
	*fitting
}

func InitBulletin(myBulletin *bulletin) {
	width, height := ui.TerminalDimensions()
	speaks := widgets.NewParagraph()
	//speaks.SetRect(2*width/3, 0, width+1, height/2+1)
	speaks.SetRect(myBulletin.fitting.getRect())
	speaks.Border = myBulletin.border
	speaks.Title = myBulletin.title
	speaks.TextStyle = ui.NewStyle(colorWheel[selectedTheme])
	speaks.Text = welcomeText
	ui.Render(speaks)

	handle := func(input *bulletin) { // Closure is so cool
		widthRefresh, heightRefresh := ui.TerminalDimensions()
		if widthRefresh != width || heightRefresh != height {
			width = widthRefresh
			height = heightRefresh
			speaks.SetRect(input.fitting.getRect())
		}
		speaks.Text = input.message
		speaks.TextStyle = ui.NewStyle(theme())
		if input.title != "" {
			speaks.Title = input.title
		}
		speaks.Border = input.border
		ui.Render(speaks)
	}
	myBulletin.speaker = handle
}

func (myBulletin *bulletin) Refresh() {
	myBulletin.speaker(myBulletin)
}

func (myBulletin *bulletin) Post(str string) {
	myBulletin.message = str

	myBulletin.speaker(myBulletin)
}
func (myBulletin *bulletin) Error(message string, err error) {
	message = fmt.Sprintf(message+"\n", err)
	myBulletin.message = message
	myBulletin.color = ui.ColorRed
	myBulletin.speaker(myBulletin)
}

func NewBulletin() bulletin {
	fit := CreateFitting([3]int{5, 6, 0}, [3]int{0, 0, 0}, [3]int{1, 1, 1}, [3]int{2, 3, 0})
	return bulletin{title: "Status", color: theme(), fitting: fit, message: ""}
}

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
