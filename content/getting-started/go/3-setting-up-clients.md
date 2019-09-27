---
title: "3) Setting up E4 clients"
date: "2019-09-06"
lastmod: "2019-09-27"
draft: false
weight: 3
---

Previously, we've updated our application to integrate the E4 library, and protect and unprotect the exchanged messages. But we could not communicate yet, since the clients didn't hold any keys necessary to encrypt or decrypt the messages. We'll fix this now.

E4 clients can receive commands, meant to update their internal state, like the list of topic keys they can uses. So to fix our issue, we'll need to:

* generate a topic key for /e4demo/messages topic
* send this key to each clients, on their respective E4 receiving topics

Once clients have received the key, `alice` will be able to protect message she send, and unprotect messages from `bob`. Also, `bob` can protect messages he send, and unprotect `alice` messages.


Let's start by booting up our 2 clients, so they are listening on their topics:
```bash
# In a first terminal, start Alice client:
go run e4demo.go  -client alice -password alice-super-secret-password
# And in another, start Bob client:
go run e4demo.go  -client bob -password bob-super-secret-password
```

Now, let's write another small `initKeys.go` script to send SetTopicKey commands to our clients:

```text
touch ./initKeys.go
```

```go
package main

import (
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/teserakt-io/e4go"
	e4crypto "github.com/teserakt-io/e4go/crypto"
)

func main() {
	// Generate a key for the topic
	messageTopicKey := e4crypto.RandomKey()

	// Create Alice and Bob keys from their passwords
	aliceKey, err := e4crypto.DeriveSymKey("super-secret-alice-password")
	if err != nil {
		panic(fmt.Sprintf("failed to derivate alice key from password: %v", err))
	}
	bobKey, err := e4crypto.DeriveSymKey("super-secret-bob-password")
	if err != nil {
		panic(fmt.Sprintf("failed to derivate bob key from password: %v", err))
	}

	// Create a E4 command to set the topic key:
	setTopicKeyCmd, err := e4go.CmdSetTopicKey(messageTopicKey, "/e4go/demo/messages")
	if err != nil {
		panic(fmt.Sprintf("failed to create setTopicKeyCmd: %v", err))
	}

	// Connect to MQTT broker
	opts := mqtt.NewClientOptions()
	opts.AddBroker("mqtt.teserakt.io:1883")
	opts.SetCleanSession(true)

	mqttClient := mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.WaitTimeout(time.Second) && token.Error() != nil {
		panic(fmt.Sprintf("failed to connect to mqtt broker: %v", token.Error()))
	}

	// Protect and send the command to our 2 clients via MQTT
	if err := protectAndSendCommand(mqttClient, "alice", aliceKey, setTopicKeyCmd); err != nil {
		panic(fmt.Sprintf("failed to protect command: %v", err))
	}
	if err := protectAndSendCommand(mqttClient, "bob", bobKey, setTopicKeyCmd); err != nil {
		panic(fmt.Sprintf("failed to protect command: %v", err))
	}

	fmt.Println("Topic keys have been set!")
}

func protectAndSendCommand(mqttClient mqtt.Client, clientName string, clientKey, command []byte) error {
	protectedCommand, err := e4crypto.ProtectSymKey(command, clientKey)
	if err != nil {
		return fmt.Errorf("failed to protect command: %v", err)
	}

	clientReceivingTopic := e4go.TopicForID(e4crypto.HashIDAlias(clientName))
	token := mqttClient.Publish(clientReceivingTopic, 2, true, protectedCommand)
	if !token.WaitTimeout(time.Second) {
		return fmt.Errorf("failed to publish command: %v", token.Error())
	}

	return nil
}
```

[Click here to download the full source of this script](../initKeys-step3.go)

Let's run this script:
```bash
$ go run initKeys.go
Topic key have been set for alice!
Topic key have been set for bob!
```

And we can observe on the client sides:
```bash
# Alice:
< received raw message on e4/a7dcef9aef26202fce82a7c7d6672afb: ǎr]�z{�����ʣ_�v�����^����m��>�Cｃs����U3$�˥�T���]�sʁ>�D�}
< unprotected message:
# Bob:
< received raw message on e4/b5d577dc9ce59725e29886632e69ecdf: Ȏr]s��
�Vu�3�#%
       A������7����'�������a�I���KZ�
< unprotected message:
```

We see raw binary being printed (our encrypted commands) and nothing in the unprotected message. This is good, it means our E4 clients have processed the commands successfully (otherwise an error would have been returned).

We're now ready to exchange messages!
```bash
# In alice's client:
Hello, I am alice and this is a secret message for bob!
> message published successfully

# And see in bob client:
< received raw message on /e4go/demo/messages: Z�r]�w�f�>�`��~$1��6���l���_�a��ւX�x��ES��%�����V6��uҲ+�z�����
< unprotected message: Hello, I'm alice and this is a secret message for bob!
```

It works! Now let's repeat our experiment with `eve`, trying to intercept messages from `alice`:
```bash
$ go run e4demo.go  -client eve -password super-secret-eve-password
connected to mqtt.teserakt.io:1883
> subscribed to MQTT topic /e4go/demo/messages
> type anything and press enter to send the message to /e4go/demo/messages:
< received raw message on /e4go/demo/messages: ûö]îªò(ð\þ¿lYj:Kò"Ó
failed to unprotect message: topic key not found
```

All good, unauthorized clients cannot unprotect the messages and read their content. `alice` and `bob` can now exchange private messages.
Now, feel free to authorize `eve` by sending to its receiving topic a the setTopicKeyCommand!

We'll do it in the next section anyway, to simulate `eve` having managed to steal the topic key. And we'll see another way of using the E4 client to protect even further our messages.
