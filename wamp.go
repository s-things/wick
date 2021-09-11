package main

import (
	"context"
	"fmt"
	"github.com/gammazero/nexus/v3/client"
	"github.com/gammazero/nexus/v3/wamp"
	"log"
	"os"
	"os/signal"
	"reflect"
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
	}
	defer subscriber.Close()

	// Define function to handle events received.
	eventHandler := func(event *wamp.Event) {
		if len(event.Arguments) != 0 {
			fmt.Println("args :", event.Arguments[0])
		}
		if len(event.Arguments) != 0 {
			m := event.ArgumentsKw
			fmt.Println(reflect.ValueOf(m).MapKeys()[0], ": ", m["kwargs"])
		}
	}

	// Subscribe to topic.
	err = subscriber.Subscribe(topicSub, eventHandler, nil)
	if err != nil {
		logger.Fatal("subscribe error:", err)
	}
	logger.Println("Subscribed to", topicSub)

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

	// Publish to topic.
	err = publisher.Publish(topicPub, nil,wamp.List{argsList}, wamp.Dict{"kwargs": kwargsMap})
	if err != nil {
		logger.Fatal("publish error:", err)
	}
	logger.Println("Published", topicPub, "event")
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
			fmt.Println("args :", inv.Arguments[0])
		}
		if len(inv.Arguments) != 0 {
			m := inv.ArgumentsKw
			fmt.Println(reflect.ValueOf(m).MapKeys()[0], ": ", m["kwargs"])
		}
		return client.InvokeResult{Args: wamp.List{inv.Arguments}}
	}

	if err = register.Register(procedureReg, eventHandler, nil);
		err != nil {
		logger.Fatal("Failed to publish procedure:", err)
	}
	logger.Println("Registered procedure", procedureReg, "with router")

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

	caller.Call(ctx, procedureCal, nil, wamp.List{argsList}, wamp.Dict{"kwargs": kwargsMap}, nil)

	logger.Println("Call the procedure ", procedureCal)
}
