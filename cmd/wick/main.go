// MIT License
//
// Copyright (c) 2021 CODEBASE
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package main

import (
	"github.com/gammazero/nexus/v3/client"
	"github.com/gammazero/nexus/v3/transport/serialize"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/codebasepk/wick/wamp"
)

var (
	url = kingpin.Flag("url", "WAMP URL to connect to").
		Default("ws://localhost:8080/ws").Envar("WICK_URL").String()
	realm      = kingpin.Flag("realm", "The WAMP realm to join").Default("realm1").
		Envar("WICK_REALM").String()
	authMethod = kingpin.Flag("authmethod", "The authentication method to use").Envar("WICK_AUTHMETHOD").
			Default("anonymous").Enum("anonymous", "ticket", "wampcra", "cryptosign")
	authid   = kingpin.Flag("authid", "The authid to use, if authenticating").Envar("WICK_AUTHID").
		String()
	authrole = kingpin.Flag("authrole", "The authrole to use, if authenticating").
		Envar("WICK_AUTHROLE").String()
	secret   = kingpin.Flag("secret", "The secret to use in Challenge-Response Auth.").
			Envar("WICK_SECRET").String()
	privateKey = kingpin.Flag("private-key", "The ed25519 private key hex for cryptosign").
			Envar("WICK_PRIVATE_KEY").String()
	ticket     = kingpin.Flag("ticket", "The ticket when using ticket authentication").
		Envar("WICK_TICKET").String()
	serializer = kingpin.Flag("serializer", "The serializer to use").Envar("WICK_SERIALIZER").
		Default("json").Enum("json", "msgpack", "cbor")

	subscribe      = kingpin.Command("subscribe", "subscribe a topic.")
	subscribeTopic = subscribe.Arg("topic", "Topic to subscribe to").Required().String()

	publish            = kingpin.Command("publish", "Publish to a topic.")
	publishTopic       = publish.Arg("topic", "topic name").Required().String()
	publishArgs        = publish.Arg("args", "give the arguments").Strings()
	publishKeywordArgs = publish.Flag("kwarg", "give the keyword arguments").Short('k').StringMap()

	register          = kingpin.Command("register", "Register a procedure.")
	registerProcedure = register.Arg("procedure", "procedure name").Required().String()
	onInvocationCmd   = register.Arg("command", "Shell command to run and return it's output").String()

	call            = kingpin.Command("call", "Call a procedure.")
	callProcedure   = call.Arg("procedure", "Procedure to call").Required().String()
	callArgs        = call.Arg("args", "give the arguments").Strings()
	callKeywordArgs = call.Flag("kwarg", "give the keyword arguments").Short('k').StringMap()
)

const versionString = "0.3.0"

func main() {
	kingpin.Version(versionString).VersionFlag.Short('v')
	cmd := kingpin.Parse()

	serializerToUse := serialize.JSON

	switch *serializer {
	case "json":
	case "msgpack":
		serializerToUse = serialize.MSGPACK
	case "cbor":
		serializerToUse = serialize.CBOR
	}

	logger := log.New()

	if *privateKey != "" && *ticket != "" {
		logger.Fatal("Provide only one of private key, ticket or secret")
	} else if *ticket != "" && *secret != "" {
		logger.Fatal("Provide only one of private key, ticket or secret")
	} else if *privateKey != "" && *secret != "" {
		logger.Fatal("Provide only one of private key, ticket or secret")
	}

	if *privateKey != "" {
		*authMethod = "cryptosign"
	} else if *ticket != "" {
		*authMethod = "ticket"
	} else if *secret != "" {
		*authMethod = "wampcra"
	}

	var session *client.Client

	switch *authMethod {
	case "anonymous":
		if *privateKey != "" {
			logger.Fatal("Private key not needed for anonymous auth")
		}
		if *ticket != "" {
			logger.Fatal("ticket not needed for anonymous auth")
		}
		if *secret != "" {
			logger.Fatal("secret not needed for anonymous auth")
		}
		session = wamp.ConnectAnonymous(*url, *realm, serializerToUse, *authid, *authrole, logger)
	case "ticket":
		if *ticket == "" {
			logger.Fatal("Must provide ticket when authMethod is ticket")
		}
		session = wamp.ConnectTicket(*url, *realm, serializerToUse, *authid, *authrole, *ticket, logger)
	case "wampcra":
		if *secret == "" {
			logger.Fatal("Must provide secret when authMethod is wampcra")
		}
		session = wamp.ConnectCRA(*url, *realm, serializerToUse, *authid, *authrole, *secret, logger)
	case "cryptosign":
		if *privateKey == "" {
			logger.Fatal("Must provide private key when authMethod is cryptosign")
		}
		session = wamp.ConnectCryptoSign(*url, *realm, serializerToUse, *authid, *authrole, *privateKey, logger)
	}

	defer session.Close()

	switch cmd {
	case subscribe.FullCommand():
		wamp.Subscribe(session, logger, *subscribeTopic)
	case publish.FullCommand():
		wamp.Publish(session, logger, *publishTopic, *publishArgs, *publishKeywordArgs)
	case register.FullCommand():
		wamp.Register(session, logger, *registerProcedure, *onInvocationCmd)
	case call.FullCommand():
		wamp.Call(session, logger, *callProcedure, *callArgs, *callKeywordArgs)
	}
}
