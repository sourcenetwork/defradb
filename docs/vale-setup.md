# Vale Documentation Linting

DefraDB uses [Vale](https://vale.sh/) to enforce documentation style, branding consistency, and writing quality across all markdown files.

## Quick Start

### Install Vale

**macOS:**
```bash
brew install vale
```

**Linux:**
```bash
# Download latest release
wget https://github.com/errata-ai/vale/releases/download/v2.29.0/vale_2.29.0_Linux_64-bit.tar.gz
tar -xzf vale_2.29.0_Linux_64-bit.tar.gz
sudo mv vale /usr/local/bin/
```

**Or use Make:**
```bash
make deps:lint-docs
```

### Sync Style Packages

Vale uses external style packages (Google, write-good). Sync them before first use:

```bash
make vale:sync
# or
vale sync
```

### Run Vale

**Check all documentation:**
```bash
make lint:docs
```

**Check only errors (strict mode):**
```bash
make lint:docs:strict
```

**Check specific file:**
```bash
vale docs/website/guides/peer-to-peer.md
```

**Check specific directory:**
```bash
vale docs/website/guides/
```

## Understanding Vale Output

Vale reports issues with three severity levels:

- 🔴 **Error**: Branding violations, incorrect product names (must fix)
- 🟡 **Warning**: Style guide violations, corporate speak (should fix)
- 🔵 **Suggestion**: Improvements, voice consistency (consider fixing)

Example output:
```
docs/website/guides/example.md
 15:23  error    Use 'DefraDB' instead of        DefraDB.ProductNames
                 'Defra' for correct product
                 branding
 42:15  warning  Avoid corporate buzzword        DefraDB.CorporateSpeak
                 'leverage' - be specific
                 instead
 68:8   suggestion  Remove hedging word 'simply'  DefraDB.Hedging
                     - be direct
```

## Vale Rules

DefraDB has 11 custom Vale rules organized by category:

### Branding & Terminology (Errors)

- **ProductNames.yml**: Enforces correct capitalization of DefraDB, SourceHub, Orbis, LensVM
- **TechTerms.yml**: Ensures consistent technical terminology (edge-first, Edge AI, etc.)

### Style & Voice (Warnings)

- **CorporateSpeak.yml**: Flags buzzwords (leverage, synergistic, cutting-edge)
- **GrandioseClaims.yml**: Catches unsupported claims ("the best", "leading")
- **SourceStack.yml**: Enforces "Source stack" not "Source platform"
- **Acronyms.yml**: Requires acronym definitions on first use

### Writing Quality (Suggestions)

- **Hedging.yml**: Removes weak language (simply, just, easily)
- **PreferredTerms.yml**: Suggests brand-consistent alternatives
- **Voice.yml**: Encourages second-person voice for developer docs
- **Headings.yml**: Validates title case in headings
- **NoteFormat.yml**: Standardizes note formatting

## Disabling Rules

### Disable for a section

```markdown
<!-- vale off -->

This content is ignored by Vale.

<!-- vale on -->
```

### Disable specific rule

```markdown
<!-- vale DefraDB.Hedging = NO -->

This section can use "simply" or "just" without warnings.

<!-- vale DefraDB.Hedging = YES -->
```

### Disable entire style

```markdown
<!-- vale DefraDB = NO -->

All DefraDB rules disabled in this section.

<!-- vale DefraDB = YES -->
```

## Editor Integration

### VS Code

1. Install the [Vale VSCode extension](https://marketplace.visualstudio.com/items?itemName=errata-ai.vale-server)
2. Vale will automatically detect `.vale.ini` and provide real-time linting

### Neovim

```lua
require('lspconfig').vale_ls.setup{
  filetypes = {'markdown', 'text'},
  init_options = {
    installVale = true,
  }
}
```

### Other Editors

Vale provides LSP support for most editors. See: https://vale.sh/docs/integrations/guide/

## CI/CD Integration

Vale runs automatically on all pull requests that modify markdown files:

- **GitHub Actions**: `.github/workflows/lint-docs.yml`
- **Inline comments**: Vale adds review comments to PRs
- **Fail on errors**: PRs fail if error-level rules are violated
- **Filter mode**: Only checks changed lines (not entire files)

## Vocabulary

Vale uses a custom vocabulary for DefraDB-specific terms:

- **Accept list** (`styles/config/vocabularies/DefraDB/accept.txt`): Approved technical terms
- **Reject list** (`styles/config/vocabularies/DefraDB/reject.txt`): Forbidden terms

To add new technical terms:
1. Edit `styles/config/vocabularies/DefraDB/accept.txt`
2. Add one term per line (regex patterns supported)
3. Run `vale sync` to reload

## Troubleshooting

### "Package not found" error

```bash
make vale:sync
```

### Vale not finding .vale.ini

Ensure you're running Vale from the repository root, or specify config:

```bash
vale --config=.vale.ini docs/
```

### False positives

If Vale incorrectly flags something:
1. Check if it's a technical term that should be in the vocabulary
2. Add to `styles/config/vocabularies/DefraDB/accept.txt`
3. Or use inline disabling for specific instances

### Performance issues

Vale might be slow on very large files. To optimize:
- Use scopes in rules to limit where they apply
- Exclude large generated files in `.vale.ini`

## Contributing

When adding new documentation:
1. Write your content
2. Run `make lint:docs` locally
3. Fix error-level issues before committing
4. Address warnings and suggestions as appropriate
5. Use inline disabling sparingly (explain why in comments)

## Resources

- [Vale Documentation](https://vale.sh/docs/)
- [DefraDB Style Guide](https://github.com/sourcenetwork/defradb/blob/develop/CONTRIBUTING.md)
- [Vale Rule Syntax](https://vale.sh/docs/topics/styles/)
- [Google Style Guide](https://developers.google.com/style)

## Questions?

Ask in the `#docs` channel on [Discord](https://discord.gg/w7jYQVJ) or open an issue.
