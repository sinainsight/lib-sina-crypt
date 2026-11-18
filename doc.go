// Package sinacrypt provides field-level AES-256-GCM encryption with a
// patented deterministic monotonic nonce (INPI FR2601046).
//
// # Overview
//
// lib_sina_crypt encrypts individual database fields with AES-256-GCM using
// pre-shared keys. Its strictly monotonic, deterministic nonce eliminates
// the main weakness of AES-GCM (nonce reuse) by design.
//
// Since it uses pre-shared keys with no asymmetric key exchange, there is no
// key-exchange step for a quantum computer to attack.
//
// # Searchable Encryption
//
// Two index types enable SQL queries on encrypted columns:
//
//   - Blind indices (HMAC-SHA256) for exact-match search
//   - N-gram indices for partial search (LIKE '%substring%')
//
// Existing application queries work unchanged on encrypted columns.
//
// # Performance
//
// Measured on commodity hardware with AES-NI:
//
//   - Encrypt: 0.02 ms per field
//   - Decrypt: 0.02 ms per field
//   - Blind index: 0.01 ms per field
//
// # Language Bindings
//
// The core library is written in C. Bindings are available for:
//
//   - Go (this package, via cgo)
//   - Python (C extension module)
//   - Rust (FFI bindings)
//   - Node.js (N-API bindings)
//
// # Patent
//
// The deterministic monotonic nonce mechanism is covered by French patent
// INPI FR2601046 (registered, grant pending).
//
// # More Information
//
//   - Documentation: https://github.com/sinainsight/lib-sina-crypt
//   - Company: https://www.sinainsight.pro
//   - Full technical reference: https://www.sinainsight.pro/llms-full.txt
package sinacrypt
