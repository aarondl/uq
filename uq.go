package main // import "github.com/aarondl/uq"

import (
	"fmt"
	"log"
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
	_ "github.com/aarondl/uq/reminder"

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

type ciLogger struct {
	b *bot.Bot
}

func (c ciLogger) Write(msg []byte) (int, error) {
	c.b.Logger.Error(string(msg))
	return len(msg), nil
}

func main() {
	rand.Seed(time.Now().UnixNano())

	h := &Handler{}

	err := bot.Run(func(b *bot.Bot) {
		var cinotifyNet, cinotifyChan, cinotifyBind string
		b.ReadConfig(func(cfg *config.Config) {
			cinotifyNet, _ = cfg.ExtGlobal().ConfigVal("", "", "cinotify_network")
			cinotifyChan, _ = cfg.ExtGlobal().ConfigVal("", "", "cinotify_channel")
			cinotifyBind, _ = cfg.ExtGlobal().ConfigVal("", "", "cinotify_bind")
		})

		b.Register("", "", irc.PRIVMSG, h)

		if len(cinotifyNet) != 0 && len(cinotifyChan) != 0 {
			cinotify.Logger = log.New(ciLogger{b: b}, "", 0)

			b.Logger.Info("cinotify", "net", cinotifyNet, "chan", cinotifyChan)
			cinotify.To(cinotify.NotifyFunc(func(name string, notification fmt.Stringer) {
				writer := b.NetworkWriter(cinotifyNet)
				if writer == nil {
					return
				}

				writer.Privmsgln(cinotifyChan, notification)
			}))

			go func() {
				if err := cinotify.StartServer(cinotifyBind); err != nil {
					b.Logger.Error("cinotify", "err", err)
				}
			}()
		}

	})

	if err != nil {
		fmt.Println(err)
	}
}
