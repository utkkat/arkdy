# arkdy

arkdy is a decentralized p2p WEB-3.0 messenger that aims to be the best of all.

## Architecture

The [original architecture diagram](docs/architecture.pdf) defines the component
boundaries. See the [package mapping and data flows](docs/architecture.md) for how
the source tree follows that diagram.

The [selected libraries](docs/dependencies.md) document the implementation stack,
the reasons for each choice, and the remaining integration constraints.

## Project layout

The layout follows the [Go module organization guidance](https://go.dev/doc/modules/layout):
executable entry points live in `cmd/`, and application packages live in `internal/`.

```text
.
|-- cmd/
|   |-- arkdy/
|   |   `-- main.go               # client entry point (placeholder)
|   `-- arkdy-server/
|       `-- main.go               # libp2p relay entry point (placeholder)
|-- internal/
|   |-- core/                     # client workflow orchestration
|   |-- transport/                # ephemeral identity, host, direct/relay delivery
|   |-- ratchet/                  # IK, prekeys, padding, encryption/decryption
|   |-- vault/                    # local storage protection and sessions
|   |-- bchain/                   # publish/update signed identity records
|   |-- bchainfinder/             # find/verify public identity records
|   |-- mailbox/                  # encrypted queues, bundles, forwarding
|   |-- push/                     # device registration and push gateway
|   |-- db/                       # local persistence and outbox
|   `-- relay/                    # standalone libp2p relay server helpers
|-- docs/
|   |-- architecture.md           # component responsibilities and data flows
|   |-- architecture.pdf          # unchanged source diagram
|   `-- dependencies.md           # selected libraries and integration constraints
|-- go.mod
`-- go.sum
```

Package names use lowercase words; filenames use lowercase words separated by
underscores where helpful. Tests should live beside the package they exercise.
Application packages are internal because this repository does not currently
expose a public Go library.

The module path is `github.com/utkkat/arkdy`, matching the repository's current
Git `origin`.

## Development status

This repository is an early scaffold. Both command entry points are placeholders;
they do not start a client or a server yet. Packages containing only declarations
or documentation reserve the corresponding architecture component for future work.

`transport.NewHost(ctx, transport.HostConfig{...})` creates a libp2p host with a
fresh ephemeral identity. `ListenAddrs` configures listening multiaddrs; an empty
list uses free TCP and QUIC ports on all IPv4 interfaces. For local runs, use
`[]string{"/ip4/127.0.0.1/tcp/0"}`. An empty `RelayAddr` skips relay connection;
otherwise provide a full multiaddr ending in `/p2p/<PeerID>`.
`RelayConnectTimeout` defaults to 15 seconds and respects the caller's context.
Initialization errors close the created host; after success, the caller must
call `Close`. The context controls initialization, not the returned host's lifetime.
Connecting to a relay does not reserve an inbound circuit or renew reservations.

Direct and relay send helpers now share the same application protocol ID, but
message delivery is still unfinished: direct sending treats a PeerID as an
address, relay routing is not configured, and receive/framing are not implemented.

`internal/relay.GetPrivateKey(path)` loads an existing libp2p key or generates and
saves a new Ed25519 key when the file is absent. It preserves unreadable or invalid
files and returns their error. The parent directory must already exist; saving
uses a hard link to publish the completed file without replacing another key.
The destination filesystem must support hard links (such as NTFS or ext4).

Each Go file contains dependency notes and links to the libraries' documentation.
Active imports match the existing code. Unimplemented files contain commented
import examples: enable a dependency when its API is actually used, since Go
rejects unused imports. Library choices are recorded in `docs/dependencies.md`;
choosing a library does not implement its adapter or finalize the wire formats.
The comments distinguish library APIs from the X3DH and Double Ratchet
specifications and from this project's own package responsibilities.

Check package discovery, build, and tests from the repository root:

```sh
go list ./...
go build ./...
go test ./...
```
