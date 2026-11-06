# Architecture and package boundaries

The [architecture diagram](architecture.pdf) is an unchanged copy of the
[PDF linked by the original README](https://github.com/user-attachments/files/33307452/architecture.2.drawio.4.pdf).
The table below maps its components to Go packages. It describes the intended
responsibilities; most of these operations are still unimplemented.

The [selected libraries](dependencies.md) fit within these package boundaries and
record which integration details remain unfinished.

| Diagram component | Go package | Responsibility |
| --- | --- | --- |
| transport | `internal/transport` | Generate ephemeral transport keys, start a libp2p host, receive direct envelopes, and send direct or via relay. |
| ratchet | `internal/ratchet` | Generate long-term IK and prekeys, initiate/respond to X3DH, pad messages, and encrypt/decrypt using Double Ratchet state. |
| vault | `internal/vault` | Create a vault, wrap/unwrap secrets, seal/open local history, and unlock/lock storage sessions. |
| bchain | `internal/bchain` | Create and update signed public username, IK public key, and introduction mailbox records. |
| bchain_finder | `internal/bchainfinder` | Find identity records by username and verify signatures. |
| mailbox | `internal/mailbox` | Create queues, publish/get prekey bundles, put/forward/fetch/delete encrypted envelopes. |
| push | `internal/push` | Register device tokens and notify through a gateway using an opaque push blob. |
| db | `internal/db` | Store sealed history, settings, contacts, wrapped ratchet state and prekeys, and the encrypted outbox. |

`internal/core` preserves the existing client core placeholder. It is the place
for coordinating the first-contact and message workflows shown at the bottom of
the diagram; it does not replace the individual components above.

`internal/relay` preserves the existing `arkdy-server` key helpers for a standalone
libp2p relay. A libp2p circuit relay and the diagram's mailbox forwarding node have
different responsibilities. Mailbox queues and bundle storage belong to
`internal/mailbox`.

`cmd/arkdy` and `cmd/arkdy-server` are reserved for application startup and wiring.
Their `main` functions are placeholders until startup is implemented.

## Placement changes

- `cmd/arkdy/core.go` becomes `internal/core/core.go`.
- `cmd/arkdy/transport` becomes `internal/transport`; filenames are normalized to
  lowercase with underscores, including the corrected `receive_direct.go`.
- `cmd/arkdy/bchain/bchain.go` becomes `internal/bchain/bchain.go`.
- `cmd/arkdy/bchain/bchain_finder.go` becomes `internal/bchainfinder/finder.go`,
  retaining the diagram's separate lookup component.
- The old `cmd/arkdy/vault` files describe IK/prekey generation, padding, and
  message encryption/decryption. They belong to `internal/ratchet` according to
  the diagram. `internal/vault` is reserved for local storage protection.
- The helpers previously in `cmd/arkdy-server/main.go` become
  `internal/relay/identity.go`.
- `vault`, `mailbox`, `push`, and `db` have package documentation placeholders;
  the reorganization does not implement these services.

## Data flows to preserve

Transport creates a new ephemeral PeerID at startup. No transport PeerID or IP
address is published on the blockchain. Long-term identity keys belong to
`ratchet`; wrapping and persistent storage belong to `vault` and `db`.

For first contact, look up and verify the recipient's identity, obtain and verify
their mailbox prekey bundle, initiate X3DH, and exchange contact requests through
the introduction queue. The recipient responds to X3DH and creates a contact
queue, as shown in the diagram.

For outgoing messages, seal a local pending copy, pad/encrypt the message, and
store its envelope in the outbox. Normal delivery uses the sender's mailbox and
then the recipient's mailbox, followed by an empty push notification. On receipt,
fetch/decrypt the envelope, seal/store the local message, and delete/acknowledge
the delivered envelope. Direct transport is an opt-in path using the same
encrypted envelope.

The database stores protected representations: sealed history, wrapped ratchet
state and prekeys, and encrypted outbox envelopes. Decrypting message envelopes
is a ratchet responsibility; opening sealed local history is a vault
responsibility.
