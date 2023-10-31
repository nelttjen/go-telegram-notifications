package models

type TelegramBot struct {
	ID       uint64 `gorm:"primary_key"`
	BotToken string `gorm:"not null;unique"`
	BotHost  string `gorm:"not null;unique"`
	Enabled  bool   `gorm:"not null;default:true"`
}
