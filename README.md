## Claimcheck

Webservice for verifying redemption claims on the stellar network

## TODO
- [ ] Send proper error codes.
- [ ] Log traffic and errors. Use care not to log the given secret.
- [ ] Handle asset codes that are alpha4 and alpha12
- [ ] Reject transactions involving native assets
- [ ] Setup CI pipeline
- [ ] Optimize docker build
- [ ] Mitigate DDOS attacks
    - ban by ip if too many invalid transaction hashes are sent.
    - adjust rate limit settings.
    - cache response from horizon using bolt with the transaction hash as the cache key.
    - cloudflare?
- [ ] Setup asset_code as an optional parameter
- [ ] Setup supervisor to keep server alive
- [ ] Configure for testnet and mainnet
- [ ] Setup for use at claimcheck.app/stellar/v1 and claimcheck.app/stellar-testnet/v1
- [ ] Setup horizon servers for testnet and mainnet
- [ ] Look into querying a stellar core postgres database directly
