package basics

import (
	"errors"

	"github.com/aarondl/ultimateq/bot"
	"github.com/aarondl/ultimateq/data"
	"github.com/aarondl/ultimateq/dispatch"
	"github.com/aarondl/ultimateq/dispatch/cmd"
	"github.com/aarondl/ultimateq/irc"
)

func init() {
	bot.RegisterExtension("basics", &Handler{})
}

// Handler extension
type Handler struct {
	b                *bot.Bot
	privmsgHandlerID uint64
	joinHandlerID    uint64
	opID             uint64
	pingID           uint64
}

// Init the extension
func (h *Handler) Init(b *bot.Bot) error {
	h.b = b

	h.privmsgHandlerID = b.Register("", "", irc.PRIVMSG, h)
	h.joinHandlerID = b.Register("", "", irc.JOIN, h)
	var err error
	h.opID, err = b.RegisterCmd("", "", cmd.NewAuthed(
		"basics",
		"up",
		"Ops or voices a user if they have o or v flags respectively.",
		h,
		cmd.Privmsg, cmd.AnyScope, 0, "", "#chan",
	))
	if err != nil {
		return err
	}
	h.pingID, err = b.RegisterCmd("", "", cmd.New(
		"basics",
		"ping",
		"Responds to ping commands",
		h,
		cmd.Privmsg, cmd.AnyScope,
	))
	if err != nil {
		return err
	}

	return nil
}

// Deinit the extension
func (h *Handler) Deinit(b *bot.Bot) error {
	b.Unregister(h.joinHandlerID)
	b.Unregister(h.privmsgHandlerID)
	b.UnregisterCmd(h.opID)
	b.UnregisterCmd(h.pingID)
	return nil
}

// Cmd handler
func (*Handler) Cmd(string, irc.Writer, *cmd.Event) error {
	return nil
}

// Ping from a user, let's pong
func (h *Handler) Ping(w irc.Writer, ev *cmd.Event) error {
	nick := ev.Nick()
	w.Notifyf(ev.Event, nick, "\x02%s:\x02 pong!", nick)
	return nil
}

// Up lets a user with proper access voice/op themselves.
func (h *Handler) Up(w irc.Writer, ev *cmd.Event) error {
	user := ev.StoredUser
	ch := ev.TargetChannel
	if ch == nil {
		return errors.New("Must be a channel that the bot is on")
	}
	chname := ch.Name

	if !putPeopleUp(ev.Event, chname, user, w) {
		return dispatch.MakeFlagsError("ov")
	}
	return nil
}

// Handle to check for join messages to auto-op auto-voice people on
func (h *Handler) Handle(w irc.Writer, ev *irc.Event) {
	if ev.Name == irc.JOIN {
		store := h.b.Store()
		a := store.AuthedUser(ev.NetworkID, ev.Sender)
		ch := ev.Target()
		putPeopleUp(ev, ch, a, w)
	}
}

func putPeopleUp(ev *irc.Event, ch string,
	a *data.StoredUser, w irc.Writer) (did bool) {
	if a == nil {
		return false
	}

	nick := ev.Nick()
	if a.HasFlags(ev.NetworkID, ch, "o") {
		w.Sendf("MODE %s +o :%s", ch, nick)
		return true
	} else if a.HasFlags(ev.NetworkID, ch, "v") {
		w.Sendf("MODE %s +v :%s", ch, nick)
		return true
	}

	return false
}
