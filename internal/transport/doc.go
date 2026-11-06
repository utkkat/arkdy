// Документация используемых библиотек:
//   - libp2p: создание и настройка сетевого узла.
//     https://pkg.go.dev/github.com/libp2p/go-libp2p@v0.50.0
//   - core/host: соединения, потоки и регистрация обработчиков протокола.
//     https://pkg.go.dev/github.com/libp2p/go-libp2p@v0.50.0/core/host
//   - core/network: поток Stream и его жизненный цикл.
//     https://pkg.go.dev/github.com/libp2p/go-libp2p@v0.50.0/core/network
//   - go-multiaddr: разбор составных адресов libp2p.
//     https://pkg.go.dev/github.com/multiformats/go-multiaddr@v0.16.1
//   - go-msgio: выбранный framing с varint-длиной и ограничением размера при чтении.
//     https://pkg.go.dev/github.com/libp2p/go-msgio#NewVarintReaderSize
//   - protobuf: выбранная бинарная кодировка конвертов, типы генерируются из .proto.
//     https://pkg.go.dev/google.golang.org/protobuf/proto
package transport
