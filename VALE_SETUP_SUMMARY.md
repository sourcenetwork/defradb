# Vale Documentation Linting - Setup Complete! ✨

Vale has been successfully configured for DefraDB with custom rules based on your brand guidelines.

## 📦 What Was Created

### Configuration Files
- **`.vale.ini`** - Main Vale configuration
- **`.github/workflows/lint-docs.yml`** - GitHub Actions workflow for PR linting
- **`.gitignore`** - Updated to exclude Vale packages

### Custom DefraDB Style Rules (11 rules)

#### 🔴 Error Level (Branding & Consistency)
1. **ProductNames.yml** - Enforces correct product names (DefraDB, SourceHub, Orbis, LensVM)
2. **TechTerms.yml** - Ensures consistent technical terminology (edge-first, Edge AI, etc.)

#### 🟡 Warning Level (Style & Voice)
3. **CorporateSpeak.yml** - Flags buzzwords (leverage, synergistic, cutting-edge)
4. **GrandioseClaims.yml** - Catches unsupported claims ("the best", "leading")
5. **SourceStack.yml** - Enforces "Source stack" not "Source platform"
6. **Acronyms.yml** - Requires acronym definitions on first use

#### 🔵 Suggestion Level (Writing Quality)
7. **Hedging.yml** - Removes weak language (simply, just, easily)
8. **PreferredTerms.yml** - Suggests brand-consistent alternatives
9. **Voice.yml** - Encourages second-person voice for developer docs
10. **Headings.yml** - Validates title case in headings
11. **NoteFormat.yml** - Standardizes note formatting

### Vocabulary Files
- **`accept.txt`** - 60+ approved technical terms and product names
- **`reject.txt`** - Forbidden terms and weak language

### Documentation
- **`docs/vale-setup.md`** - Complete usage guide

### Makefile Integration
- **`make deps:lint-docs`** - Install Vale
- **`make vale:sync`** - Sync style packages
- **`make lint:docs`** - Run Vale on all documentation
- **`make lint:docs:strict`** - Run Vale (errors only)

## 🎯 Current Status

Vale found issues across your documentation:
- **705 errors** (branding, Google style guide violations)
- **93 warnings** (style guide, voice, acronyms)
- **20 suggestions** (hedging language, voice consistency)

This is **expected and good**! Vale is catching real issues that need fixing.

## 🚀 Quick Start

### 1. Run Vale Locally

```bash
# Sync packages (first time)
make vale:sync

# Lint all docs
make lint:docs

# Lint specific file
vale docs/website/guides/peer-to-peer.md

# Strict mode (errors only)
make lint:docs:strict
```

### 2. Understand Output

```
docs/website/guides/example.md
 15:23  error    Use 'DefraDB' instead of        DefraDB.ProductNames
                 'Defra' for correct product
                 branding
 42:15  warning  Avoid corporate buzzword        DefraDB.CorporateSpeak
                 'leverage' - be specific
 68:8   suggestion  Remove hedging word 'simply'  DefraDB.Hedging
                     - be direct
```

### 3. Fix Issues Gradually

**Priority 1: Fix Errors** (branding, critical)
```bash
# See all errors
vale --minAlertLevel=error docs/
```

**Priority 2: Fix Warnings** (style guide violations)
**Priority 3: Consider Suggestions** (improvements)

### 4. Disable Rules When Needed

```markdown
<!-- vale off -->
This section is ignored by Vale.
<!-- vale on -->

<!-- vale DefraDB.Hedging = NO -->
In this context, "simply" is acceptable.
<!-- vale DefraDB.Hedging = YES -->
```

## 🤖 CI/CD Integration

Vale runs automatically on all PRs that modify markdown files:

- **Inline comments** on PR diffs
- **Fails** if error-level rules are violated
- **Filter mode**: Only checks changed lines (not entire files)
- **Workflow**: `.github/workflows/lint-docs.yml`

## 📝 Rule Examples from Brand Guidelines

### Product Names (Error)
```
❌ Defra is a database
✅ DefraDB is a database

❌ Sourcehub provides trust
✅ SourceHub provides trust
```

### Tech Terms (Error)
```
❌ Edge-First Software
✅ edge-first software

❌ edge ai
✅ Edge AI
```

### Corporate Speak (Warning)
```
❌ Leverage our cutting-edge platform
✅ Use our data management stack

❌ Synergistic enterprise solution
✅ Integrated data management tools
```

### Hedging Language (Suggestion)
```
❌ Simply install the package
✅ Install the package

❌ Just add the configuration
✅ Add the configuration
```

### Voice (Suggestion)
```
❌ Developers can use DefraDB to build apps
✅ You can use DefraDB to build your apps
```

## 📊 Next Steps

### 1. Triage Issues (Recommended Approach)

**Week 1: Fix Critical Errors**
- Product name consistency (DefraDB, SourceHub)
- Brand terminology violations
- Run: `vale --minAlertLevel=error docs/`

**Week 2-3: Address Warnings**
- Corporate buzzwords
- Unsupported claims
- Acronym definitions
- Run: `vale --minAlertLevel=warning docs/`

**Week 4: Review Suggestions**
- Hedging language
- Voice consistency
- Note formatting

### 2. Update Vocabulary

As you encounter false positives:
1. Add terms to `styles/config/vocabularies/DefraDB/accept.txt`
2. Run `vale sync`
3. Test again

### 3. Refine Rules

If a rule is too strict:
1. Edit the rule file in `styles/DefraDB/`
2. Adjust `level` (error → warning → suggestion)
3. Add `exceptions` for known valid cases
4. Test changes locally

### 4. Integrate into Workflow

- **Pre-commit**: Set up pre-commit hooks
- **Editor**: Install Vale VSCode extension for real-time linting
- **Team**: Share `docs/vale-setup.md` with contributors

## 🔧 Customization

### Add New Rules

Create new YAML file in `styles/DefraDB/`:

```yaml
# styles/DefraDB/MyNewRule.yml
extends: existence
message: "Avoid using '%s'"
level: warning
scope: text
tokens:
  - foo
  - bar
```

### Adjust Severity

Edit rule file and change `level`:
- `error` - Must fix (branding, breaking issues)
- `warning` - Should fix (style guide)
- `suggestion` - Consider fixing (improvements)

### Update Vocabulary

Add new technical terms to `accept.txt`:
```
[Rr]ust
[Gg]o(?:lang)?
Kubernetes
```

## 🐛 Troubleshooting

### "Package not found"
```bash
make vale:sync
```

### Too many errors initially
```bash
# Start strict, fix errors only
vale --minAlertLevel=error docs/

# Gradually lower threshold
vale --minAlertLevel=warning docs/
```

### False positives
- Add to vocabulary: `styles/config/vocabularies/DefraDB/accept.txt`
- Or use inline disabling: `<!-- vale off -->`

## 📚 Resources

- **Vale Documentation**: https://vale.sh/docs/
- **DefraDB Vale Guide**: `docs/vale-setup.md`
- **Style Rule Syntax**: https://vale.sh/docs/topics/styles/
- **Google Style Guide**: https://developers.google.com/style
- **write-good**: https://github.com/btford/write-good

## 🎉 Success Metrics

After implementing Vale:
- ✅ Consistent product name capitalization across all docs
- ✅ Removed corporate buzzwords and weak language
- ✅ Standardized technical terminology
- ✅ Improved developer-focused voice
- ✅ Caught documentation drift in CI before merge

## 💬 Questions?

- Review `docs/vale-setup.md` for detailed usage
- Check rule files in `styles/DefraDB/` for examples
- Ask in `#docs` channel on Discord

Happy linting! ✨
