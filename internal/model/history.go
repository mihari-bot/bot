package model

import "gorm.io/gorm"

type Round struct {
	gorm.Model

	Role       string
	UserID     int64
	DataInJSON string
}
