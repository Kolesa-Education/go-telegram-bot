# Веб сервис на go. Telegram Bot

Телеграм бот, который сохраняет пользователя в БД и умеет играть в Камень-Ножницы-Бумага

### Команды бота
- /start - Приветствие и сохранение пользователя
- /game - Описание игры
- /try {вариант} - Отправка своего варианта

### Шаги для написания веб-сервиса
#### 1) Создание модуля

```shell
go mod init upgrade
```
_________________
#### 2) Подключаем файл конфигурации
```shell
go get github.com/BurntSushi/toml@latest
```

main.go
```go
package main

import (
   "flag"
   "log"

   "github.com/BurntSushi/toml"
)

type Config struct {
   Env string
}

func main() {
   configPath := flag.String("config", "", "Path to config file")
   flag.Parse()

   cfg := &Config{}
   _, err := toml.DecodeFile(*configPath, cfg)

   if err != nil {
       log.Fatalf("Ошибка декодирования файла конфигов %v", err)
   }

   log.Println(cfg.Env)
}
```
local.toml

```toml
Env="local"
```
_________________
#### 3) Создаем бота
```shell
go get gopkg.in/telebot.v3
```

config/local.toml
```toml
Env="local"
BotToken="5667090127:AAHj1OTBYvj2KJ_MA0-cL1Ys9oa019npuPw"
```

cmd/bot/bot.go
```go
package bot

import (
	"log"
	"time"

	"gopkg.in/telebot.v3"
)

type UpgradeBot struct {
	Bot *telebot.Bot
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
```

main.go
```go
package main

import (
	"flag"
	"log"
	"upgrade/cmd/bot"

	"github.com/BurntSushi/toml"
	"gopkg.in/telebot.v3"
)

type Config struct {
	Env      string
	BotToken string
}

func main() {
	configPath := flag.String("config", "", "Path to config file")
	flag.Parse()

	cfg := &Config{}
	_, err := toml.DecodeFile(*configPath, cfg)

	if err != nil {
		log.Fatalf("Ошибка декодирования файла конфигов %v", err)
	}

	upgradeBot := bot.UpgradeBot{
		Bot: bot.InitBot(cfg.BotToken),
	}

	upgradeBot.Bot.Handle("/start", func(ctx telebot.Context) error {
		return ctx.Send("Привет, " + ctx.Sender().FirstName)
	})
	upgradeBot.Bot.Start()
}
```
_________________
#### 4) Вынесем обработку команды в отдельную функцию, чтобы нам было удобно подключать БД
main.go
```go
upgradeBot.Bot.Handle("/start", upgradeBot.StartHandler)
```

cmd/bot/bot.go
```go
package bot

import (
	"log"
	"time"

	"gopkg.in/telebot.v3"
)

type UpgradeBot struct {
	Bot *telebot.Bot
}

func (bot *UpgradeBot) StartHandler(ctx telebot.Context) error {
	return ctx.Send("Привет, " + ctx.Sender().FirstName)
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
```
_________________
#### 5) Создаем БД

```shell
sqlite3 upgrade.db

create table users
(
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name        varchar(255),
    telegram_id INT,
    first_name  varchar(255),
    last_name   varchar(255),
    chat_id     INT,
    created_at  datetime default CURRENT_TIMESTAMP,
    updated_at  datetime,
    deleted_at  datetime
);

.quit
```

Расширение для VsCode ***SQLite***
_________________
#### 6) Подключим БД
```shell
go get gorm.io/gorm  
go get gorm.io/driver/sqlite
```

config/local.toml
```toml
Env="local"
BotToken="5667090127:AAHj1OTBYvj2KJ_MA0-cL1Ys9oa019npuPw"
Dsn="upgrade.db"
```

internal/models/users.go
```go
package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Name       string `json:"name"`
	TelegramId int64  `json:"telegram_id"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	ChatId     int64  `json:"chat_id"`
}

type UserModel struct {
	Db *gorm.DB
}

func (m *UserModel) Create(user User) error {

	result := m.Db.Create(&user)

	return result.Error
}
```

main.go
```go
package main

