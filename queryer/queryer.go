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
	youtubeID       uint64
	googleHandlerID uint64
	calcHandlerID   uint64
	yrID            uint64
	shortenID       uint64
	githubID        uint64
}

// Init the extension
func (q *Queryer) Init(b *bot.Bot) error {
	if conf := query.NewConfig("query.toml"); conf != nil {
		queryConf = conf
	} else {
		return errors.New("error loading queryer configuration")
	}

	var err error
	q.youtubeID = b.Register("", "", irc.PRIVMSG, q)
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
	q.googleHandlerID, err = b.RegisterCmd("", "", cmd.New(
		"query",
		"bing",
		"Submits a query to Bing.",
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
	q.yrID, err = b.RegisterCmd("", "", cmd.New(
		"query",
		"yr",
		"Get weather from norway based sources",
		q,
		cmd.Privmsg, cmd.AnyScope, "query...",
	))
	if err != nil {
		return err
	}
	q.shortenID, err = b.RegisterCmd("", "", cmd.New(
		"query",
		"shorten",
		"Shorten a URL with the goo.gl url shortener",
		q,
		cmd.Privmsg, cmd.AnyScope, "query...",
	))
	if err != nil {
		return err
	}
	q.githubID, err = b.RegisterCmd("", "", cmd.New(
		"query",
		"stars",
		"Count the number of stars a user or repo has on github",
		q,
		cmd.Privmsg, cmd.AnyScope, "userorrepo",
	))
	if err != nil {
		return err
	}

	return nil
}

// Deinit the extension
func (q *Queryer) Deinit(b *bot.Bot) error {
	b.Unregister(q.youtubeID)
	b.UnregisterCmd(q.googleHandlerID)
	b.UnregisterCmd(q.calcHandlerID)
	b.UnregisterCmd(q.yrID)
	b.UnregisterCmd(q.shortenID)
	b.UnregisterCmd(q.githubID)
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

	if out, err := query.YouTube(ev.Message(), queryConf); len(out) != 0 {
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

// Bing some query and return the first result
func (Queryer) Bing(w irc.Writer, ev *cmd.Event) error {
	q := ev.Args["query"]
	nick := ev.Nick()

	if out, err := query.Bing(q, queryConf); len(out) != 0 {
		out = sanitize(out)
		w.Notify(ev.Event, nick, out)
	} else if err != nil {
		w.Notice(nick, err.Error())
	}

	return nil
}

// Yr weather for norway based residents
func (Queryer) Yr(w irc.Writer, ev *cmd.Event) error {
	q := ev.Args["query"]
	nick := ev.Nick()

	if out, err := query.WeatherYR(q, queryConf); len(out) != 0 {
		out = sanitize(out)
		w.Notify(ev.Event, nick, out)
	} else if err != nil {
		w.Notice(nick, err.Error())
	}

	return nil
}

// Shorten a url
func (Queryer) Shorten(w irc.Writer, ev *cmd.Event) error {
	q := ev.Args["query"]
	nick := ev.Nick()

	if out, err := query.GetShortURL(q, queryConf); len(out) != 0 {
		w.Notifyf(ev.Event, nick, "\x02Shorten:\x02 %s", sanitize(out))
	} else if err != nil {
		w.Notice(nick, err.Error())
	}

	return nil
}

// Stars counts github stars
func (Queryer) Stars(w irc.Writer, ev *cmd.Event) error {
	q := ev.Args["userorrepo"]
	nick := ev.Nick()

	if out, err := query.GithubStars(q, queryConf); err != nil {
		w.Notice(nick, err.Error())
	} else if out != 0 {
		w.Notifyf(ev.Event, nick, "\x02Github stars (%s):\x02 %d", strings.ToLower(q), out)
	} else {
		w.Notify(ev.Event, nick, "\x02Github stars:\x02 could not find repo %s", strings.ToLower(q))
	}

	return nil
}

func sanitize(str string) string {
	return rgxSpace.ReplaceAllString(sanitizeNewline.Replace(str), " ")
}
