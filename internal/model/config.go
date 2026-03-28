package model

import "gorm.io/gorm"

type Config struct {
	gorm.Model

	UserID int64

	Role               string
	APIBaseURL         string
	APIKey             string
	APIModel           string
	Rounds             int
	VoiceEnabled       bool
	VoiceBaseURL       string
	VoiceAuthorization string
	VoiceRole          string
}
