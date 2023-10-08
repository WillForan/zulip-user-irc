.PHONY: run
run: zulip-user-irc
	./$^

debug: zulip-user-irc
	dlv exec ./zulip-user-irc

zulip-user-irc: zulip-user-irc.go
	go build
