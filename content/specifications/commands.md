---
title: Commands
weight: 1
---

## What are E4 commands ?

Commands are protected messages (understand: encrypted sequence of bytes), which will, once passed to a client `Unprotect` method, update the client state.
They can be distinguished from the other regular messages because they are exchanged on a client `ReceivingTopic`, which is a unique topic, created from a hash of the client ID, so each clients can have their own. Messages on those topics must be protected using the client private key to ensure legitimate key holders are the only ones able to send commands to clients.

## Creating a command

Creating a command is easy. A command is composed of bytes like so:
* A command identifier (indicating the type of command, and determining what parameters are expected)
* A concatenation of fixed length parameters (for command specific parameters, please refer to the available commands list below)

Once the command identifier and the parameters bytes are concatenated, the command is then protected, using the client's private key (refer to your E4 client library implementation for helper functions). This command is then ready to be sent on the client `ReceivingTopic`.

## Available commands

### Managing client topics

#### SetTopicKey

Set a topic and its private key on the client.

| Name | Length (bytes) | Description |
| --- | ---: | --- |
| Command ID | 1 | must be set to `0x3` |
| TopicKey | 32 | The topic private key |
| TopicHash | 16 | A sha3-256sum hash of the topic |


#### RemoveTopic

Removes a topic and its private key from the client.

| Name | Length (bytes) | Description |
| --- | ---: | --- |
| Command ID | 1 | must be set to `0x0` |
| TopicHash | 16 | A sha3-256sum hash of the topic |


#### ResetTopics

Removes all topics and their private keys from the client.

| Name | Length (bytes) | Description |
| --- | ---: | --- |
| Command ID | 1 | must be set to `0x1` |


### Managing client private key

#### SetIDKey

Set the client private key. The old key will be replaced, so next commands after this one will need to be protected using the updated key.

| Name | Length (bytes) | Description |
| --- | ---:| --- |
| Command ID | 1 | must be set to `0x2` |
| ClientKey | 32 | The new client private key |

### Managing client public keys

The public key commands are not available on every clients implementation. For instance, the SymKeyClient does not support them.é

#### SetPubKey

Add the public key of another client on the current client.

| Name | Length (bytes) | Description |
| --- | ---:| --- |
| Command ID | 1 | must be set to `0x6` |
| ClientPubKey | 32 | A client public key |
| ClientID | 16 | The client ID to set the public key for  |

#### RemovePubKey

Removes the public key of the given client.

| Name | Length (bytes) | Description |
| --- | ---:| --- |
| Command ID | 1 | must be set to `0x4` |
| ClientID | 16 | The client ID to set the public key for  |

#### ResetPubKeys

Removes all public keys stored on the client.

| Name | Length (bytes) | Description |
| --- | ---:| --- |
| Command ID | 1 | must be set to `0x5` |
