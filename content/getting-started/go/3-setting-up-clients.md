---
title: "3) Setting up E4 clients"
date: "2019-09-06"
lastmod: "2019-09-06"
draft: false
---

Previously, we've updated our application to integrate E4 library, and protect and unprotect the exchanged messages. But we could not communicate yet, since the clients didn't hold any keys necessary to encrypt or decrypt the messages. We'll fix this now.

E4 clients can receive commands, meant to update their internal state, like the list of topic keys they can uses. So to fix our issue, we'll need to:

* generate 2 topic keys for /e4demo/alice/messages and /e4demo/bob/messages
* send thoses keys to each clients, on their respective E4 receiving topics

Once clients have received the keys, Alice will be able to protect message she send, and unprotect messages from Bob, and Bob can protect messages he send, and unprotect Alice's messages.


Let's start by booting up our 2 clients, so they are listening on their topics:
```bash
# In a first terminal, start Alice client:
go run e4demo.go  -client alice -peer bob -password alice-super-secret-password
# And in another, start Bob client:
go run e4demo.go  -client bob -peer alice -password bob-super-secret-password
```

Now, let's write another small `initKey.go` script to send SetTopicKey commands to our clients:

```go
package main

import (
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	e4c "gitlab.com/teserakt/e4common"
	e4crypto "gitlab.com/teserakt/e4common/crypto"
)

func main() {
	// Generate alice and Bob's topic keys
	aliceTopicKey := e4crypto.RandomKey()
	bobTopicKey := e4crypto.RandomKey()

	// Create Alice and Bob keys from their passwords
	aliceKey, err := e4crypto.DeriveSymKey("alice-super-secret-password")
	if err != nil {
		panic(fmt.Sprintf("failed to derivate alice key from password: %v", err))
	}
	bobKey, err := e4crypto.DeriveSymKey("bob-super-secret-password")
	if err != nil {
		panic(fmt.Sprintf("failed to derivate bob key from password: %v", err))
	}

	// Create commands:
	setAliceTopicKeyCmd := []byte{e4c.SetTopicKey.ToByte()}
	setAliceTopicKeyCmd = append(setAliceTopicKeyCmd, aliceTopicKey...)
	setAliceTopicKeyCmd = append(setAliceTopicKeyCmd, e4crypto.HashTopic("/e4demo/alice/messages")...)

	setBobTopicKeyCmd := []byte{e4c.SetTopicKey.ToByte()}
	setBobTopicKeyCmd = append(setBobTopicKeyCmd, bobTopicKey...)
	setBobTopicKeyCmd = append(setBobTopicKeyCmd, e4crypto.HashTopic("/e4demo/bob/messages")...)

	// Connect to MQTT broker
	opts := mqtt.NewClientOptions()
	opts.AddBroker("mqtt.teserakt.io:1883")
	opts.SetCleanSession(true)

	mqttClient := mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.WaitTimeout(time.Second) && token.Error() != nil {
		panic(fmt.Sprintf("failed to connect to mqtt broker: %v", token.Error()))
	}

	clientKeys := map[string][]byte{
		"alice": aliceKey,
		"bob":   bobKey,
	}

	// Protect and send the 2 commands to our 2 clients via MQTT
	for client, key := range clientKeys {
		if err := protectAndSendCommand(mqttClient, client, key, setAliceTopicKeyCmd); err != nil {
			panic(fmt.Sprintf("failed to protect command: %v", err))
		}
		if err := protectAndSendCommand(mqttClient, client, key, setBobTopicKeyCmd); err != nil {
			panic(fmt.Sprintf("failed to protect command: %v", err))
		}
	}

	fmt.Println("TopicKeys have been set!")
}

func protectAndSendCommand(mqttClient mqtt.Client, clientName string, clientKey, command []byte) error {
	protectedCommand, err := e4crypto.ProtectSymKey(command, clientKey)
	if err != nil {
		return fmt.Errorf("failed to protect command: %v", err)
	}

	clientReceivingTopic := e4c.TopicForID(e4crypto.HashIDAlias(clientName))
	token := mqttClient.Publish(clientReceivingTopic, 2, true, protectedCommand)
	if !token.WaitTimeout(time.Second) {
		return fmt.Errorf("failed to publish command: %v", token.Error())
	}

	return nil
}
```

[Click here to download the full source of this script](../e4demo-initKeys.go)

Let's run this script:
```bash
 $ go run initKeys.go
TopicKeys have been set!
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
< received raw message on /e4go/demo/alice/messages: Z�r]�w�f�>�`��~$1��6���l���_�a��ւX�x��ES��%�����V6��uҲ+�z�����
< unprotected message: Hello, I'm alice and this is a secret message for bob!
```

It works! Now let's repeat our experiment with Eve, trying to intercept messages from Alice:
```bash
$ go run e4demo.go  -client eve -peer alice -password eve-super-secret-password
connected to mqtt.teserakt.io:1883
subscribed to peer topic /e4go/demo/alice/messages
type anything and press enter to publish a message on to /e4go/demo/eve/messages:
< receive raw message on /e4go/demo/alice/messages: Z�r]�w�f�>�`��~$1��6���l���_�a��ւX�x��ES��%�����V6��uҲ+�z�����
failed to unprotect message: topic key not found
```

All good, unauthorized clients cannot unprotect the messages and read their content. Alice and Bob can now exchange private messages.
Now, feel free to authorize Eve by sending to its client the aliceTopicKey!
