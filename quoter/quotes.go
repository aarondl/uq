package quoter

import (
	"database/sql"
	"strconv"
	"strings"

	"github.com/aarondl/quotes"
	"github.com/aarondl/ultimateq/bot"
	"github.com/aarondl/ultimateq/dispatch/cmd"
	"github.com/aarondl/ultimateq/irc"
)

const (
	dateFormat = "January 02, 2006 at 3:04pm MST"
)

func init() {
	bot.RegisterExtension("quoter", &Quoter{
		Listen:    ":8000",
		ServerURI: "https://quotes.bitforge.ca",
	})
}

// Quoter extension
type Quoter struct {
	Listen    string
	ServerURI string

	db *quotes.QuoteDB

	quoteID     uint64
	quotesID    uint64
	detailsID   uint64
	addQuoteID  uint64
	delQuoteID  uint64
	editQuoteID uint64
	quoteWebID  uint64
	upvoteID    uint64
	downvoteID  uint64
	unvoteID    uint64
}

// Cmd lets reflection hook up the commands, instead of doing it here.
func (q *Quoter) Cmd(_ string, _ irc.Writer, _ *cmd.Event) error {
	return nil
}

// Init the extension
func (q *Quoter) Init(b *bot.Bot) error {
	qdb, err := quotes.OpenDB("quotes.sqlite3")
	if err != nil {
		return err
	}

	q.db = qdb
	qdb.StartServer(q.Listen)

	q.quoteID, err = b.RegisterCmd("", "", cmd.New(
		"quote",
		"quote",
		"Retrieves a quote. Randomly selects a quote if no id is provided.",
		q,
		cmd.Privmsg, cmd.AnyScope, "[id]",
	))
	if err != nil {
		return nil
	}
	q.quotesID, err = b.RegisterCmd("", "", cmd.New(
		"quote",
		"quotes",
		"Shows the number of quotes in the database.",
		q,
		cmd.Privmsg, cmd.AnyScope,
	))
	if err != nil {
		return nil
	}
	q.detailsID, err = b.RegisterCmd("", "", cmd.New(
		"quote",
		"details",
		"Gets the details for a specific quote.",
		q,
		cmd.Privmsg, cmd.AnyScope, "id",
	))
	if err != nil {
		return nil
	}
	q.addQuoteID, err = b.RegisterCmd("", "", cmd.New(
		"quote",
		"addquote",
		"Adds a quote to the database.",
		q,
		cmd.Privmsg, cmd.AnyScope, "quote...",
	))
	if err != nil {
		return nil
	}
	q.delQuoteID, err = b.RegisterCmd("", "", cmd.NewAuthed(
		"quote",
		"delquote",
		"Removes a quote from the database.",
		q,
		cmd.Privmsg, cmd.AnyScope, 0, "Q", "id",
	))
	if err != nil {
		return nil
	}
	q.editQuoteID, err = b.RegisterCmd("", "", cmd.NewAuthed(
		"quote",
		"editquote",
		"Edits an existing quote.",
		q,
		cmd.Privmsg, cmd.AnyScope, 0, "Q", "id", "quote...",
	))
	if err != nil {
		return nil
	}
	q.quoteWebID, err = b.RegisterCmd("", "", cmd.New(
		"quote",
		"quoteweb",
		"Shows the address for the quote webserver.",
		q,
		cmd.Privmsg, cmd.AnyScope,
	))
	if err != nil {
		return nil
	}
	q.upvoteID, err = b.RegisterCmd("", "", cmd.New(
		"quote",
		"up",
		"Upvotes a quote",
		q,
		cmd.Privmsg, cmd.AnyScope,
		"id",
	))
	if err != nil {
		return nil
	}
	q.downvoteID, err = b.RegisterCmd("", "", cmd.New(
		"quote",
		"down",
		"Downvotes a quote",
		q,
		cmd.Privmsg, cmd.AnyScope,
		"id",
	))
	if err != nil {
		return nil
	}
	q.unvoteID, err = b.RegisterCmd("", "", cmd.New(
		"quote",
		"unvote",
		"Unvotes a quote",
		q,
		cmd.Privmsg, cmd.AnyScope,
		"id",
	))
	if err != nil {
		return nil
	}

	return nil
}

