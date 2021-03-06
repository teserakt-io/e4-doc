---
title: "Going further"
date: "2019-09-25"
lastmod: "2019-09-27"
draft: false
weight: 4
---

At this point, we have a working solution to secure communication over a shared topic, where all topic's members can send and receive encrypted messages from every other members. But E4 also allow us to explore another scenario, where we can choose who a client is authorized to receive messages from. Let's imagine `eve` have been given the topic key so she can send and receive messages, but she's not allowed to send messages to `alice` and `bob`.

We'll go through the following steps, modifying our previous application to achieve it:

1. Generate an admin (for us) public / private key pair
2. Generate public/private key pairs for `alice`, `bob` and `eve`
3. Switch the E4 SymClient to a PubKeyClient
4. Update our previous `init/initKey.go` to
 - protect commands using our new admin key
 - share the topic key to `eve`
 - send a new command to `alice`, to set  `bob` public key on its client
 - send a new command to `bob`, to set  `alice` public key on its client

Let's get started!

First, let's use E4 key generator for all our key generations. We'll generate Ed25519 keys, and write the resulting public and private keys to a file:

```text
# Install the E4 keygen:
$ go install github.com/teserakt-io/e4go/cmd/e4keygen
$ e4keygen -type ed25519 -out ./admin
private key successfully written at ./admin
public key successfully written at ./admin.pub
```

Since we're at it, let's also generate other keys for `alice`, `bob` and `eve`, we'll use them in a moment:
```text
$ e4keygen -type ed25519 -out ./alice
private key successfully written at ./alice
public key successfully written at ./alice.pub
$ e4keygen -type ed25519 -out ./bob
private key successfully written at ./bob
public key successfully written at ./bob.pub
$ e4keygen -type ed25519 -out ./eve
private key successfully written at ./eve
public key successfully written at ./eve.pub
```

Now that we have keys for everyone, let's update our application.
We start by adding 2 new helper functions to load our previous keys in `e4demo.go`:

```go
func loadPublicKey(name string) ed25519.PublicKey {
	pubKey, err := ioutil.ReadFile(fmt.Sprintf("%s.pub", name))
	if err != nil {
		panic(fmt.Sprintf("failed to load key %s: %v", name, err))
	}

	return ed25519.PublicKey(pubKey)
}

func loadPrivateKey(name string) ed25519.PrivateKey {
	privKey, err := ioutil.ReadFile(name)
	if err != nil {
		panic(fmt.Sprintf("failed to load key %s: %v", name, err))
	}

	return ed25519.PrivateKey(privKey)
}
```

Then we continue by removing the now useless `clientPassword` flag, and switch our `SymNameAndPassword` config to the new `PubIDAndKey`. Notice how the pubKeyClient is created given the `admin` public key (after converting it to a curve25519 key), such as it could authenticate the commands we'll send later.

