package transport

// Выбраны go-msgio и Protobuf; шаблон импортов будущего обработчика потока.
/*
import (
	"time" // Дедлайн чтения потока: https://pkg.go.dev/time

	// network.Stream предоставляет Read, SetReadDeadline, Close и Reset.
	// https://pkg.go.dev/github.com/libp2p/go-libp2p@v0.50.0/core/network#Stream
	"github.com/libp2p/go-libp2p/core/network"
	// NewVarintReaderSize ограничивает длину входящего кадра до чтения его тела.
	// https://pkg.go.dev/github.com/libp2p/go-msgio#NewVarintReaderSize
	"github.com/libp2p/go-msgio"
	// Unmarshal декодирует кадр в сгенерированный тип Envelope после проверки размера.
	// https://pkg.go.dev/google.golang.org/protobuf/proto#Unmarshal
	"google.golang.org/protobuf/proto"
)
*/
// Выбран кадр varint(length) + protobuf bytes; .proto-схема и обработчик ещё не реализованы.
// При использовании Reader.ReleaseMsg буфер освобождается после обработки сообщения.
// После чтения конверт передаётся в сценарий расшифровки; transport не хранит ratchet-ключи.
