/*
*
* Copyright 2021-2022 Simple Things Inc.
*
* Permission is hereby granted, free of charge, to any person obtaining a copy
* of this software and associated documentation files (the "Software"), to deal
* in the Software without restriction, including without limitation the rights
* to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
* copies of the Software, and to permit persons to whom the Software is
* furnished to do so, subject to the following conditions:
*
* The above copyright notice and this permission notice shall be included in all
* copies or substantial portions of the Software.
*
* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
* IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
* FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
* LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
* OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
* SOFTWARE.
*
 */

package core

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/gammazero/nexus/v3/client"
	"github.com/gammazero/nexus/v3/transport/serialize"
	"github.com/gammazero/nexus/v3/wamp"
)

var logger *logrus.Logger

func init() {
	logger = logrus.New()
}

func connect(url string, cfg client.Config) *client.Client {
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

func ConnectAnonymous(url string, realm string, serializer serialize.Serialization, authid string,
	authrole string) *client.Client {

	cfg := getAnonymousAuthConfig(realm, serializer, authid, authrole)

	return connect(url, cfg)
}

func ConnectTicket(url string, realm string, serializer serialize.Serialization, authid string, authrole string,
	ticket string) *client.Client {

	cfg := getTicketAuthConfig(realm, serializer, authid, authrole, ticket)

	return connect(url, cfg)
}

func ConnectCRA(url string, realm string, serializer serialize.Serialization, authid string, authrole string,
	secret string) *client.Client {

	cfg := getCRAAuthConfig(realm, serializer, authid, authrole, secret)

	return connect(url, cfg)
}

func ConnectCryptoSign(url string, realm string, serializer serialize.Serialization, authid string, authrole string,
	privateKey string) *client.Client {

	cfg := getCryptosignAuthConfig(realm, serializer, authid, authrole, privateKey)

	return connect(url, cfg)
}

func Subscribe(session *client.Client, topic string, subscribeOptions map[string]string, printDetails bool) {
	// Define function to handle events received.
	eventHandler := func(event *wamp.Event) {
		if printDetails {
			argsKWArgs(event.Arguments, event.ArgumentsKw, event.Details)
		} else {
			argsKWArgs(event.Arguments, event.ArgumentsKw, nil)
		}
	}

	// Subscribe to topic.
	err := session.Subscribe(topic, eventHandler, dictToWampDict(subscribeOptions))
	if err != nil {
		logger.Fatal("subscribe error:", err)
	} else {
		logger.Printf("Subscribed to topic '%s'\n", topic)
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

func actualPublish(session *client.Client, topic string, args []string, kwargs map[string]string, logPublishTime bool,
	delayPublish int, group *sync.WaitGroup, publishOptions map[string]string) {
	if group != nil {
		defer group.Done()
	}
	if delayPublish > 0 {
		time.Sleep(time.Duration(delayPublish) * time.Millisecond)
	}

	startTime := time.Now().UnixMilli()

	// Publish to topic.
	err := session.Publish(topic, dictToWampDict(publishOptions), listToWampList(args), dictToWampDict(kwargs))
	if err != nil {
		logger.Fatal("Publish error:", err)
	} else {
		logger.Printf("Published to topic '%s'\n", topic)
	}

	if logPublishTime {
		endTime := time.Now().UnixMilli()
		logger.Printf("call took %dms\n", endTime-startTime)
	}
}

func Publish(session *client.Client, topic string, args []string, kwargs map[string]string, publishOptions map[string]string,
	logPublishTime bool, repeatPublish int, delayPublish int, parallelPublish bool) {

	startTime := time.Now().UnixMilli()

	if parallelPublish {
		var wg sync.WaitGroup
		wg.Add(repeatPublish)

		for i := 0; i < repeatPublish; i++ {
			go actualPublish(session, topic, args, kwargs, logPublishTime, delayPublish, &wg, publishOptions)
		}

		wg.Wait()
	} else {
		for i := 0; i < repeatPublish; i++ {
			actualPublish(session, topic, args, kwargs, logPublishTime, delayPublish, nil, publishOptions)
		}
	}

	if logPublishTime && repeatPublish > 1 {
		endTime := time.Now().UnixMilli()
		logger.Printf("%d calls took %dms\n", repeatPublish, endTime-startTime)
	}
}

func Register(session *client.Client, procedure string, command string, delay int, invokeCount int, registerOptions map[string]string) {

	// If the user has called with --invoke-count
	hasMaxInvokeCount := invokeCount > 0

	eventHandler := func(ctx context.Context, inv *wamp.Invocation) client.InvokeResult {

		argsKWArgs(inv.Arguments, inv.ArgumentsKw, nil)

		result := ""

		if command != "" {
			err, out, _ := shellOut(command)
			if err != nil {
				logger.Println("error: ", err)
			}
			result = out
		}

		if hasMaxInvokeCount {
			invokeCount--
			if invokeCount == 0 {
				session.Unregister(procedure)
				time.AfterFunc(1*time.Second, func() {
					logger.Println("session closing")
					session.Close()
				})
			}
		}

		return client.InvokeResult{Args: wamp.List{result}}

	}

	if delay > 0 {
		logger.Printf("procedure will be registered after %d seconds.\n", delay)
		time.Sleep(time.Duration(delay) * time.Second)
	}

	if err := session.Register(procedure, eventHandler, dictToWampDict(registerOptions)); err != nil {
		logger.Fatal("Failed to register procedure:", err)
	} else {
		logger.Printf("Registered procedure '%s'\n", procedure)
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

func actuallyCall(session *client.Client, procedure string, args []string, kwargs map[string]string, logCallTime bool,
	delayCall int, group *sync.WaitGroup, callOptions map[string]string) {

	if group != nil {
		defer group.Done()
	}

	if delayCall > 0 {
		time.Sleep(time.Duration(delayCall) * time.Millisecond)
	}

	startTime := time.Now().UnixMilli()
	result, err := session.Call(context.Background(), procedure, dictToWampDict(callOptions), listToWampList(args), dictToWampDict(kwargs), nil)
	if err != nil {
		logger.Fatal(err)
	} else if result != nil && len(result.Arguments) > 0 {
		jsonString, err := json.MarshalIndent(result.Arguments[0], "", "    ")
		if err != nil {
			logger.Fatal(err)
		}
		fmt.Println(string(jsonString))
	}

	if logCallTime {
		endTime := time.Now().UnixMilli()
		logger.Printf("call took %dms\n", endTime-startTime)
	}
}

func Call(session *client.Client, procedure string, args []string, kwargs map[string]string,
	logCallTime bool, repeatCount int, delayCall int, parallelCall bool, callOptions map[string]string) {

	startTime := time.Now().UnixMilli()

	if parallelCall {
		var wg sync.WaitGroup
		wg.Add(repeatCount)

		for i := 0; i < repeatCount; i++ {
			go actuallyCall(session, procedure, args, kwargs, logCallTime, delayCall, &wg, callOptions)
		}

		wg.Wait()
	} else {
		for i := 0; i < repeatCount; i++ {
			actuallyCall(session, procedure, args, kwargs, logCallTime, delayCall, nil, callOptions)
		}
	}

	if logCallTime && repeatCount > 1 {
		endTime := time.Now().UnixMilli()
		logger.Printf("%d calls took %dms\n", repeatCount, endTime-startTime)
	}
}
