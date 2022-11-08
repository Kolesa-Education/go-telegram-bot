package models

import "gorm.io/gorm"

type ToDo struct {
	gorm.Model
	ID 			uint 		`json:"id"`
	Title      	string 		`json:"title"`
	Description string 		`json:"description"`
	EndDate  	string 		`json:"end_date"`
	UserId      uint   		`json:"user_id"`
}

type ToDoModel struct {
	Db *gorm.DB
}

func (m *ToDoModel) Create(addtoo ToDo) error {
	save := m.Db.Create(&addtoo)
	return save.Error
}

func (m *ToDoModel) Delete(todoid int) error {
	toDos := ToDo{}
	result := m.Db.Delete(&toDos, todoid)
	return result.Error
}