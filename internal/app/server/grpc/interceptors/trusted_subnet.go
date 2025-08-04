// internal/server/grpc/interceptors/trusted_subnet.go
package interceptors

import (
	"context"
	"net"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// TrustedSubnetConfig содержит конфигурацию для интерцептора проверки доверенных подсетей.
//
// Используется для ограничения доступа к gRPC-методам только из определенных IP-подсетей.
// Применяется в TrustedSubnetInterceptor.
//
// Поля:
//
//   - TrustedSubnet: строка в CIDR-нотации, определяющая доверенную подсеть.
//     Формат: "IPv4/маска" или "IPv6/маска" (например "192.168.1.0/24").
//     Пустая строка означает отсутствие настроенной подсети.
//
//   - ProtectedMethods: map[string]bool, где ключи - это имена gRPC-методов,
//     которые требуют проверки доступа. Поддерживает два формата:
//
//   - Полное имя метода (например "/shortener.Shortener/Stats")
//
//   - Префикс сервиса (например "/shortener.Admin/" для всех методов сервиса)
//     Примечание: регистрозависимый поиск.
//
//   - DenyIfNotConfigured: флаг, определяющий поведение при отсутствии настроенной подсети.
//
//   - true - возвращать ошибку PermissionDenied
//
//   - false - пропускать запрос без проверки
//     Рекомендуемое значение для production: true.
type TrustedSubnetConfig struct {
	TrustedSubnet string
	// Методы, требующие проверки (например: ["/shortener.Shortener/Stats"])
	ProtectedMethods map[string]bool
	// Блокировать если подсеть не настроена (true) или пропускать (false)
	DenyIfNotConfigured bool
}

// TrustedSubnetInterceptor возвращает gRPC-интерцептор для контроля доступа по доверенным подсетям.
//
// Интерцептор выполняет:
//   - Проверку принадлежности IP-адреса клиента к указанной доверенной подсети
//   - Выборочное применение проверки только к защищенным методам (см. TrustedSubnetConfig)
//   - Гибкую настройку поведения при отсутствии конфигурации подсети
//
// Параметры:
//   - cfg: конфигурация интерцептора (TrustedSubnetConfig):
//   - TrustedSubnet: CIDR-нотация подсети (например "192.168.1.0/24")
//   - ProtectedMethods: список защищаемых gRPC-методов
//   - DenyIfNotConfigured: блокировать вызовы если подсеть не настроена
//
// Возвращает:
//   - grpc.UnaryServerInterceptor: настроенный интерцептор контроля доступа
//
// Логика работы:
//  1. Проверка метода на наличие в ProtectedMethods
//  2. Если подсеть не задана:
//     - DenyIfNotConfigured=true: возвращает PermissionDenied
//     - DenyIfNotConfigured=false: пропускает запрос
//  3. Определение IP-адреса клиента (X-Forwarded-For → X-Real-IP → peer)
//  4. Проверка принадлежности IP к доверенной подсети
//  5. При отказе: возвращает PermissionDenied с детальным описанием
//
// Пример использования:
//
//	interceptor := TrustedSubnetInterceptor(TrustedSubnetConfig{
//	    TrustedSubnet: "10.0.0.0/8",
//	    ProtectedMethods: map[string]bool{
//	        "/shortener.Shortener/Stats": true,
//	    },
//	    DenyIfNotConfigured: true,
//	})
//	server := grpc.NewServer(grpc.ChainUnaryInterceptor(interceptor))
//
// Особенности:
//   - Поддерживает CIDR-нотацию для IPv4/IPv6
//   - Совместим с прокси (X-Forwarded-For/X-Real-IP)
//   - Гибкая настройка через ProtectedMethods (поддержка wildcards)
//   - Детализированные сообщения об ошибках
//
// Рекомендации:
//   - Для публичных методов исключайте проверку через ProtectedMethods
//   - В production всегда устанавливайте DenyIfNotConfigured=true
//   - Логируйте факты отказа в доступе
func TrustedSubnetInterceptor(cfg TrustedSubnetConfig) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Проверяем, требуется ли проверка для этого метода
		if !cfg.isMethodProtected(info.FullMethod) {
			return handler(ctx, req)
		}

		// Если подсеть не задана
		if cfg.TrustedSubnet == "" {
			if cfg.DenyIfNotConfigured {
				return nil, status.Error(codes.PermissionDenied, "trusted subnet not configured")
			}
			return handler(ctx, req)
		}

		// Получаем IP клиента
		clientIP, err := getClientIP(ctx)
		if err != nil {
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}

		// Парсим доверенную подсеть
		_, subnet, err := net.ParseCIDR(cfg.TrustedSubnet)
		if err != nil {
			return nil, status.Error(codes.Internal, "invalid trusted subnet configuration")
		}

		// Проверяем принадлежность IP к подсети
		ip := net.ParseIP(clientIP)
		if ip == nil || !subnet.Contains(ip) {
			return nil, status.Error(codes.PermissionDenied, "access denied: IP not in trusted subnet")
		}

		return handler(ctx, req)
	}
}

func (c *TrustedSubnetConfig) isMethodProtected(method string) bool {
	// Проверяем полное совпадение метода
	if c.ProtectedMethods[method] {
		return true
	}

	// Проверяем совпадение по префиксу (для группировки методов)
	for m := range c.ProtectedMethods {
		if strings.HasPrefix(method, m) {
			return true
		}
	}

	return false
}

func getClientIP(ctx context.Context) (string, error) {
	// 1. Пробуем получить из X-Forwarded-For (если есть прокси)
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if forwardedIPs := md.Get("x-forwarded-for"); len(forwardedIPs) > 0 {
			return strings.Split(forwardedIPs[0], ",")[0], nil
		}
		if realIPs := md.Get("x-real-ip"); len(realIPs) > 0 {
			return realIPs[0], nil
		}
	}

	// 2. Получаем из peer информации
	if p, ok := peer.FromContext(ctx); ok {
		switch addr := p.Addr.(type) {
		case *net.TCPAddr:
			return addr.IP.String(), nil
		case *net.UDPAddr:
			return addr.IP.String(), nil
		default:
			return addr.String(), nil
		}
	}

	return "", status.Error(codes.PermissionDenied, "could not determine client IP")
}
