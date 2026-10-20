package presenter

func ExpenseCancelled() Reply {
	return Reply{Text: "Cancelado. Nada foi registrado."}
}

func CallbackInvalid() Reply {
	return Reply{Text: "⚠️ Esse botão não vale mais. Mande o áudio de novo."}
}
