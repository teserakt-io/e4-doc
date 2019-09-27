package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/teserakt-io/e4go"
	e4crypto "github.com/teserakt-io/e4go/crypto"
	"golang.org/x/crypto/ed25519"
)

func main() {
	// 1 - Read a client identifier from a command line flag
	var clientName string
	//var clientPassword string
	flag.StringVar(&clientName, "client", "", "the client name")
	// flag.StringVar(&clientPassword, "password", "", "the client password")
	flag.Parse()

	if len(clientName) == 0 {
		fmt.Println("-client is required")
		os.Exit(1)
	}
	// if len(clientPassword) < 16 {
	// 	panic("-password is required and must contains at least 16 characters")
	// }

	// 2 - Connect to a MQTT broker (we'll use our public mqtt.teserakt.io:1338)
	brokerEndpoint := "mqtt.teserakt.io:1883"
	mqttClient, err := initMQTT(brokerEndpoint, clientName)
	if err != nil {
		fmt.Printf("failed to init mqtt client: %v\n", err)
	}
	fmt.Printf("connected to %s\n", brokerEndpoint)

	// e4Client, err := e4go.NewSymKeyClientPretty(clientName, clientPassword, fmt.Sprintf("%s.json", clientName))
	adminPubCurveKey := e4crypto.PublicEd25519KeyToCurve25519(loadPublicKey("admin"))
	e4Client, err := e4go.NewPubKeyClient(
		e4crypto.HashIDAlias(clientName),
		loadPrivateKey(clientName),
		fmt.Sprintf("%s.json", clientName),
		adminPubCurveKey[:],
	)

	// 3 - Subscribe to message MQTT topic and print incoming messages to stdout
	messageTopic := "/e4go/demo/messages"
	// token := mqttClient.Subscribe(messageTopic, 1, func(_ mqtt.Client, msg mqtt.Message) {
	// 	fmt.Printf("< received raw message on %s: %s\n", msg.Topic(), msg.Payload())
	// })
	topics := map[string]byte{
		messageTopic:                 1,
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

		protectedMessage, err := e4Client.ProtectMessage([]byte(message), messageTopic)
		if err != nil {
			fmt.Printf("> failed to protect message: %v\n", err)
			continue
		}
		if token := mqttClient.Publish(messageTopic, 1, true, protectedMessage); token.Error() != nil {
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
