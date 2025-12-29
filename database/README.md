# Database Layer (Off-Chain)

This directory manages off-chain storage requirements.

## Purpose

While the blockchain acts as the immutable ledger, a traditional database is often used for:
- Caching world state for complex queries.
- Storing user profile data not suitable for the ledger.
- Analytics and reporting.

## Contents

- **migrations/**: SQL/NoSQL migration scripts.
- **seeds/**: Initial data seeding.
