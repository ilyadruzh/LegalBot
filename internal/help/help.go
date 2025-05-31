package help

// messages holds help text in different languages.
var messages = map[string]string{
	"en": `Available commands:
/start - start the bot
/help - show this message
/claim - submit a claim
/status - check status
/delete - delete your history
/lang - switch language

Data policy: https://github.com/owner/legalbot/blob/main/DATA_POLICY.md`,
	"ru": `Доступные команды:
/start - запустить бота
/help - показать это сообщение
/claim - подать обращение
/status - проверить статус
/delete - удалить историю
/lang - сменить язык

Политика данных: https://github.com/owner/legalbot/blob/main/DATA_POLICY.md`,
}

// Message returns bot help text in the requested language. Defaults to English.
func Message(lang string) string {
	if msg, ok := messages[lang]; ok {
		return msg
	}
	return messages["en"]
}
