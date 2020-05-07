# Teserakt E4

Teserakt's E4 solution makes IoT data protection painless thanks to:

* **E4 client library**, which encrypts and decrypts data in IoT devices
* **C2 server**, which manages devices' keys thanks to:
    - **Automation** of key distribution and key rotation
    - **Scalability** to millions of devices and messages
    - **Monitoring** of the data to detect network anomalies

For example, the C2 server can ensure that each device will use

* a first key to decrypt firmware and configuration updates
* a second key to encrypt data sent to the analytics back-end

in a way that these keys are automatically rotated every week in order
to mitigate a potential breach.

C2 can run on-premise or as SaaS and talks HTTP/JSON and
gRPC/protobuf. We provide a CLI interface and a web GUI.

Don't hesitate to <a href="mailto:contact@teserakt.io">contact us</a>
for more information or to try E4.

The client library is open-source and can be used for free without C2,
by following our [Getting started](getting-started) pages.

## The E4 client library

The E4 client library enables lightweight and secure communication once integrated into IoT devices.
It plug itself right between your business logic and the device transport layer, making it easy to integrate.

Requirements

The E4 client library can be used in two ways:

* Unmanaged:

* Managed mode: Your fleet of devices is managed by E4's

The E4 client library is developed by [Teserakt](https://teserakt.io) and is free to use.

Please let us know in the [issues](https://github.com/Teserakt-io/e4common/issues) or via [email](mailto:team@teserakt.io) if you stumble upon a bug, see a possible enhancement, or have a comment on design.

## Features

* Works on ARM, AVR, MIPS, x86
* Minimal code and RAM footprint
* Strong cryptography

## Motivations

Machine-to-machine communications are generally not peer-to-peer, but instead via a host that relays messages from a sender to one or more recipients.
Such a class of protocols is *publish/subscribe* protocols, wherein "published" messages are labelled with a “topic”  and receiving devices “subscribe” to a topic in order to receive all messages labelled with the said topic.

One such protocol, and allegedly the most popular one, is Message Queue Telemetry Transport (MQTT), wherein the relay server is called a “message broker”.
In MQTT, communication security can be ensured by two means:

* *Client–server security*, using Transport Layer Security (TLS): With this technique, data can be protected (in particular, encrypted and authenticated) in transit between a client and a server, but the server must decrypt the data and can therefore read clear data and tamper with it. Client–server security therefore requires total trust in the server and in its security against third-party attackers.

* *End-to-end security*: With this technique, data can be protected in transit all the way through from a client to another client, in such a way that the broker does not hold the cryptographic keys necessary to decrypt or tamper with the data. However, this technique is more complex to deploy than client–server encryption, because it requires 1) custom choices of encryption algorithms and 2) a process to distribute and manage keys.

To summarize, client–server security is insufficient to offer complete protection, while end-to-end security is hard to deploy and therefore is rarely used.
A major challenge is indeed key management, or operations consistent of creation and update of security keys, as well as key revocation.

## E4 in a nutshell

E4 simplifies the use and deployment of end-to-end security for MQTT and other IoT protocols, thanks to

1. A special MQTT client, called *command-and-control (C2)* remotely manages per-topic keys to other clients via MQTT messages, which are encrypted using clients' unique keys. C2 stores all clients' keys, all topic keys, as well as the list of topics associated to each client. C2 has a web interface and a REST API to perform key management operations.

2. A clients software library to encrypt and decrypt messages as well as process C2's commands. Encryption and decryption can also be done in a local proxy service.

Keys can then be assigned to a group of clients in order to prevent clients that are not in this group from decrypting their messages.
Keys can also be updated regularly, manually or automatically, to mitigate the risk of a compromise of a previous key.

Key update can also be used to ensure that clients added to a group would not be able to decrypt previous messages, and that clients removed from a group will not longer be able to decrypt future messages in the group.

Key benefits of E4 include:

* **Unmodified broker**: No modification to the broker is needed, brokers such as AWS IoT Core or Google Cloud IoT Core are supported out of the box.

* **Simple integration**: E4 sits between the transport layer and the business logic, with almost no required interaction with these protocols.

* **Lightweight**: E4 is optimized for platforms with constrained performance profile, such as embedded GNU/Linux or AVR-based platforms.

* **Built-in key management**: A server called C2 (command-and-control) provides a web UI and a REST API to manage per-topic keys. C2 acts as an MQTT client, hence communications between C2 and devices occur via a broker.

* **Protocol-agnostic**: E4 was designed with MQTT in mind, but can be adapted to other transport protocols (AMQP, CoAP, and so on). Specifically, the client library has nothing specific to MQTT besides its use of the topics semantics, common to other pubsub protocols.

Security features of E4 are:

* **Full protection**: E4 protects the confidentiality and integrity of messages, and prevents replay attacks.

* **Key agility**: Keys are unique per MQTT topic, and can be changed/rotated at any time to provide forward and backward secrecy.

* **Group communications**: Ad hoc groups of devices communicating securely with each other can be created.

* **Post-quantum**: E4's security would not be altered by a quantum computer, thanks to its use of 256-bit symmetric cryptography security.

* **Secure defaults**: E4 is opinionated, so that users don't need to pick correct secure options. The protocol is built on high security from the ground up.

Security limitations of E4 are:

* **Single point of failure**: C2 holds the client and topic keys, and can therefore impersonate any client, and decrypt any encrypted message. This can be mitigated by storing keys in encrypted form and storing the decryption key in an HSM, or in other some software vault technology. C2 should only be accessible after proper authentication (and preferably multi-factor authentication), and all its operations must be logged.

* **Risk from untrusted clients**: A malicious client may ignore a `RemoveTopic` command and keep the decryption keys in their table. To ensure that a client loses access to a given topic, the topic key must therefore be updated on all clients.


## E4 modes

E4 comes in two main modes:

* Symmetric-key crypto (**SK mode**), which does not use public-key cryptography and is therefore the lightest in terms of computation, code size, message overhead. Like with mobile telecommunication, TLS PSK, or Kerberos, this mode only employs symmetric-key primitives and establishes trust via pre-shared keys.

* Public-key crypto (**PK mode**), which uses public-key key agreement and digital signatures, which provides non-repudiation and mitigates the risk of database compromise, by storing clients' public keys rather than shared symmetric keys. Like PKI-based systems, this mode establishes trust by pre-provisioning trusted public keys of other identities (which can be seen as simplified certificates in E4).

In the specification below, parts specific to a mode are signalled with "**[SK mode]**" or "**[PK mode]**".

E4 in addition supports variants of the above mode:

* **FIPS mode**, which uses only algorithms that are compliant with FIPS 140-2. Both the symmetric-key and public-key mode can be instantiated in FIPS mode.

* **Fast mode**, which accelerates symmetric-key primitives by making fewer rounds, following the analysis from the Too Much Crypto paper: in this mode, the AES-256 core makes 10 rounds instead of 12, and SHA-3 makes 10 rounds instead of 24. Note that this is incompatible with FIPS mode, as these versions are no longer FIPS 140-2-compliant.

* **Post-quantum mode**, which is a variant of the public-key mode that uses post-quantum primitives instead of primitives broken by quantum algorithms.

These modes are defined in the section [Protocol Variants](protocol-variants) below.
