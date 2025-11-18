# Backlog

## Future Improvements

Currently empty - this fix is complete with no identified technical debt or deferred work.

## Potential Optimizations (Not Planned)

The following are observations, not actionable items:

1. **Monitor cosmos-sdk updates**: Future versions of cosmossdk.io/log may update their sonic dependency. Monitor for updates in cosmos-sdk v0.51+.

2. **Sonic alternatives**: If sonic causes issues in the future, consider alternatives:
   - Standard library `encoding/json` (slower but always compatible)
   - `github.com/goccy/go-json` (another fast JSON library)
   - Wait for cosmossdk.io/log to support configurable JSON encoders

3. **Go 1.25 planning**: When Go 1.25 is released, verify sonic v1.14.2 compatibility or upgrade if needed.

---

**Note**: All of the above are hypothetical future scenarios, not current action items. The current fix is complete and requires no follow-up work.
