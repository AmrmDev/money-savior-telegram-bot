package presenter

func Welcome(username string) Reply {
	if username == "" {
		return Reply{Text: "Olá! 👋 Bem-vindo ao Money Savior!"}
	}
	return Reply{Text: "Olá, @" + username + "! 👋 Bem-vindo ao Money Savior!"}
}

func Help() Reply {
	return Reply{Text: "Comandos disponíveis:\n/start - Iniciar o bot\n/help - Ver esta ajuda"}
}
