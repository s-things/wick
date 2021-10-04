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
	"bytes"
	"context"
	"fmt"
	"github.com/gammazero/nexus/v3/client"
	"github.com/gammazero/nexus/v3/wamp"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
)

func subscribe(URLSub string, realmSub string, topicSub string,  authidFlag string, authSecretFlag string) {
	logger := log.New(os.Stdout, "Subscriber> ", 0)
	var cfg client.Config
	if authidFlag != "" && authSecretFlag != "" {
		cfg = client.Config{
			Realm:  realmSub,
			Logger: logger,
			HelloDetails: wamp.Dict{
				"authid": authidFlag,
			},
			AuthHandlers: map[string]client.AuthFunc{
				"wampcra": clientAuthFunc,
			},
		}
	} else {
		cfg = client.Config{
			Realm:  realmSub,
			Logger: logger,
		}
	}

	goodSecret = authSecretFlag

	// Connect subscriber session.
	subscriber, err := client.ConnectNet(context.Background(), URLSub, cfg)
	if err != nil {
		logger.Fatal(err)
	} else {
		logger.Println("Connected to ", URLSub)
	}
	defer subscriber.Close()

	// Define function to handle events received.
	eventHandler := func(event *wamp.Event) {
		if len(event.Arguments) != 0 {
			fmt.Print("args : ")
			for index, value := range event.Arguments {
				if index != len(event.Arguments)-1 {
					fmt.Print(value, ", ")
				} else {
					fmt.Println(value)
				}
			}
		} else {
			fmt.Println("args : {}")
		}
		i := 1
		if len(event.ArgumentsKw) != 0 {
			fmt.Print("kwargs : ")
			for key, value := range event.ArgumentsKw {
				if i == len(event.ArgumentsKw) {
					fmt.Print(key, "=", value, "\n")
				} else {
					fmt.Print(key, "=", value, ", ")
				}
				i++
			}
		} else {
			fmt.Println("kwargs : {}")
		}
	}

	// Subscribe to topic.
	err = subscriber.Subscribe(topicSub, eventHandler, nil)
	if err != nil {
		logger.Fatal("subscribe error:", err)
	} else {
		logger.Println("Subscribed to", topicSub)
	}
	// Wait for CTRL-c or client close while handling events.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	select {
	case <-sigChan:
	case <-subscriber.Done():
		logger.Print("Router gone, exiting")
		return // router gone, just exit
	}

	// Unsubscribe from topic.
	if err = subscriber.Unsubscribe(topicSub); err != nil {
		logger.Println("Failed to unsubscribe:", err)
	}
}

func publish(URLPub string, realmPub string, topicPub string, argsList []string, kwargsMap map[string]string,
	authidFlag string, authSecretFlag string) {
	logger := log.New(os.Stdout, "Publisher> ", 0)
	var cfg client.Config
	if authidFlag != "" && authSecretFlag != "" {
		cfg = client.Config{
			Realm:  realmPub,
			Logger: logger,
			HelloDetails: wamp.Dict{
				"authid": authidFlag,
			},
			AuthHandlers: map[string]client.AuthFunc{
				"wampcra": clientAuthFunc,
			},
		}
	} else {
		cfg = client.Config{
			Realm:  realmPub,
			Logger: logger,
		}
	}
	goodSecret = authSecretFlag


	// Connect publisher session.
	publisher, err := client.ConnectNet(context.Background(), URLPub, cfg)
	if err != nil {
		logger.Fatal(err)
	}
	defer publisher.Close()

	var arguments wamp.List
	for _, value := range argsList {
		arguments = append(arguments, value)
	}

	var keywordArguments wamp.Dict = make(map[string]interface{})
	for key, value := range kwargsMap {
		keywordArguments[key] = value
	}
	// Publish to topic.
	err = publisher.Publish(topicPub, nil, arguments, keywordArguments)
	if err != nil {
		logger.Fatal("publish error:", err)
	} else {
		logger.Println("Published", topicPub, "event")
	}
}

