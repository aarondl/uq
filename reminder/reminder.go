package reminder

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/aarondl/ultimateq/bot"
	"github.com/aarondl/ultimateq/dispatch/cmd"
	"github.com/aarondl/ultimateq/irc"
	"github.com/bradj/remindme"
)

const (
	dateFormat = "January 02, 2006 at 3:04pm MST"
)

func init() {
	bot.RegisterExtension("remindme", &Reminder{})
}

// Reminder extension
type Reminder struct {
	db *remindme.DB

	cmdRemindme uint64
}

// Cmd lets reflection hook up the commands, instead of doing it here.
func (r *Reminder) Cmd(_ string, _ irc.Writer, _ *cmd.Event) error {
	return nil
}

// Init the extension
func (r *Reminder) Init(b *bot.Bot) error {
	var err error

	r.db = remindme.New()

	r.cmdRemindme, err = b.RegisterCmd("", "", cmd.New(
		"remindme",
		"remindme",
		"Sets a reminder that is associated with your current nick.",
		r,
		cmd.Privmsg, cmd.AnyScope, "duration", "message...",
	))

	if err != nil {
		return nil
	}

	go r.db.WaitForReminders()
	go r.Listener(b)

	return nil
}

// Listener listens for expired reminders
func (r *Reminder) Listener(b *bot.Bot) {
	for rem := range r.db.ExpiredReminders {
		fmt.Println(rem)
		w := b.NetworkWriter(rem.Network)

		if len(rem.Channel) == 0 {
			w.Noticef(rem.Author, "\x02Remindme:\x02 %v", rem.Body)
			continue
		}

		w.Privmsgf(rem.Channel, "\x02Remindme (\x02%s\x02):\x02 %s", rem.Author, rem.Body)
	}
}

// Deinit the extension
func (r *Reminder) Deinit(b *bot.Bot) error {
	b.UnregisterCmd(r.cmdRemindme)

	return nil
}

// Remindme creates a reminder
func (r *Reminder) Remindme(w irc.Writer, ev *cmd.Event) error {
	nick := ev.Nick()
	duration := ev.Args["duration"]
	message := ev.Args["message"]

	if len(message) == 0 {
		w.Notifyf(ev.Event, nick, "\x02Remindme:\x02 You didn't supply a message")
		return nil
	}

	end, err := getEndTime(duration)

	if err != nil {
		w.Notifyf(ev.Event, nick, err.Error())
		return nil
	}

	var channel string

	if ev.Event.IsTargetChan() {
		channel = ev.Event.Target()
	}

	fmt.Println("channel", ev.Event.Target(), ev.Event.IsTargetChan())

	r.db.Add(remindme.Reminder{
		Author:  nick,
		Body:    message,
		EndTime: end,
		Network: ev.NetworkID,
		Channel: channel,
	})

	w.Notifyf(ev.Event, nick, "\x02Remindme:\x02 You will be notified at %s", end.Format("2006-01-02 15:04"))

	return nil
}

func getEndTime(duration string) (end time.Time, err error) {
	const formattingError = "\x02Remindme:\x02 Improperly formatted duration. Example: 5d, 1h, 3m, 12w"

	numberStr := duration[:len(duration)-1]
	number, err := strconv.ParseInt(numberStr, 10, 64)

	if err != nil {
		return end, errors.New(formattingError)
	}

	units := duration[len(duration)-1]

	switch units {
	case 'm':
		end = time.Now().Add(time.Duration(number) * time.Minute)
	case 'h':
		end = time.Now().Add(time.Duration(number) * time.Hour)
	case 'd':
		end = time.Now().AddDate(0, 0, int(number))
	case 'w':
		end = time.Now().AddDate(0, 0, int(7*number))
	default:
		return end, errors.New(formattingError)
	}

	return end, nil
}
