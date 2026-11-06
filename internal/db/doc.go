// Package db owns local persistence for sealed messages, general settings,
// contacts, wrapped ratchet state, wrapped prekeys, and the encrypted outbox.
// Database operations are not implemented yet.
//
// Документация библиотек для будущего SQL-хранилища:
//   - database/sql: DB, QueryContext, ExecContext и BeginTx для запросов и транзакций.
//     https://pkg.go.dev/database/sql
//   - context: отмена и дедлайны операций с базой.
//     https://pkg.go.dev/context
//   - modernc.org/sqlite: выбранный SQLite-драйвер без CGO.
//     https://pkg.go.dev/modernc.org/sqlite
//
// Начальные настройки: WAL, короткие транзакции, ограниченный пул и индексы запросов.
// Клиентская база отделена от базы mailbox. Выбор драйвера и ограничения мобильной
// сборки описаны в ../../docs/dependencies.md.
// Сохраняются sealed/wrapped/ciphertext-данные, подготовленные vault и ratchet.
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
