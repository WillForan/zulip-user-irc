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
	"fmt"
	gzb "github.com/ifo/gozulipbot"
	"github.com/pelletier/go-toml"
	"log"
	"math"
	"sync/atomic"
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
	return b
}

func connect(file string) {
	config, err := toml.LoadFile(file)
	if err != nil {
		panic(err)
	}
	user := config.Get("zulip.key").(string)
	print(user, "\n")
	bot := bot_config(config)

	//bot.GetConfigFromEnvironment()

	fmt.Println(bot)
	bot.Init()
	fmt.Println(bot)

	backoffTime := time.Now().Add(bot.Backoff * time.Duration(math.Pow10(int(atomic.LoadInt64(&bot.Retries)))))
	fmt.Println("backoff time:", backoffTime)

	bot.RegisterAt()
	q, err := bot.RegisterAt()
	if err != nil {
		log.Println("register error:", err)
	}

	stopFunc := q.EventsCallback(respondToMessage)

	time.Sleep(1 * time.Minute)
	stopFunc()

}

func respondToMessage(em gzb.EventMessage, err error) {
	if err != nil {
		log.Println("error in respond to message:", err)
		return
	}

	log.Println("message received")
	//log.Println(em)

	//em.Queue.Bot.Respond(em, "hi forever!")
}

func main() {
	config_file := flag.String("config", "config.toml", "[zulip] and [irc] configuration")
	flag.Parse()
	fmt.Println(*config_file)
	connect(*config_file)
}
