package main

import (
	ui "github.com/gizak/termui"
	"github.com/gizak/termui/widgets"
	"time"
)

type menu struct {
	options []string
	title   string
	color   ui.Color
	selectedColor ui.Color
	fitting
	associatedList *widgets.List

	selection int
}

func NewMenu() menu {
	return menu{color: ui.ColorYellow, selectedColor:ui.ColorClear,selection:-1}
}
func InitMenu(theMenu *menu) {
	menu := widgets.NewList()
	menu.Rows = theMenu.options
	menu.Title = theMenu.title
	menu.TextStyle = ui.NewStyle(theMenu.color)
	menu.SetRect(theMenu.fitting.getRect())
	theMenu.associatedList = menu
}

//type Selection struct {
//	*menu
//	 row int
//	 executed bool
//}
//func NewSelection() Selection {
//return Selection{row:-1}
//}

func (theMenu *menu) Poller(askedToPoll <-chan bool) {
	polling := false

	for {
		polling = getRequest(askedToPoll, polling)

		if polling {

			switch AskForKeyPress() {
			case "":
				time.Sleep(time.Millisecond * 20)
			case "<Up>","j":
				theMenu.associatedList.ScrollUp()
				//theMenu.associatedList.TextStyle = ui.NewStyle()
			case "<Down>","k":
				theMenu.associatedList.ScrollDown()
			case "<Enter>":
				theMenu.selection = theMenu.associatedList.SelectedRow
			case "<End>":
				theMenu.associatedList.ScrollBottom()
			case "<Home>":
				theMenu.associatedList.ScrollTop()
			default:
				continue
			}
			theMenu.associatedList.SelectedRowStyle = ui.NewStyle(theMenu.selectedColor)
			ui.Render(theMenu.associatedList)
		} else {
			theMenu.associatedList.SelectedRowStyle = ui.NewStyle(theMenu.color)
		}
	}
}

type fitting struct {
	widthStart  [2]int
	heightStart [2]int
	widthEnd    [2]int
	heightEnd   [2]int
}

func (P fitting) getRect() (int, int, int, int) {
	width, height := ui.TerminalDimensions()
	if P.widthStart[0] == 0 {
		P.widthStart[0] = 99999999
	}
	if P.heightStart[0] == 0 {
		P.heightStart[0] = 99999999
	}
	return width/P.widthStart[0] + P.widthStart[1], height/P.heightStart[0] + P.heightStart[1], width/P.widthEnd[0] + P.widthEnd[1], height/P.heightEnd[0] + P.heightEnd[1]
}
