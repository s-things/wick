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

package wamp

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/pbkdf2"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"

	"github.com/gammazero/nexus/v3/client"
	"github.com/gammazero/nexus/v3/transport/serialize"
	"github.com/gammazero/nexus/v3/wamp"
	"github.com/gammazero/nexus/v3/wamp/crsign"
)

func connect(url string, cfg client.Config, logger *log.Logger) *client.Client {
	//baseUrl := url
	if strings.HasPrefix(url, "rs") {
		url = "tcp" + strings.TrimPrefix(url, "rs")
	} else if strings.HasPrefix(url, "rss") {
		url = "tcp" + strings.TrimPrefix(url, "rss")
	}
	session, err := client.ConnectNet(context.Background(), url, cfg)
	if err != nil {
		logger.Fatal(err)
	} else {
		// FIXME: use a better logger and only print such messages in debug mode.
		//logger.Println("Connected to ", baseUrl)
	}

	return session
}

func ConnectAnonymous(url string, realm string, serializer serialize.Serialization, authid string, authrole string,
	logger *log.Logger) *client.Client {

	helloDict := wamp.Dict{}
	if authid != "" {
		helloDict["authid"] = authid
	}

	if authrole != "" {
		helloDict["authrole"] = authrole
	}

	cfg := client.Config{
		Realm:         realm,
		Logger:        logger,
		HelloDetails:  helloDict,
		Serialization: serializer,
	}

	return connect(url, cfg, logger)
}

func ConnectTicket(url string, realm string, serializer serialize.Serialization, authid string, authrole string,
	ticket string, logger *log.Logger) *client.Client {

	helloDict := wamp.Dict{}
	if authid != "" {
		helloDict["authid"] = authid
	}

	if authrole != "" {
		helloDict["authrole"] = authrole
	}

	cfg := client.Config{
		Realm:        realm,
		Logger:       logger,
		HelloDetails: helloDict,
		AuthHandlers: map[string]client.AuthFunc{
			"ticket": func(c *wamp.Challenge) (string, wamp.Dict) {
				return ticket, wamp.Dict{}
			},
		},
		Serialization: serializer,
	}

	return connect(url, cfg, logger)
}

func ConnectCRA(url string, realm string, serializer serialize.Serialization, authid string, authrole string,
	secret string, logger *log.Logger) *client.Client {

	helloDict := wamp.Dict{}
	if authid != "" {
		helloDict["authid"] = authid
	}

	if authrole != "" {
		helloDict["authrole"] = authrole
	}

	cfg := client.Config{
		Realm:        realm,
		Logger:       logger,
		HelloDetails: helloDict,
		AuthHandlers: map[string]client.AuthFunc{
			"wampcra": func(c *wamp.Challenge) (string, wamp.Dict) {
				ch, _ := wamp.AsString(c.Extra["challenge"])
				// If the client needed to lookup a user's key, this would require decoding
				// the JSON-encoded challenge string and getting the authid.  For this
				// example assume that client only operates as one user and knows the key
				// to use.
				saltStr, _ := wamp.AsString(c.Extra["salt"])
				// If no salt given, use raw password as key.
				if saltStr == "" {
					return crsign.SignChallenge(ch, []byte(secret)), wamp.Dict{}
				}

				// If salting info give, then compute a derived key using PBKDF2.
				salt := []byte(saltStr)
				iters, _ := wamp.AsInt64(c.Extra["iterations"])
				keylen, _ := wamp.AsInt64(c.Extra["keylen"])

				if iters == 0 {
					iters = 1000
				}
				if keylen == 0 {
					keylen = 32
				}

				// Compute derived key.
				dk := pbkdf2.Key([]byte(secret), salt, int(iters), int(keylen), sha256.New)
				// Get base64 bytes. see https://github.com/gammazero/nexus/issues/252
				derivedKey := []byte(base64.StdEncoding.EncodeToString(dk))

				return crsign.SignChallenge(ch, derivedKey), wamp.Dict{}
			},
		},
		Serialization: serializer,
	}

	return connect(url, cfg, logger)
}

