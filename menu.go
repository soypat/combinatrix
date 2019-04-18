package main

import (
	ui "github.com/gizak/termui"
	"github.com/gizak/termui/widgets"
	"time"
)

type menu struct {
	options       []string
	title         string
	color         ui.Color
	border        bool
	selectedColor ui.Color
	*fitting
	associatedList *widgets.List
	// asociado a una accion/seleccion:
	selection int
	action    string
}

func NewMenu() menu {
	return menu{border: true, color: ui.ColorYellow, selectedColor: ui.ColorClear, selection: -1}
}
func (theMenu *menu) GetSelection() int {
	return theMenu.associatedList.SelectedRow
}

func InitMenu(theMenu *menu) {
	menu := widgets.NewList()
	menu.Rows = theMenu.options
	menu.Title = theMenu.title
	menu.TextStyle = ui.NewStyle(theMenu.color)
	menu.SelectedRowStyle = ui.NewStyle(theMenu.selectedColor)
	menu.SetRect(theMenu.fitting.getRect())
	menu.Border = theMenu.border


	theMenu.associatedList = menu // LAST LINE
}

func RenderMenu(theMenu *menu) {
	ui.Render(theMenu.associatedList)
}

func (theMenu *menu) Poller(askedToPoll <-chan bool) {
	polling := false

	for {
		polling = getRequest(askedToPoll, polling)

		if polling {
			keyIdentifier := AskForKeyPress()
			switch keyIdentifier {
			case "":
				time.Sleep(time.Millisecond * 20)
			case "<Up>", "j":
				theMenu.associatedList.ScrollUp()
				//theMenu.associatedList.TextStyle = ui.NewStyle()
			case "<Down>", "k":
				theMenu.associatedList.ScrollDown()
			case "<Enter>":
				theMenu.selection = theMenu.associatedList.SelectedRow
				theMenu.action = keyIdentifier
			case "<End>":
				theMenu.associatedList.ScrollBottom()
			case "<Home>":
				theMenu.associatedList.ScrollTop()
			default:
				theMenu.selection = theMenu.associatedList.SelectedRow
				theMenu.action = keyIdentifier
				continue
			}
			theMenu.associatedList.SelectedRowStyle = ui.NewStyle(theMenu.selectedColor)
			ui.Render(theMenu.associatedList)
		} else if askedToPoll == nil {
			//time.Sleep(1000*time.Millisecond) // TODO make this even cleaner. a return would be fantastic
			return
		} else {
			_, ok := <- askedToPoll // CHECK IF CHANNEL CLOSED
			if !ok {
				time.Sleep(20*time.Millisecond)
				return
			}
			time.Sleep(1000*time.Millisecond)
			theMenu.associatedList.SelectedRowStyle = ui.NewStyle(theMenu.color)
		}
	}
}

type fitting struct {
	widthStart  [3]int
	heightStart [3]int
	widthEnd    [3]int
	heightEnd   [3]int
}

func CreateFitting(wS [3]int, hS [3]int, wE [3]int, hE [3]int) *fitting {
	var P fitting
	if wS[1] == 0 {
		wS[1] = 1
	}
	if wE[1] == 0 {
		wE[1] = 1
	}
	if hS[1] == 0 {
		hS[1] = 1
	}
	if hE[1] == 0 {
		hE[1] = 1
	}
	P.widthStart = wS
	P.heightStart = hS
	P.widthEnd = wE
	P.heightEnd = hE

	return &P
}

func (theMenu menu) GetDims() (X int, Y int) {
	x1, y1, x2, y2 := theMenu.getRect()
	if x1 > x2 {
		X = x1 - x2
	} else {
		X = x2 - x1
	}
	if y1 > y2 {
		Y = y1 - y2
	} else {
		Y = y2 - y1
	}
	return X, Y
}

func (P fitting) getRect() (int, int, int, int) {
	width, height := ui.TerminalDimensions()
	return width*P.widthStart[0]/P.widthStart[1] + P.widthStart[2], height*P.heightStart[0]/P.heightStart[1] + P.heightStart[2], width*P.widthEnd[0]/P.widthEnd[1] + P.widthEnd[2], height*P.heightEnd[0]/P.heightEnd[1] + P.heightEnd[2]
}
