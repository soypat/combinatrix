package main

import (
	"errors"
	"fmt"
	ui "github.com/gizak/termui"
)

const tableHeight int = 30

func RenderCursada(classes []*Class, cursada *Cursada, week []*menu) {
	//height,width := ui.TerminalDimensions()
	var dayInt int8

	for _, v := range week {
		for h := 0; h < tableHeight; h++ {
			v.options = append(v.options, "")
		}
		fit := CreateFitting([3]int{int(dayInt), 6, 0}, [3]int{0, 0, 0}, [3]int{int(dayInt) + 1, 6, 0}, [3]int{0, 0, tableHeight})
		v.title = Days[dayInt]
		v.fitting = fit
		v.border = false
		v.color = ui.ColorWhite
		fillSchedule(classes,v,dayInt,cursada )
		dayInt++

	}

	for _,v := range week {
		InitMenu(v)
	}
	//fit := CreateFitting([3]int{0, 1, 0}, [3]int{0, 1, 0}, [3]int{5, 6, 0}, [3]int{2, 3, 0})

}

func fillSchedule(classes []*Class, schedule *menu, dayInt int8, cursada *Cursada) error {
	for i, com := range *cursada {
		for _, sched := range com.schedules {
			if sched.day == dayInt { // TODO verifyCursada debería asegurar que no me den caso end.hour<start.hour
				for h := sched.start.hour - 8; h <= sched.end.hour-8; h++ {
					if h < 0 {
						return errors.New("Indice negativo. Error en comienzo/fin de algún horario.")
					}
					schedule.options[h] = printScheduleLine(schedule, h, classes[i])
				}

			}
		}
	}
	return nil
}

func printScheduleLine(schedule *menu, index int, class *Class) string {
	//textWidth, _ := schedule.GetDims()
	//if schedule.options[index] //TODO add superposicion
	return fmt.Sprintf("▓%s", class.name)
}
