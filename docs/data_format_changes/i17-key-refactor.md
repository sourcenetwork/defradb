# Key refactor

String based keys converted to type-safe(er) system.

Commit key multicodec changed from raw to dag-pb - code is cleaner and codec is now consistent with the rest of our keys.