{{<tabs>}}
{{<tab after>}}
{{<highlight go>}}
func main() {
	var clientName string
	flag.StringVar(&clientName, "client", "", "the client name")
	flag.Parse()

	if len(clientName) == 0 {
		fmt.Println("-client is required")
		os.Exit(1)
	}

	adminPubCurveKey := e4crypto.PublicEd25519KeyToCurve25519(loadPublicKey("admin"))
	e4Client, err := e4.NewClient(&e4.PubIDAndKey{
		ID:       e4crypto.HashIDAlias(clientName),
		Key:      loadPrivateKey(clientName),
		C2PubKey: adminPubCurveKey[:],
	}, e4.NewInMemoryStore(nil))
	// ...
{{</highlight>}}
{{</tab>}}
{{<tab before>}}
{{<highlight go>}}
func main() {
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
		panic("-password is required and must contains at least 16 characters")
	}

	e4Client, err := e4.NewClient(&e4.SymNameAndPassword{Name: clientName, Password: clientPassword}, e4.NewInMemoryStore(nil))
	// ...
{{</highlight>}}
{{</tab>}}
{{</tabs>}}

[Click here to download the full source of this script](../e4demo-step4.go)

And that's all we need!

Let's now modify then `init/initKeys.go` script to protect and send the commands using the new keys.
We'll start by adding 3 helpers, reusing our previous 2 key loading functions, and a new `pubKeyProtectAndSendCommand`. We also comment out the `protectAndSendCommand` function as we'll not need it anymore:

{{<tabs>}}
{{<tab after>}}
{{<highlight go>}}
func pubKeyProtectAndSendCommand(mqttClient mqtt.Client, clientName string, command []byte) error {
	// Load ed25519 keys, and convert them to curve25519 keys
	clientPublicCurveKey := e4crypto.PublicEd25519KeyToCurve25519(loadPublicKey(clientName))
	adminPrivateCurveKey := e4crypto.PrivateEd25519KeyToCurve25519(loadPrivateKey("admin"))

	shared, err := curve25519.X25519(adminPrivateCurveKey, clientPublicCurveKey)
	if err != nil {
		return fmt.Errorf("curve25519 X25519 failed: %v", err)
	}

	protectedCommand, err := e4crypto.ProtectSymKey(command, e4crypto.Sha3Sum256(shared))
	if err != nil {
		return fmt.Errorf("failed to protect command: %v", err)
	}

	clientReceivingTopic := e4.TopicForID(e4crypto.HashIDAlias(clientName))
	token := mqttClient.Publish(clientReceivingTopic, 2, true, protectedCommand)
	if !token.WaitTimeout(time.Second) {
		return fmt.Errorf("failed to publish command: %v", token.Error())
	}

	return nil
}

func loadPublicKey(name string) ed25519.PublicKey {
	pubKey, err := ioutil.ReadFile(fmt.Sprintf("%s.pub", name))
	if err != nil {
		panic(fmt.Sprintf("failed to load key %s: %v", name, err))
	}

	return ed25519.PublicKey(pubKey)
}

func loadPrivateKey(name string) ed25519.PrivateKey {
	privKey, err := ioutil.ReadFile(name)
	if err != nil {
		panic(fmt.Sprintf("failed to load key %s: %v", name, err))
	}

	return ed25519.PrivateKey(privKey)
}
{{</highlight>}}
{{</tab>}}
{{<tab before>}}
{{<highlight go>}}
func protectAndSendCommand(mqttClient mqtt.Client, clientName string, clientKey, command []byte) error {
	protectedCommand, err := e4crypto.ProtectSymKey(command, clientKey)
	if err != nil {
		return fmt.Errorf("failed to protect command: %v", err)
	}

	clientReceivingTopic := e4.TopicForID(e4crypto.HashIDAlias(clientName))
	token := mqttClient.Publish(clientReceivingTopic, 2, true, protectedCommand)
	if !token.WaitTimeout(time.Second) {
		return fmt.Errorf("failed to publish command: %v", token.Error())
	}

	return nil
}
{{</highlight>}}
{{</tab>}}
{{</tabs>}}

Next, we update the main function to create the commands and send them over mqtt, as we did before. But this time we'll give the topic key to `alice`, `bob` and `eve`, and add an extra command to give `bob` public key to `alice`, and `alice` public key to `bob`:

{{<tabs>}}
{{<tab after>}}
{{<highlight go>}}
package main

import (
	"fmt"
	"io/ioutil"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"golang.org/x/crypto/ed25519"

	e4 "github.com/teserakt-io/e4go"
	e4crypto "github.com/teserakt-io/e4go/crypto"
)

func main() {
	// Generate a key for the topic
	messageTopicKey := e4crypto.RandomKey()

	// Create a E4 command to set the topic key:
	setTopicKeyCmd, err := e4.CmdSetTopicKey(messageTopicKey, "/e4go/demo/messages")
	if err != nil {
		panic(fmt.Sprintf("failed to create setTopicKeyCmd: %v", err))
	}

	// Connect to MQTT broker
	opts := mqtt.NewClientOptions()
	opts.AddBroker("mqtt.eclipse.org:1338")
	opts.SetCleanSession(true)

	mqttClient := mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.WaitTimeout(time.Second) && token.Error() != nil {
		panic(fmt.Sprintf("failed to connect to mqtt broker: %v", token.Error()))
	}

	// Protect and send alice's topic key to our 3 clients:
	for _, client := range []string{"alice", "bob", "eve"} {
		if err := pubKeyProtectAndSendCommand(mqttClient, client, setTopicKeyCmd); err != nil {
			panic(fmt.Sprintf("failed to protect and send command: %v", err))
		}
		fmt.Printf("topic key have been set for %s!\n", client)
	}

	// Now gives bob's public key to alice
	setBobPubKeyCmd, err := e4.CmdSetPubKey(loadPublicKey("bob"), "bob")
	if err != nil {
		panic(fmt.Sprintf("failed to create setBobPubKeyCmd: %v", err))
	}
	if err := pubKeyProtectAndSendCommand(mqttClient, "alice", setBobPubKeyCmd); err != nil {
		panic(fmt.Sprintf("failed to send bob's public key to alice: %v", err))
	}
	fmt.Println("alice now have bob's public key!")

	// And alice's public key to bob
	setAlicePubKeyCmd, err := e4.CmdSetPubKey(loadPublicKey("alice"), "alice")
	if err != nil {
		panic(fmt.Sprintf("failed to create setAlicePubKeyCmd: %v", err))
	}
	if err := pubKeyProtectAndSendCommand(mqttClient, "bob", setAlicePubKeyCmd); err != nil {
		panic(fmt.Sprintf("failed to send alice's public key to bob: %v", err))
	}
	fmt.Println("bob now have alice's public key!")
}
{{</highlight>}}
{{</tab>}}
{{<tab before>}}
{{<highlight go>}}
package main

import (
	"fmt"
	"io/ioutil"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	e4 "github.com/teserakt-io/e4go"
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
	setTopicKeyCmd, err := e4.CmdSetTopicKey(messageTopicKey, "/e4go/demo/messages")
	if err != nil {
		panic(fmt.Sprintf("failed to create setTopicKeyCmd: %v", err))
	}

	// Connect to MQTT broker
	opts := mqtt.NewClientOptions()
	opts.AddBroker("mqtt.eclipse.org:1338")
	opts.SetCleanSession(true)

	mqttClient := mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.WaitTimeout(time.Second) && token.Error() != nil {
		panic(fmt.Sprintf("failed to connect to mqtt broker: %v", token.Error()))
	}

	// Protect and send the command to our 2 clients via MQTT
	if err := protectAndSendCommand(mqttClient, "alice", aliceKey, setTopicKeyCmd); err != nil {
		panic(fmt.Sprintf("failed to protect command: %v", err))
	}
	fmt.Println("Topic key have been set for alice!")

	if err := protectAndSendCommand(mqttClient, "bob", bobKey, setTopicKeyCmd); err != nil {
		panic(fmt.Sprintf("failed to protect command: %v", err))
	}
	fmt.Println("Topic key have been set for bob!")
}
{{</highlight>}}
{{</tab>}}
{{</tabs>}}

[Click here to download the full source of this script](../initKeys-step4.go)

Now we're all set! Let's try it out.

First, we open 3 terminal and start our clients:
```text
# Alice
$ go run e4demo.go -client alice
connected to mqtt.eclipse.org:1338
> subscribed to MQTT topic /e4go/demo/messages
> type anything and press enter to send the message to /e4go/demo/messages:

# Bob
$ go run e4demo.go -client bob
connected to mqtt.eclipse.org:1338
> subscribed to MQTT topic /e4go/demo/messages
> type anything and press enter to send the message to /e4go/demo/messages:

# Eve
$ go run e4demo.go -client eve
connected to mqtt.eclipse.org:1338
> subscribed to MQTT topic /e4go/demo/messages
> type anything and press enter to send the message to /e4go/demo/messages:
```

And in another terminal, run the `initKey.go` script:

```text
$ go run initKeys.go
topic key have been set for alice!
topic key have been set for bob!
topic key have been set for eve!
alice now have bob's public key!
bob now have alice's public key!
```

`alice` and `bob` clients have received 2 messages, topic key in first, and the public key in the second:
```text
< received raw message on e4/a7dcef9aef26202fce82a7c7d6672afb: <raw bytes>
< unprotected message:
< received raw message on e4/a7dcef9aef26202fce82a7c7d6672afb: <raw bytes>
< unprotected message:
```

And `eve` only one, the topic key:
```text
< received raw message on e4/fcf4908516e5f2aa8d07a01b093fd4ef: <raw bytes>
< unprotected message:
```

From `alice` client, we can now send a message:

```text
# Alice
Hello, I'm alice and this is a secret message for bob!
> message published successfully
```

And observe the result in `bob` and `eve` clients:
```text
# Bob
< received raw message on /e4go/demo/messages: <raw bytes>
< unprotected message:  Hello, I'm alice and this is a secret message for bob!

# Eve
< received raw message on /e4go/demo/messages: <raw bytes>
failed to unprotect message: signer public key not found
```

So `bob` properly receive, unprotect and validate the message he received, but `eve` can't, since she doesn't have the sender public key.
Now let's try to send a message on the topic as `eve`:

```text
# Eve
Hello I'm Eve!
> message published successfully

# From Alice or Bob clients:
< received raw message on /e4go/demo/messages: <raw bytes>
failed to unprotect message: signer public key not found
```

E4 properly discard `eve` messages as neither `alice` or `bob` have been given her public key.

That's it for our E4 introduction with the Go client. If you have any questions, or other use cases which are not covered yet, feel free to open an issue on the [Github tracker](https://github.com/teserakt-io/e4go/issues)
