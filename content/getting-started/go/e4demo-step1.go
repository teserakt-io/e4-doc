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
