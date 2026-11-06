package ratchet

// Выбран session.Cipher из go.mau.fi/libsignal; шаблон импортов для реализации.
/*
import (
	"context" // Контекст операции Encrypt: https://pkg.go.dev/context

	// CiphertextMessage представляет результат шифрования, SignalAddress - адрес сессии.
	// https://pkg.go.dev/go.mau.fi/libsignal/protocol
	"go.mau.fi/libsignal/protocol"
	// Cipher.Encrypt выполняет шифрование через состояние сессии.
	// https://pkg.go.dev/go.mau.fi/libsignal/session#Cipher.Encrypt
	"go.mau.fi/libsignal/session"
)
*/
// Хранилища сессий реализуются через vault/db: https://pkg.go.dev/go.mau.fi/libsignal/state/store
// Состояние и outbox нужно сохранять согласованно. Самостоятельная замена алгоритмов
// движка на XChaCha из vault не предусмотрена.
// Сериализация libsignal содержит заголовки и не является готовым приватным конвертом:
// внешний формат и защита метаданных описаны как задача интеграции в ../../docs/dependencies.md.
// Алгоритм и состояние: https://signal.org/docs/specifications/doubleratchet/
// На вход поступает padded message; на выходе ожидаются header, ciphertext и новое state.
