package sinacrypt

// CipherConfig holds the configuration for a sinacrypt encryption context.
//
// The key must be exactly 32 bytes (AES-256). The nonce state file tracks
// the monotonic counter that guarantees nonce uniqueness.
type CipherConfig struct {
	// Key is the 32-byte AES-256 pre-shared key.
	Key []byte

	// NonceStateFile is the path to the file that persists the monotonic
	// nonce counter across process restarts. The file is created
	// automatically if it does not exist.
	NonceStateFile string
}

// BlindIndexConfig configures blind index generation for searchable encryption.
type BlindIndexConfig struct {
	// HMACKey is the key used for HMAC-SHA256 blind index computation.
	// It must be independent of the encryption key.
	HMACKey []byte
}

// NgramIndexConfig configures N-gram index generation for partial search
// on encrypted fields (supports LIKE '%substring%' queries).
type NgramIndexConfig struct {
	// HMACKey is the key used for HMAC-SHA256 computation of each N-gram.
	HMACKey []byte

	// N is the N-gram size (number of characters per gram). Typical value: 3.
	N int

	// MinLength is the minimum plaintext length to index. Shorter values
	// produce a single blind index instead of N-grams.
	MinLength int
}

// EncryptedField represents a field encrypted with lib_sina_crypt.
//
// The ciphertext is AES-256-GCM with a deterministic monotonic nonce.
// BlindIndex and NgramIndices are optional and present only when the
// corresponding index configuration was provided at encryption time.
type EncryptedField struct {
	// Ciphertext is the AES-256-GCM encrypted value including the nonce
	// and authentication tag.
	Ciphertext []byte

	// BlindIndex is the HMAC-SHA256 of the plaintext, used for exact-match
	// search on the encrypted column. Empty if no BlindIndexConfig was set.
	BlindIndex []byte

	// NgramIndices are the HMAC-SHA256 values of each N-gram of the
	// plaintext, used for partial search (LIKE queries). Nil if no
	// NgramIndexConfig was set.
	NgramIndices [][]byte
}

// Cipher performs field-level encryption and decryption using AES-256-GCM
// with a deterministic monotonic nonce.
//
// A Cipher is safe for concurrent use by multiple goroutines. The nonce
// counter is protected by an internal mutex and persisted to disk after
// each increment.
//
// This is the documentation-only package. The full implementation requires
// the lib_sina_crypt C library linked via cgo. Contact SINAINSIGHT for
// licensing: contact@sinainsight.eu
type Cipher struct {
	config      CipherConfig
	blindConfig *BlindIndexConfig
	ngramConfig *NgramIndexConfig
}
