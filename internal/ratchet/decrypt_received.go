package ratchet

// Выбран go.mau.fi/libsignal; шаблон импортов для расшифровки.
/*
import (
	"context" // Контекст операции расшифровки: https://pkg.go.dev/context

	// SignalMessage и PreKeySignalMessage описывают обычное и начальное сообщение.
	// https://pkg.go.dev/go.mau.fi/libsignal/protocol
	"go.mau.fi/libsignal/protocol"
	// Cipher.Decrypt и DecryptMessage обрабатывают соответствующий тип сообщения.
	// https://pkg.go.dev/go.mau.fi/libsignal/session
	"go.mau.fi/libsignal/session"
)
*/
// Сначала адаптер разбирает защищённый внешний конверт, затем передаёт данные движку.
// Stores получают и сохраняют защищённое состояние через vault/db, без открытых records на диске.
// Работа с out-of-order сообщениями и skipped keys относится к состоянию ratchet:
// https://signal.org/docs/specifications/doubleratchet/
// Этот файл расшифровывает сетевой конверт; открытие локальной sealed-истории относится к vault.
