package ratchet

// Выбран go.mau.fi/libsignal; шаблон импортов для SPK/OPK.
/*
import (
	// NewProtoBufSerializer задаёт внутреннюю сериализацию records.
	// https://pkg.go.dev/go.mau.fi/libsignal/serialize#NewProtoBufSerializer
	"go.mau.fi/libsignal/serialize"
	// GeneratePreKeys и GenerateSignedPreKey создают OPK и подписанный SPK.
	// https://pkg.go.dev/go.mau.fi/libsignal/util/keyhelper
	"go.mau.fi/libsignal/util/keyhelper"
)
*/
// SPK подписывается IK через выбранный движок; OPK используется однократно.
// Спецификация состава и проверки bundle: https://signal.org/docs/specifications/x3dh/
// Библиотечные records содержат секреты и перед сохранением должны проходить через vault.
// Публикация bundle относится к mailbox, защита закрытых ключей при хранении - к vault/db.
