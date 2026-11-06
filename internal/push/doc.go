// Package push owns device token registration and gateway notifications.
// Push notifications contain no message content. Push operations are not
// implemented yet.
//
// Документация библиотек для HTTP-варианта gateway:
//   - context: отмена регистрации и отправки уведомлений.
//     https://pkg.go.dev/context
//   - net/http: Client.Do, NewRequestWithContext и обработчики HTTP.
//     https://pkg.go.dev/net/http
//   - firebase.google.com/go/v4/messaging: выбранный официальный SDK для FCM.
//     https://pkg.go.dev/firebase.google.com/go/v4/messaging
//   - github.com/sideshow/apns2: выбранный HTTP/2 APNs-клиент с token authentication.
//     https://pkg.go.dev/github.com/sideshow/apns2
//
// Выбраны net/http для gateway API и указанные SDK для провайдеров.
// При реализации SDK из блока ниже относятся к серверному адаптеру gateway;
// клиентская регистрация должна быть отдельным пакетом без серверных SDK/credentials.
// Формат зашифрованного push_blob и авторизация gateway ещё требуют реализации.
// Gateway отправляет пустое уведомление; текст сообщения остаётся в ciphertext.
package push

// Шаблон импортов серверного gateway; включить при реализации в отдельном пакете.
/*
import (
	"context"
	"net/http"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/sideshow/apns2"
)
*/
