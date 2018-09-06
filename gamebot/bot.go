package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"gopkg.in/telegram-bot-api.v4"
)

var (
	// @BotFather gives you this
	BotToken   = "487620392:AAE3yUBEVLQao_n3xD63xta1MvzpvaJIUYE"
	WebhookURL = "https://525f2cb5.ngrok.io"
)

type User struct {
	ID    int64
	Room  string
	Level int
	Bag   []string
}

var (
	rooms = map[string][]string{
		"кухня":   []string{"коридор"},
		"коридор": []string{"кухня", "комната", "улица"},
		"комната": []string{"коридор"},
		"улица":   []string{"домой"},
	}

	gameUsers = make(map[int64]*User)
)

// LEVEL:
// 0 - старт
// 1 - одеть рюкзак
// 2 - взять ключи
// 3 - взять конспекты
// 4 - применить ключи дверь

func startGameBot(ctx context.Context) error {
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

	go func() { log.Fatal(http.ListenAndServe(":8081", nil)) }()
	fmt.Println("start listen :8080")

	for update := range updates {
		msg := update.Message.Text

		if msg == "/start" {
			if _, ok := gameUsers[update.Message.Chat.ID]; !ok {
				gameUsers[update.Message.Chat.ID] = &User{
					ID:    update.Message.Chat.ID,
					Room:  "кухня",
					Level: 0,
				}
			}

			bot.Send(tgbotapi.NewMessage(
				update.Message.Chat.ID,
				"добро пожаловать в игру!",
			))
		} else if msg == "осмотреться" {
			replyMsg := ""
			switch gameUsers[update.Message.Chat.ID].Room {
			case "кухня":
				if gameUsers[update.Message.Chat.ID].Level < 3 {
					replyMsg = "ты находишься на кухне, на столе чай, надо собрать рюкзак и идти в универ. можно пройти - коридор."
				} else {
					replyMsg = "ты находишься на кухне, на столе чай, надо идти в универ. можно пройти - коридор."
				}
			case "комната":
				switch gameUsers[update.Message.Chat.ID].Level {
				case 0:
					replyMsg = "на столе: ключи, конспекты, на стуле - рюкзак. можно пройти - коридор."
				case 1:
					replyMsg = "на столе: ключи, конспекты. можно пройти - коридор."
				case 2:
					replyMsg = "на столе: конспекты. можно пройти - коридор."
				default:
					replyMsg = "пустая комната. можно пройти - коридор."
				}
			default:
				replyMsg = "нет описания."
			}

			bot.Send(tgbotapi.NewMessage(
				update.Message.Chat.ID,
				replyMsg,
			))
		} else if strings.HasPrefix(msg, "идти ") {
			room := strings.TrimPrefix(msg, "идти ")

			canGo := false
			for _, availableRoom := range rooms[gameUsers[update.Message.Chat.ID].Room] {
				if availableRoom == room {
					canGo = true
				}
			}

			replyMsg := ""
			if canGo {
				if room == "домой" {
					room = "коридор"
				}

				switch room {
				case "коридор":
					replyMsg = "ничего интересного. можно пройти - кухня, комната, улица."
					gameUsers[update.Message.Chat.ID].Room = room
				case "комната":
					replyMsg = "ты в своей комнате. можно пройти - коридор."
					gameUsers[update.Message.Chat.ID].Room = room
				case "кухня":
					replyMsg = "кухня, ничего интересного. можно пройти - коридор."
					gameUsers[update.Message.Chat.ID].Room = room
				case "улица":
					if gameUsers[update.Message.Chat.ID].Level >= 4 {
						replyMsg = "на улице уже вовсю готовятся к новому году. можно пройти - домой."
						gameUsers[update.Message.Chat.ID].Room = room
					} else {
						replyMsg = "дверь закрыта"
					}
				}
			} else {
				replyMsg = "нет пути в " + room
			}

			bot.Send(tgbotapi.NewMessage(
				update.Message.Chat.ID,
				replyMsg,
			))
		} else if strings.HasPrefix(msg, "взять ") {
			thing := strings.TrimPrefix(msg, "взять ")

			replyMsg := ""
			if gameUsers[update.Message.Chat.ID].Level == 0 {
				replyMsg = "некуда класть"
			} else if gameUsers[update.Message.Chat.ID].Level == 1 && thing == "ключи" {
				gameUsers[update.Message.Chat.ID].Level++
				replyMsg = "предмет добавлен в инвентарь: ключи"
				gameUsers[update.Message.Chat.ID].Bag = append(gameUsers[update.Message.Chat.ID].Bag, "ключи")
			} else if gameUsers[update.Message.Chat.ID].Level == 2 && thing == "конспекты" {
				gameUsers[update.Message.Chat.ID].Level++
				replyMsg = "предмет добавлен в инвентарь: конспекты"
				gameUsers[update.Message.Chat.ID].Bag = append(gameUsers[update.Message.Chat.ID].Bag, "конспекты")
			} else {
				replyMsg = "нет такого"
			}

			bot.Send(tgbotapi.NewMessage(
				update.Message.Chat.ID,
				replyMsg,
			))
		} else if strings.HasPrefix(msg, "одеть ") {
			thing := strings.TrimPrefix(msg, "одеть ")

			replyMsg := ""
			if gameUsers[update.Message.Chat.ID].Level == 0 && thing == "рюкзак" {
				gameUsers[update.Message.Chat.ID].Level++
				replyMsg = "вы одели: рюкзак"
			} else {
				replyMsg = "нет такого"
			}

			bot.Send(tgbotapi.NewMessage(
				update.Message.Chat.ID,
				replyMsg,
			))
		} else if strings.HasPrefix(msg, "применить ") {
			things := strings.Split(strings.TrimPrefix(msg, "применить "), " ")

			inBag := false
			for _, thing := range gameUsers[update.Message.Chat.ID].Bag {
				if thing == things[0] {
					inBag = true
				}
			}

			replyMsg := ""
			if gameUsers[update.Message.Chat.ID].Level == 3 && things[0] == "ключи" && things[1] == "дверь" && inBag {
				gameUsers[update.Message.Chat.ID].Level++
				replyMsg = "дверь открыта"
			} else if !inBag {
				replyMsg = "нет предмета в инвентаре - " + things[0]

			} else {
				replyMsg = "не к чему применить"
			}

			bot.Send(tgbotapi.NewMessage(
				update.Message.Chat.ID,
				replyMsg,
			))
		} else if strings.HasPrefix(msg, "/reset") {
			gameUsers[update.Message.Chat.ID].Room = "кухня"
			gameUsers[update.Message.Chat.ID].Level = 0
			gameUsers[update.Message.Chat.ID].Bag = []string{}

			bot.Send(tgbotapi.NewMessage(
				update.Message.Chat.ID,
				"состояние игры сброшено",
			))
		} else {
			bot.Send(tgbotapi.NewMessage(
				update.Message.Chat.ID,
				"неизвестная команда",
			))
		}
	}

	return nil
}

func main() {
	err := startGameBot(context.Background())
	if err != nil {
		panic(err)
	}
}
