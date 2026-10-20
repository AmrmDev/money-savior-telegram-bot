package presenter

import (
	"fmt"

	"money-savior-telegram-bot/app/src/core/domain"
)

func VoiceTooLong(maxSeconds int) Reply {
	return Reply{Text: fmt.Sprintf("⚠️ Áudio muito longo. Mande até %d segundos.", maxSeconds)}
}

func VoiceNotUnderstood() Reply {
	return Reply{Text: "🤔 Não entendi o áudio. Tenta de novo, por exemplo: \"quanto eu gastei essa semana?\""}
}

func VoiceExpenseUseCommand() Reply {
	return Reply{Text: "🤔 Não consegui montar a confirmação desse gasto. Use:\n/gastei 21,90 uber pix"}
}

func VoiceError() Reply {
	return Reply{Text: "❌ Não consegui processar o áudio agora. Tenta de novo em instantes."}
}

func ExpensePreview(e domain.Expense, confirmData, cancelData string) Reply {
	method := ""
	if e.Method != "" {
		method = " (" + e.Method + ")"
	}

	return Reply{
		Text: fmt.Sprintf("🎙️ Entendi: %s, %s%s.\nRegistrar?", e.Category, formatMoney(e.AmountCents), method),
		Buttons: []Button{
			{Label: "✅ Confirmar", Data: confirmData},
			{Label: "❌ Cancelar", Data: cancelData},
		},
	}
}