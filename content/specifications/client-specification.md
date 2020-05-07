---
title: Client specification
weight: 3
---

* [Client internal state](#client-internal-state)
* [Client local API (rx)](#client-local-api-rx)
    * [protect_message(message, topic)](#protect-messagemessage-topic)
    * [unprotect(message, topic)](#unprotectmessage-topic)
* [Client local API (tx)](#client-local-api-tx)
* [Client remote API (rx)](#client-remote-api-rx)
    * [RemoveTopic(topichash)](#removetopictopichash)
    * [ResetTopics()](#resettopics)
    * [SetIdKey(key)](#setidkeykey)
    * [SetTopicKey(topichash, key)](#settopickeytopichash-key)
    * [RemovePubKey(id) [PK mode]](#removepubkeyid-pk-mode)
    * [ResetPubKeys() [PK mode]](#resetpubkeys-pk-mode)
    * [SetPubKey(id, pubkey) [PK mode]](#setpubkeyid-pubkey-pk-mode)
    * [SetC2Key(pubkey) [PK mode]](#setc2keypubkey-pk-mode)
* [Client remote API (tx)](#client-remote-api-tx)
* [Key transition logic](#key-transition-logic)

E4 clients are for example sensors devices, vehicles, mobile phones, backend servers or any type of device that participates in the communication.

In order to receive commands from the C2, a client's MQTT service must subscribe to the topic `E4/<id>`, with their own id in hexadecimal (lower-case).

## Client internal state

E4 clients have a state composed of:

* `id`, a 16-byte unique identifier (constant).

* `topickeys`, a table mapping topic hashes (as 32-byte strings) to 32-byte keys, initially either empty or pre-filled.

* A key unique per `id`:

  - **[SK mode]**: `key`, a 32-byte symmetric key, shared with C2.
  - **[PK mode]**: `key`, a 32-byte Ed25519 private key, whose corresponding public key is shared with C2.

* **[PK mode]**: `clientkeys`, a table mapping client ids to their respective Ed25519 public key.

* **[PK mode]**: `c2key`, the 32-byte Curve25519 public key of C2.

Note that `topickeys` lists topic hashes rather than topic names.
This is in order to work with fixed-length values rather than variable-size, potentially long (up to 65535 bytes) topic names.
This also minimizes the length of commands transmitted by C2.

## Client local API (rx)

Clients expose the following APIs to the client application:

### `protect_message(message, topic)`

* *When*: For every message to be transmitted over MQTT by the application, under a given topic.

* *Returns*: A protected message, or a `TopicKeyMissing` error.

* *How*: If a key exists in `topickeys` for the hash of the given topic, then the message is protected using this key and the protected message is returned.

The protected message is always slightly longer than the plaintext, because it includes the initial value, a timestamp, and in **[PK mode]** a signature.

If a `TopicKeyMissing` error is returned, by default the message is discarded (strict encryption).
But an application can change this policy if the business logic requires it, and choose to send unprotected messages when no topic key is known.


### `unprotect(message, topic)`

* *When*: For every payload received under a topic that is not `E4/<id>`, that is, all messages except commands from C2.

* *Returns*: A message or an error in `InvalidProtectedLen`, `InvalidSignature`, `NotAuthentic`, `TimestampInFuture`, `TimestampTooOld`,  `TooShortCiphertext`, `TopicKeyMissing`.

* *How*: `unprotect()` handles both application messages and control messages:
  - If the topic is *not* `E4/<id>`: In **[PK mode]**, the signature of the message is first verified. Then, for both modes: If a key exists for the hash of given topic, then the message is verified using this key. Verification includes verification of the timestamp, then decryption and verification of the tag. Upon successful verification, the unprotected message is returned to the application, otherwise the message is discarded.
- If the topic is `E4/<id>` (control message): The client verifies that the timestamp and discards the message if it is too old, otherwise the message is decrypted and authenticated using the client's `key` in **[SK mode]** or the Diffie-Hellman shared secret in **[PK mode]**, from the C2's public key and the client's private key.

Verification of the timestamp consists in checking that the timestamp in the message received is no older that a prescribed value (for example, 10 minutes).
The maximal delay authorized will depend on the use case, and is a parameter of E4.

If a `TopicKeyMissing` error is returned, by default the message is discarded (strict encryption).
But an application can change this policy if the business logic requires it, and choose to received the process the message when no topic key is known.

## Client local API (tx)

**TODO**


## Client remote API (rx)

The client remote API is exposed via C2 commands, which are protected using the client's key (symmetric key or public key, depending on the mode).

After being successfully unprotected, the following commands are supported by a client, given in the format `CommandName(parameters)`.
We also provide the encoding of the commands, where each command is prefixed by a 1-byte command identifier.
In the encoding format, byte lengths are given in parentheses.

Additional user specific extensions and commands can be added to the protocol if needed.

### `RemoveTopic(topichash)`

Removes the given topic hash and key from the client's `topickeys` table.

Encoding: `0x00 | topicHash (32)`

### `ResetTopics()`

Empties the client's `topickeys` table.

Encoding: `0x01`.

### `SetIdKey(key)`


Updates the client's key with the key included in the command, overwriting the current value.

In **[SK mode]**, this sets a symmetric key, and in **[PK mode]** this sets a private key.

Encoding (**[SK mode]**): `0x02 | key (32)`.

Encoding (**[PK mode]**): `0x02 | key (64)`.

In **[PK mode]**, this command may be used to rotate client identity
keys when the client's pseudorandom generator is not trusted, and such
that the C2 only keeps the public key, and discards the private key.

### `SetTopicKey(topichash, key)`

Adds the given topic hash and key to the client's `topickeys` table, erasing any previous entry.

Encoding: `0x03 | key (32) | topichash (32)`

Here `topichash` is a 32-byte hash of the topic.
This ensures that the commands are of short, constant size.

### `RemovePubKey(id)` [PK mode]

Removes the given client id and associated public key from the client's `clientkeys` table.

Encoding: `0x04 | id (16)`

### `ResetPubKeys()` [PK mode]

Empties the client's `clientkeys`  table.

Encoding: `0x05`

### `SetPubKey(id, pubkey)` [PK mode]

Adds the given id and public key to the client's `clientkeys` table, erasing any previous entry.

Encoding: `0x06 | id (16) | pubkey (32) )`

### `SetC2Key(pubkey)` [PK mode]

Instructs the device to replace the current C2 public key with the newly transmitted one (must be use with care).

Encoding: `0x07 | pubkey (32)`

This command should be use with great care, and in such a way that devices are unlikely to miss the message. The new C2 key should also be recoverable at any time, for example through instantaneous backups of the private key generated.

## Client remote API (tx)

**TODO**

## Key transition logic

When a new topic key is sent to multiple devices, not all of them will received it simultaneously, because some might be offline, different network latency, or unreceived C2 messages.

Risks are therefore that the first devices to switch to the new key be unable to decrypt messages sent with the previous key, and that devices switching late to the new key be unable to decrypt messages sent with new key before they received it.

As a mitigation, E4 clients use a *key transition* mechanism, whereby devices switch to a new key while keeping the old key during a given transition period.

The client must then behave as follows:

When receiving a `SetTopicKey` command, the previous key followed by its encoded timestamp is added to the client's `topickey` under the hash of the topic's hash, erasing any previous entry.

When receiving a message under the given topic, the device attempts to decrypt it with the current key, and if decryption fails it tries with the old key if and only if 1) it exists 2) the timestamp is not older than a given transition period (unless the device has no reliable clock, in which case it ignores the timestamp).

This mechanism guarantees that if a device that  received the new key gets a message encrypted with the old one,  it will still be able to unprotect it.

If a client  receives a `SetTopicKey` command including the same key as the current one, the `topickey` table is not modified, and therefore the client retains the previous (distinct) key.

***
With the above strategy, we may want to prioritize the key transmission to devices receiving messages first, over devices emitting messages to minimize decryption errors due to invalid key. We can define 3 devices *type*:
 - Publishers: only send messages
 - Consumers: only receive messages
 - Publishers & Consumers: send and receive messages

Adding this type to every devices in C2 would allow to dispatch a new key in this order:
   Consumers -> Publishers & Consumers -> Publishers

This could prevent most of decryption errors by not having the right key, except in scenarios having multiple *Publishers & Consumers* devices communicating together.
***
