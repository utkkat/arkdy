// Документация библиотек для будущей реализации:
//   - context: отмена Put, Fetch, Forward и запросов prekey bundle.
//     https://pkg.go.dev/context
//   - crypto/rand: случайные секретные идентификаторы очередей.
//     https://pkg.go.dev/crypto/rand#Read
//   - time: срок хранения конвертов и тайм-ауты операций.
//     https://pkg.go.dev/time
//   - database/sql + modernc.org/sqlite: отдельная база узла для очередей и bundles.
//     https://pkg.go.dev/database/sql
//     https://pkg.go.dev/modernc.org/sqlite
//   - google.golang.org/protobuf/proto: выбранная кодировка запросов и конвертов.
//     https://pkg.go.dev/google.golang.org/protobuf/proto
package mailbox

// Планируемые импорты; включить при реализации соответствующих операций.
/*
import (
	"context"
	"crypto/rand"
	"database/sql"
	"time"

	"google.golang.org/protobuf/proto"
	_ "modernc.org/sqlite" // Регистрация драйвера отдельной базы mailbox.
)
*/
