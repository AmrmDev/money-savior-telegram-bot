package presenter

import "sync"

const invitePlaceholder = "Cole seu código de convite"

const guideFallbackText = "📖 Como usar o Money Savior\n\n" +
	"1. Registrar digitando:\n/gastei 45,90 mercado pix\n\n" +
	"2. Registrar por áudio:\nmande um áudio, por exemplo: “comprei uma água por 3 reais e paguei no crédito”\n\n" +
	"3. Ver o resumo:\n/consulta semana (ou hoje, ontem, mes, mes passado)\n\n" +
	"O método de pagamento é opcional."

var inviteCard = cardContent{
	Eyebrow:  "ACESSO POR CONVITE",
	Title:    "Opa, tudo bem?",
	Subtitle: "Pra começar, preciso do seu código de convite.",
	Steps: []cardStep{
		{Title: "Copie o código que te mandaram", Detail: "do jeito que veio, sem mudar nada"},
		{Title: "Cole aqui no chat e envie", Detail: "é só mandar como uma mensagem normal"},
	},
	Hint: "RESPONDA AQUI EMBAIXO ↓",
	Note: "Sem código? Peça a quem te convidou.",
}

var guideCard = cardContent{
	Eyebrow: "GUIA RÁPIDO",
	Title:   "Como usar o Money Savior",
	Steps: []cardStep{
		{Title: "Registrar digitando", Detail: "/gastei 45,90 mercado pix", Mono: true},
		{Title: "Registrar por áudio", Detail: "“comprei uma água por 3 reais e paguei no crédito”", Mono: true},
		{Title: "Ver o resumo", Detail: "/consulta semana · hoje · ontem · mes · mes passado", Mono: true},
	},
	FooterNote: "O MÉTODO DE PAGAMENTO É OPCIONAL",
}

type cachedCard struct {
	once    sync.Once
	content cardContent
	data    []byte
}

func (c *cachedCard) png() []byte {
	c.once.Do(func() {
		if data, err := renderCard(c.content); err == nil {
			c.data = data
		}
	})
	return c.data
}

var (
	inviteImage = &cachedCard{content: inviteCard}
	guideImage  = &cachedCard{content: guideCard}
)

func InviteRequest() Reply {
	return Reply{
		Text:             "Opa, tudo bem? 👋 Cola aqui o seu código de convite pra gente seguir com os seus gastos!",
		Image:            inviteImage.png(),
		ReplyPlaceHolder: invitePlaceholder,
	}
}

func InviteAccepted() Reply {
	return Reply{Text: "Agora sim! ✅ Seu convite foi verificado e você já pode usar o Money Savior.\n\nVou te mandar um guia rápido de como usar, fechou?"}
}

func InviteRejected() Reply {
	return Reply{Text: "Puts... esse código não funcionou. 😕\n\nConfere se copiou direitinho e manda de novo. Se continuar assim, fala com quem te convidou."}
}

func UsageGuide() Reply {
	if image := guideImage.png(); image != nil {
		return Reply{Text: "📖 Guia rápido do Money Savior", Image: image}
	}
	return Reply{Text: guideFallbackText}
}