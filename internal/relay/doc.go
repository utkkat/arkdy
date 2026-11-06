// Package relay contains the existing standalone libp2p relay server helpers.
// It is separate from the mailbox queues and forwarding described in the diagram.
// Server initialization is not implemented yet.
// GetPrivateKey(path) loads the server identity or creates and saves it if absent.
//
// Документация существующих ключевых операций:
//   - os.ReadFile: чтение файла ключа.
//     https://pkg.go.dev/os#ReadFile
//   - libp2p crypto: генерация, сериализация и загрузка ключей узла.
//     https://pkg.go.dev/github.com/libp2p/go-libp2p@v0.50.0/core/crypto
//
// Для будущего запуска сервера libp2p.New создаёт host, а circuit relay.New
// включает сервис ретрансляции. Освобождение ресурсов выполняется через Close.
//   - https://pkg.go.dev/github.com/libp2p/go-libp2p@v0.50.0#New
//   - https://pkg.go.dev/github.com/libp2p/go-libp2p@v0.50.0/p2p/protocol/circuitv2/relay#New
//
// Circuit relay обслуживает транспортные соединения; очереди mailbox здесь не создаются.
package relay

// Пример импортов для будущего файла запуска сервиса; doc.go содержит только документацию.
/*
import (
	"github.com/libp2p/go-libp2p"
	circuitrelay "github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
)
*/
