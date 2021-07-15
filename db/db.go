package db

import (
	"math/rand"
	"time"
)

func GetRandomDinnerPlace() string {
	places := []string{
		"Сковородовна",
		"Мантоварка",
		"Вьетнамка",
		"Столовая",
		"Гриль №1",
		"Узбечка",
		"КФС",
	}
	rand.Seed(time.Now().UnixNano())
	n := rand.Intn(len(places))
	return places[n]
}
