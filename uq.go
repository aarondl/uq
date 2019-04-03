package main // import "github.com/aarondl/uq"

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/aarondl/cinotify"
	"github.com/aarondl/ultimateq/bot"
	"github.com/aarondl/ultimateq/config"
	"github.com/aarondl/ultimateq/irc"

	_ "github.com/aarondl/uq/basics"
	_ "github.com/aarondl/uq/queryer"
	_ "github.com/aarondl/uq/quoter"
	_ "github.com/aarondl/uq/runner"

	_ "github.com/knivey/gitbot"
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
		var cinotifyNet, cinotifyChan string
		b.ReadConfig(func(cfg *config.Config) {
			cinotifyNet, _ = cfg.ExtGlobal().ConfigVal("", "", "cinotify_network")
			cinotifyChan, _ = cfg.ExtGlobal().ConfigVal("", "", "cinotify_channel")
		})

		b.Register("", "", irc.PRIVMSG, h)

		if len(cinotifyNet) != 0 && len(cinotifyChan) != 0 {
			b.Logger.Info("cinotify", "net", cinotifyNet, "chan", cinotifyChan)
			cinotify.ToFunc(func(name string, notification fmt.Stringer) {
				writer := b.NetworkWriter(cinotifyNet)
				if writer == nil {
					return
				}

				writer.Privmsgln(cinotifyChan, notification)
			})
		}

		cinotify.StartServer(5000)
	})

	if err != nil {
		fmt.Println(err)
	}
}
