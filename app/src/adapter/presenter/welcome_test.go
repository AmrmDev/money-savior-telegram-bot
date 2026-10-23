package presenter

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestInviteRequest(t *testing.T) {
	reply := InviteRequest()

	if len(reply.Image) == 0 {
		t.Error("expected the invite card image")
	}
	if !strings.Contains(reply.Text, "código de convite") {
		t.Errorf("unexpected caption: %q", reply.Text)
	}

	size := utf8.RuneCountInString(reply.ReplyPlaceHolder)
	if size < 1 || size > 64 {
		t.Errorf("the placeholder must have between 1 and 64 characters, got %d", size)
	}
}

func TestInviteAccepted(t *testing.T) {
	reply := InviteAccepted()

	if reply.Image != nil {
		t.Error("the confirmation is text only")
	}
	if !strings.Contains(reply.Text, "verificado") || !strings.Contains(reply.Text, "guia rápido") {
		t.Errorf("unexpected text: %q", reply.Text)
	}
}

func TestInviteRejected(t *testing.T) {
	reply := InviteRejected()

	if reply.Image != nil || reply.ReplyPlaceHolder != "" {
		t.Error("the rejection is a plain text message")
	}
	if !strings.Contains(reply.Text, "não funcionou") {
		t.Errorf("unexpected text: %q", reply.Text)
	}
}

func TestUsageGuide(t *testing.T) {
	reply := UsageGuide()

	if len(reply.Image) == 0 {
		t.Error("expected the guide image")
	}
	if reply.Text == "" {
		t.Error("expected a caption")
	}
}

func TestGuideFallbackText_ExplainsEveryCommand(t *testing.T) {
	for _, want := range []string{"/gastei", "/consulta", "áudio"} {
		if !strings.Contains(guideFallbackText, want) {
			t.Errorf("the text fallback should mention %q", want)
		}
	}
}