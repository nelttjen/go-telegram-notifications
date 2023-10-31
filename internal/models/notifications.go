package models

type TelegramMessageNotification struct {
	ID   uint64 `gorm:"primary_key"`
	Text string `gorm:"not null"`
}

type TelegramNotification struct {
	ID        uint64                       `gorm:"primary_key"`
	ChatID    uint64                       `gorm:"not null"`
	BotID     *uint64                      `gorm:"not null"`
	Bot       *TelegramBot                 `gorm:"foreignkey:BotID"`
	TextID    *uint64                      `gorm:"not null"`
	Text      *TelegramMessageNotification `gorm:"foreignkey:TextID"`
	Processed bool                         `gorm:"not null;default:false"`
	Sent      bool                         `gorm:"not null;default:false"`
}
