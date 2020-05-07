---
title: Additional components
weight: 8
---

* [Automation engine](#automation-engine)
* [Monitoring &amp; analytics](#monitoring--analytics)

This section briefly describes additional server components, implementing other features than the core C2 key management logic:

## Automation engine

The automation engine (AE) is a component of the C2 server which rotates topic keys according to a crypto period defined by the user, or according to events (devices joining/leaving).

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

We use ElasticSearch and Kibana to store and analyze+visualize data collected.
