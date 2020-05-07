---
title: Protocol variants
weight: 7
---

* [FIPS mode](#fips-mode)
* [Post-quantum mode](#post-quantum-mode)

## FIPS mode

E4's FIPS allows the application that uses it to obtain a FIPS 140-2 validation (CMVP).
For this, only FIPS-compliant algorithms must be used.

In its default version, E4 uses SHA-3 (FIPS-compliant), however it uses AES in SIV mode, a mode that is not FIPS-compliant.
E4's FIPS mode therefore replaces AES-SIV with AES-GCM.

In the public-key version of E4,

* Ed25519 is replaced with FIPS-compliant probabilistic ECDSA.
* Curve25519 is replaced with nistp256 for key agreement

## Post-quantum mode

The post-quantum mode is a variant of the public-key version of E4 wherein elliptic-curve cryptography is replaced with post-quantum algorithms.

For signature, E4 can use Dilithium instead of ECDSA, which implies that

* Public keys are 1472-byte instead of 32-byte
* Private keys are 3504-byte instead of 32-byte
* Signatures are 2701-byte instead of 64-byte

For key agreement, E4 can use..  TBD
