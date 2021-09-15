package main

import (
	"context"
	"fmt"
	"github.com/gammazero/nexus/v3/client"
	"github.com/gammazero/nexus/v3/wamp"
	"log"
	"os"
	"os/signal"
)

func subscribe(URLSub string, realmSub string, topicSub string){
	logger := log.New(os.Stdout, "Subscriber> ", 0)
	cfg := client.Config{
		Realm:  realmSub,
		Logger: logger,
	}

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
			for index,value := range event.Arguments {
				if index != len(event.Arguments) -1 {
					fmt.Print(value,", ")
				} else {
					fmt.Println(value)
				}
			}
		} else{
			fmt.Println("args : {}")
		}
		i := 1
		if len(event.ArgumentsKw) != 0 {
			fmt.Print("kwargs : ")
			for key,value := range event.ArgumentsKw{
				if i == len(event.ArgumentsKw) {
					fmt.Print(key ,"=", value, "\n")
				} else {
					fmt.Print(key ,"=", value ,", ")
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

func publish(URLPub string, realmPub string, topicPub string, argsList []string, kwargsMap map[string]string) {
	logger := log.New(os.Stdout, "Publisher> ", 0)
	cfg := client.Config{
		Realm:  realmPub,
		Logger: logger,
	}

	// Connect publisher session.
	publisher, err := client.ConnectNet(context.Background(), URLPub, cfg)
	if err != nil {
		logger.Fatal(err)
	}
	defer publisher.Close()

	var arguments wamp.List
	for _,value := range argsList {
		arguments = append(arguments,value)
	}

	var keywordArguments wamp.Dict = make(map[string]interface{})
	for key,value := range kwargsMap {
		keywordArguments[key] = value
	}
	// Publish to topic.
	err = publisher.Publish(topicPub, nil ,arguments, keywordArguments)
	if err != nil {
		logger.Fatal("publish error:", err)
	} else {
		logger.Println("Published", topicPub, "event")
	}
}

func register(URLReg string, realmReg string, procedureReg string){
	logger := log.New(os.Stdout, "Register> ", 0)
	cfg := client.Config{
		Realm:  realmReg,
		Logger: logger,
	}
	register, err := client.ConnectNet(context.Background(), URLReg, cfg)
	logger.Println("Connected to ", URLReg)
	if err != nil {
		logger.Fatal(err)
	}
	defer register.Close()

	eventHandler:= func(ctx context.Context, inv *wamp.Invocation) client.InvokeResult {
		if len(inv.Arguments) != 0 {
			fmt.Print("args : ")
			for index,value := range inv.Arguments {
				if index != len(inv.Arguments) -1 {
					fmt.Print(value,", ")
				} else {
					fmt.Println(value)
				}
			}
		}else{
			fmt.Println("args : {}")
		}
		i := 1
		if len(inv.ArgumentsKw) != 0 {
			fmt.Print("kwargs : ")
			for key,value := range inv.ArgumentsKw{
				if i == len(inv.ArgumentsKw) {
					fmt.Print(key ,"=", value, "\n")
				} else {
					fmt.Print(key ,"=", value ,", ")
				}
				i++
			}
		} else {
			fmt.Println("kwargs : {}")
		}
		return client.InvokeResult{Args: wamp.List{inv.Arguments}}
	}

	if err = register.Register(procedureReg, eventHandler, nil);
		err != nil {
		logger.Fatal("Failed to publish procedure:", err)
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

func call(URLCal string, realmCal string, procedureCal string, argsList []string, kwargsMap map[string]string){
	logger := log.New(os.Stderr, "Caller> ", 0)

	cfg := client.Config{
		Realm:  realmCal,
		Logger: logger,
	}

	// Connect caller client
	caller, err := client.ConnectNet(context.Background(), URLCal, cfg)
	if err != nil {
		logger.Fatal(err)
	}
	defer caller.Close()

	ctx := context.Background()

	var arguments wamp.List
	for _,value := range argsList {
		arguments = append(arguments,value)
	}

	var keywordArguments wamp.Dict = make(map[string]interface{})
	for key,value := range kwargsMap {
		keywordArguments[key] = value
	}

	result, err := caller.Call(ctx, procedureCal, nil, arguments, keywordArguments, nil)
	if err != nil {
		logger.Println("Failed to call ", err)
	} else if result != nil{
		logger.Println("Call the procedure ", procedureCal)
	}
}
