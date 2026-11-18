# lib_sina_crypt

Field-level encryption library with deterministic monotonic nonce, by [SINAINSIGHT](https://www.sinainsight.pro).

## Overview

`lib_sina_crypt` is an AES-256-GCM encryption library with a **strictly monotonic, deterministic, cross-platform nonce** that is guaranteed unique. It removes the main weakness of AES-GCM — nonce reuse — by design.

**Patent**: INPI FR2601046 (registered, grant pending).

## Key Features

- **AES-256-GCM** with hardware acceleration (AES-NI)
- **Deterministic monotonic nonce** — guaranteed unique across platforms (patented)
- **Pre-shared keys only** — no key exchange step, no asymmetric cryptography
- **Quantum-resistant by design** — quantum algorithms (Shor's) threaten the key-exchange step of classical protocols (RSA, DH, ECDH); lib_sina_crypt has no such step
- **Blind indices** (HMAC-SHA256) for exact-match search on encrypted columns
- **N-gram indices** for partial search (`LIKE '%dupont%'`) on encrypted data
- **SQL transparency** — existing application queries work unchanged on encrypted columns
- **0.02 ms per field** encryption (measured)

## Language Bindings

| Language | Status | Notes |
|----------|--------|-------|
| C | Core library | Reference implementation |
| Go | Production | cgo bindings |
| Python | Production | C extension module |
| Rust | Production | FFI bindings |
| Node.js | Production | N-API bindings |

## How It Works

### Deterministic Monotonic Nonce

Standard AES-GCM uses a random 96-bit nonce. With random generation, the birthday paradox creates a collision risk after ~2³² encryptions per key — a catastrophic failure that leaks plaintext.

lib_sina_crypt replaces the random nonce with a **deterministic, strictly monotonic counter** that is:

1. **Guaranteed unique** — monotonically increasing, never repeats
2. **Cross-platform** — same sequence regardless of OS, architecture, or language binding
3. **Deterministic** — reproducible from the same starting state

This eliminates nonce reuse entirely, making AES-GCM safe for high-volume encryption (billions of fields per key).

### Searchable Encryption

For database column encryption that preserves query capability:

- **Exact match**: HMAC-SHA256 blind index stored alongside the ciphertext. Query: `WHERE blind_index = HMAC(search_term)`.
- **Partial match**: N-gram decomposition of the plaintext, each n-gram indexed with HMAC. Supports `LIKE '%substring%'` queries on encrypted data.

Both approaches are **deterministic** — the same plaintext always produces the same index — enabling standard SQL queries without decryption.

### Performance

| Operation | Time | Notes |
|-----------|------|-------|
| Encrypt one field | 0.02 ms | AES-NI hardware acceleration |
| Decrypt one field | 0.02 ms | AES-NI hardware acceleration |
| Blind index (HMAC) | 0.01 ms | Per field |
| N-gram index | 0.05 ms | Per field (depends on length) |

## Use Cases

- **GDPR compliance**: Encrypt PII columns (names, emails, addresses) while preserving search
- **PCI DSS**: Encrypt cardholder data at rest
- **Credential storage**: Encrypt stored passwords and API keys
- **Message transport**: End-to-end encryption with pre-shared keys (Kafka, AMQP)
- **Database migration**: Add encryption to existing columns without changing application queries

## Deployment

lib_sina_crypt is deployed **on-premise only**, inside the customer's information system. No data leaves the customer's perimeter. The library ships as a static or shared object linked into the application.

In production since 2024.

## Related Products

lib_sina_crypt is the encryption foundation for all SINAINSIGHT products:

- **[SinaShield](https://www.sinainsight.pro/sinashield/)** — Pre-SIEM telemetry control plane
- **[SinaStack](https://www.sinainsight.pro/sinastack/)** — Agentless infrastructure observability (Kafka transport encryption)
- **[SinaApp](https://www.sinainsight.pro/sinaapp/)** — Stateful synthetic monitoring (sina_mem encrypted storage)

## Links

- [SINAINSIGHT](https://www.sinainsight.pro) — Company website
- [llms.txt](https://www.sinainsight.pro/llms-full.txt) — Full technical reference for AI agents
- [Fraud detection POC](https://fraud-poc.sinainsight.eu/) — Live dashboard using lib_sina_crypt

## License

Documentation: [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/).
lib_sina_crypt is proprietary software by SINAINSIGHT SAS, covered by patent INPI FR2601046.
