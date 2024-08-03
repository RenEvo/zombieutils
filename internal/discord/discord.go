package discord

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/bwmarrin/discordgo"
)

type Discord struct {
	serverID  string
	channelID string
	token     string
	session   *discordgo.Session
}

func New() (*Discord, error) {
	d := &Discord{
		token:     os.Getenv("DISCORD_SERVER_TOKEN"),
		serverID:  os.Getenv("DISCORD_SERVER_ID"),
		channelID: os.Getenv("DISCORD_SERVER_CHANNEL_ID"),
	}

	if d.token == "" || d.channelID == "" {
		return d, nil
	}

	session, err := discordgo.New("Bot " + d.token)
	if err != nil {
		return nil, fmt.Errorf("failed to create discord session: %w", err)
	}

	session.Identify.Intents = discordgo.IntentGuildMessages

	session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		slog.Info("Discord logged in as: %s#%s", s.State.User.Username, s.State.User.Discriminator)
	})
	session.AddHandler(d.handleMessage)

	if err := session.Open(); err != nil {
		return nil, fmt.Errorf("failed to connect to discord: %w", err)
	}

	d.session = session

	slog.Info("Connected to discord")

	return d, nil
}

func (d *Discord) Publish(msgs <-chan string) {
	// loop until channel is closed - but make sure we consume them all
	for msg := range msgs {
		d.Message(msg)
	}
}

func (d *Discord) handleMessage(s *discordgo.Session, msg *discordgo.MessageCreate) {
	// only specified channel
	if msg.ChannelID != d.channelID {
		return
	}

	// don't get into an echo loop....
	if msg.Author.ID == s.State.User.ID {
		return
	}

	// render the contents away from @<id> things to @User things
	content := msg.ContentWithMentionsReplaced()

	// some things can be just images, etc... lets not spam with nothingness
	if len(content) == 0 {
		return
	}

	slog.Info("Discord chat", "author", msg.Author.Username, "content", content)
}

func (d *Discord) Message(msg string, args ...any) {
	if d.session == nil {
		return
	}

	_, _ = d.session.ChannelMessageSend(d.channelID, fmt.Sprintf(msg, args...))
}

func (d *Discord) Close() error {
	if d.session == nil {
		return nil
	}

	return d.session.Close()
}
