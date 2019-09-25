package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/teserakt-io/e4go"
)

func main() {
	// 1. Read a client and a peer identifiers from command line flags
	var clientName string
	var clientPassword string
	var peerName string
	flag.StringVar(&clientName, "client", "", "the client name")
	flag.StringVar(&clientPassword, "password", "", "the client password")
	flag.StringVar(&peerName, "peer", "", "the peer name to read messages from")
	flag.Parse()

	if len(clientName) == 0 {
		panic("-client is required")
	}
	if len(clientPassword) < 16 {
		panic("-password is required and must contains at least 16 characters")
	}
	if len(peerName) == 0 {
		panic("-peer is required")
	}

	// 2. Connect to a MQTT broker (we'll use our public mqtt.teserakt.io:1338)
	brokerEndpoint := "mqtt.teserakt.io:1883"
	mqttClient, err := initMQTT(brokerEndpoint, clientName)
	if err != nil {
		panic(fmt.Sprintf("failed to init mqtt client: %v\n", err))
	}
	fmt.Printf("> connected to %s\n", brokerEndpoint)

	e4Client, err := e4go.NewSymKeyClientPretty(clientName, clientPassword, fmt.Sprintf("%s.json", clientName))
	if err != nil {
		panic(fmt.Sprintf("failed to create E4 client: %v\n", err))
	}

	// 3. Subscribe to the peer MQTT topic /e4go/demo/<peerID>/messages and print any incoming messages to stdout
	peerTopic := fmt.Sprintf("/e4go/demo/%s/messages", peerName)
	topics := map[string]byte{
		peerTopic:                    1,
		e4Client.GetReceivingTopic(): 2,
	}
	token := mqttClient.SubscribeMultiple(topics, func(_ mqtt.Client, msg mqtt.Message) {
		fmt.Printf("< received raw message on %s: %s\n", msg.Topic(), msg.Payload())
		clearMessage, err := e4Client.Unprotect(msg.Payload(), msg.Topic())
		if err != nil {
			fmt.Printf("failed to unprotect message: %v\n", err)
			return
		}

		fmt.Printf("< unprotected message: %s\n", clearMessage)
	})
	if !token.WaitTimeout(1 * time.Second) {
		panic(fmt.Sprintf("failed to mqtt subscribe: %v\n", token.Error()))
	}
	fmt.Printf("> subscribed to peer topic %s\n", peerTopic)

	// 4. Wait for user input on stdin, so user can type in a message and press enter.
	// Messages will then be publish on a MQTT topic /e4go/demo/<clientID>/messages.
	publishTopic := fmt.Sprintf("/e4go/demo/%s/messages", clientName)
	fmt.Printf("> type anything and press enter to publish a message on to %s:\n", publishTopic)
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		message := scanner.Text()
		if len(message) == 0 { // Skip empty messages
			continue
		}

		protectedMessage, err := e4Client.ProtectMessage([]byte(message), publishTopic)
		if err != nil {
			fmt.Printf("failed to protect message: %v\n", err)
			continue
		}
		if token := mqttClient.Publish(publishTopic, 1, true, protectedMessage); token.Error() != nil {
			fmt.Printf("failed to publish message: %v\n", token.Error())
			continue
		}

		fmt.Println("> message published successfully")
	}
}

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