// Deinit the extension
func (q *Quoter) Deinit(b *bot.Bot) error {
	defer q.db.Close()

	b.UnregisterCmd(q.quotesID)
	b.UnregisterCmd(q.detailsID)
	b.UnregisterCmd(q.addQuoteID)
	b.UnregisterCmd(q.delQuoteID)
	b.UnregisterCmd(q.editQuoteID)
	b.UnregisterCmd(q.quoteWebID)
	b.UnregisterCmd(q.upvoteID)
	b.UnregisterCmd(q.downvoteID)
	b.UnregisterCmd(q.unvoteID)

	return nil
}

// Addquote to db
func (q *Quoter) Addquote(w irc.Writer, ev *cmd.Event) error {
	nick := ev.Nick()
	quote := ev.Args["quote"]
	if len(quote) == 0 {
		return nil
	}

	id, err := q.db.AddQuote(nick, quote)
	if err != nil {
		w.Noticef(nick, "\x02Quote:\x02 %v", err)
	} else {
		w.Notifyf(ev.Event, nick, "\x02Quote:\x02 Added quote #%d", id)
	}

	if id != 0 && strings.HasPrefix(strings.ToLower(nick), "scott") {
		// Don't care if this fails
		_, _ = q.db.Downvote(int(id), "autodown")
	}

	return nil
}

// Delquote from db
func (q *Quoter) Delquote(w irc.Writer, ev *cmd.Event) error {
	nick := ev.Nick()
	id, err := strconv.Atoi(ev.Args["id"])

	if err != nil {
		w.Notice(nick, "\x02Quote:\x02 Not a valid id.")
		return nil
	}
	if did, err := q.db.DelQuote(int(id)); err != nil {
		w.Noticef(nick, "\x02Quote:\x02 %v", err)
	} else if !did {
		w.Noticef(nick, "\x02Quote:\x02 Could not find quote %d.", id)
	} else {
		w.Noticef(nick, "\x02Quote:\x02 Quote %d deleted.", id)
	}
	return nil
}

// Editquote in db
func (q *Quoter) Editquote(w irc.Writer, ev *cmd.Event) error {
	nick := ev.Nick()
	quote := ev.Args["quote"]
	id, err := strconv.Atoi(ev.Args["id"])

	if len(quote) == 0 {
		return nil
	}

	if err != nil {
		w.Notice(nick, "\x02Quote:\x02 Not a valid id.")
		return nil
	}
	if did, err := q.db.EditQuote(int(id), quote); err != nil {
		w.Noticef(nick, "\x02Quote:\x02 %v", err)
	} else if !did {
		w.Noticef(nick, "\x02Quote:\x02 Could not find quote %d.", id)
	} else {
		w.Noticef(nick, "\x02Quote:\x02 Quote %d updated.", id)
	}
	return nil
}

// Quote returns a random quote
func (q *Quoter) Quote(w irc.Writer, ev *cmd.Event) error {
	strid := ev.Args["id"]
	nick := ev.Nick()

	var quote quotes.Quote
	var id int
	var err error
	if len(strid) > 0 {
		getid, err := strconv.Atoi(strid)
		id = int(getid)
		if err != nil {
			w.Notice(nick, "\x02Quote:\x02 Not a valid id.")
			return nil
		}
		quote, err = q.db.GetQuote(id)
	} else {
		quote, err = q.db.RandomQuote()
	}
	if err != nil {
		if err == sql.ErrNoRows {
			w.Notice(nick, "\x02Quote:\x02 No quotes to display.")
			return nil
		}
		w.Noticef(nick, "\x02Quote:\x02 %v", err)
		return nil
	}

	if len(quote.Quote) == 0 {
		w.Notify(ev.Event, nick, "\x02Quote:\x02 Does not exist.")
	} else {
		w.Notifyf(ev.Event, nick, "\x02Quote (\x02#%d|%+d\x02):\x02 %s",
			quote.ID, quote.Upvotes-quote.Downvotes, quote.Quote)
	}
	return nil
}

