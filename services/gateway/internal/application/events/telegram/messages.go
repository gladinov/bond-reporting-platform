package telegram

import "fmt"

const msgHelp = `Я бот для получение аналитики с Т-Инвестиций. 
В данный момент обладаю следующими командами:
/start - для запуска тг-бота,
/help - хелп, сейчас мы тут,
/accounts - получение списка счетов по предоставленому токену`

// const msgHello = "Приветствую. Для дальнейшей работы пришли токен от Тинькофф АПИ 👾\n\n" + msgHelp
const msgHello = "Приветствую. Для дальнейшей работы пришлите токен от Тинькофф АПИ 👾\n\n"

const (
	msgUnknownCommand = "Unknown command 🤔"
	msgBadToken       = "Ошибка подключения по переданному токену"
	msgNoToken        = "Не предоставлен токен. Для дальнейшей работы пришлите токен от Тинькофф АПИ 👾\n\n"
	msgIncorrectToken = "Некорректный токен 👾\n\n"
	msgTrueToken      = "Токен верный и сохранен для работы в этом чате"
	msgInternalErr    = "Внутренняя ошибка. Попробуйте еще раз"
)

func msgKafka(traceId string) string {
	return fmt.Sprintf("Запрос %s находится в обработке", traceId)
}
