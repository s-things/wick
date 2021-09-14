package main

import (
	"gopkg.in/alecthomas/kingpin.v2"
)
var (
	subscribeCommand = kingpin.Command("subscribe", "subscribe a topic.")
	URLArgSub   = subscribeCommand.Arg("url", "url").Required().String()
	realmArgSub = subscribeCommand.Arg("realm", "realmSub").Required().String()
	topicArgSub = subscribeCommand.Arg("topic", "topic name").Required().String()

	publishCommand  = kingpin.Command("publish", "publishing a topic.")
	urlArgPub   = publishCommand.Arg("url", "url").Required().String()
	realmArgPub = publishCommand.Arg("realm", "realmSub").Required().String()
	topicArgPub = publishCommand.Arg("topic", "topic name").Required().String()
	argumentsFlagPub = publishCommand.Flag("args","give the arguments").Short('a').Strings()
	kwargsFlagPub = publishCommand.Flag("kwargs", "give the keyword arguments").Short('k').StringMap()

	registerCommand  = kingpin.Command("register", "registering a procedure.")
	urlArgReg   = registerCommand.Arg("url", "url").Required().String()
	realmArgReg = registerCommand.Arg("realm", "realmSub").Required().String()
	topicArgReg = registerCommand.Arg("procedure", "procedure name").Required().String()

	callCommand  = kingpin.Command("call", "calling a procedure.")
	urlArgCal   = callCommand.Arg("url", "url").Required().String()
	realmArgCal = callCommand.Arg("realm", "realmSub").Required().String()
	topicArgCal = callCommand.Arg("procedure", "procedure name").Required().String()
	argumentsFlagCal = callCommand.Flag("args","give the arguments").Short('a').Strings()
	kwargsFlagCal = callCommand.Flag("kwargs", "give the keyword arguments").Short('k').StringMap()
)

func main() {
	switch kingpin.Parse() {
		case "subscribe":
			subscribe(*URLArgSub, *realmArgSub, *topicArgSub)
		case "publish":
			publish(*urlArgPub, *realmArgPub, *topicArgPub, *argumentsFlagPub, *kwargsFlagPub)
		case "register":
			register(*urlArgReg, *realmArgReg, *topicArgReg)
		case "call":
			call(*urlArgCal, *realmArgCal, *topicArgCal, *argumentsFlagCal, *kwargsFlagCal)
	}
}