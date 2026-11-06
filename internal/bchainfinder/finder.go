package bchainfinder

// Выбраны ethclient для EVM-поиска и libsignal/ecc для проверки подписи IK.
/*
import (
	"context" // Отмена и дедлайн поиска: https://pkg.go.dev/context
	"errors"  // Ошибки отсутствующей записи или неверной подписи: https://pkg.go.dev/errors

	// Чтение данных из EVM RPC; конкретная сеть и контракт ещё не заданы.
	// https://pkg.go.dev/github.com/ethereum/go-ethereum/ethclient
	"github.com/ethereum/go-ethereum/ethclient"
	// VerifySignature проверяет запись в том же формате ключей, что bchain и ratchet.
	// https://pkg.go.dev/go.mau.fi/libsignal/ecc#VerifySignature
	"go.mau.fi/libsignal/ecc"
)
*/
// Формат записи и проверка цепочки обновлений требуют реализации; выбор SDK не
// подтверждает принадлежность имени автоматически. См. ../../docs/dependencies.md.
