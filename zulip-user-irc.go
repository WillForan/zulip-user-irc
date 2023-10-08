/*
A go port of https://github.com/Stantheman/zulip-irc-ii/
with aspirations to be like https://github.com/42wim/matterircd for zulip
but could be a bitlbee or weechat plugin

	subscribe to zulip events, push them to irc, listen for responses

thoughts:

		default to replyng to last message. but check time of last message. dont send if within 1s(?)
	   populate nicks with DM options, and stream names?
		+n:emoji: for reactions
*/
package main

import (
	"flag"
	"fmt"
	gzb "github.com/ifo/gozulipbot"
	hbot "github.com/whyrusleeping/hellabot"
	"regexp"
	//log2 "gopkg.in/inconshreveable/log15.v2"
	"github.com/pelletier/go-toml"
	"log"
	"time"
)

// gzb has it's own GetConfigFromFlags() which would have been used
// had it been seen earlier. But I like config.toml better away.
func zulip_config(config *toml.Tree) gzb.Bot {
	b := gzb.Bot{}
	b.APIKey = config.Get("zulip.key").(string)
	b.APIURL = config.Get("zulip.site").(string)
	b.Email = config.Get("zulip.email").(string)
	b.Backoff = 1 * time.Second
	b.Retries = 0

	b.Init()
	return b
}

// zulip bot wants a call back function but we wont have an irc bot or user until later
// so we'll generate a function that includes those in the scope
func zulip_to_irc(bot *hbot.Bot, user string) func(gzb.EventMessage, error) {
	var msgfunc = func(message string) { bot.Msg(user, message) }
	return func(em gzb.EventMessage, err error) {
		zulip_recieve_message(msgfunc, em, err)
	}
}

// callback for zulip events. interfaces with irc bot via function that sends message
func zulip_recieve_message(irc_msg func(string), em gzb.EventMessage, err error) {
	if err != nil {
		log.Println("error in respond to message:", err)
		return
	}

	//does not get reactions
	//em.Content = "message"
	//em.Timestamp = 1696694596
	//em.Client = "website"
	//em.SenderEmail = "foranw@upmc.edu"
	//em.Subject = ""
	//em.SenderID    = 642506
	//em.Type        = "Private"
	//fmt.Sprintf("a")
	log.Println(em.SenderEmail, "/", em.Type, ": ", em.Content)
	irc_msg(fmt.Sprintf("%s/%s: %s", em.SenderEmail, em.Type, em.Content))

	//em.Queue.Bot.Respond(em, "hi forever!")
}

//TODO: what does stream/topic look like? reactions?
//TODO: capture mutliple emails for group DM
var irc_to_z_patt = regexp.MustCompile(`^(?P<Email>[^@ ]+@[^/:]+)/?(?P<Topic>[^:]+})?:(?P<Msg>.*)`)

func main() {
	config_file := flag.String("config", "config.toml", "[zulip] and [irc] configuration")
	flag.Parse()
	config, err := toml.LoadFile(*config_file)
	if err != nil {
		log.Println("failed to load configuration file", *config_file)
		panic(err)
	}
	//user := config.Get("zulip.key").(string)

	/** irc **/
	serv := config.Get("irc.host").(string)
	nick := config.Get("irc.bot").(string)
	irc_to := config.Get("irc.user").(string)
	log.Println("connecting to", serv, "as", nick)
	irc_bot, err := hbot.NewBot(serv, nick)
	//irc_bot.Logger.SetHandler(log2.StdoutHandler)

	if err != nil {
		log.Println("irc start error:", err)
		panic(err)
	}

	/** zulip **/
	zulip_bot := zulip_config(config) //alt: bot.GetConfigFromEnvironment()
	q, err := zulip_bot.RegisterAll()
	if err != nil {
		log.Println("zulip register error:", err)
		panic(err)
	}

	/** RUN **/
	// TODO: collect and return function handle(s)
	// to execute recieveMessages not yet processed?
	q.EventsCallback(zulip_to_irc(irc_bot, irc_to))

	irc_user := config.Get("irc.user").(string)
	var irc_recieved_message = hbot.Trigger{
		Condition: func(b *hbot.Bot, m *hbot.Message) bool {
			return m.From == irc_user
		},
		Action: func(b *hbot.Bot, m *hbot.Message) bool {
			matches := irc_to_z_patt.FindStringSubmatch(m.Content)
			log.Println("message from irc:", m, "\n\tmatchs:", matches)
			if len(matches) > 1 {
				email := matches[irc_to_z_patt.SubexpIndex("Email")]
				msg_text := matches[irc_to_z_patt.SubexpIndex("Msg")]
				log.Println("to zulip email:", email, "\n\tmsg:", msg_text)
				msg := gzb.Message{
					Stream:  "",
					Topic:   "",
					Emails:  []string{email},
					Content: msg_text}
				zulip_bot.Message(msg)
			} else {
				b.Reply(m, fmt.Sprintf("don't know what to do with '%s'", m.Content))
			}
			return false
		},
	}
	irc_bot.AddTrigger(irc_recieved_message)

	log.Println("running bots, irc watching for", irc_user)
	irc_bot.Run() // Blocks until exit
	log.Println("done")

	// todo run forever
	//time.Sleep(5 * time.Minute)

}
