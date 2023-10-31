package telegram

type Bot interface {
	SendMessage(chatID uint64, text string) error
	InitBot() error
	RunExecutor()
}

type bot struct {
	token string
	Bot
}

func NewBot(token string) Bot {
	return &bot{
		token: token,
	}
}

func (b *bot) SendMessage(chatID uint64, text string) error {
	return nil
}

func (b *bot) InitBot() error {
	return nil
}

func (b *bot) RunExecutor() {

}
