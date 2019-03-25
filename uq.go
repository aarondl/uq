package main // import "github.com/aarondl/uq"

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/aarondl/ultimateq/bot"
	"github.com/aarondl/ultimateq/irc"

	_ "github.com/aarondl/uq/basics"
	_ "github.com/aarondl/uq/queryer"
	_ "github.com/aarondl/uq/quoter"
	_ "github.com/aarondl/uq/runner"
)

// Handler extension
type Handler struct {
}

// Handle allows the "do" command from a hardcoded bot owner
func (h *Handler) Handle(w irc.Writer, ev *irc.Event) {
	flds := strings.Fields(ev.Message())
	if ev.Nick() == "Aaron" && flds[0] == "do" {
		w.Send(strings.Join(flds[1:], " "))
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	h := &Handler{}

	err := bot.Run(func(b *bot.Bot) {
		b.Register("", "", irc.PRIVMSG, h)
	})

	if err != nil {
		fmt.Println(err)
	}
}
