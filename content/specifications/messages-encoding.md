---
title: Message encoding
weight: 5
---

* [Client and C2 messages encoding [SK mode]](#client-and-c2-messages-encoding-sk-mode)
* [Client messages encoding [PK mode]](#client-messages-encoding-pk-mode)
* [C2 messages encoding [PK mode]](#c2-messages-encoding-pk-mode)

This section describes the format of protected messages and how they are (de)serialized:

* Messages exchanged between E4 clients under an arbitrary topic for which keys have been provisioned.

* Messages sent from C2 to clients. These messages contain a command whose format is specified below.

The QoS of messages sent by clients is defined by the application, and does not affect E4's encryption/decryption operations, which are stateless.
However, messages sent by C2 (that is, commands) should be sent with QoS 2, to ensure that a command is delivered exactly once.

## Client and C2 messages encoding [SK mode]

Given an `n`-byte clear message, a protected message has the following format, with byte lengths in parentheses:

```
Timestamp (8) | IV (16) | Ciphertext (n)
```

where:

* `Timestamp` is current time in seconds (since Unix epoch) little-endian encoded used for replay protection.

* `SIV` is the initial value computed during AES-SIV encryption, taking the timestamp as associated data,

* `Ciphertext` is the encrypted message, of same length as the plaintext.

The overhead is therefore of 24 bytes per message.

When C2 sends a message, the plaintext is a command aimed for a specific client, and the message is sent to the client's control topic.

Whereas clients only have control on the payload sent in an MQTT packet, C2 can also control the MQTT parameters and in particular QoS.

## Client messages encoding [PK mode]

Compared to SK mode, in PK mode, the `protect_message()` function signs the message using Ed25519:

The protected message format is thus augmented with

* `id`, the identifier of the sender (since a protocol such as MQTT does not include the sender ID with the message).

* `sig`, the 64-byte signature of the ciphertext.

```
Timestamp (8) | id (16) | sig (64) | IV (16) | Ciphertext (n)
```

The overhead is therefore of 104 bytes per message.

`unprotect()` will then verify the signature if the receiver knows the public key of the sender, identified from `id`.
Depending on the application, the receiver may either discard the message or process it if the signing key is not known.
When encrypting the message, the additional data is the timestamp.


## C2 messages encoding [PK mode]

To protect C2 commands sent to client devices, we used public-key authenticated encryption similar to what NacCl's `crypto_box` [is doing](https://nacl.cr.yp.to/box.html), namely:

* Multiply the sender's private scalar with the recipient's public point, following the X25519 definition

* Hash the resulting point coordinates fixed-length encoding with SHA-3-256

* Use the 256-bit result as a symmetric key with the specified authenticated encryption logic.

The protected message format can be left unchanged compared to the SK mode, since the client will already know that the message is coming from C2, and therefore will know what key to used.

Note that the client precompute and store the shared key, since it will not change unless C2 changes its key pair.
