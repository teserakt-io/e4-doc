package main

import (
	"fmt"
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
	opts.AddBroker("mqtt.teserakt.io:1883")
	opts.SetCleanSession(true)

	mqttClient := mqtt.NewClient(opts)
	timeout := time.Second
	if token := mqttClient.Connect(); token.WaitTimeout(timeout) && token.Error() != nil {
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

	clientReceivingTopic := e4.TopicForID(e4crypto.HashIDAlias(clientName))
	token := mqttClient.Publish(clientReceivingTopic, 2, true, protectedCommand)
	timeout := time.Second
	if !token.WaitTimeout(timeout) {
		return fmt.Errorf("failed to publish command: %v", token.Error())
	}

	return nil
}
