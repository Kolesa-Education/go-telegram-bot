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
	Tasks *models.ToDoModel
}

func (bot *UpgradeBot) StartHandler(ctx telebot.Context) error {
	addUser := models.User{
		Name:       ctx.Sender().Username,
		TelegramId: ctx.Chat().ID,
		FirstName:  ctx.Sender().FirstName,
		LastName:   ctx.Sender().LastName,
		ChatId:     ctx.Chat().ID,
	}

	existUser, err := bot.Users.FindOne(ctx.Chat().ID)

	if err != nil {
		log.Printf("Ошибка получения пользователя %v", err)
	}

	if existUser == nil {
		err := bot.Users.Create(addUser)

		if err != nil {
			log.Printf("Ошибка создания пользователя %v", err)
		}
	}

	return ctx.Send("Привет, " + ctx.Sender().FirstName)
}


func (bot *UpgradeBot) AddTaskHandler(ctx telebot.Context) error {
	args := ctx.Args()

	if len(args) != 3 {
		return ctx.Send("Недостаточно аргументов. Введите title description date")
	}
	existUser, err := bot.Users.FindOne(ctx.Chat().ID)
	if err != nil {
		log.Printf("Ошибка получения пользователя %v", err)
	}

	newTask := models.ToDo{
		Title:       args[0],
		Description: args[1],
		EndDate:     args[2],
		UserId:      existUser.ID,
	}

	err = bot.Tasks.Create(newTask)

	if err != nil {
		log.Printf("Ошибка создания задачи %v", err)
	}

	return ctx.Send("Задача создана, " + ctx.Sender().FirstName)
}

func (bot *UpgradeBot) DeleteTaskHandler(ctx telebot.Context) error {
	args := ctx.Args()

	if len(args) != 1 {
		return ctx.Send("Недостаточно аргументов. Введите task id")
	}
	var id = args[0]
	taskId, err := strconv.Atoi(id)
	if err != nil {
		return ctx.Send("Неправильный аргумент. Нужен integer")
	}

	err = bot.Tasks.Delete(taskId)

	if err != nil {
		return ctx.Send("Не получилось удалить задачку")
	}
	return ctx.Send("Задача успешно удалена")
}

func (bot *UpgradeBot) ShowAllTasksHandler(ctx telebot.Context) error {
	existUser, err := bot.Users.FindOne(ctx.Chat().ID)
	if err != nil {
		log.Printf("Ошибка получения пользователя %v", err)
	}

	tasks, err := bot.Users.FindAllTasks(*existUser)

	if err != nil {
		return ctx.Send("Не получилось найти задачи")
	}

	var allTasks string
	for i := 0; i < len(tasks); i++ {
		var task = tasks[i]
        allTasks += fmt.Sprintf("ID: %d Заголовок: %s Описание: %s Срок: %s\n",
		task.ID, task.Title, task.Description, task.EndDate,
		)
    }
	return ctx.Send(allTasks)
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