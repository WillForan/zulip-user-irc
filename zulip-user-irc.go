/*
A go port of https://github.com/Stantheman/zulip-irc-ii/

	subscribe to zulip events, push them to irc, listen for responses

thoughts:

		default to replyng to last message. but check time of last message. dont send if within 1s(?)
	   populate nicks with DM options, and stream names?
		+n:emoji: for reactions
*/
package main

import (
	"flag"
	//"fmt"
	gzb "github.com/ifo/gozulipbot"
	"github.com/pelletier/go-toml"
	"log"
	"time"
)

// gzb has it's own GetConfigFromFlags() which would have been used
// had it been seen earlier. But I like config.toml better away.
func bot_config(config *toml.Tree) gzb.Bot {
	b := gzb.Bot{}
	b.APIKey = config.Get("zulip.key").(string)
	b.APIURL = config.Get("zulip.site").(string)
	b.Email = config.Get("zulip.email").(string)
	b.Backoff = 1 * time.Second
	b.Retries = 0
	return b
}

func connect_zulip(file string) {
	config, err := toml.LoadFile(file)
	if err != nil {
		panic(err)
	}
	//user := config.Get("zulip.key").(string)
	//print(user, "\n")
	bot := bot_config(config)
	//alt: bot.GetConfigFromEnvironment()

	bot.Init()
	q, err := bot.RegisterAll()
	if err != nil {
		log.Println("register error:", err)
	}
   // TODO: collect and return function handle(s)
	// to execute recieveMessages not yet processed?
	q.EventsCallback(recieveMessage)
}

func recieveMessage(em gzb.EventMessage, err error) {
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
	log.Println(em.SenderEmail,"/",em.Type,": ",em.Content)

	//em.Queue.Bot.Respond(em, "hi forever!")
}

func main() {
	config_file := flag.String("config", "config.toml", "[zulip] and [irc] configuration")
	flag.Parse()
	connect_zulip(*config_file)

	//TODO: sleep forever?
	time.Sleep(1 * time.Minute)
}
