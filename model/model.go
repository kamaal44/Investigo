package model

import "time"

// Model base model definition, including fields `ID`, `CreatedAt`, `UpdatedAt`, `DeletedAt`, which could be embedded in your models
//    type User struct {
//      gorm.Model
//    }
type Model struct {
	ID        uint       `gorm:"primary_key" yaml:"-" json:"-" csv:"'-"`
	CreatedAt time.Time  `yaml:"-" json:"-" csv:"'-"`
	UpdatedAt time.Time  `yaml:"-" json:"-" csv:"'-"`
	DeletedAt *time.Time `sql:"index" yaml:"-" json:"-" csv:"'-"`
}
