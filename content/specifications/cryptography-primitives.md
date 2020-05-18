---
title: Cryptography primitives
weight: 6
---

* [Authenticated encryption](#authenticated-encryption)
* [Hashing](#hashing)
* [Password-based key derivation](#password-based-key-derivation)
* [Signature [PK mode]](#signature-pk-mode)
* [Key agreement [PK mode]](#key-agreement-pk-mode)

This section lists the default cryptography algorithms used in E4.
We can also integrate other algorithms based on customers' preferences, as well as custom proprietary algorithms, both as a replacement or as an additional encryption layer.

## Authenticated encryption

E4 uses the AES-SIV authenticated encryption mode as specified in [RFC 5297](https://tools.ietf.org/html/rfc5297), with a same 256-bit key for the internal MAC and CTR instance, with by default no nonce.

AES-SIV can leverage hardware AES instructions or other cryptographic accelerators.

## Hashing

Hashing is used for hashing topic names, and uses SHA-3 with a 32-byte output.

Hashing can also be used to generate client id's from a string alias.

## Password-based key derivation

The C2 human interface components use Argon2id to derive a symmetric key, or to derive the seem from which a public/private key pair is generated.

## Signature [PK mode]

E4 client use Ed25519 to sign messages.s

## Key agreement [PK mode]

C2 and E4 clients use X25519 to derive a symmetric key from a private and a public key, hashing the result with SHA-3-256.
