// Package ratchet owns identity keys, prekeys, X3DH, message padding, and
// Double Ratchet encryption and decryption. These operations are not implemented yet.
package ratchet

// Выбран движок go.mau.fi/libsignal для текущего прототипа; решение: ../../docs/dependencies.md.
// Шаблон импортов: включить вместе с реализацией.
/*
import (
	// GenerateIdentityKeyPair создаёт IK в формате выбранного движка.
	// https://pkg.go.dev/go.mau.fi/libsignal/util/keyhelper#GenerateIdentityKeyPair
	"go.mau.fi/libsignal/util/keyhelper"
)
*/
// IK является долгосрочной идентичностью пользователя, отдельно от libp2p PeerID.
// Для подписи identity-записей используется тот же формат ключей и libsignal/ecc:
// https://pkg.go.dev/go.mau.fi/libsignal/ecc
// Это сторонняя библиотека с лицензией GPL-3.0; интеграция ещё не реализована.
// Спецификация X3DH: https://signal.org/docs/specifications/x3dh/
// Состояние Double Ratchet: https://signal.org/docs/specifications/doubleratchet/
