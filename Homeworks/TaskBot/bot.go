package main

// сюда писать код

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

type Task struct {
	Name   string // задача
	User   string // автор задачи
	Assign string // логин, на кого назначена задача
}

var (
	// @BotFather в телеграме даст вам это
	BotToken = "1911432734:AAHx58EmDiONrkjWH_gGW-kErlLRwmWpBKo"

	// урл выдаст вам игрок или хероку
	WebhookURL                = "https://525f2cb5.ngrok.io"
	Users      map[string]int = make(map[string]int)
	TaskList   map[int]Task   = make(map[int]Task)

	iCount = 1
)

// выводит список всех активных задач
func GetTasks(author string) string {
	var Result string

	if len(TaskList) == 0 {
		Result = "Нет задач"
	} else {
		for i := 1; i <= iCount; i++ {
			curTask, bOK := TaskList[i]
			if bOK != false { // если ключ существует в карте
				Result = strconv.Itoa(i) + ". " + curTask.Name + " by @" + curTask.User
			}
			switch curTask.Assign {
			case author:
				Result += "\nassignee: я" + "\n/unassign_" + strconv.Itoa(i) + " /resolve_" + strconv.Itoa(i)
			case "":
				Result += "\nassign_" + strconv.Itoa(i)
			default:
				Result += "\nassignee: @" + curTask.Assign
			}
		}
	}

	return Result
}

// /new XXX YYY ZZZ - создаёт новую задачу
func CreateTask(author string, TName string) string {
	var NewTask Task
	if len(TaskList) != 0 {
		for i := 1; i <= iCount; i++ {
			if TName == TaskList[i].Name {
				return "Задача была создана ранее"
			}
		}
	}

	NewTask.Name = TName
	NewTask.User = author
	TaskList[iCount] = NewTask
	iCount++ // увеличиваем количество задач

	return "Задача \"" + TName + "\"" + "создана, id=" + strconv.Itoa(iCount-1)
}

// /assign_$ID - делаеть пользователя исполнителем задачи
func DoAssign(user string, index int) (string, string) {
	var first, second string
	curTask, bOK := TaskList[index]
	if bOK == false {
		first = "Задачи не существует:с"
		return first, second
	}
	curTask.Assign = user
	first = "Задача \"" + curTask.Name + "\"" + " назначена на вас"
	if curTask.User != user {
		first = "Задача \"" + curTask.Name + "\"" + " назначена на @" + user
	}

	return first, second
}

// /unassign_$ID - снимает задачу с текущего исполнителя
func UnassignTask(user string, index int) (string, string) {
	var first, second string
	curTask, bOK := TaskList[index]
	if bOK == false {
		first = "Задачи не существует:с"
		return first, second
	}

	if curTask.Assign != user {
		first = "Задача не на вас"
		return first, second
	} else {
		first = "Принято"
		second = "Задача \"" + curTask.Name + "\"" + "осталась без исполнителя"
		curTask.Assign = ""
		return first, second
	}

}

func startTaskBot(ctx context.Context) error {
	// инициализация BotAPI
	bot, err := tgbotapi.NewBotAPI(BotToken)
	if err != nil {
		log.Fatalf("NewBotAPI failed: &s", err)
	}

	bot.Debug = true
	fmt.Printf("Authorized on account %s\n", bot.Self.UserName)

	_, err = bot.SetWebhook(tgbotapi.NewWebhook(WebhookURL)) // идем на WebhookURL
	if err != nil {
		log.Fatalf("SetWebhook failed: %s", err)
	}

	updates := bot.ListenForWebhook("/")
	http.HandleFunc("/state", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("All is working!"))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8585"
	}
	go func() {
		log.Fatalln("http err: ", http.ListenAndServe(":"+port, nil))
	}()
	fmt.Println("Start listen :" + port)

	return nil
}

func main() {
	err := startTaskBot(context.Background())
	if err != nil {
		panic(err)
	}
}
