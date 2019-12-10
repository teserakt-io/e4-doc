---
title: "2) Protecting messages with E4"
date: "2019-09-06"
lastmod: "2019-09-27"
draft: false
weight: 2
---

In previous part, we made a simple application where `alice` and `bob` could exchange messages. But now we want them to be able to communicate privately, even if `eve` subscribe to the MQTT topic.

To do so, we'll integrate the E4 library in our application, and create a symmetric key, and securely share it with `alice` and `bob`, so they can encrypt their messages with it. After this, only key holders could read the exchanged messages.

First, let's modify our previous application to create an E4 client and read a client password from the flags.
{{< highlight go "hl_lines=7 13 15 22-25 31" >}}
package main

import (
    // ...

	mqtt "github.com/eclipse/paho.mqtt.golang"
	e4 "github.com/teserakt-io/e4go"
)

func main() {
	// 1. Read a client and a peer identifiers from command line flags
	var clientName string
	var clientPassword string
	flag.StringVar(&clientName, "client", "", "the client name")
	flag.StringVar(&clientPassword, "password", "", "the client password")
	flag.Parse()

	if len(clientName) == 0 {
		fmt.Println("-client is required")
		os.Exit(1)
	}
	if len(clientPassword) < 16 {
		fmt.Println("-password is required and must be longer than 16 characters")
		os.Exit(1)
	}

	// 2. Connect to a MQTT broker (we'll use our public mqtt.teserakt.io:1338)
	// ...
	fmt.Printf("> connected to %s\n", brokerEndpoint)

	e4Client, err := e4.NewSymKeyClientPretty(clientName, clientPassword, fmt.Sprintf("%s.json", clientName))

	// 3 - Subscribe to message MQTT topic and print incoming messages to stdout
	// ...
}
{{< / highlight >}}

Now we'll replace the previous `mqtt.Subscribe()` to subscribe to the peer topic and the e4 receiving topic
We also feed the incoming messages to the `e4Client.Unprotect()` in the MQTT messages reception callback.

{{<tabs>}}
{{<tab after>}}
{{<highlight go>}}
	// ...
	// 3 - Subscribe to message MQTT topic and print incoming messages to stdout
	messageTopic := "/e4go/demo/messages"
	topics := map[string]byte{
		messageTopic:                 1,
		e4Client.GetReceivingTopic(): 2,
	}
	token := mqttClient.SubscribeMultiple(topics, func(_ mqtt.Client, msg mqtt.Message) {
		fmt.Printf("< receive raw message on %s: %s\n", msg.Topic(), msg.Payload())
		clearMessage, err := e4Client.Unprotect(msg.Payload(), msg.Topic())
		if err != nil {
			fmt.Printf("failed to unprotect message: %v\n", err)
			return
		}

		fmt.Printf("< unprotected message: %s\n", clearMessage)
	})
	// ...
{{</highlight>}}
{{</tab>}}
{{<tab before>}}
{{<highlight go>}}
	// ...
	// 3 - Subscribe to message MQTT topic and print incoming messages to stdout
	messageTopic := "/e4go/demo/messages"
	token := mqttClient.Subscribe(messageTopic, 1, func(_ mqtt.Client, msg mqtt.Message) {
	 	fmt.Printf("< received raw message on %s: %s\n", msg.Topic(), msg.Payload())
	 })
	// ...
{{</highlight>}}
{{</tab>}}
{{</tabs>}}


And last, we also update the payload we pass to `mqtt.Publish`, to protect the message first :

{{< highlight go "hl_lines=7-12" >}}
        // ...
		message := scanner.Text()
		if len(message) == 0 { // Skip empty messages
			continue
		}

		protectedMessage, err := e4Client.ProtectMessage([]byte(message), messageTopic)
		if err != nil {
			fmt.Printf("> failed to protect message: %v\n", err)
			continue
		}
		if token := mqttClient.Publish(messageTopic, 1, true, protectedMessage); token.Error() != nil {
			fmt.Printf("> failed to publish message: %v\n", token.Error())
			continue
		}
        // ...
{{< / highlight >}}

[Click here to download the full source at this point](../e4demo-step2.go)

And voila, our application have now integrated E4 and is ready to communicate securely.

But something goes wrong when we try it:
```text
# Alice
$ go run e4demo.go -client alice -password super-secret-alice-password
> connected to mqtt.teserakt.io:1883
> subscribed to MQTT topic /e4go/demo/messages
> type anything and press enter to send the message to /e4go/demo/messages:
Hello, I'm alice and this is a secret message for bob!
> failed to protect message: topic key not found

# Bob
$ go run e4demo.go -client bob -password super-secret-bob-password
connected to mqtt.teserakt.io:1883
> subscribed to MQTT topic /e4go/demo/messages
> type anything and press enter to send the message to /e4go/demo/messages:
```

Right, we still need to transmit the key to the clients so they can encrypt and decrypt messages with it.
in E4, each topics need its own key. And we can see that our updated client now isn't listening only for incoming messages from the peer, but also from `e4Client.ReceivingTopic()` topic. And this is how we'll transmit the topic key to each clients.

We'll cover this in the next section.
