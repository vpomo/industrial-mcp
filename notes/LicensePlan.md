# План реализации системы лицензирования

## Обзор архитектуры

Используется асимметричная криптография (RSA 2048 бит):
- **Закрытый ключ** — хранится у издателя лицензий, никогда не распространяется
- **Публичный ключ** — встроен в приложение, используется для проверки

Лицензия содержит: hardware_hash, expires_at, features, signature (подпись закрытым ключом)

---

## Этап 1: Генерация ключей ✅

### 1.1 Структура проекта
```
pkg/license/
├── keys/
│   ├── private.pem      # издатель генерирует и хранит в секрете
│   └── public.pem       # вшивается в приложение
├── gen.go              # генератор ключей
├── crypto.go           # RSA операции (Sign/Verify)
├── hardware.go         # сбор HW информации
├── license_file.go     # работа с файлом лицензии
├── validator.go        # валидация лицензии
├── generator.go        # создание лицензий
└── obfuscate.go        # обфускация публичного ключа
```

### 1.2 Алгоритм генерации ключей
```
1. openssl genrsa -out private.pem 2048
2. openssl rsa -in private.pem -pubout -out public.pem
```

**Статус**: ✅ Выполнено. Ключи сгенерированы в `pkg/license/keys/`

---

## Этап 2: Сбор аппаратной информации ✅

### 2.1 HardwareInfo struct
```go
type HardwareInfo struct {
    CPUID       string  // CPU processor ID
    MACAddr     string  // Primary network MAC
    VolumeID    string  // OS volume ID
    Motherboard string  // Motherboard serial
}
```

### 2.2 Сбор данных
- **Windows**: WMI запросы через powershell
- **Linux**: /proc/cpuinfo, ip link, blkid, dmidecode
- **Cross-platform**: machine-id fallback

### 2.3 HWID формирование
```
HWID = SHA256(CPUID + MACAddr + VolumeID + Motherboard)
```

**Статус**: ✅ Выполнено в `pkg/license/hardware.go`

---

## Этап 3: Структура файла лицензии ✅

### 3.1 license.dat (JSON)
```json
{
    "version": 1,
    "hardware_hash": "a1b2c3d4...",
    "issued_at": "2024-01-15T00:00:00Z",
    "expires_at": "2025-01-15T23:59:59Z",
    "features": ["basic", "pro", "enterprise"],
    "signature": "base64_encoded_rsa_signature",
    "issuer": "awwantil Licensing"
}
```

### 3.2 Подпись
```
payload = hardware_hash + "|" + expires_at + "|" + features_json
signature = RSA_Sign_SHA256(private_key, payload)
```

**Статус**: ✅ Выполнено в `pkg/license/license_file.go`

---

## Этап 4: Генерация лицензий (издатель) ✅

### 4.1 CLI утилита генератора
```
./license-tool create \
    --hardware-hash <hwid> \
    --expires "2027-01-01" \
    --features basic,pro \
    --output license.dat \
    --private-key keys/private.pem
```

### 4.2 Алгоритм
```
1. Получить hardware_hash от клиента
2. Создать JSON payload
3. Подписать payload закрытым ключом
4. Записать в license.dat
```

**Статус**: ✅ Выполнено. CLI утилита: `cmd/license-tool/main.go`

---

## Этап 5: Валидация лицензии (приложение) ✅

### 5.1 Проверки при старте
```
1. Загрузить license.dat
2. Проверить JSON синтаксис
3. Извлечь signature и данные
4. Verify signature = RSA_Verify(public_key, data, signature)
5. Проверить hardware_hash совпадает с текущим HWID
6. Проверить expires_at > now
7. Проверить features содержат необходимые
```

### 5.2 Периодическая проверка
- Каждые 60 минут перепроверять файл
- При изменении файла — перезагрузить и проверить

### 5.3 Обработка ошибок
```
ErrLicenseNotFound      -> Лицензия не найдена
ErrLicenseExpired       -> Срок истёк
ErrHardwareMismatch     -> HWID не совпадает
ErrInvalidSignature     -> Подпись неверна
ErrLicenseCorrupted     -> JSON повреждён
ErrFeatureMissing       -> Требуемая функция недоступна
```

**Статус**: ✅ Выполнено в `pkg/license/validator.go`

---

## Этап 6: Защита публичного ключа ✅

### 6.1 Проблема
Публичный ключ вшит в бинарник — злоумышленник может подменить.

### 6.2 Решение: Обфускация
```
1. Разбить ключ на 4 части
2. XOR с константой awwantil_license_2024_secure_key_factor
3. Собрать в runtime
```

