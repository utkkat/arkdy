// Package transport manages ephemeral libp2p identities, hosts, and direct or
// relayed envelope delivery. Transport identities are not stored on the blockchain.
//
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
//
// Этот файл содержит только документацию и не требует импортов.
// Импорты в Go действуют в пределах файла: используемые API подключены в файлах реализации.
// Send-функции пока пишут сырые байты. Новый framing и кодировка ещё не реализованы;
// при подключении go-msgio нужно одновременно обновить отправку и приём.
package transport
