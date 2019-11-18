# Teserakt E4

Teserakt's E4 solution makes IoT data protection painless thanks to:

* **E4 client library**, which encrypts and decrypts data in IoT devices

* **C2 server**, which manages devices' keys thanks to:
    - **Automation** of key distribution and key rotation
    - **Scalability** to millions of devices and messages
    - **Monitoring** of the data to detect network abnomalies

For example, the C2 server can ensure that each device will use

* a first key to decrypt firmware and configuration updates
* a second key to encrypt data sent to the analytics back-end

in a way that these keys are automatically rotated every week in order
to mitigate a potential breach.

C2 can run on-premise or as SaaS and talks HTTP/JSON and
gRPC/protobuf. We provide a CLI interface and a web GUI, which
you can test on our "free-for-all" platform at <a
href="https://console.demo.teserakt.io">console.demo.teserakt.io</a>, by
following <a href="TODO">our instructions</a>.

Don't hesitate to <a href="mailto:contact@teserakt.io">contact us</a>
for more information or to try E4.

The client library is open-source and can be used for free without C2,
by following our <a href="TODO">Getting started</a> pages.


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

## Support

Contact us.
