package main

import (
	"errors"
	"fmt"
	ui "github.com/gizak/termui"
	"time"
)


const startHour int = 8
const endHour int = 22
const tableHeight int = endHour-startHour+2
const infoMenuWidth int = 30
var horariosMenu = menu{}
var cursadaInfoMenu = menu{}

func RenderCursada(classes []*Class, cursada *Cursada, week []*menu) error {
	var dayInt int8

	totalDays := len(week)

	horariosMenu.title = "Hora"
	horariosMenu.border = true
	horariosMenu.fitting = CreateFitting([3]int{0, 6, -1}, [3]int{0, 1, 0},[3]int{0, 1, totalDays + 1},[3]int{0, 1, tableHeight} )
	horariosMenu.color = theme()
	horariosMenu.selectedColor = theme()

	cursadaInfoMenu.title = ""
	cursadaInfoMenu.border=false

	cursadaInfoMenu.color = ui.ColorYellow
	cursadaInfoMenu.selectedColor = ui.ColorYellow

	for i, v := range *cursada {
		cursadaInfoMenu.options = append(cursadaInfoMenu.options, "")
		if len(classes[i].name)>infoMenuWidth-6{
			cursadaInfoMenu.options[i] = fmt.Sprintf("%s Com:%s",classes[i].name[:infoMenuWidth-7],v.label)
		} else{
			cursadaInfoMenu.options[i] = fmt.Sprintf("%s Com:%s",classes[i].name,v.label)
		}
	}
	cursadaInfoMenu.fitting = CreateFitting([3]int{0, 6, -1}, [3]int{0, 1, tableHeight},[3]int{0, 1, infoMenuWidth},[3]int{0, 1, tableHeight+len(*cursada)+2} )

	for i := startHour; i <= endHour; i++ {
		horariosMenu.options = append(horariosMenu.options, "")
		horariosMenu.options[i-startHour] = fmt.Sprintf("%d-%d",i,i+1)
	}
	InitMenu(&horariosMenu)
	InitMenu(&cursadaInfoMenu)
	for _, v := range week {
		for h := 0; h < tableHeight; h++ {
			v.options = append(v.options, "")
		}
		fit := CreateFitting([3]int{int(dayInt), 6, totalDays - int(dayInt)}, [3]int{0, 1, 0}, [3]int{int(dayInt) + 1, 6, totalDays - int(dayInt)}, [3]int{0, 1, tableHeight})
		v.title = Days[dayInt]
		v.fitting = fit
		v.color = ui.ColorWhite
		err := fillSchedule(classes, v, dayInt, cursada)
		if err != nil {
			return err
		}

		dayInt++
	}
	for _, v := range week {
		InitMenu(v)
	}
	RenderMenu(&cursadaInfoMenu)
	RenderMenu(&horariosMenu)
	for _, v := range week {
		RenderMenu(v)
	}
	//horariosMenu.associatedList

	//fit := CreateFitting([3]int{0, 1, 0}, [3]int{0, 1, 0}, [3]int{5, 6, 0}, [3]int{2, 3, 0})
	return nil
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

func UnrenderMenuSlice(menus []*menu) {
	if len(menus) == 0 {
		return
	}
	for _,v := range menus {
		if v.associatedList == nil {
			continue
		}
		v.title = ""
		for i,_ := range v.options {
			v.options[i] = ""
		}
	}
	ui.Clear()
	time.Sleep(time.Microsecond*100)
}