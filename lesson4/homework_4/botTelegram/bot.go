package main

import (
	"encoding/json"
	"firstGoProject/lesson4/homework_4/handler"
	handler2 "firstGoProject/lesson4/microPet/handler"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/gorilla/mux"
	tgbotapi "github.com/skinass/telegram-bot-api/v5"
)

const (
	BotToken   = "8897544923:AAF5nOsxJgwcRr4UfmIovzvy-0RceTCJEcw"
	WebhookURL = "https://894ba73c97c14cab-151-242-104-153.serveousercontent.com"
)

func sendTextToTelegram(botTG *tgbotapi.BotAPI, update tgbotapi.Update, title string) {
	reply := tgbotapi.NewMessage(update.Message.Chat.ID, title)

	_, err := botTG.Send(reply)
	if err != nil {
		log.Println("Не удалось отправить сообщение в Телеграм. Ошибка: ", err)
	}
}

func main() {

	bot, err := tgbotapi.NewBotAPI(BotToken)
	if err != nil {
		fmt.Println("Can't start BOT with Token! Error:", err)
		return
	}

	wh, err := tgbotapi.NewWebhook(WebhookURL)

	if err != nil {
		fmt.Println("Can't create new WebHook! Error:", err)
		return
	}

	_, err = bot.Request(wh)

	if err != nil {
		fmt.Println("Bot can't request with WebHook! Error:", err)
		return
	}

	updates := make(chan tgbotapi.Update, 1000) // Создали свой канал, куда будут прилетать сообщения

	tasksForHW := make(map[int]*handler.TaskForHW)
	router := mux.NewRouter().StrictSlash(true)
	mu := &sync.Mutex{}

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var update tgbotapi.Update

		err := json.NewDecoder(r.Body).Decode(&update)
		if err != nil {
			handler2.EncodeMessageAndCheckError(w, http.StatusBadRequest, "Bad Request")
			return
		}

		// Если все нормально и получилось распарсить сообщение от телеграма, то кладем его в канал updates
		updates <- update

		// Сообщаем телеграм, что обновление принято
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodPost)

	// Настроили путь для проверки работы сервера
	router.HandleFunc("/state", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("All is working!"))
		if err != nil {
			fmt.Println("Can't write state as response! Error:", err)
		}
	})

	handlerForHW := &handler.HandlerForHW{
		Tasks:  tasksForHW,
		Mu:     mu,
		Router: router,
		LastID: 0,
	}

	serverForHW := http.Server{
		Addr:         ":3000",
		Handler:      handlerForHW,
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  5 * time.Second,
	}

	fmt.Println("Starting server at port 3000...")

	go func() {
		log.Fatalln("http error:", serverForHW.ListenAndServe())
	}()

	for update := range updates {

		if update.Message == nil || update.Message.From == nil {
			continue
		}

		msg := strings.TrimSpace(update.Message.Text)
		command, title, _ := strings.Cut(msg, " ")
		title = strings.TrimSpace(title) // Сбрасываем лишние пробелы

		switch command {
		case "/help":
			textForCommandHelp := `Доступные команды:
			/tasks — показать все задачи
			/new Название задачи — создать задачу
			/assign_ID — назначить задачу себе
			/unassign_ID — снять с себя задачу
			/done_ID — выполнить задачу и перевести ее в статус "ВЫПОЛНЕНО"
			/undone_ID — отменить выполнение задачи и перевести ее в статус "В ПРОЦЕССЕ"
			/resolve_ID — завершить задачу и удалить её из списка
			/my — показать задачи, назначенные мне
			/owner — показать задачи, созданные мной

			Вместо ID укажи номер задачи, например: /assign_1`

			sendTextToTelegram(bot, update, textForCommandHelp)

		case "/new":
			if utf8.RuneCountInString(title) > 3 {
				newCreatedTask := handlerForHW.CreateNewTask(title, update.Message.From.ID)
				var taskAsText string
				if newCreatedTask.Status {
					taskAsText = fmt.Sprintf("ID задачи - %d\nЗадача заведена пользователем с ID %d\nВыполнение задачи назначено на пользователя с ID: %d\nОписание задачи: %s\nСтатус задачи: ВЫПОЛНЕНО", newCreatedTask.ID, newCreatedTask.OwnerID, newCreatedTask.AssigneeID, newCreatedTask.Title)
				} else {
					taskAsText = fmt.Sprintf("ID задачи - %d\nЗадача заведена пользователем с ID %d\nВыполнение задачи назначено на пользователя с ID: %d\nОписание задачи: %s\nСтатус задачи: В ПРОЦЕССЕ", newCreatedTask.ID, newCreatedTask.OwnerID, newCreatedTask.AssigneeID, newCreatedTask.Title)
				}

				sendTextToTelegram(bot, update, taskAsText)

			} else {
				sendTextToTelegram(bot, update, "Напиши после '/new' какую-то задачу через пробел. Ее длина должна быть больше 3 символов, иначе это странная задача...")
			}

		case "/tasks":
			tasksToShow := handlerForHW.ShowAllTasks()
			if len(tasksToShow) == 0 {
				sendTextToTelegram(bot, update, "На данный момент в списке нет задач")
			} else {
				var tasksToShowAsString strings.Builder
				tasksToShowAsString.WriteString("Все задачи:\n")

				for _, task := range tasksToShow {
					// Создали строку и нормально распарсили ее в "string"
					var taskToAdd string
					if task.Status {
						taskToAdd = fmt.Sprintf("ID задачи -%d\nЗадачу завел пользователь с ID %d\nЗадача назначена на пользователя с ID '%d'\nОписание задачи: %s\nСтатус задачи: ВЫПОЛНЕНО", task.ID, task.OwnerID, task.AssigneeID, task.Title)
					} else {
						taskToAdd = fmt.Sprintf("ID задачи -%d\nЗадачу завел пользователь с ID %d\nЗадача назначена на пользователя с ID '%d'\nОписание задачи: %s\nСтатус задачи: В ПРОЦЕССЕ", task.ID, task.OwnerID, task.AssigneeID, task.Title)
					}
					tasksToShowAsString.WriteString(taskToAdd)
					tasksToShowAsString.WriteString("\n\n===\n\n")
				}
				sendTextToTelegram(bot, update, tasksToShowAsString.String())
			}
		default:
			if idText, ok := strings.CutPrefix(command, "/assign_"); ok {
				taskID, err := strconv.Atoi(idText)
				if err != nil || taskID <= 0 || title != "" {
					sendTextToTelegram(bot, update, "Ошибка. Команда /assign_ID используется не так. Обратитесь к команде /help")
					continue // Некорректный ID: /assign_abc, /assign_ или /assign_0
				}
				// taskID — готовый ID задачи, например 12
				result := handlerForHW.AssigneeTask(taskID, update.Message.From.ID)
				if result {
					sendTextToTelegram(bot, update, "Задача успешно переназначена")
				} else {
					sendTextToTelegram(bot, update, "Ошибка. Задача не переназначена. Обратитесь к команде /help")
				}
			} else {
				sendTextToTelegram(bot, update, "Непонятная команда. Чтобы отобразить список существующих команд, напишите /help")
			}
		}

	}

}
