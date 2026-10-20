package presenter

import "fmt"

func VoiceTooLong(maxSeconds int) Reply {
	return Reply{Text: fmt.Sprintf("⚠️ Áudio muito longo. Mande até %d segundos.", maxSeconds)}
}

func VoiceNotUnderstood() Reply {
	return Reply{Text: "🤔 Não entendi o áudio. Tenta de novo, por exemplo: \"quanto eu gastei essa semana?\""}
}

func VoiceExpenseNotAvailable() Reply {
	return Reply{Text: "Ainda não registro gastos por voz. Por enquanto use:\n/gastei 21,90 uber pix"}
}

func VoiceError() Reply {
	return Reply{Text: "❌ Não consegui processar o áudio agora. Tenta de novo em instantes."}
}