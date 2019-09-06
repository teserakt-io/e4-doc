---
title: "2) Protecting messages with E4"
date: "2019-09-06"
lastmod: "2019-09-06"
draft: false
---

In previous part, we made a simple application where `alice` and `bob` could exchange messages. But now we want them to be able to communicate privately, even if `eve` subscribe to their respective topics.

To do so, we'll create a symmetric key, and securely share it with `alice` and `bob`, so they can encrypt their messages.

First, let's modify our previous application to create an E4 client and read a client password from the flags.
{{< highlight go "hl_lines=7 13 17 24-27 37" >}}
package main

import (
    // ...

	mqtt "github.com/eclipse/paho.mqtt.golang"
	e4 "gitlab.com/teserakt/e4common"
)

func main() {
	// 1 - Read a client identifier from a command line flag
	var clientName string
	var clientPassword string
	var peerName string
	flag.StringVar(&clientName, "client", "", "the client name")
	flag.StringVar(&clientPassword, "password", "", "the client password")
	flag.StringVar(&peerName, "peer", "", "the peer name to read messages from")
	flag.Parse()

	if len(clientName) == 0 {
		fmt.Println("-client is required")
		os.Exit(1)
	}
	if len(clientPassword) < 16 {
		fmt.Println("-password is required and must contains at least 16 characters")
		os.Exit(1)
	}
	if len(peerName) == 0 {
		fmt.Println("-peer is required")
		os.Exit(1)
	}

	// 2 - Connect to a MQTT broker (we'll use our public mqtt.teserakt.io:1338)
	// ...
	fmt.Printf("connected to %s\n", brokerEndpoint)

	e4Client, err := e4.NewSymKeyClientPretty(clientName, clientPassword, fmt.Sprintf("%s.json", clientName))

	// 3 - Subscribe to peer MQTT topic and print incoming messages to stdout
	// ...
}
{{< / highlight >}}

Now we'll update the `mqtt.Subscribe()` to subscribe to the peer topic and the e4 receiving topic
We also feed the incoming messages to the `e4Client.Unprotect()` in the message's reception callback.

```go
    // ...
	peerTopic := fmt.Sprintf("/e4go/demo/%s/messages", peerName)
	topics := map[string]byte{
		peerTopic:                    1,
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
```

And last, we also update the `mqtt.Publish` call to protect the message :

{{< highlight go "hl_lines=7-12" >}}
        // ...
		message := scanner.Text()
		if len(message) == 0 { // Skip empty messages
			continue
		}

		protectedMessage, err := e4Client.ProtectMessage([]byte(message), publishTopic)
		if err != nil {
			fmt.Printf("> failed to protect message: %v\n", err)
			continue
		}
		if token := mqttClient.Publish(publishTopic, 1, true, protectedMessage); token.Error() != nil {
			fmt.Printf("> failed to publish message: %v\n", token.Error())
			continue
		}
        // ...
{{< / highlight >}}

[Click here to download the full source at this point](../e4demo-step2.go)

And voila, our application have now integrated E4 and is ready to communicate securely.

But something goes wrong when we try it:
```bash
$ go run e4demo.go  -client alice -peer bob -password alice-super-secret-password
connected to mqtt.teserakt.io:1883
subscribed to peer topic /e4go/demo/bob/messages
type anything and press enter to publish a message on to /e4go/demo/alice/messages:
Hello, I'm alice and this is a secret message for bob!
> failed to protect message: topic key not found
```

Right, we still need to transmit the key to the clients so they can encrypt and decrypt messages with it.
in E4, each topics need its own key. And we can see that our updated client now isn't listening only for incoming messages from the peer, but also from `e4Client.ReceivingTopic()` topic. And this is how we'll transmit the publishTopic and peerTopics keys to each clients.

We'll cover this in the next section.
