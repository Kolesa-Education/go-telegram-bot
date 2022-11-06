package models

import (
	"gorm.io/gorm"
)

type Task struct {
	gorm.Model
	Title       string `json:"title"`
	Description string `json:"description"`
	EndDate     int64  `json:"end_date"`
	TelegramId  int64  `json:"Telegram_Id"`
}

type TaskModel struct {
	Db *gorm.DB
}

func (m *TaskModel) CreateTask(task Task) error {
	result := m.Db.Create(&task)
	return result.Error
}

func (m *TaskModel) DeleteTask(id int) error {
	result := m.Db.Delete(&Task{}, id)
	return result.Error
}