func register(URLReg string, realmReg string, procedureReg string, commands []string, shell string,
	authidFlag string, authSecretFlag string) {
	logger := log.New(os.Stdout, "Register> ", 0)
	var cfg client.Config
	if authidFlag != "" && authSecretFlag != "" {
		cfg = client.Config{
			Realm:  realmReg,
			Logger: logger,
			HelloDetails: wamp.Dict{
				"authid": authidFlag,
			},
			AuthHandlers: map[string]client.AuthFunc{
				"wampcra": clientAuthFunc,
			},
		}
	} else {
		cfg = client.Config{
			Realm:  realmReg,
			Logger: logger,
		}
	}
	goodSecret = authSecretFlag

	register, err := client.ConnectNet(context.Background(), URLReg, cfg)
	logger.Println("Connected to ", URLReg)
	if err != nil {
		logger.Fatal(err)
	}

	defer register.Close()

	eventHandler := func(ctx context.Context, inv *wamp.Invocation) client.InvokeResult {
		if len(inv.Arguments) != 0 {
			fmt.Print("args : ")
			for index, value := range inv.Arguments {
				if index != len(inv.Arguments)-1 {
					fmt.Print(value, ", ")
				} else {
					fmt.Println(value)
				}
			}
		} else {
			fmt.Println("args : {}")
		}
		i := 1
		if len(inv.ArgumentsKw) != 0 {
			fmt.Print("kwargs : ")
			for key, value := range inv.ArgumentsKw {
				if i == len(inv.ArgumentsKw) {
					fmt.Print(key, "=", value, "\n")
				} else {
					fmt.Print(key, "=", value, ", ")
				}
				i++
			}
		} else {
			fmt.Println("kwargs : {}")
		}

		if commands != nil {
			var command string
			for _, com := range commands {
				if command != "" {
					command = command + "; " + com
				} else {
					command = command + com
				}
			}

			err, out, _ := shellOut(command, shell)
			if err != nil {
				log.Println("error: ", err)
			}

			return client.InvokeResult{Args: wamp.List{out}}
		} else if shell != "" && commands == nil {
			err, out, _ := execute(shell)
			if err != nil {
				log.Println("error: ", err)
			}
			return client.InvokeResult{Args: wamp.List{out}}
		}

		return client.InvokeResult{Args: wamp.List{""}}
	}

	if err = register.Register(procedureReg, eventHandler, nil);
		err != nil {
		logger.Fatal("Failed to register procedure:", err)
	} else {
		logger.Println("Registered procedure", procedureReg, "with router")
	}

	// Wait for CTRL-c or client close while handling remote procedure calls.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	select {
	case <-sigChan:
	case <-register.Done():
		logger.Print("Router gone, exiting")
		return // router gone, just exit
	}

	if err = register.Unregister(procedureReg); err != nil {
		logger.Println("Failed to unregister procedure:", err)
	}

	logger.Println("Registered procedure with router")

}

func call(URLCal string, realmCal string, procedureCal string, argsList []string, kwargsMap map[string]string,
	authidFlag string, authSecretFlag string) {
	logger := log.New(os.Stderr, "Caller> ", 0)

	var cfg client.Config
	if authidFlag != "" && authSecretFlag != "" {
		cfg = client.Config{
			Realm:  realmCal,
			Logger: logger,
			HelloDetails: wamp.Dict{
				"authid": authidFlag,
			},
			AuthHandlers: map[string]client.AuthFunc{
				"wampcra": clientAuthFunc,
			},
		}
	} else {
		cfg = client.Config{
			Realm:  realmCal,
			Logger: logger,
		}
	}

	goodSecret = authSecretFlag

	// Connect caller client
	caller, err := client.ConnectNet(context.Background(), URLCal, cfg)
	if err != nil {
		logger.Fatal(err)
	}
	defer caller.Close()

	ctx := context.Background()

	var arguments wamp.List
	for _, value := range argsList {
		arguments = append(arguments, value)
	}

	var keywordArguments wamp.Dict = make(map[string]interface{})
	for key, value := range kwargsMap {
		keywordArguments[key] = value
	}

	result, err := caller.Call(ctx, procedureCal, nil, arguments, keywordArguments, nil)
	if err != nil {
		logger.Println("Failed to call ", err)
	} else if result != nil {
		fmt.Println(result.Arguments[0])
	}
}

func shellOut(command string, ShellToUse string) (error, string, string) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	var cmd *exec.Cmd
	if strings.HasSuffix(command, ".sh") || strings.HasSuffix(command, ".py") {
		cmd = exec.Command(ShellToUse, command)
	} else {
		cmd = exec.Command(ShellToUse, "-c", command)
	}

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return err, stdout.String(), stderr.String()
}

func execute(command string) (error, string, string) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	var cmd *exec.Cmd

	cmd = exec.Command(command)

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return err, stdout.String(), stderr.String()
}
