package main

import (
	ui "github.com/gizak/termui"
	"github.com/gizak/termui/widgets"
	"github.com/soypat/Combinatrix"
	"log"
	"time"
)

func main() {

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}


	defer ui.Close()
	speak := initStatus()
	wait(3)
	speak("Hey fucker")
	wait(1)
}

func wait(seconds Combinatrix.time.Duration) {
	time.Sleep(time.Millisecond*1000*seconds)
}

func initStatus() func(string) {
	speaks := widgets.NewParagraph()
	speaks.SetRect(60, 0, 100, 10)
	speaks.Title = "Status"
	speaks.Text = `Programa Iniciado.

Programado en Go.
Patricio Whittingslow 2019`
	ui.Render(speaks)
	handle := func(statusString string) { // Closure is so cool
		speaks.Text = statusString
		ui.Render(speaks)
	}
	return handle
}