import (
	"flag"
	"log"
	"upgrade/cmd/bot"
	"upgrade/internal/models"

	"github.com/BurntSushi/toml"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Config struct {
	Env      string
	BotToken string
	Dsn      string
}

func main() {
	configPath := flag.String("config", "", "Path to config file")
	flag.Parse()

	cfg := &Config{}
	_, err := toml.DecodeFile(*configPath, cfg)

	if err != nil {
		log.Fatalf("Ошибка декодирования файла конфигов %v", err)
	}

	db, err := gorm.Open(sqlite.Open(cfg.Dsn), &gorm.Config{})

	if err != nil {
		log.Fatalf("Ошибка подключения к БД %v", err)
	}

	upgradeBot := bot.UpgradeBot{
		Bot:   bot.InitBot(cfg.BotToken),
		Users: &models.UserModel{Db: db},
	}

	upgradeBot.Bot.Handle("/start", upgradeBot.StartHandler)

	upgradeBot.Bot.Start()
}
```

cmd/bot/bot.go
```go
package bot

import (
	"log"
	"time"
	"upgrade/internal/models"

	"gopkg.in/telebot.v3"
)

type UpgradeBot struct {
	Bot   *telebot.Bot
	Users *models.UserModel
}

func (bot *UpgradeBot) StartHandler(ctx telebot.Context) error {
	newUser := models.User{
		Name:       ctx.Sender().Username,
		TelegramId: ctx.Chat().ID,
		FirstName:  ctx.Sender().FirstName,
		LastName:   ctx.Sender().LastName,
		ChatId:     ctx.Chat().ID,
	}

	err := bot.Users.Create(newUser)

	if err != nil {
		log.Printf("Ошибка создания пользователя %v", err)
	}

	return ctx.Send("Привет, " + ctx.Sender().FirstName)
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
```
_________________
#### 7) Добавляем проверку на существование пользователя
internal/models/users.go
```go
func (m *UserModel) FindOne(telegramId int64) (*User, error) {
	existUser := User{}
  
	result := m.Db.First(&existUser, User{TelegramId: telegramId})
  
	if result.Error != nil {
		return nil, result.Error
	}
  
	return &existUser, nil
 }
```

cmd/bot/bot.go
```go
func (bot *UpgradeBot) StartHandler(ctx telebot.Context) error {
	newUser := models.User{
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
		err := bot.Users.Create(newUser)

		if err != nil {
			log.Printf("Ошибка создания пользователя %v", err)
		}
	}

	return ctx.Send("Привет, " + ctx.Sender().FirstName)
}
```
_________________
#### 8) Делаем еще одну команду для реализации игры
cmd/bot/bot.go
```go
package bot

import (
	"log"
	"time"
	"upgrade/internal/models"

	"gopkg.in/telebot.v3"
)

type UpgradeBot struct {
	Bot   *telebot.Bot
	Users *models.UserModel
}

func (bot *UpgradeBot) StartHandler(ctx telebot.Context) error {
	newUser := models.User{
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
		err := bot.Users.Create(newUser)

		if err != nil {
			log.Printf("Ошибка создания пользователя %v", err)
		}
	}

	return ctx.Send("Привет, " + ctx.Sender().FirstName)
}

func (bot *UpgradeBot) GameHandler(ctx telebot.Context) error {
	return ctx.Send("Сыграем в камень-ножницы-бумага " +
		"Введи твой вариант в формате /try камень")
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
```

main.go
```go
package main

import (
	"flag"
	"log"
	"upgrade/cmd/bot"
	"upgrade/internal/models"

	"github.com/BurntSushi/toml"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Config struct {
	Env      string
	BotToken string
	Dsn      string
}

func main() {
	configPath := flag.String("config", "", "Path to config file")
	flag.Parse()

	cfg := &Config{}
	_, err := toml.DecodeFile(*configPath, cfg)

	if err != nil {
		log.Fatalf("Ошибка декодирования файла конфигов %v", err)
	}

	db, err := gorm.Open(sqlite.Open(cfg.Dsn), &gorm.Config{})

	if err != nil {
		log.Fatalf("Ошибка подключения к БД %v", err)
	}

	upgradeBot := bot.UpgradeBot{
		Bot:   bot.InitBot(cfg.BotToken),
		Users: &models.UserModel{Db: db},
	}

	upgradeBot.Bot.Handle("/start", upgradeBot.StartHandler)
	upgradeBot.Bot.Handle("/game", upgradeBot.GameHandler)

	upgradeBot.Bot.Start()
}
```
_________________
#### 9) Делаем команду для обработки ответа
cmd/bot/bot.go
```go
package bot

import (
	"log"
	"math/rand"
	"strings"
	"time"
	"upgrade/internal/models"

	"gopkg.in/telebot.v3"
)

type UpgradeBot struct {
	Bot   *telebot.Bot
	Users *models.UserModel
}

var gameItems = [3]string{
	"камень",
	"ножницы",
	"бумага",
}

var winSticker = &telebot.Sticker{
	File: telebot.File{
		FileID: "CAACAgIAAxkBAAEGMEZjVspD4JulorxoH7nIwco5PGoCsAACJwADr8ZRGpVmnh4Ye-0RKgQ",
	},
	Width:    512,
	Height:   512,
	Animated: true,
}

var loseSticker = &telebot.Sticker{
	File: telebot.File{
		FileID: "CAACAgIAAxkBAAEGMEhjVsqoRriJRO_d-hrqguHNlLyLvQACogADFkJrCuweM-Hw5ackKgQ",
	},
	Width:    512,
	Height:   512,
	Animated: true,
}

func (bot *UpgradeBot) StartHandler(ctx telebot.Context) error {
	newUser := models.User{
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
		err := bot.Users.Create(newUser)

		if err != nil {
			log.Printf("Ошибка создания пользователя %v", err)
		}
	}

	return ctx.Send("Привет, " + ctx.Sender().FirstName)
}

func (bot *UpgradeBot) GameHandler(ctx telebot.Context) error {
	return ctx.Send("Сыграем в камень-ножницы-бумага " +
		"Введи твой вариант в формате /try камень")
}

func (bot *UpgradeBot) TryHandler(ctx telebot.Context) error {
	attempts := ctx.Args()

	if len(attempts) == 0 {
		return ctx.Send("Вы не ввели ваш вариант")
	}

	if len(attempts) > 1 {
		return ctx.Send("Вы ввели больше одного варианта")
	}

	try := strings.ToLower(attempts[0])
	botTry := gameItems[rand.Intn(len(gameItems))]

	if botTry == "камень" {
		switch try {
		case "ножницы":
			ctx.Send("🪨")
			return ctx.Send("Камень! Ты проиграл!")
		case "бумага":
			ctx.Send("🪨")
			return ctx.Send("Камень! Ты выиграл!")
		}
	}

	if botTry == "ножницы" {
		switch try {
		case "камень":
			ctx.Send("✂️")
			return ctx.Send("Ножницы! Ты выиграл!")
		case "бумага":
			ctx.Send("✂️")
			return ctx.Send("Ножницы! Ты проиграл!")
		}
	}

	if botTry == "бумага" {
		switch try {
		case "ножницы":
			ctx.Send("📃")
			return ctx.Send("Бумага! Ты выиграл!")
		case "камень":
			ctx.Send("📃")
			return ctx.Send("Бумага! Ты проиграл!")
		}
	}

	if botTry == try {
		return ctx.Send("Ничья!")
	}

	return ctx.Send("Кажется вы ввели неверный вариант!")
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
```

main.go
```go
package main

import (
	"flag"
	"log"
	"upgrade/cmd/bot"
	"upgrade/internal/models"

	"github.com/BurntSushi/toml"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Config struct {
	Env      string
	BotToken string
	Dsn      string
}

func main() {
	configPath := flag.String("config", "", "Path to config file")
	flag.Parse()

	cfg := &Config{}
	_, err := toml.DecodeFile(*configPath, cfg)

	if err != nil {
		log.Fatalf("Ошибка декодирования файла конфигов %v", err)
	}

	db, err := gorm.Open(sqlite.Open(cfg.Dsn), &gorm.Config{})

	if err != nil {
		log.Fatalf("Ошибка подключения к БД %v", err)
	}

	upgradeBot := bot.UpgradeBot{
		Bot:   bot.InitBot(cfg.BotToken),
		Users: &models.UserModel{Db: db},
	}

	upgradeBot.Bot.Handle("/start", upgradeBot.StartHandler)
	upgradeBot.Bot.Handle("/game", upgradeBot.GameHandler)
	upgradeBot.Bot.Handle("/try", upgradeBot.TryHandler)

	upgradeBot.Bot.Start()
}
```