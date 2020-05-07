---
title: C2 specification
weight: 4
---

* [C2 internal state](#c2-internal-state)
* [C2 API](#c2-api)
    * [new_client(id, key) [local]](#new_clientid-key-local)
    * [remove_client(id)  [local]](#remove_clientid--local)
    * [new_topic_client(id, topic) [remote]](#new_topic_clientid-topic-remote)
    * [remove_topic_client(id, topic) [remote]](#remove_topic_clientid-topic-remote)
    * [reset_client(id) [remote]](#reset_clientid-remote)
    * [new_topic(topic) [local]](#new_topictopic-local)
    * [remove_topic(topic) [local]](#remove_topictopic-local)
    * [new_client_key(id) [local &amp; remote, SK mode]](#new_client_keyid-local--remote-sk-mode)
    * [reset_client_pubkeys(id) [remote, PK mode]](#reset_client_pubkeysid-remote-pk-mode)
    * [send_client_pubkey(iddst, idsrc) [remote, PK mode]](#send_client_pubkeyiddst-idsrc-remote-pk-mode)
    * [remove_client_pubkey(iddst, idsrc) [remote, PK mode]](#remove_client_pubkeyiddst-idsrc-remote-pk-mode)
    * [new_c2_key() [local &amp; remote, PK mode]](#new_c2_key-local--remote-pk-mode)
* [C2 human interface](#c2-human-interface)

C2 (command-and-control) is the host that manages clients' keys remotely, by sending commands to clients over MQTT.

E4's C2 holds client keys and topic keys, and manages clients by sending them commands.

## C2 internal state

C2's state is composed of:

* `clientkeys`, a map from client identities (as 16-byte strings) to their keys (symmetric keys or public keys, depending on the mode).

* `topickeys`, a map from topics (as strings) to 32-byte topics keys.

This state is typically stored as a database.

Unlike clients whose actions are triggered by events (commands received, messages going in and out), C2's actions can be triggered by an operator, via a web UI or command-line interface, or by another service via C2's HTTP or gRPC APIs, for example by E4's automation engine.


## C2 API

Operations performed by C2 are of three types

* Operations modifying the local state (local `clientkeys` and `topickeys` tables)

* Operations modifying clients' states (remote `key` value and `topickeys` table), by sending a command as an MQTT messages (with QoS 2).

* Operations modifying both the local state and clients' states

This is just the minimal set of functionalities that the C2 API can implement, via HTTP and/or gRPC interfaces.

Other helper endpoints can be added, as needed by other components (command-line client, automation engine, etc.).

The API might also be extended by having endpoints supporting batch commands (for example, to reset multiple clients with a single request).


### `new_client(id, key)` [local]

Adds `(id, key)` to `clientkeys`; overwrite the existing key if `id` already exists. Depending on the mode, `key` is either a symmetric key or a public key.

### `remove_client(id)`  [local]

Removes `id` from C2's `clientkeys` map; fails if `id` does not exist.

### `new_topic_client(id, topic)` [remote]

Sends the key of `topic` to the `id` using a `SetTopicKey` command; fails if no key is known for `topic` or for `id`.

### `remove_topic_client(id, topic)` [remote]

Removes the given topic from the client's `topickeys` table using a `RemoveTopic` command.

### `reset_client(id)` [remote]

Resets the `topickeys` table of the given client, using a `ResetTopics` command.

### `new_topic(topic)` [local]

A random key is generated for `topic`, erasing any previous entry in `topickeys`. The key is then distributed to all clients that have the key for this topic.

### `remove_topic(topic)` [local]

If the given topic is in `topickeys`, then the topic and its key are removed from it; fails if `topic` does not exist.


### `new_client_key(id)` [local & remote, SK mode]

If `id` exists, random key is generated for `id` and sent to it using a `SetIdKey` command. The `(id, key)` pair is then added to C2's `clientkeys`.

A risk when updating the client key is the client being "in the dark" if it did not receive the  key update.

Here there is no need for the client to retain the old key.

However, C2 may need to retain the previous key, but this only shifts the DoS problem to the next generation of key update, unless C2 can get a confirmation that the client registered the new key.

Depending on the use case and associated risk, we can implement a failure recovery mechanism, whereby clients can recover from desynchronized keys.

### `reset_client_pubkeys(id)` [remote, PK mode]

Resets the `clientkeys` table of the given client, using a `ResetPubKeys` command.


### `send_client_pubkey(iddst, idsrc)` [remote, PK mode]

Sends the public key of `idsrc` to `iddst`, to add it to its `clientkeys` table.

Note that when a client is added to a topic, other devices from this topic don't automatically receive the client's public key. This would be necessary when emulating group messaging (many-to-many communications), but not in many-to-one network. In the former case, public key sharing can be automated through simple scripts. This behavior may be modified according to the users' needs.

### `remove_client_pubkey(iddst, idsrc)` [remote, PK mode]

Remove the public key of `idsrc` to `iddst`, to add it to its `clientkeys` table.

### `new_c2_key()` [local & remote, PK mode]

A random key pair is generated for C2 and its public key is sent to all clients using a `SetC2Key` command. The new key pair is then used by C2 and replaces the previous one.

## C2 human interface

C2 human interfaces can be command-line clients or a web-based GUI, using the API to translate user actions into C2 operations.

To facilitate manual operation, the following user-friendly mechanisms can be implemented:

* `id` can be generated from a string alias, as a hash of the alias

* The key or key pair can be derived from a passphrase, using a password-based key derivation function