**Статус**: ✅ Выполнено в `pkg/license/obfuscate.go`

---

## Этап 7: Интеграция в приложение ✅

### 7.1 REST endpoints
```
GET /api/v1/license/status       - статус лицензии
GET /api/v1/license/hwid         - HWID машины
GET /api/v1/license/check?feature=pro - проверка функции
```

### 7.2 Инициализация
```go
validator, err := license.New(publicKeyPEM, "/app/license/license.dat")
if err := validator.Validate(); err != nil {
    log.Fatal("License invalid:", err)
}
```

**Статус**: ✅ Выполнено. REST handlers: `internal/rest/c_license/`

---

## Этап 8: CLI утилиты ✅

### 8.1 Экспорт HWID
```
./license-tool export-hwid
> Hardware ID: a1b2c3d4e5f6...
```

### 8.2 Проверка лицензии
```
./license-tool verify --file license.dat --public-key public.pem
> License valid until 2027-01-01
> Features: basic, pro
> Signature: VALID
```

**Статус**: ✅ Выполнено. CLI утилита: `cmd/license-tool/main.go`

---

## Реализованные файлы

| Файл | Назначение | Статус |
|------|------------|--------|
| `pkg/license/keys/private.pem` | Закрытый ключ (издатель) | ✅ |
| `pkg/license/keys/public.pem` | Публичный ключ (приложение) | ✅ |
| `pkg/license/keys/gen.go` | Генератор ключей | ✅ |
| `pkg/license/hardware.go` | Сбор HW информации | ✅ |
| `pkg/license/crypto.go` | RSA подпись/верификация | ✅ |
| `pkg/license/license_file.go` | Чтение/запись .dat файла | ✅ |
| `pkg/license/validator.go` | Валидация лицензии | ✅ |
| `pkg/license/generator.go` | Генератор лицензий | ✅ |
| `pkg/license/obfuscate.go` | Обфускация ключа | ✅ |
| `cmd/license-tool/main.go` | CLI для создания/проверки лицензий | ✅ |
| `internal/rest/c_license/license_handler.go` | REST endpoints | ✅ |

---

## Тестирование ✅

1. ✅ Генерация ключей (openssl)
2. ✅ Создание лицензии с валидной подписью
3. ✅ Валидация валидной лицензии
4. ✅ Проверка отклонения при:
   - Истёкшем сроке ✅
   - Неверном HWID ✅
   - Неверной подписи ✅

---

## Развертывание

1. Издатель генерирует ключи (1 раз) — `pkg/license/keys/`
2. Публичный ключ компилируется в приложение
3. Клиент отправляет HWID: `./license-tool export-hwid`
4. Издатель генерирует: `./license-tool create --hardware-hash <hwid> --expires 2027-01-01 --features basic,pro`
5. Клиент кладёт license.dat в `/app/license/`

---

## Пример использования

### Издатель (создание лицензии):
```bash
# 1. Получить HWID от клиента
$ ./license-tool export-hwid
Hardware ID: d41db333ccaae17355d7dd370981d3adbd85b16d74380506d07e3bf58e276c28

# 2. Создать лицензию
$ ./license-tool create \
    --hardware-hash d41db333ccaae17355d7dd370981d3adbd85b16d74380506d07e3bf58e276c28 \
    --expires 2027-01-01 \
    --features basic,pro,enterprise \
    --output license.dat
License created successfully:
  Hardware Hash: d41db333ccaae17355d7dd370981d3adbd85b16d74380506d07e3bf58e276c28
  Expires: 2027-01-01 00:00:00
  Features: [basic pro enterprise]

# 3. Передать license.dat клиенту
```

### Клиент (установка лицензии):
```bash
# 1. Скопировать файл
$ sudo cp license.dat /app/license/

# 2. Проверить статус
$ curl http://localhost:8080/api/v1/license/status
{"enabled":true,"valid":true,"expires_at":"2027-01-01T00:00:00Z","features":["basic","pro","enterprise"],"days_remaining":227}
```

---

## Статус выполнения

| Этап | Описание | Статус |
|------|----------|--------|
| 1 | Генерация ключей | ✅ Выполнено |
| 2 | Сбор HW информации | ✅ Выполнено |
| 3 | Структура license.dat | ✅ Выполнено |
| 4 | Генератор лицензий | ✅ Выполнено |
| 5 | Валидация лицензии | ✅ Выполнено |
| 6 | Обфускация ключа | ✅ Выполнено |
| 7 | REST endpoints | ✅ Выполнено |
| 8 | CLI утилиты | ✅ Выполнено |
| Тестирование | | ✅ Выполнено |