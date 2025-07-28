/*
Package trustednet предоставляет security-мидлвары для HTTP-сервера.

Обеспечивает контроль доступа по IP-адресу через проверку вхождения
в доверенную подсеть. Основные сценарии использования:

1. Защита внутренних API эндпоинтов
2. Ограничение доступа к административным интерфейсам

Пример конфигурации nginx для корректной работы:

	location / {
	    proxy_set_header X-Real-IP $remote_addr;
	    proxy_pass http://backend;
	}
*/
package trustednet

import (
	"net"
	"net/http"
)

// CheckTrustedSubnet создает middleware для проверки доступа по доверенной подсети.
//
// Параметры:
//   - trustedSubnet: строка в формате CIDR (например, "192.168.1.0/24"),
//     определяющая доверенную подсеть. Если пустая строка - доступ запрещен для всех.
//
// Возвращает:
//   - Middleware функцию для chi.Router или стандартного http.Handler
//
// Логика работы:
//  1. Если trustedSubnet пустой - все запросы отклоняются с 403 Forbidden
//  2. Проверяет наличие заголовка X-Real-IP
//  3. Валидирует переданный IP-адрес
//  4. Проверяет вхождение IP в доверенную подсеть
//
// Коды ответа:
//   - 403 Forbidden:
//   - trustedSubnet пустой
//   - Отсутствует X-Real-IP
//   - IP не входит в доверенную подсеть
//   - 500 Internal Server Error: невалидный формат trustedSubnet
//
// Пример использования:
//
//	r := chi.NewRouter()
//	r.Use(CheckTrustedSubnet("192.168.1.0/24"))
//	r.Get("/admin", adminHandler)
//
// Требования:
//   - Перед middleware должен быть установлен proxy (например, nginx), который
//     добавляет заголовок X-Real-IP с реальным IP клиента
//   - Формат trustedSubnet должен быть валидным CIDR
func CheckTrustedSubnet(trustedSubnet string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if trustedSubnet == "" {
				http.Error(w, "Access denied", http.StatusForbidden)
				return
			}

			realIP := r.Header.Get("X-Real-IP")
			if realIP == "" {
				http.Error(w, "X-Real-IP header required", http.StatusForbidden)
				return
			}

			_, subnet, err := net.ParseCIDR(trustedSubnet)
			if err != nil {
				http.Error(w, "Invalid trusted subnet configuration", http.StatusInternalServerError)
				return
			}

			ip := net.ParseIP(realIP)
			if ip == nil || !subnet.Contains(ip) {
				http.Error(w, "Access denied", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
