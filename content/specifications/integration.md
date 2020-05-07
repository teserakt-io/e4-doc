---
title: Integration
weight: 2
---

* [Client integration](#client-integration)
* [Server integration (C2)](#server-integration-c2)
  * [Services](#services)
  * [Database](#database)
  * [Requirements](#requirements)
  * [Disaster recovery](#disaster-recovery)
  * [High availability](#high-availability)
  * [Fault tolerance](#fault-tolerance)
  * [Logging](#logging)
* [Key provisioning](#key-provisioning)

## Client integration

The E4 client can be integrated in one of the following ways:

1. *Software library*: We currently have a C and Go version of the E4 library, which developers can integrate in their applications. The E4 logic sits between the application logic and the transport layer: when transmitting a message, the code of the application logic will first call the E4 library's `protect()` function before calling MQTT's sending function, instead of sending the unencrypted message directly.

2. *Local protection proxy*: For cases where modifying the existing codebase is not acceptable, but the system can accommodate a local piece of software that can receive messages and then forward them to the broker. Then by reconfiguring the legacy application to just forward the data to the local proxy, which will protect the data and forward it to the broker, we have the same benefit of protection, but without the added clarity of a patched solution.

3. *Gateway protection proxy*: is a piece of software that is deployed on a gateway that forwards data from several devices to the broker. In this case, the gateway will transparently perform the protect/unprotect operations, and the end devices will not need any change.

Technical requirements are minimal: the client software requires 2KiB of RAM, plus 64 bytes of storage per topic that is encrypted.
The exact RAM consumption also depend on the size of messages received (e.g., a 1KiB message will need an additional 1KiB of RAM).

## Server integration (C2)

### Services

E4's server includes the core C2 service as well as other services:

* Database: PostgreSQL
* E4's automation engine
* Web application back-end
* Monitoring components: ElasticSearch, Kibana
* Observability components: OpenCensus agent, Prometheus, Jaeger

Of these components, only the C2 core and database are mandatory for a basic deployment of E4.
To deploy these in an existing infrastructure as Docker containers, we can provide Ansible scripts that will automatically download, configure, and start the required  services.
Alternatively, we can provide virtual appliances.

Once up and running, users can interact with C2 and its component via their respective  web interfaces, or use our command-line application.
Other services can leverage C2's HTTP and gRPC APIs, as well as APIs exposed by other services.

C2 should be connected to the same network as the MQTT broker used by the clients.
If a broker is not yet running, we can provide it as part of our deployment scripts.

### Database

C2 can either use its own database service, or use an SQL-like database already available in the organization's database infrastructure.
Any SQL-like database is directly supported. Other database types can be supported on demand.

Part of the database content is encrypted, so that the client keys in the database are not exposed to other components than C2.
Encryption can be performed in software directly, using an HSM (for which we would provide a custom functional module; we can support the Gemalto ProtectServer devices, having experience writing custom firmware for these models), or using a secret management service such as Hashicorp's Vault.

### Requirements

For a fleet of thousands of devices, the C2 core service will run on one core and 2GB RAM.

Monitoring services' requirements depends on the amount of messages, analytics type, and parameters such as messages retention period.

Observability components' requirements depend on the C2's activity level, and parameters such as the sampling rate.

### Disaster recovery

Disaster recovery concerns can be addressed by having regular off-site back-ups of the databases (C2's SQL database, and ElasticSearch if monitoring is used), as well as copies of the services configuration files.
With these back-ups, the C2 server can be rapidly operational after restoring the latest version of the database.
However, risks are the following:

* Depending on key rotation period, need to update topic keys

* Loss of contact with a device if identity key rotate (SK mode)

* Analytics data unavailable for the clients' activity during the period after the back-up when the C2 is not operational

However, when only basic features are used (manual device/topic management and no automated key rotation), the C2 can be shut down or unavailable with no impact on the operation of the clients.
This is convenient or maintenance, migration or just fault tolerance.
When automated features are used, these might be disabled for the duration of a planned outage of C2 components.

### High availability

To mitigate the risks from service and database outage, a high-availability architecture might be implemented, by deploying a second C2 application, with a load balancer / heartbeat type strategy for automation, for example as an active–active setup over different sites and a distributed database.

### Fault tolerance

A highly available setup does not eliminate all the risks related to a distributed system:

* Non-atomic operations, if run concurrently, can lead to an inconsistent state. For example, a first topic key rotation is sent, and a second key rotation request is sent milliseconds afterwards. Then, because of unpredictable CPU and network latencies, and of possible out-of-order message delivery, some devices might receive the first key update after the second one, whereas the C2 will consider the the topic key is  the second  one.

* Network-level failures preventing control messages to be delivered might require resending messages several times, either automatically (for example leveraging MQTT QoS and resending until a PUBACK is received, with a timeout and a cap to the number of retransmissions) or manually (for example if only QoS 0 is available). In such scenarios, one should empirically assess the risk of failure and configure the system (e.g, automation engine crypto period) in a way that minimizes such risks.


### Logging

C2 will generate detailed and structured system logs in JSON, which can be integrated in the organization's log management or SIEM, such as Splunk.


## Key provisioning

Here are possibilities for key initial provisioning:

* **Pre-keying in a trusted environment**:
  * Each device is connected in a trusted network, and a helper service/device provisions the devices with keys, configurations etc. The keys can then be transferred to C2.
  * Something like a QR code or NFC tag can be used on the devices to identify them (and link them to a factory set key) and help with key rotation. On first use, a key rotation command is issued, and the keys get updated.

* **Pre-keying in the field**: The operator has a trusted device that will use i2c, iButton, NFC, BLE, light or something to "fill" the device with the initial key. Said interface must be authorized and require authentication of the device pushing the key.

* **Trust on first use (TOFU)**: Devices are initialized a unique key and on first use C2 force-updates the key. This is the simplest and the least safe solution.
