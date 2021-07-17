package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
)

// ----- Look -----
// ----- Go -----
// ----- PutOn -----
// ----- Take -----
// ----- Apply -----

// 4 locations - kitchen, room, street, hallway
var location map[string]int = map[string]int{
	"кухня":   0,
	"комната": 1,
	"улица":   2,
	"коридор": 3,
}

var InvAll map[string]int = map[string]int{
	"ключи":     0,
	"конспекты": 1,
}

var player Player

// Player is
type Player struct {
	Cmd         string
	inventory   []Inv // инвентарь на текущий момент игры
	curLocation int   // текущая локация
	backpack    bool  // рюкзак, true - есть, false - нет
	rLocation   Room
	bDoor       bool // true - дверь открыта, false - дверь закрыта
}

// Inv is struct for inventory
type Inv struct {
	name string
}

// Room is
type Room struct {
	status  string         // kitchen, room, street, hallway
	objects map[string]int // объекты, которые есть в текущей комнате (рюкзак не включается)
	CurObj  []string       // объекты, которые остались в комнате на текущий момент времени
}

func (r Room) IsEmptyRoom() (bOK bool) {
	iCount := 0
	for _, keyExist := range r.objects {
		if keyExist != 0 {
			iCount++
		}
	}
	if iCount != 0 {
		bOK = false
	} else {
		bOK = true
	}
	return bOK
}

// Сортировка map по ключам
func SortKeys(sMap map[string]int) []string {
	keys := make([]string, len(sMap))
	i := 0

	for keyValue := range sMap {
		keys[i] = keyValue
		i++
	}

	sort.Strings(keys)

	return keys
}

// Look is function for printing location
func (p Player) Look() string {
	var loc int
	var str string
	str = ""
	loc = p.curLocation
	switch loc {
	case 0: // кухня
		if p.backpack {
			str = "ты находишься на кухне, на столе: чай, надо идти в универ. можно пройти - коридор"
		} else {
			str = "ты находишься на кухне, на столе: чай, надо собрать рюкзак и идти в универ. можно пройти - коридор"
		}
	case 1: // комната
		// функция, которая проверяет, пуста ли комната
		if p.rLocation.IsEmptyRoom() {
			str = "на столе: "
			p.CurrentObjects()
			mapSize := len(p.rLocation.CurObj)
			i := 0

			sInvAll := SortKeys(InvAll)

			for _, keyStr := range sInvAll {
				i++
				var bFind bool
				iSize := 0
				for _, keyEx := range p.inventory {
					if !(keyStr == keyEx.name) {
						bFind = false
					} else {
						bFind = true
						break
					}

					iSize++
				}
				if !bFind {
					if i != mapSize-1 {
						str = str + keyStr + ", "
					} else {
						str = str + keyStr + ". "
					}
				}
			}

		} else {
			str = "пустая комната. "
		}
		if !p.backpack {
			if str[len(str)-2:] == ". " {
				str = str[:len(str)-2]
				str = str + ", "
			}
			str = str + "на стуле: рюкзак. " + "можно пройти - коридор"
			return str
		} else {
			if str[len(str)-2:] == ", " {
				str = str[:len(str)-2]
				str = str + ". "
			}
		}
		if str == "на столе: " {
			str = "пустая комната. "
		}
		str = str + "можно пройти - коридор"

	default:
		// error or in other functions doesn't exist this function
	}
	return str
}

// Изменяет поле CurObj - текущие предметы в комнате
func (p *Player) CurrentObjects() {
	var ObjRoom = p.rLocation.objects
	var current = []string{}

	for keyValueRoom, _ := range ObjRoom {
		for _, keyExist := range p.inventory {
			bOK := (keyValueRoom == keyExist.name)
			if !bOK {
				current = append(current, keyExist.name)
			}
		}
	}

	p.rLocation.CurObj = make([]string, len(current), len(current))
	copy(p.rLocation.CurObj, current)
}

// Go is function for go to next location, cmd - next location
func (p *Player) Go(cmd string) string {
	var str string
	var loc = p.curLocation
	switch cmd {
	case "кухня":
		// to kitchen
		if loc == 3 {
			p.curLocation = 0
			p.rLocation.status = "кухня"
			str = "кухня, ничего интересного. можно пройти - коридор"
		} else {
			// error
		}
	case "комната": // to room
		if loc == 3 {
			p.curLocation = 1
			p.rLocation.status = "комната"
			p.CurrentObjects()
			str = "ты в своей комнате. можно пройти - коридор"
		} else {
			// error +
			str = "нет пути в " + cmd
		}
	case "улица": // to street
		if loc == 3 && p.bDoor {
			p.curLocation = 2
			p.rLocation.status = "улица"
			p.CurrentObjects()
			str = "на улице весна. можно пройти - домой"
		} else {
			str = "дверь закрыта"
		}
	case "коридор": // to hallway
		p.curLocation = 3
		p.rLocation.status = "коридор"
		p.CurrentObjects()
		str = "ничего интересного. можно пройти - кухня, комната, улица"
	}

	return str
}

