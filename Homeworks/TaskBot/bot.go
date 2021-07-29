package main

// сюда писать код

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

type Task struct {
	Name   string // задача
	User   string // автор задачи
	Assign string // логин, на кого назначена задача
}

var (
	// @BotFather в телеграме даст вам это
	BotToken = ""

	// урл выдаст вам игрок или хероку
	WebhookURL                  = " http://8da3e1a9c640.ngrok.io"
	Users      map[string]int64 = make(map[string]int64)
	TaskList   map[int]Task     = make(map[int]Task)

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

// /assign_$ID - делает пользователя исполнителем задачи
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

// /resolve_$ID - выполняет задачу, удаляет её из списка
func DoTask(user string, index int) string {
	var Result string
	curTask, bOK := TaskList[index]
	if bOK == false {
		Result = "Задачи не существует:с"
		return Result
	}
	// если задачу выполнил другой человек
	if curTask.User != user {
		Result = "Задача \"" + curTask.Name + "\"" + "выполнена @" + user
	} else {
		// выполнена автором
		Result = "Задача \"" + curTask.Name + "\"" + "выполнена"
	}
	delete(TaskList, index)
	return Result
}

// /my - показывает задачи, которые назначены на меня
func MyTask(user string) string {
	var Result string
	if len(TaskList) == 0 {
		return "Нет задач"
	}

	for index, curTask := range TaskList {
		if curTask.Assign == user {
			Result += strconv.Itoa(index) + ". " + curTask.Name + " by @" + curTask.User +
				"\nunassign_" + strconv.Itoa(index) + " /resolve_" + strconv.Itoa(index)
		}
	}

	return Result
}

// /owner - показывает задачи,которые были созданы пользователем
func OtherTask(user string) string {
	var Result string
	if len(TaskList) == 0 {
		return "Нет задач"
	}

	for index, curTask := range TaskList {
		if curTask.User == user {
			Result += strconv.Itoa(index) + ". " + curTask.Name + " by @" + curTask.User +
				"\nassign_" + strconv.Itoa(index)
		}
	}

	return Result
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
		port = "8081"
	}
	go func() {
		log.Fatalln("http err: ", http.ListenAndServe(":"+port, nil))
	}()
	fmt.Println("Start listen :" + port)

	// получаем все обновления из канала updates
	for update := range updates {
		UserName := update.Message.From.UserName
		ChatID := update.Message.Chat.ID
		MessageText := update.Message.Text
		Users[UserName] = ChatID

		cmd := strings.Split(MessageText, " ") // считываем всю команду
		MainCmd := strings.Split(cmd[0], "_")

		switch cmd[0] {
		case "/tasks":
			NewMessage := tgbotapi.NewMessage(ChatID, GetTasks(UserName))
			bot.Send(NewMessage)
		case "/new":
			NewMessage := tgbotapi.NewMessage(ChatID, CreateTask(UserName, MessageText[5:]))
			bot.Send(NewMessage)
		case "/my":
			NewMessage := tgbotapi.NewMessage(ChatID, MyTask(UserName))
			bot.Send(NewMessage)
		case "/owner":
			NewMessage := tgbotapi.NewMessage(ChatID, OtherTask(UserName))
			bot.Send(NewMessage)
		default:
			switch MainCmd[0] {
			case "/assign":
				var tmp int64
				var first, second string
				ind, _ := strconv.Atoi(MainCmd[1]) // индекс
				// есть исполнитель
				if TaskList[ind].Assign != "" {
					tmp = Users[TaskList[ind].Assign] // ChatID
				} else {
					// нет исполнителя
					tmp = Users[TaskList[ind].User] // ChatID
				}
				first, second = DoAssign(UserName, ind)
				NewMessage := tgbotapi.NewMessage(ChatID, first)
				bot.Send(NewMessage)

				// если пользователь не автор задачи, посылаем сообщение о том,
				// что задача назначена на пользователя
				if TaskList[ind].User != UserName {
					NewMessage := tgbotapi.NewMessage(tmp, second)
					bot.Send(NewMessage)
				}
			case "/unassign":
				var tmp int64
				var first, second string
				ind, _ := strconv.Atoi(MainCmd[1]) // индекс

				first, second = UnassignTask(UserName, ind)
				NewMessage := tgbotapi.NewMessage(ChatID, first)
				bot.Send(NewMessage)

				// если нет исполнителя
				if TaskList[ind].Assign == "" {
					tmp = Users[TaskList[ind].User] // ChatID
					NewMessage := tgbotapi.NewMessage(tmp, second)
					bot.Send(NewMessage)
				}
			case "/resolve":
				var tmp int64
				var first string
				ind, _ := strconv.Atoi(MainCmd[1]) // индекс

				first = DoTask(UserName, ind)
				NewMessage := tgbotapi.NewMessage(ChatID, first)
				bot.Send(NewMessage)

				// если пользователь не автор задачи, то отправляем сообщение,
				// что задача принадлежит @Name
				if TaskList[ind].User != UserName {
					tmp = Users[TaskList[ind].User] // ChatID
					Msg := first + " @" + TaskList[ind].Assign
					NewMessage := tgbotapi.NewMessage(tmp, Msg)
					bot.Send(NewMessage)
				}
			}
		}
	}

	return nil
}

func main() {
	err := startTaskBot(context.Background())
	if err != nil {
		panic(err)
	}
}
