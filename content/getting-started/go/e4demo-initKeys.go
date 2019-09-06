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
	setAliceTopicKeyCmd = append(setAliceTopicKeyCmd, e4crypto.HashTopic("/e4go/demo/alice/messages")...)

	setBobTopicKeyCmd := []byte{e4c.SetTopicKey.ToByte()}
	setBobTopicKeyCmd = append(setBobTopicKeyCmd, bobTopicKey...)
	setBobTopicKeyCmd = append(setBobTopicKeyCmd, e4crypto.HashTopic("/e4go/demo/bob/messages")...)

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