// PutOn is function for put on backpack
func (p *Player) PutOn() string {
	var str string
	if !p.backpack {
		p.backpack = true
		str = "вы надели: рюкзак"
	} else {
		str = "вы уже надели рюкзак"
		// need in second test
	}

	return str
}

// Возвращает true, если elem есть в инвентаре, иначе - false
func (p *Player) Check(elem string) bool {
	var bOK bool
	for _, keyExist := range p.inventory {
		if elem == keyExist.name {
			bOK = true
			break
		} else {
			bOK = false
		}
	}
	return bOK
}

// Take is function for take thing
func (p *Player) Take(sElem string) string {
	var str string
	var elem Inv
	elem.name = sElem
	if p.backpack {
		if _, keyExist := InvAll[sElem]; !keyExist {
			str = "нет такого" // неизвестный предмет
		} else if p.Check(sElem) {
			str = "нет такого" // предмета в комнате нет, мы его уже взяли
		} else {
			p.inventory = append(p.inventory, elem)
			delete(p.rLocation.objects, sElem)
			str = "предмет добавлен в инвентарь: " + sElem
		}

	} else {
		// need in second test
		str = "некуда класть"
	}
	p.CurrentObjects()
	return str
}

// Apply is function of applying one item to another
func (p *Player) Apply(cmdWhat, cmdFor string) string {
	var str string
	var bOK bool
	for _, keyExist := range p.inventory {
		if keyExist.name == cmdWhat {
			bOK = true
			break
		} else {
			bOK = false
		}
	}
	if !bOK {
		return "нет предмета в инвентаре - " + cmdWhat
	} else {
		switch cmdWhat {
		case "ключи":
			if cmdFor == "дверь" {
				p.bDoor = true
				str = "дверь открыта"
			} else {
				str = "не к чему применить"
			}

		}
	}

	return str
}

//
/*
	код писать в этом файле
	наверняка у вас будут какие-то структуры с методами, глобальные перменные ( тут можно ), функции
*/

func main() {
	/*
		в этой функции можно ничего не писать
		но тогда у вас не будет работать через go run main.go
		очень круто будет сделать построчный ввод команд тут, хотя это и не требуется по заданию
	*/
	initGame()
	cmd := bufio.NewScanner(os.Stdin)

	var inCMD string
	for cmd.Scan() {
		inCMD = cmd.Text()
		fmt.Println(handleCommand(inCMD))

	}

}

var kitchen Room
var room Room
var street Room
var hallway Room

func initGame() {
	/*
		эта функция инициализирует игровой мир - все команты
		если что-то было - оно корректно перезатирается
	*/
	kitchen.status = "кухня"
	kitchen.objects = map[string]int{
		//"стол": 0,
		"чай": 0,
	}
	kitchen.CurObj = []string{"чай"}
	room.status = "комната"
	room.objects = map[string]int{
		"ключи":     0,
		"конспекты": 1,
		//"стол":      2,
		//"стул":      3,
	}
	room.CurObj = []string{"ключи", "конспекты"}
	street.status = "улица"
	hallway.status = "коридор"
	hallway.objects = map[string]int{
		"дверь": 0,
	}
	hallway.CurObj = []string{"дверь"}

	player.bDoor = false
	player.backpack = false
	player.curLocation = 0
	player.rLocation.status = "кухня"
	player.inventory = make([]Inv, 3)
	player.rLocation.objects = kitchen.objects

}

func handleCommand(command string) string {
	/*
		данная функция принимает команду от "пользователя"
		и наверняка вызывает какой-то другой метод или функцию у "мира" - списка комнат
	*/
	sCmd := strings.Split(command, " ")
	sCommand := sCmd[0]
	sWhatAndFor := sCmd[1:]
	switch sCommand {
	case "осмотреться":
		return player.Look()
	case "идти":
		return player.Go(sWhatAndFor[0])
	case "надеть":
		return player.PutOn()
	case "взять":
		return player.Take(sWhatAndFor[0])
	case "применить":
		return player.Apply(sWhatAndFor[0], sWhatAndFor[1])
	}
	return "неизвестная команда"
}
