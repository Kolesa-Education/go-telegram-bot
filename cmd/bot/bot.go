package bot

import (
	"fmt"
	"log"
	"strconv"
	"time"
	"upgrade/internal/models"

	"gopkg.in/telebot.v3"
)

type UpgradeBot struct {
	Bot   *telebot.Bot
	Users *models.UserModel
	Tasks *models.TaskModel
}

func (bot *UpgradeBot) StartHandler(ctx telebot.Context) error {
	newUser := models.User{
		Name:       ctx.Sender().Username,
		TelegramId: ctx.Sender().ID,
		FirstName:  ctx.Sender().FirstName,
		LastName:   ctx.Sender().LastName,
		ChatId:     ctx.Chat().ID,
	}

	existUser, err := bot.Users.FindOne(ctx.Chat().ID)

	if err != nil {
		log.Printf("Ошибка получения пользователя %v", err)
	}

	if existUser == nil {
		err := bot.Users.Create(newUser)

		if err != nil {
			log.Printf("Ошибка создания пользователя %v", err)
		}
	}

	return ctx.Send("Привет, " + ctx.Sender().FirstName)

}

func (bot *UpgradeBot) HelpHandler(ctx telebot.Context) error {
	return ctx.Send("Введи данные о задаче в следующем виде: /название, описание, дедлайн(гггг-мм-дд)")
}

func (bot *UpgradeBot) NewTaskHandler(ctx telebot.Context) error {
	attempts := ctx.Text()
	newTask := models.Task{
		Title:       string(attempts[0]),
		Description: string(attempts[1]),
		EndDate:     int64(attempts[2]),
		TelegramId:  ctx.Chat().ID,
	}
	err := bot.Tasks.CreateTask(newTask)
	if err != nil {
		log.Printf("Ошибка создания задачи %v", err)
	}
	return ctx.Send("Задача успешно добавлена, ")
}

func (bot *UpgradeBot) ShowTaskHandler(ctx telebot.Context) error {
	users, err := bot.Users.GetAll()
	if err != nil {
		return ctx.Send("Error: ", err.Error())
	}
	var tasksMsg string

	for _, user := range users {
		if user.ChatId == ctx.Chat().ID {
			for _, task := range user.Tasks {
				tasksMsg += fmt.Sprintf("Title: %v\nDescription: %v\nDeadline: %v\n",
					task.Title, task.Description, task.EndDate)
			}
		}
	}

	if tasksMsg == "" {
		return ctx.Send("Задачи не найдены")
	}
	return ctx.Send(tasksMsg)
}

func (bot *UpgradeBot) DeleteTaskHandler(ctx telebot.Context) error {
	arg := ctx.Args()
	if len(arg) == 0 {
		return ctx.Send("Укажи id задачи")
	}
	if len(arg) > 1 {
		return ctx.Send("Укажи один id задачи")
	}
	id, err := strconv.Atoi(arg[0])
	if err != nil {
		return ctx.Send("Некорректный id задачи")
	}
	err = bot.Tasks.DeleteTask(id)
	if err != nil {
		return ctx.Send("Ошибка удаления")
	}
	return ctx.Send("Задача успешно удалена")
}

func InitBot(token string) *telebot.Bot {
	pref := telebot.Settings{
		Token:  token,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}
	b, err := telebot.NewBot(pref)
	if err != nil {
		log.Fatalf("Ошибка при инициализации бота %v", err)
	}
	return b
}
