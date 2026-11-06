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
|   |   `-- main.go               # local console client
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
|-- tests/
|   |-- cli/                      # console process tests
|   |-- core/                     # client workflow tests
|   |-- ratchet/                  # encryption and padding tests
|   `-- transport/                # host and message transport tests
|-- docs/
|   |-- architecture.md           # component responsibilities and data flows
|   |-- architecture.pdf          # unchanged source diagram
|   `-- dependencies.md           # selected libraries and integration constraints
|-- go.mod
`-- go.sum
```

Package names use lowercase words; filenames use lowercase words separated by
underscores where helpful. Tests live in `tests/`, grouped by component, and
exercise package APIs or the compiled console command.
Application packages are internal because this repository does not currently
expose a public Go library.

The module path is `github.com/utkkat/arkdy`, matching the repository's current
Git `origin`.

## Локальный чат из консоли

Нужен Go версии не ниже указанной в `go.mod`. Запускайте команды из корня проекта.

В первом терминале:

```sh
go run ./cmd/arkdy
```

Скопируйте полный адрес из строки `Адрес:` (включая `/p2p/<PeerID>`).
Во втором терминале подставьте его в команду:

```sh
go run ./cmd/arkdy -peer '/ip4/127.0.0.1/tcp/ПОРТ/p2p/PeerID'
```

Напишите сообщение во втором терминале и нажмите Enter. Оно появится в первом;
после первого входящего сообщения можно сразу отвечать. `[доставлено]` означает
подтверждение приёма вторым клиентом. Для выхода используйте `/quit`, Ctrl+C или
Ctrl+D (EOF). Ключи и переписка исчезают после завершения процесса.

Можно запустить оба клиента без `-peer`, затем в одном ввести
`/connect <полный адрес второго клиента>`. По умолчанию используется свободный
TCP-порт на `127.0.0.1`. Для фиксированного порта:

```sh
go run ./cmd/arkdy -listen /ip4/127.0.0.1/tcp/4001
```

| Команда | Действие |
| --- | --- |
| `/connect <multiaddr>` | Подключиться и выбрать собеседника. |
| `/to <PeerID>` | Выбрать известного собеседника для следующих сообщений. |
| `/peers` | Показать известных собеседников и текущий выбор. |
| `/addr` | Показать свой полный адрес. |
| `/help` | Показать команды. |
| `/quit` или `/exit` | Завершить клиент. |

Обычная строка отправляется выбранному собеседнику; строки с `/` являются командами.
Входящие сообщения от других узлов не меняют уже выбранного собеседника.
Сервер `arkdy-server`, блокчейн и внешние сервисы для этого сценария не нужны.

## Development status

The client now supports local encrypted console conversations. The standalone
relay command remains a placeholder; a relay is not needed for local chat.
Packages containing only declarations or documentation reserve the corresponding
architecture component for future work.

`transport.NewHost(ctx, transport.HostConfig{...})` creates a libp2p host with a
fresh ephemeral identity. `ListenAddrs` configures listening multiaddrs; an empty
list uses free TCP and QUIC ports on all IPv4 interfaces. For local runs, use
`[]string{"/ip4/127.0.0.1/tcp/0"}`. An empty `RelayAddr` skips relay connection;
otherwise provide a full multiaddr ending in `/p2p/<PeerID>`.
`RelayConnectTimeout` defaults to 15 seconds and respects the caller's context.
Initialization errors close the created host; after success, the caller must
call `Close`. The context controls initialization, not the returned host's lifetime.
Connecting to a relay does not reserve an inbound circuit or renew reservations.

`core.Client` connects to a full peer multiaddr and coordinates prekey exchange,
padding, libsignal encryption/decryption, and direct delivery. The transport
uses bounded go-msgio varint frames with Protobuf `BytesValue` payloads (20 KiB
including Protobuf overhead). Both sides use `/arkdy/direct/1.0.0`; bootstrap
bundles use `/arkdy/prekeys/1.0.0`. A successful send means the receiving core
accepted the decrypted message into its in-memory queue. It does not mean the
user read it or that it was saved. A lost acknowledgement can report an error
even after delivery; there is no automatic retry or persistent outbox.

`ratchet` uses `go.mau.fi/libsignal` v0.2.3 for signed prekeys and sessions, with
padding buckets of 256 B / 1 KiB / 4 KiB / 16 KiB. Text is limited to 16,380 UTF-8
bytes. A distinct one-time prekey is reserved for each bootstrap peer (up to 100
peers per process); libsignal consumes it on the first accepted prekey message.
Keys, pinned peer identities, session records and messages remain in memory.
Restarting changes both the transport identity and the prototype's Signal identity.
The eventual persistent IK belongs to ratchet with storage through vault/db.

The manually copied PeerID authenticates the libp2p connection. The Signal IK is
trusted on first use for that PeerID during the process lifetime. This local mode
does not implement username discovery, mailbox queues, hidden-sender envelopes,
persistent history, vault, or push. The local Signal payload contains protocol
metadata protected in transit by TLS/Noise; it must not be reused as a public
mailbox envelope. Circuit relay routing/reservations remain unimplemented.

`internal/relay.GetPrivateKey(path)` loads an existing libp2p key or generates and
saves a new Ed25519 key when the file is absent. It preserves unreadable or invalid
files and returns their error. The parent directory must already exist; saving
uses a hard link to publish the completed file without replacing another key.
The destination filesystem must support hard links (such as NTFS or ext4).

Unimplemented components retain dependency notes and commented
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
go test -race ./...
```

Только тесты из отдельной папки:

```sh
go test ./tests/...
```

Отдельно тест создания host:

```sh
go test ./tests/transport -run '^TestNewHost$' -v
```

Консольные тесты собирают клиент во временной папке и запускают отдельные процессы.
Для проверки покрытия исходных пакетов при внешнем расположении тестов:

```sh
go test -coverpkg=./internal/... ./tests/...
```
