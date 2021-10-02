//MIT License
//
//Copyright (c) 2021 CODEBASE
//
//Permission is hereby granted, free of charge, to any person obtaining a copy
//of this software and associated documentation files (the "Software"), to deal
//in the Software without restriction, including without limitation the rights
//to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
//copies of the Software, and to permit persons to whom the Software is
//furnished to do so, subject to the following conditions:
//
//The above copyright notice and this permission notice shall be included in all
//copies or substantial portions of the Software.
//
//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
//IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
//AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
//LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
//OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
//SOFTWARE.

package main

import (
	"fmt"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	subscribeCommand = kingpin.Command("subscribe", "subscribe a topic.")
	URLFlag = kingpin.Flag("url", "A WAMP URL to connect to, like ws://127.0.0.1:8080/ws or rs://localhost:1234").Required().Short('u').String()
	realmFlag = kingpin.Flag("realm", "The realm to join").Required().Short('r').String()
	privateKeyFlag = kingpin.Flag("private-key","Hex-encoded private key").String()
	authidFlag = kingpin.Flag("authid","The authid to use, if authenticating").String()
	authRoleFlag = kingpin.Flag("authrole", "The role to use, if authenticating").String()

	topicArgSub  = subscribeCommand.Arg("topic", "topic name").Required().String()

	publishCommand = kingpin.Command("publish", "publishing a topic.")
	topicArgPub  = publishCommand.Arg("topic", "topic name").Required().String()
	argumentsArgPub = publishCommand.Arg("args","give the arguments").Strings()
	kwArgsFlagPub   = publishCommand.Flag("kwarg", "give the keyword arguments").Short('k').StringMap()

	registerCommand = kingpin.Command("register", "registering a procedure.")
	topicArgReg  = registerCommand.Arg("procedure", "procedure name").Required().String()
	bashFlagReg = registerCommand.Flag("bash", "enter bash script").Short('b').Strings()
	shellFlagReg = registerCommand.Flag("shell","enter the shell script").Short('s').Strings()
	pythonFlagReg = registerCommand.Flag("python","enter the python script").Short('p').Strings()
	execFlagReg = registerCommand.Flag("exec","execute any file").Short('e').String()


	callCommand = kingpin.Command("call", "calling a procedure.")
	topicArgCal  = callCommand.Arg("procedure", "procedure name").Required().String()
	argumentsArgCal = callCommand.Arg("args","give the arguments").Strings()
	kwArgsFlagCal   = callCommand.Flag("kwarg", "give the keyword arguments").Short('k').StringMap()
)

func main() {
	switch kingpin.Parse() {
		case "subscribe":
			subscribe(*URLFlag, *realmFlag, *topicArgSub)
		case "publish":
			publish(*URLFlag, *realmFlag, *topicArgPub, *argumentsArgPub, *kwArgsFlagPub)
		case "register":
			if *bashFlagReg != nil && *shellFlagReg == nil && *pythonFlagReg == nil && *execFlagReg == "" {
				register(*URLFlag, *realmFlag, *topicArgReg, *bashFlagReg,"bash")
			} else if *shellFlagReg != nil && *bashFlagReg == nil && *pythonFlagReg ==nil && *execFlagReg == "" {
				register(*URLFlag, *realmFlag, *topicArgReg, *shellFlagReg,"sh")
			} else if *pythonFlagReg != nil && *bashFlagReg == nil && *shellFlagReg == nil && *execFlagReg == "" {
				register(*URLFlag, *realmFlag, *topicArgReg, *pythonFlagReg,"python3")
			} else if execFlagReg != nil && *bashFlagReg == nil && *shellFlagReg == nil && *pythonFlagReg ==nil {
				register(*URLFlag, *realmFlag, *topicArgReg, nil, *execFlagReg)
			} else if *bashFlagReg == nil && *shellFlagReg == nil && *pythonFlagReg == nil && *execFlagReg == "" {
				register(*URLFlag, *realmFlag, *topicArgReg, nil,"")
			}else {
				fmt.Println("Please use one type for running script")
			}
		case "call":
			call(*URLFlag, *realmFlag, *topicArgCal, *argumentsArgCal, *kwArgsFlagCal)
	}
}