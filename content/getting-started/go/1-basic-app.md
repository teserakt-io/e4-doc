---
title: "1) Basic application setup"
date: "2019-09-06"
lastmod: "2019-09-27"
draft: false
weight: 1
---

Let's start by creating a basic Go client application. It will:

1. Read a client identifier from command line flags
2. Connect to a MQTT broker (we'll use our public `mqtt.teserakt.io:1338`)
3. Subscribe to the MQTT topic `/e4go/demo/messages` and print any incoming messages to stdout
4. Wait for user input on stdin, so user can type in a message and press enter. Messages will then be publish on the peer MQTT topic `/e4go/demo/<peerName>/messages`.

Let's first move to an empty directory, and create our application file:
```bash
$ mkdir -p e4demo && cd e4demo
$ go mod init e4demo
$ touch e4demo.go
```

Now the code:
```go
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func main() {
	// 1 - Read a client identifier from a command line flag
	var clientName string
	flag.StringVar(&clientName, "client", "", "the client name")
	flag.Parse()

	if len(clientName) == 0 {
		panic("-client is required")
	}

	// 2 - Connect to a MQTT broker (we'll use our public mqtt.teserakt.io:1338)
	brokerEndpoint := "mqtt.teserakt.io:1883"
	mqttClient, err := initMQTT(brokerEndpoint, clientName)
	if err != nil {
		panic(fmt.Sprintf("failed to init mqtt client: %v", err))
	}
	fmt.Printf("> connected to %s\n", brokerEndpoint)

	// 3 - Subscribe to message MQTT topic and print incoming messages to stdout
	messageTopic := "/e4go/demo/messages"
	token := mqttClient.Subscribe(messageTopic, 1, func(_ mqtt.Client, msg mqtt.Message) {
		fmt.Printf("< received raw message on %s: %s\n", msg.Topic(), msg.Payload())
	})
	if !token.WaitTimeout(1 * time.Second) {
		panic(fmt.Sprintf("failed to subscribe to MQTT topic: %v\n", token.Error()))
	}
	fmt.Printf("> subscribed to MQTT topic %s\n", messageTopic)

	// 4 - Wait for user input on stdin and publish messages
	// on the peer MQTT topic `/e4go/demo/messages` once user press the enter key.
	fmt.Printf("> type anything and press enter to send the message to %s:\n", messageTopic)
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		message := scanner.Text()
		if len(message) == 0 { // Skip empty messages
			continue
		}

		if token := mqttClient.Publish(messageTopic, 1, true, message); token.Error() != nil {
			fmt.Printf("> failed to publish message: %v\n", token.Error())
			continue
		}

		fmt.Println("> message published successfully")
	}
}
```

We can now add our `initMQTT` function:

```go
func initMQTT(brokerEndpoint, clientID string) (mqtt.Client, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(brokerEndpoint)
	opts.SetClientID(clientID)
	opts.SetCleanSession(true)

	mqttClient := mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.WaitTimeout(1*time.Second) && token.Error() != nil {
		return nil, token.Error()
	}

	return mqttClient, nil
}
```

And we're done. We now have a basic application able to send and receive message over MQTT.

[Click here to download the full source at this point](../e4demo-step1.go)

We can now try it and exchange messages between `alice` and `bob`.

Run in a first terminal and start an instance for `alice`, and type in the message for `bob`:
```text
$ go run e4demo.go -client alice
> connected to mqtt.teserakt.io:1883
> subscribed to MQTT topic /e4go/demo/messages
> type anything and press enter to send the message to /e4go/demo/messages:
Hello, I'm alice, and this is a secret message for bob!
> message published successfully
```

And start a second one in another terminal for `bob`:

```text
$ go run e4demo.go -client bob
> connected to mqtt.teserakt.io:1883
> subscribed to MQTT topic /e4go/demo/messages
< received raw message on /e4go/demo/messages: Hello, I'm alice, and this is a secret message for bob!
> type anything and press enter to send the message to /e4go/demo/messages:
```

`alice` and `bob` can now exchange messages between each others.
But the evil `eve` can sneak in, subscribe to the topic, and read / write messages:

```text
$ go run e4demo.go -client eve
> connected to mqtt.teserakt.io:1883
< received raw message on /e4go/demo/messages: Hello, I'm alice, and this is a secret message for bob!
> subscribed to MQTT topic /e4go/demo/messages
> type anything and press enter to send the message to /e4go/demo/messages:
```

Let's jump to the next section and see how E4 can keep those message secret!
