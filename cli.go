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
	topicArgPub     = publishCommand.Arg("topic", "topic name").Required().String()
	argumentsArgPub = publishCommand.Arg("args","give the arguments").Strings()
	kwargsFlagPub   = publishCommand.Flag("kwargs", "give the keyword arguments").Short('k').StringMap()

	registerCommand  = kingpin.Command("register", "registering a procedure.")
	urlArgReg   = registerCommand.Arg("url", "url").Required().String()
	realmArgReg = registerCommand.Arg("realm", "realmSub").Required().String()
	topicArgReg = registerCommand.Arg("procedure", "procedure name").Required().String()

	callCommand  = kingpin.Command("call", "calling a procedure.")
	urlArgCal   = callCommand.Arg("url", "url").Required().String()
	realmArgCal = callCommand.Arg("realm", "realmSub").Required().String()
	topicArgCal     = callCommand.Arg("procedure", "procedure name").Required().String()
	argumentsArgCal = callCommand.Arg("args","give the arguments").Strings()
	kwargsFlagCal   = callCommand.Flag("kwargs", "give the keyword arguments").Short('k').StringMap()
)

func main() {
	switch kingpin.Parse() {
		case "subscribe":
			subscribe(*URLArgSub, *realmArgSub, *topicArgSub)
		case "publish":
			publish(*urlArgPub, *realmArgPub, *topicArgPub, *argumentsArgPub, *kwargsFlagPub)
		case "register":
			register(*urlArgReg, *realmArgReg, *topicArgReg)
		case "call":
			call(*urlArgCal, *realmArgCal, *topicArgCal, *argumentsArgCal, *kwargsFlagCal)
	}
}