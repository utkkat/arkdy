// Документация библиотек для будущего SQL-хранилища:
//   - database/sql: DB, QueryContext, ExecContext и BeginTx для запросов и транзакций.
//     https://pkg.go.dev/database/sql
//   - context: отмена и дедлайны операций с базой.
//     https://pkg.go.dev/context
//   - modernc.org/sqlite: выбранный SQLite-драйвер без CGO.
//     https://pkg.go.dev/modernc.org/sqlite
//
// Клиентская база отделена от базы mailbox. Выбор драйвера и ограничения мобильной
// сборки описаны в ../../docs/dependencies.md.
package db

// Шаблон выбранных импортов; включить вместе с реализацией SQL-хранилища.
/*
import (
	"context"
	"database/sql"

	// Blank import регистрирует SQL-драйвер; это его реальный побочный эффект.
	_ "modernc.org/sqlite"
)
*/
