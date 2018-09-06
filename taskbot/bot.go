package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"gopkg.in/telegram-bot-api.v4"
)

type Task struct {
	ID         int
	Name       string
	AssignedID int64
	CreaterID  int64
}

type TGUser struct {
	ChatID   int64
	Username string
}

var (
	// @BotFather gives you this
	BotToken   = "487620392:AAE3yUBEVLQao_n3xD63xta1MvzpvaJIUYE"
	WebhookURL = "https://b4aa8068.ngrok.io"
	// WebhookURL = "https://127.0.0.1:8080"
	taskID  = 1
	tasks   = []*Task{}
	tgUsers = make(map[int64]*TGUser)
)

func startTaskBot(ctx context.Context) error {
	// сюда пишите ваш код
	bot, err := tgbotapi.NewBotAPI(BotToken)
	if err != nil {
		panic(err)
	}

	bot.Debug = true
	fmt.Printf("Authorized on account %s\n", bot.Self.UserName)

	_, err = bot.SetWebhook(tgbotapi.NewWebhook(WebhookURL))
	if err != nil {
		panic(err)
	}

	updates := bot.ListenForWebhook("/")

	go func() { log.Fatal(http.ListenAndServe(":8080", nil)) }()
	fmt.Println("start listen :8080")

	for update := range updates {
		msg := update.Message.Text
		if _, ok := tgUsers[update.Message.Chat.ID]; !ok {
			tgUsers[update.Message.Chat.ID] = &TGUser{
				ChatID:   update.Message.Chat.ID,
				Username: update.Message.Chat.UserName,
			}
		}

		if strings.HasPrefix(msg, "/new") {
			newTaskName := strings.TrimPrefix(msg, "/new ")
			tasks = append(tasks, &Task{
				ID:        taskID,
				Name:      newTaskName,
				CreaterID: update.Message.Chat.ID,
			})

			bot.Send(tgbotapi.NewMessage(
				update.Message.Chat.ID,
				`Задача "`+newTaskName+`" создана, id=`+strconv.Itoa(taskID),
			))

			taskID++
		} else if strings.HasPrefix(msg, "/tasks") {
			sendMsg := ""
			for _, task := range tasks {
				sendMsg += strconv.Itoa(task.ID) + ". " + task.Name + " by @" + tgUsers[task.CreaterID].Username + "\n"
				if task.AssignedID != update.Message.Chat.ID {
					if task.AssignedID != 0 {
						sendMsg += "assignee: @" + tgUsers[task.AssignedID].Username + "\n"
					} else {
						sendMsg += "/assign_" + strconv.Itoa(task.ID) + "\n"
					}
				} else {
					sendMsg += "assignee: я\n"
					sendMsg += "/unassign_" + strconv.Itoa(task.ID) + " "
					sendMsg += "/resolve_" + strconv.Itoa(task.ID) + "\n"
				}
				sendMsg += "\n"
			}
			sendMsg = strings.TrimSuffix(sendMsg, "\n\n")

			if sendMsg == "" {
				sendMsg = "Нет задач"
			}
			replyMsg := tgbotapi.NewMessage(
				update.Message.Chat.ID,
				sendMsg,
			)
			bot.Send(replyMsg)
		} else if strings.HasPrefix(msg, "/assign_") {
			assignTaskID, err := strconv.Atoi(strings.TrimPrefix(msg, "/assign_"))
			if err != nil {
				panic(err)
			}

			for _, task := range tasks {
				if task.ID == assignTaskID {
					bot.Send(tgbotapi.NewMessage(
						update.Message.Chat.ID,
						`Задача "`+task.Name+`" назначена на вас`,
					))
					if task.AssignedID == 0 && update.Message.Chat.ID != task.CreaterID {
						bot.Send(tgbotapi.NewMessage(
							task.CreaterID,
							`Задача "`+task.Name+`" назначена на @`+update.Message.Chat.UserName,
						))
					}

					if task.AssignedID != 0 {
						bot.Send(tgbotapi.NewMessage(
							task.AssignedID,
							`Задача "`+task.Name+`" назначена на @`+update.Message.Chat.UserName,
						))
					}
					task.AssignedID = update.Message.Chat.ID
					break
				}
			}
		} else if strings.HasPrefix(msg, "/unassign_") {
			unassignTaskID, err := strconv.Atoi(strings.TrimPrefix(msg, "/unassign_"))
			if err != nil {
				panic(err)
			}

			for _, task := range tasks {
				if task.ID == unassignTaskID {
					if task.AssignedID != update.Message.Chat.ID {
						bot.Send(tgbotapi.NewMessage(
							update.Message.Chat.ID,
							`Задача не на вас`,
						))
					} else {
						bot.Send(tgbotapi.NewMessage(
							task.AssignedID,
							`Принято`,
						))
						bot.Send(tgbotapi.NewMessage(
							task.CreaterID,
							`Задача "`+task.Name+`" осталась без исполнителя`,
						))
						task.AssignedID = 0
					}
					break
				}
			}
		} else if strings.HasPrefix(msg, "/resolve_") {
			resolveTaskID, err := strconv.Atoi(strings.TrimPrefix(msg, "/resolve_"))
			if err != nil {
				panic(err)
			}

			for i, task := range tasks {
				if task.ID == resolveTaskID {
					if task.AssignedID != update.Message.Chat.ID {
						bot.Send(tgbotapi.NewMessage(
							task.AssignedID,
							`Задача не на вас`,
						))
					} else {
						bot.Send(tgbotapi.NewMessage(
							task.AssignedID,
							`Задача "`+task.Name+`" выполнена`,
						))
						bot.Send(tgbotapi.NewMessage(
							task.CreaterID,
							`Задача "`+task.Name+`" выполнена @`+update.Message.Chat.UserName,
						))
						task.AssignedID = 0
					}
					tasks = append(tasks[:i], tasks[i+1:]...)
					break
				}
			}
		} else if strings.HasPrefix(msg, "/my") {
			sendMsg := ""
			for _, task := range tasks {
				if task.AssignedID == update.Message.Chat.ID {
					sendMsg += strconv.Itoa(task.ID) + ". " + task.Name + " by @" + tgUsers[task.CreaterID].Username + "\n"
					sendMsg += "/unassign_" + strconv.Itoa(task.ID) + " "
					sendMsg += "/resolve_" + strconv.Itoa(task.ID) + "\n"
				}
			}
			sendMsg = strings.TrimSuffix(sendMsg, "\n")
			bot.Send(tgbotapi.NewMessage(
				update.Message.Chat.ID,
				sendMsg,
			))
		} else if strings.HasPrefix(msg, "/owner") {
			sendMsg := ""
			for _, task := range tasks {
				if task.CreaterID == update.Message.Chat.ID {
					sendMsg += strconv.Itoa(task.ID) + ". " + task.Name + " by @" + tgUsers[task.CreaterID].Username + "\n"
					if task.AssignedID != update.Message.Chat.ID {
						sendMsg += "/assign_" + strconv.Itoa(task.ID) + "\n"
					} else {
						sendMsg += "/unassign_" + strconv.Itoa(task.ID) + " "
						sendMsg += "/resolve_" + strconv.Itoa(task.ID) + "\n"
					}
				}
			}
			sendMsg = strings.TrimSuffix(sendMsg, "\n")
			bot.Send(tgbotapi.NewMessage(
				update.Message.Chat.ID,
				sendMsg,
			))
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
