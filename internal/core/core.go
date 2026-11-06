// Package core coordinates the client workflows between the architecture's
// transport, ratchet, vault, blockchain, mailbox, push, and database components.
// Client orchestration is not implemented yet.
package core

// Планируемые импорты координатора; подключать по мере реализации сценариев.
// Это внутренние пакеты проекта. Их обязанности описаны в ../../docs/architecture.md
// и комментариях соответствующих пакетов; команды просмотра: go doc ./internal/<пакет>.
/*
import (
	"context" // Отмена и тайм-ауты операций: https://pkg.go.dev/context

	"github.com/utkkat/arkdy/internal/bchain"       // Публикация и обновление identity-записи.
	"github.com/utkkat/arkdy/internal/bchainfinder" // Поиск и проверка identity-записи.
	"github.com/utkkat/arkdy/internal/db"           // Сохранение защищённых данных и outbox.
	"github.com/utkkat/arkdy/internal/mailbox"      // Очереди и доставка конвертов.
	"github.com/utkkat/arkdy/internal/push"         // Регистрация устройства для уведомлений.
	"github.com/utkkat/arkdy/internal/ratchet"      // X3DH и шифрование сообщений.
	"github.com/utkkat/arkdy/internal/transport"    // Сетевые соединения и прямой обмен.
	"github.com/utkkat/arkdy/internal/vault"        // Защита локальной истории и секретов.
)
*/
