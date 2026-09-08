# Breaking Change Log

This log records intentional one-off compatibility breaks that DefraDB ships without waiting for a major release, as described by the [versioning policy](./VERSIONING.md). These are genuine breaking changes, not the standing exceptions in that policy.

Unreleased entries are grouped under their release version when they ship.

## Unreleased

### Rename SourceHub to Vera and SourceHub Document ACP to Remote DAC

The SourceHub project has been renamed to [Vera](https://github.com/sourcenetwork/vera). DefraDB now calls the Vera-backed Document ACP implementation Remote DAC, alongside the existing Local DAC. DefraDB does not retain the former configuration values or API names as compatibility aliases.

DefraDB also adopts Vera's planned `LogID` terminology for the Remote DAC log identifier instead of retaining SourceHub's `ChainID` terminology. The current Vera SDK still accepts this value through `WithChainID`; that implementation detail remains inside the Remote DAC adapter so DefraDB users will not face another rename when Vera updates its API.

Update Remote DAC deployments and integrations to use Vera endpoints and dependencies.

#### Configuration and tooling

| Surface | Previous | Replacement |
|---------|----------|-------------|
| Document ACP type | `source-hub` | `remote` |
| Document ACP type flag | `--document-acp-type source-hub` | `--document-acp-type remote` |
| Document ACP type environment variable | `DEFRA_ACP_DOCUMENT_TYPE=source-hub` | `DEFRA_ACP_DOCUMENT_TYPE=remote` |
| Document ACP configuration namespace | `acp.document.sourceHub.*` | `acp.document.remote.*` |
| Remote DAC log ID configuration | `acp.document.sourceHub.ChainID` | `acp.document.remote.LogID` |
| Environment-based configuration namespace | `DEFRA_ACP_DOCUMENT_SOURCEHUB_*` | `DEFRA_ACP_DOCUMENT_REMOTE_*` |
| Remote DAC log ID environment variable | `DEFRA_ACP_DOCUMENT_SOURCEHUB_CHAINID` | `DEFRA_ACP_DOCUMENT_REMOTE_LOGID` |
| Client address flag | `--source-hub-address` | `--remote-dac-address` |
| Go module | `github.com/sourcenetwork/sourcehub` | `github.com/sourcenetwork/vera` |
| Container image | `ghcr.io/sourcenetwork/sourcehub` | `ghcr.io/sourcenetwork/vera` |
| Integration test ACP type | `DEFRA_DOCUMENT_ACP_TYPE=source-hub` | `DEFRA_DOCUMENT_ACP_TYPE=remote` |
| Integration test image | `DEFRA_SOURCEHUB_IMAGE` | `DEFRA_VERA_IMAGE` |
| Make target | `make test:source-hub` | `make test:remote-dac` |

The namespace-only change applies to `GRPCAddress`, `CometRPCAddress`, `KeyName`, and `address`. The `ChainID` setting instead becomes `LogID`, as shown above.

Addresses supplied through `acp.document.remote.address` or `--remote-dac-address`, including the JWT `authorized_account` claim, must use Vera's `vera` Bech32 prefix instead of SourceHub's `source` prefix.

#### Go API

| Previous | Replacement |
|----------|-------------|
| `dac.NewSourceHubACP` | `dac.NewRemoteDocumentACP` |
| `dac.NewACPSourceHub` | `dac.NewRemoteDocumentACPClient` |
| `dac.NewSourceHubDocumentACP` (JS/WASM) | `dac.NewRemoteDocumentACP` |
| `dac.SourceHubDocumentACP` | `dac.RemoteDocumentACP` |
| `SourceHubDocumentACP` in `acp/types` | `RemoteDocumentACP` in `acp/types` |
| `options.NodeSourceHubDocumentACPType` | `options.NodeRemoteDocumentACPType` |
| `NodeDocumentACPOptions.SourceHubChainID` | `NodeDocumentACPOptions.RemoteDACLogID` |
| `NodeDocumentACPOptions.SourceHubGRPCAddress` | `NodeDocumentACPOptions.RemoteDACGRPCAddress` |
| `NodeDocumentACPOptions.SourceHubCometRPCAddress` | `NodeDocumentACPOptions.RemoteDACCometRPCAddress` |
| `NodeDocumentACPOptionsBuilder.SetChainID` | `NodeDocumentACPOptionsBuilder.SetLogID` |
| `node.ErrSignerMissingForSourceHubACP` | `node.ErrSignerMissingForRemoteDAC` |
| `sourcehub.TxSigner` | `vera.TxSigner` |

#### C API

These mappings are included for migration completeness and do not change the C bindings' [current versioning status](./VERSIONING.md#the-c-embedded-client).

The Document ACP type stored in `NodeInitOptions.documentACPType` changes from `"source-hub"` to `"remote"`. The related fields are renamed as follows:

| Previous | Replacement |
|----------|-------------|
| `sourceHubChainID` | `remoteDACLogID` |
| `sourceHubGRPCAddress` | `remoteDACGRPCAddress` |
| `sourceHubCometRPCAddress` | `remoteDACCometRPCAddress` |
