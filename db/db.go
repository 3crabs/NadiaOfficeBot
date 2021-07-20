package db

import (
	"math/rand"
	"time"
)

var IsSelectedDinner = false
var SelectedDinner = ""

func GetRandomDinnerPlace() string {
	if IsSelectedDinner {
		return SelectedDinner
	}
	places := []string{
		"Старый базар",
		"Сковородовна",
		"Время обеда",
		"Мантоварка",
		"Вьетнамка",
		"Столовая",
		"Гриль №1",
		"Василек",
		"Узбечка",
		"Опера",
		"КФС",
	}
	rand.Seed(time.Now().UnixNano())
	n := rand.Intn(len(places))
	IsSelectedDinner = true
	SelectedDinner = places[n]
	return places[n]
}
