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

## FreeMium — Version d'évaluation

Une version FreeMium est disponible en téléchargement libre :

- **Bibliothèque statique** `.a` (Linux x86_64, macOS x86_64)
- **Python wheel** : `pip install sina_crypt-2.1.0-cp312-cp312-manylinux_2_38_x86_64.whl`
- **Démo SQLite** interactive (Go + Python)
- **Whitepaper** : [GDPR Encryption for Databases (EN)](whitepaper-en.html) | [Chiffrement RGPD des bases de données (FR)](whitepaper.html) | [PDF](https://github.com/sinainsight/lib-sina-crypt/releases/download/v2.1.0-freemium/lib_sina_crypt-chiffrement-rgpd-bases-de-donnees.pdf)

👉 [Télécharger la release v2.1.0-freemium](https://github.com/sinainsight/lib-sina-crypt/releases/tag/v2.1.0-freemium)

> **⚠️ La version FreeMium utilise un CLIENT_NUMBER partagé.** Tout utilisateur FreeMium peut déchiffrer les données d'un autre FreeMium. Évaluation uniquement — licence commerciale avec clé dédiée : contact@sinainsight.eu

## RGPD Database Encryption — Example

### Schema (PostgreSQL)

```sql
CREATE TABLE clients (
    id            SERIAL PRIMARY KEY,
    -- Encrypted fields (BLOB/BYTEA) — opaque to anyone without the key
    nom_cipher    BYTEA NOT NULL,
    prenom_cipher BYTEA NOT NULL,
    email_cipher  BYTEA NOT NULL,
    iban_cipher   BYTEA NOT NULL,
    tel_cipher    BYTEA NOT NULL,
    -- Blind indices (HMAC-SHA256) — enable exact-match search without decryption
    nom_hmac      TEXT NOT NULL,
    email_hmac    TEXT NOT NULL,
    iban_hmac     TEXT NOT NULL,
    -- Metadata (not encrypted)
    created_at    TIMESTAMPTZ DEFAULT NOW(),
    status        TEXT DEFAULT 'actif'
);

CREATE INDEX idx_nom_hmac   ON clients(nom_hmac);
CREATE INDEX idx_email_hmac ON clients(email_hmac);
CREATE INDEX idx_iban_hmac  ON clients(iban_hmac);

-- Trigram index table for partial search (LIKE '%dupont%')
CREATE TABLE client_trigrams (
    client_id   INTEGER REFERENCES clients(id) ON DELETE CASCADE,
    field_name  TEXT NOT NULL,
    trigram_idx TEXT NOT NULL
);

CREATE INDEX idx_trigrams ON client_trigrams(field_name, trigram_idx);
```

### INSERT — Encrypt fields before storing

```python
import sina_crypt
import hmac, hashlib

INDEX_KEY = b"your-secret-hmac-key-here"

def blind_index(value):
    return hmac.new(INDEX_KEY, value.lower().encode(), hashlib.sha256).hexdigest()

# Encrypt each RGPD field with AES-256-GCM
nom_enc   = sina_crypt.rgpd_crypt(b"Dupont")
email_enc = sina_crypt.rgpd_crypt(b"jean.dupont@example.com")
iban_enc  = sina_crypt.rgpd_crypt(b"FR76 3000 6000 0112 3456 7890 189")

# Compute blind indices for searchable fields
nom_idx   = blind_index("Dupont")
email_idx = blind_index("jean.dupont@example.com")
iban_idx  = blind_index("FR76 3000 6000 0112 3456 7890 189")

cursor.execute("""
    INSERT INTO clients (nom_cipher, email_cipher, iban_cipher,
                         nom_hmac, email_hmac, iban_hmac)
    VALUES (%s, %s, %s, %s, %s, %s)
""", (nom_enc, email_enc, iban_enc, nom_idx, email_idx, iban_idx))
```

### SELECT — Search by blind index, decrypt result

```sql
-- Find by email: compare HMAC, never decrypt in the WHERE clause
SELECT nom_cipher, prenom_cipher, email_cipher
FROM clients
WHERE email_hmac = 'a1b2c3...';  -- HMAC('jean.dupont@example.com')
```

```python
# Application decrypts the result
row = cursor.fetchone()
nom   = sina_crypt.rgpd_decrypt(row[0])   # b"Dupont"
email = sina_crypt.rgpd_decrypt(row[2])   # b"jean.dupont@example.com"
```

### SELECT partial — Trigram search (LIKE '%pont%')

```sql
-- Decompose 'pont' into trigrams: {pon, ont}
-- Each trigram is HMAC'd and stored in client_trigrams
SELECT c.nom_cipher, c.email_cipher
FROM clients c
WHERE c.id IN (
    SELECT client_id FROM client_trigrams
    WHERE trigram_idx IN ('hmac_pon', 'hmac_ont')
    GROUP BY client_id
    HAVING COUNT(DISTINCT trigram_idx) = 2
);
```

### DELETE — RGPD Art. 17 Right to Erasure

```sql
-- Delete without decryption: use the blind index
DELETE FROM clients
WHERE email_hmac = 'a1b2c3...';  -- HMAC('jean.dupont@example.com')

-- Cascade deletes trigrams (ON DELETE CASCADE)
```

### What the DBA sees

| id | nom_cipher | email_hmac | status |
|----|------------|------------|--------|
| 1 | `\x51514b77466a...` | `d6957904c772cce3...` | actif |
| 2 | `\x51536f347456...` | `a61ec88bb93eabc5...` | actif |

The `_cipher` columns are opaque binary. The `_hmac` columns are deterministic hashes — they enable search but cannot be reversed to recover the plaintext.

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
