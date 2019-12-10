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
	opts.AddBroker("mqtt.teserakt.io:1883")
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

func pubKeyProtectAndSendCommand(mqttClient mqtt.Client, clientName string, command []byte) error {
	// Load ed25519 keys, and convert them to curve25519 keys
	clientPublicCurveKey := e4crypto.PublicEd25519KeyToCurve25519(loadPublicKey(clientName))
	adminPrivateCurveKey := e4crypto.PrivateEd25519KeyToCurve25519(loadPrivateKey("admin"))

	protectedCommand, err := e4crypto.ProtectCommandPubKey(command, &clientPublicCurveKey, &adminPrivateCurveKey)
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