func ConnectCryptoSign(url string, realm string, serializer serialize.Serialization, authid string, authrole string,
	privateKey string, publicKey string, logger *log.Logger) *client.Client {

	helloDict := wamp.Dict{}
	if authid != "" {
		helloDict["authid"] = authid
	}

	if authrole != "" {
		helloDict["authrole"] = authrole
	}

	privkey, _ := hex.DecodeString(privateKey)
	pvk := ed25519.NewKeyFromSeed(privkey)
	key := pvk.Public().(ed25519.PublicKey)
	pubKeyExtracted := hex.EncodeToString(key)

	if publicKey == "" {
		publicKey = pubKeyExtracted
	} else {
		if publicKey != pubKeyExtracted {
			logger.Fatal("Provided public-key does not correspond to the private key")
		}
	}

	helloDict["authextra"] = wamp.Dict{"pubkey": publicKey}

	cfg := client.Config{
		Realm:        realm,
		Logger:       logger,
		HelloDetails: helloDict,
		AuthHandlers: map[string]client.AuthFunc{
			"cryptosign": func(c *wamp.Challenge) (string, wamp.Dict) {
				challengeHex, _ := wamp.AsString(c.Extra["challenge"])
				challengeBytes, _ := hex.DecodeString(challengeHex)

				signed := ed25519.Sign(pvk, challengeBytes)
				signedHex := hex.EncodeToString(signed)
				result := signedHex + challengeHex
				return result, wamp.Dict{}
			},
		},
		Serialization: serializer,
	}

	return connect(url, cfg, logger)
}

func Subscribe(session *client.Client, logger *log.Logger, topic string) {
	// Define function to handle events received.
	eventHandler := func(event *wamp.Event) {
		argsKWArgs(event.Arguments, event.ArgumentsKw)
	}

	// Subscribe to topic.
	err := session.Subscribe(topic, eventHandler, nil)
	if err != nil {
		logger.Fatal("subscribe error:", err)
	} else {
		logger.Println("Subscribed to", topic)
	}
	// Wait for CTRL-c or client close while handling events.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	select {
	case <-sigChan:
	case <-session.Done():
		logger.Print("Router gone, exiting")
		return // router gone, just exit
	}

	// Unsubscribe from topic.
	if err = session.Unsubscribe(topic); err != nil {
		logger.Println("Failed to unsubscribe:", err)
	}
}

func Publish(session *client.Client, logger *log.Logger, topic string, args []string, kwargs map[string]string) {

	// Publish to topic.
	err := session.Publish(topic, nil, listToWampList(args), dictToWampDict(kwargs))
	if err != nil {
		logger.Fatal("publish error:", err)
	} else {
		logger.Println("Published", topic, "event")
	}
}

func Register(session *client.Client, logger *log.Logger, procedure string, command string) {
	eventHandler := func(ctx context.Context, inv *wamp.Invocation) client.InvokeResult {

		argsKWArgs(inv.Arguments, inv.ArgumentsKw)

		if command != "" {
			err, out, _ := shellOut(command)
			if err != nil {
				log.Println("error: ", err)
			}

			return client.InvokeResult{Args: wamp.List{out}}
		}

		return client.InvokeResult{Args: wamp.List{""}}
	}

	if err := session.Register(procedure, eventHandler, nil); err != nil {
		logger.Fatal("Failed to register procedure:", err)
	} else {
		logger.Println("Registered procedure", procedure, "with router")
	}

	// Wait for CTRL-c or client close while handling remote procedure calls.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	select {
	case <-sigChan:
	case <-session.Done():
		logger.Print("Router gone, exiting")
		return // router gone, just exit
	}

	if err := session.Unregister(procedure); err != nil {
		logger.Println("Failed to unregister procedure:", err)
	}

	logger.Println("Registered procedure with router")

}

func Call(session *client.Client, logger *log.Logger, procedure string, args []string, kwargs map[string]string) {
	ctx := context.Background()
	result, err := session.Call(ctx, procedure, nil, listToWampList(args), dictToWampDict(kwargs), nil)
	if err != nil {
		logger.Println("Failed to call ", err)
	} else if result != nil {
		jsonString, err := json.MarshalIndent(result.Arguments[0], "", "    ")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(jsonString))
	}
}

func listToWampList(args []string) wamp.List {
	var arguments wamp.List

	if args == nil {
		return arguments
	}

	for _, value := range args {
		arguments = append(arguments, value)
	}

	return arguments
}

func dictToWampDict(kwargs map[string]string) wamp.Dict {
	var keywordArguments wamp.Dict = make(map[string]interface{})
	for key, value := range kwargs {
		keywordArguments[key] = value
	}
	return keywordArguments
}

func argsKWArgs(args wamp.List, kwArgs wamp.Dict) {
	if len(args) != 0 {
		fmt.Println("args:")
		jsonString, err := json.MarshalIndent(args, "", "    ")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(jsonString))
	}

	if len(kwArgs) != 0 {
		fmt.Println("kwargs:")
		jsonString, err := json.MarshalIndent(kwArgs, "", "    ")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(jsonString))
	}

	if len(args) == 0 && len(kwArgs) == 0 {
		fmt.Println("args: []")
		fmt.Println("kwargs: {}")
	}
}

func shellOut(command string) (error, string, string) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	var cmd *exec.Cmd
	cmd = exec.Command("bash", "-c", command)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return err, stdout.String(), stderr.String()
}
