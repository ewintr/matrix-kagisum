package bot

import (
	"regexp"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/exp/slog"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/crypto/cryptohelper"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/format"
	"maunium.net/go/mautrix/id"
)

type MatrixConfig struct {
	Homeserver    string
	UserID        string
	UserAccessKey string
	UserPassword  string
	RoomID        string
	DBPath        string
	Pickle        string
	AcceptInvites bool
}

type Bot struct {
	config       MatrixConfig
	client       *mautrix.Client
	cryptoHelper *cryptohelper.CryptoHelper
	kagiCient    *Kagi
	logger       *slog.Logger
	messages     map[string]string
}

func NewBot(cfg MatrixConfig, kagi *Kagi, logger *slog.Logger) *Bot {
	return &Bot{
		config:    cfg,
		kagiCient: kagi,
		logger:    logger,
		messages:  make(map[string]string),
	}
}

func (m *Bot) Init() error {
	client, err := mautrix.NewClient(m.config.Homeserver, id.UserID(m.config.UserID), m.config.UserAccessKey)
	if err != nil {
		return err
	}
	var oei mautrix.OldEventIgnorer
	oei.Register(client.Syncer.(mautrix.ExtensibleSyncer))
	m.client = client
	m.cryptoHelper, err = cryptohelper.NewCryptoHelper(client, []byte(m.config.Pickle), m.config.DBPath)
	if err != nil {
		return err
	}
	m.cryptoHelper.LoginAs = &mautrix.ReqLogin{
		Type:       mautrix.AuthTypePassword,
		Identifier: mautrix.UserIdentifier{Type: mautrix.IdentifierTypeUser, User: m.config.UserID},
		Password:   m.config.UserPassword,
	}
	if err := m.cryptoHelper.Init(); err != nil {
		return err
	}
	m.client.Crypto = m.cryptoHelper

	if m.config.AcceptInvites {
		m.AddEventHandler(m.InviteHandler())
	}
	m.AddEventHandler(m.ReactionHandler())
	m.AddEventHandler(m.MessageHandler())

	return nil
}

func (m *Bot) Run() error {
	if err := m.client.Sync(); err != nil {
		return err
	}

	return nil
}

func (m *Bot) Close() error {
	if err := m.client.Sync(); err != nil {
		return err
	}
	if err := m.cryptoHelper.Close(); err != nil {
		return err
	}

	return nil
}

func (m *Bot) AddEventHandler(eventType event.Type, handler mautrix.EventHandler) {
	syncer := m.client.Syncer.(*mautrix.DefaultSyncer)
	syncer.OnEventType(eventType, handler)
}

func (m *Bot) InviteHandler() (event.Type, mautrix.EventHandler) {
	return event.StateMember, func(source mautrix.EventSource, evt *event.Event) {
		if evt.GetStateKey() == m.client.UserID.String() && evt.Content.AsMember().Membership == event.MembershipInvite && evt.RoomID.String() == m.config.RoomID {
			_, err := m.client.JoinRoomByID(evt.RoomID)
			if err != nil {
				m.logger.Error("failed to join room after invite", slog.String("err", err.Error()), slog.String("room_id", evt.RoomID.String()), slog.String("inviter", evt.Sender.String()))
				return
			}

			m.logger.Info("joined room after invite", slog.String("room_id", evt.RoomID.String()), slog.String("inviter", evt.Sender.String()))
		}
	}
}

func (m *Bot) MessageHandler() (event.Type, mautrix.EventHandler) {
	return event.EventMessage, func(source mautrix.EventSource, evt *event.Event) {
		content := evt.Content.AsMessage()
		eventID := evt.ID
		m.logger.Info("received message", slog.String("content", content.Body))
		// ignore if the message is sent by the bot itself
		if evt.Sender == id.UserID(m.config.UserID) {
			m.logger.Info("message sent by bot itself, ignoring", slog.String("event_id", eventID.String()))
			return
		}
		url, ok := findFirstURL(content.Body)
		if !ok {
			m.logger.Info("message does not contain url, ignoring", slog.String("event_id", eventID.String()))
			return
		}
		m.messages[eventID.String()] = url
		m.logger.Info("saved url", slog.String("event_id", eventID.String()), slog.String("url", url))
	}
}

func (m *Bot) ReactionHandler() (event.Type, mautrix.EventHandler) {
	return event.EventReaction, func(source mautrix.EventSource, evt *event.Event) {
		eventID := evt.ID
		// ignore if the message is sent by the bot itself
		if evt.Sender == id.UserID(m.config.UserID) {
			m.logger.Info("message sent by bot itself, ignoring", slog.String("event_id", eventID.String()))
			return
		}
		rel := evt.Content.AsReaction().GetRelatesTo()
		relID := rel.GetAnnotationID()
		relKey := rel.GetAnnotationKey()
		m.logger.Info("received reaction", slog.String("event_id", eventID.String()), slog.String("rel_id", relID.String()), slog.String("rel_key", relKey))
		if relKey != `🗒️` && relKey != `🗒` { // different clients have different emoji
			m.logger.Info("reaction is not 🗒️ or 🗒, ignoring", slog.String("event_id", eventID.String()))
			return
		}

		wantedURL, ok := m.messages[relID.String()]
		var summary string
		if !ok {
			summary = "could not find referenced message in ternal storage, or referenced message does not contain url"
			m.logger.Info("referenced message is not known or does not contain url, ignoring", slog.String("event_id", eventID.String()))
		} else {
			var err error
			summary, err = m.kagiCient.Summarize(wantedURL)
			if err != nil {
				m.logger.Error("failed to summarize", slog.String("err", err.Error()))
				return
			}
		}

		reply := summary
		formattedReply := format.RenderMarkdown(reply, true, false)
		formattedReply.RelatesTo = &event.RelatesTo{
			InReplyTo: &event.InReplyTo{
				EventID: relID,
			},
		}
		if _, err := m.client.SendMessageEvent(evt.RoomID, event.EventMessage, &formattedReply); err != nil {
			m.logger.Error("failed to send message", slog.String("err", err.Error()))
			return
		}

		if len(reply) > 30 {
			reply = reply[:30] + "..."
		}
		m.logger.Info("sent reply", slog.String("parent_id", eventID.String()), slog.String("msg", reply))
	}
}

func findFirstURL(s string) (string, bool) {
	re := regexp.MustCompile(`https?\:\/\/[^ \)]*`)
	match := re.FindString(s)
	if match == "" {
		return "", false
	}
	return match, true
}
