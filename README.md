## Claimcheck

Webservice for verifying redemption claims on the stellar network

## TODO
- [ ] Send proper error codes.
- [ ] Log traffic and errors. Do not log the given secret.
- [x] Handle asset codes that are alpha4 and alpha12
- [ ] Reject transactions involving native assets
- [ ] Prevent DDOS attacks
    - ban by ip if too many invalid transaction hashes are sent.
    - adjust rate limit settings.
    - cache response from horizon using bolt with the transaction hash as the cache key.
- [ ] Setup asset_code as an optional parameter
