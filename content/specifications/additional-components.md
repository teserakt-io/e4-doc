---
title: Additional components
weight: 8
---

* [Automation engine](#automation-engine)
* [Monitoring &amp; analytics](#monitoring--analytics)
* [Keygen](#keygen)

This section briefly describes additional server components, implementing other features than the core C2 key management logic:

## Automation engine

The automation engine (AE) is a separate application which will connect to the C2 in order to subscribe to an event stream. It's role is to rotates topic or clients keys according to a period defined by the user, or according to events occurring on the C2 (such as devices joining/leaving a topic).

See [https://github.com/teserakt-io/automation-engine](https://github.com/teserakt-io/automation-engine)

## Monitoring & analytics

C2 can subscribe to the topics managed and collect messages and their metadata in order to:

* Performs QA and security analytics, checking for example that:
    - Messages that should be encrypted are encrypted (that is, messages sent by devices that hold the topic's key)
    - Encrypted messages successfully decrypt
    - Timestamps are accurate

* Attempts to detect malicious behavior, such as:
    - Replay attacks
    - Messages dropped
    - Timestamp manipulation
    - Malicious content being sent within encrypted payloads

E4 uses ElasticSearch and Kibana to store and analyze+visualize data collected.

The C2 and automation engine application code also integrate OpenCensus instrumentation, in order to monitor performances and provide traceability, viewable from a Jaeger web interface.

## Keygen

The E4 Go library does provide a command line key generator, allowing to ease generation of the various supported key formats in use. See [https://github.com/teserakt-io/e4go/tree/develop/cmd/e4keygen](https://github.com/teserakt-io/e4go/tree/develop/cmd/e4keygen)
