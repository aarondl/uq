package queryer

import (
	"errors"
	"regexp"
	"strings"

	"github.com/aarondl/query"
	"github.com/aarondl/ultimateq/bot"
	"github.com/aarondl/ultimateq/dispatch/cmd"
	"github.com/aarondl/ultimateq/irc"
)

var (
	sanitizeNewline = strings.NewReplacer("\r\n", " ", "\n", " ")
	rgxSpace        = regexp.MustCompile(`\s{2,}`)
	queryConf       *query.Config
)

func init() {
	bot.RegisterExtension("queryer", &Queryer{})
}

// Queryer allows for various HTTP queries to different servers.
type Queryer struct {
	privmsgHandlerID uint64
	googleHandlerID  uint64
	calcHandlerID    uint64
}

// Init the extension
func (q *Queryer) Init(b *bot.Bot) error {
	if conf := query.NewConfig("query.toml"); conf != nil {
		queryConf = conf
	} else {
		return errors.New("error loading queryer configuration")
	}

	var err error
	q.privmsgHandlerID = b.Register("", "", irc.PRIVMSG, q)
	q.googleHandlerID, err = b.RegisterCmd("", "", cmd.New(
		"query",
		"google",
		"Submits a query to Google.",
		q,
		cmd.Privmsg, cmd.AnyScope, "query...",
	))
	if err != nil {
		return err
	}
	q.calcHandlerID, err = b.RegisterCmd("", "", cmd.New(
		"query",
		"calc",
		"Submits a query to Wolfram Alpha.",
		q,
		cmd.Privmsg, cmd.AnyScope, "query...",
	))
	if err != nil {
		return err
	}

	return nil
}

// Deinit the extension
func (q *Queryer) Deinit(b *bot.Bot) error {
	b.Unregister(q.privmsgHandlerID)
	b.UnregisterCmd(q.googleHandlerID)
	b.UnregisterCmd(q.calcHandlerID)
	return nil
}

// Cmd handler to satisfy the interface, but let reflection look up
// all our methods.
func (q Queryer) Cmd(string, irc.Writer, *cmd.Event) error {
	return nil
}

// Handle traps youtube links
func (q Queryer) Handle(w irc.Writer, ev *irc.Event) {
	if !ev.IsTargetChan() {
		return
	}

	if out, err := query.YouTube(ev.Message()); len(out) != 0 {
		w.Privmsg(ev.Target(), out)
	} else if err != nil {
		nick := ev.Nick()
		w.Notice(nick, err.Error())
	}
}

// Calc something using wolfram alpha
func (Queryer) Calc(w irc.Writer, ev *cmd.Event) error {
	q := ev.Args["query"]
	nick := ev.Nick()

	if out, err := query.Wolfram(q, queryConf); len(out) != 0 {
		out = sanitize(out)

		// Ensure two lines only
		// ircmaxlen - maxhostsize - PRIVMSG - targetsize - spacing - colons
		maxlen := 2 * (510 - 62 - 7 - len(ev.Target()) - 3 - 2)
		if len(out) > maxlen {
			out = out[:maxlen-3]
			out += "..."
		}

		w.Notify(ev.Event, nick, out)
	} else if err != nil {
		w.Notice(nick, err.Error())
	}

	return nil
}

// Google some query and return the first result
func (Queryer) Google(w irc.Writer, ev *cmd.Event) error {
	q := ev.Args["query"]
	nick := ev.Nick()

	if out, err := query.Google(q, queryConf); len(out) != 0 {
		out = sanitize(out)
		w.Notify(ev.Event, nick, out)
	} else if err != nil {
		w.Notice(nick, err.Error())
	}

	return nil
}

func sanitize(str string) string {
	return rgxSpace.ReplaceAllString(sanitizeNewline.Replace(str), " ")
}
