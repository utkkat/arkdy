// Выбранные библиотеки; обоснование: ../../docs/dependencies.md.
//   - crypto/hpke: HPKE для sealed-истории на storage_pub, стандартная библиотека Go 1.26+.
//     https://pkg.go.dev/crypto/hpke
//   - crypto/ecdh: отдельная пара X25519 storage keys для HPKE.
//     https://pkg.go.dev/crypto/ecdh#X25519
//   - crypto/rand: случайные storage keys и salt.
//     https://pkg.go.dev/crypto/rand
//   - golang.org/x/crypto/argon2: IDKey выводит ключ из пароля, salt и параметров KDF.
//     https://pkg.go.dev/golang.org/x/crypto@v0.54.0/argon2#IDKey
//   - golang.org/x/crypto/chacha20poly1305: NewX для симметрического wrapping секретов.
//     https://pkg.go.dev/golang.org/x/crypto/chacha20poly1305#NewX
package vault

// Пример импортов для будущих файлов реализации; doc.go не требует активных импортов.
/*
import (
	"crypto/ecdh"
	"crypto/hpke"
	"crypto/rand"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/chacha20poly1305"
)
*/
