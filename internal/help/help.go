package help

// Message returns bot help text.
func Message() string {
	return `Available commands:
/start - start the bot
/help - show this message
/claim - submit a claim
/status - check status
/delete - delete your history

Data policy: https://github.com/owner/legalbot/blob/main/DATA_POLICY.md`
}
