// Package bchain creates and updates signed public identity records containing
// a username, an identity public key, and an introduction mailbox address.
// Blockchain publication is not implemented yet.
package bchain

// Выбран ethclient для EVM-адаптера и libsignal/ecc для подписи тем же IK, что в ratchet.
// Шаблон импортов; публикация ещё не реализована.
/*
import (
	"context" // Отмена запроса публикации: https://pkg.go.dev/context
	"errors"  // Ошибки проверки записи и результата публикации: https://pkg.go.dev/errors

	// Типизированный RPC-клиент выбранного EVM-адаптера.
	// https://pkg.go.dev/github.com/ethereum/go-ethereum/ethclient
	"github.com/ethereum/go-ethereum/ethclient"
	// CalculateSignature подписывает байты записи ключом identity выбранного движка.
	// https://pkg.go.dev/go.mau.fi/libsignal/ecc#CalculateSignature
	"go.mau.fi/libsignal/ecc"
)
*/
// Сеть и контракт ещё не выбраны. Проверка подписи текущего IK на стороне цепочки
// является отдельной задачей: подпись EVM-транзакции её не заменяет.
// Формат подписываемых байтов и кодировка IK должны быть согласованы с ratchet.
// Ограничения выбранного адаптера: ../../docs/dependencies.md.
// Здесь публикуются username, IK_pub, intro_addr и signature; PeerID/IP сюда не входят.