// Quotes gets the number of quotes
func (q *Quoter) Quotes(w irc.Writer, ev *cmd.Event) error {
	nick := ev.Nick()

	w.Notifyf(ev.Event, nick, "\x02Quote:\x02 %d quote(s) in database.",
		q.db.NQuotes())
	return nil
}

// Details provides more detail on a given quote
func (q *Quoter) Details(w irc.Writer, ev *cmd.Event) error {
	nick := ev.Nick()
	id, err := strconv.Atoi(ev.Args["id"])

	if err != nil {
		w.Notice(nick, "\x02Quote:\x02 Not a valid id.")
		return nil
	}

	if quote, err := q.db.GetQuote(int(id)); err != nil {
		w.Noticef(nick, "\x02Quote:\x02 error %v", err)
	} else {
		w.Notifyf(ev.Event, nick,
			"\x02Quote (\x02#%d\x02):\x02 Created on %s by %s, %d upvote(s), %d downvote(s)",
			id,
			quote.Date.UTC().Format(dateFormat),
			quote.Author,
			quote.Upvotes,
			quote.Downvotes,
		)
	}

	return nil
}

// Quoteweb provides a server to see the quotes
func (q *Quoter) Quoteweb(w irc.Writer, ev *cmd.Event) error {
	if len(q.ServerURI) == 0 {
		w.Notify(ev.Event, ev.Nick(), "\x02Quote:\x02 No quote server running")
		return nil
	}
	w.Notify(ev.Event, ev.Nick(), "\x02Quote:\x02 "+q.ServerURI)
	return nil
}

// Up vote a quote
func (q *Quoter) Up(w irc.Writer, ev *cmd.Event) error {
	nick := ev.Nick()
	id, err := strconv.Atoi(ev.Args["id"])

	if err != nil {
		w.Notice(nick, "\x02Quote:\x02 Not a valid id.")
		return nil
	}

	voter := strings.ToLower(nick)
	did, err := q.db.Upvote(id, voter)
	if err != nil {
		w.Noticef(nick, "\x02Quote:\x02 Error attempting to upvote: %v", err)
		return nil
	} else if !did {
		w.Noticef(nick, "\x02Quote:\x02 You have already upvoted quote #%d", id)
	} else {
		w.Noticef(nick, "\x02Quote:\x02 Upvoted quote #%d", id)
		return nil
	}

	return nil
}

// Down vote a quote
func (q *Quoter) Down(w irc.Writer, ev *cmd.Event) error {
	nick := ev.Nick()
	id, err := strconv.Atoi(ev.Args["id"])

	if err != nil {
		w.Notice(nick, "\x02Quote:\x02 Not a valid id.")
		return nil
	}

	voter := strings.ToLower(nick)
	did, err := q.db.Downvote(id, voter)
	if err != nil {
		w.Noticef(nick, "\x02Quote:\x02 Error attempting to upvote: %v", err)
		return nil
	} else if !did {
		w.Noticef(nick, "\x02Quote:\x02 You have already downvoted quote #%d", id)
	} else {
		w.Noticef(nick, "\x02Quote:\x02 Downvoted quote #%d", id)
		return nil
	}

	return nil
}

// Unvote vote a quote
func (q *Quoter) Unvote(w irc.Writer, ev *cmd.Event) error {
	nick := ev.Nick()
	id, err := strconv.Atoi(ev.Args["id"])

	if err != nil {
		w.Notice(nick, "\x02Quote:\x02 Not a valid id.")
		return nil
	}

	voter := strings.ToLower(nick)
	did, err := q.db.Unvote(id, voter)
	if err != nil {
		w.Noticef(nick, "\x02Quote:\x02 Error attempting to upvote: %v", err)
		return nil
	} else if !did {
		w.Noticef(nick, "\x02Quote:\x02 You have not voted on quote #%d", id)
	} else {
		w.Noticef(nick, "\x02Quote:\x02 Unvoted quote #%d", id)
		return nil
	}

	return nil
}
