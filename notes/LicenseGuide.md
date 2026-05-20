# Руководство по работе с лицензионной системой

## Содержание
1. [Обзор системы](#обзор-системы)
2. [Генерация ключей](#генерация-ключей)
3. [Создание лицензии](#создание-лицензии)
4. [Установка лицензии](#установка-лицензии)
5. [Проверка лицензии](#проверка-лицензии)
6. [Диагностика проблем](#диагностика-проблем)
7. [FAQ](#faq)

---

## Обзор системы

Система использует асимметричную криптографию RSA-2048:
- **Закрытый ключ** — хранится у издателя лицензий
- **Публичный ключ** — встроен в приложение

Лицензия привязана к аппаратной конфигурации ПК (HWID) и имеет срок действия.

---

## Генерация ключей

### Шаг 1: Создать директорию для ключей

```bash
mkdir -p /opt/imatic-license/keys
cd /opt/imatic-license/keys
```

### Шаг 2: Генерировать закрытый ключ

```bash
openssl genrsa -out private.pem 2048
```

### Шаг 3: Извлечь публичный ключ

```bash
openssl rsa -in private.pem -pubout -out public.pem
```

### Шаг 4: Защитить закрытый ключ

```bash
chmod 600 private.pem
# Хранить в защищённом месте, никогда не передавать клиентам
```

### Шаг 5: Проверить ключи

```bash
# Посмотреть публичный ключ
cat public.pem

# Проверить закрытый ключ
openssl rsa -in private.pem -check
```

---

## Создание лицензии

### Шаг 1: Получить HWID от клиента

Клиент запускает команду:

```bash
cd /opt/imatic-license
./imatic-license export-hwid
```

Результат:
```
Hardware ID: A1B2C3D4E5F67890...
```

Клиент передаёт этот HWID издателю.

### Шаг 2: Создать файл лицензии

```bash
cd /opt/imatic-license
./license-generator create \
    --hardware-hash A1B2C3D4E5F67890 \
    --expires 2025-12-31 \
    --features basic,pro,enterprise \
    --output /tmp/license.dat
```

### Шаг 3: Проверить созданную лицензию

```bash
./license-generator verify --file /tmp/license.dat
```

Ожидаемый вывод:
```
License valid
Expires: 2025-12-31 23:59:59
Features: basic, pro, enterprise
Signature: OK
```

### Шаг 4: Передать файл клиенту

```bash
# Клиенту отправляется только license.dat
scp /tmp/license.dat client@example.com:/tmp/
```

---

## Установка лицензии

### На сервере приложения

```bash
# Создать директорию для лицензии
sudo mkdir -p /app/license

# Скопировать файл лицензии
sudo cp license.dat /app/license/

# Установить права доступа
sudo chmod 644 /app/license/license.dat
```

### Проверка установки

```bash
ls -la /app/license/
cat /app/license/license.dat
```

### Структура license.dat

```json
{
    "version": 1,
    "hardware_hash": "A1B2C3D4E5F67890...",
    "issued_at": "2024-01-15T10:30:00Z",
    "expires_at": "2025-12-31T23:59:59Z",
    "features": ["basic", "pro", "enterprise"],
    "signature": "base64_encoded_signature...",
    "issuer": "iMatic Licensing"
}
```

---

## Проверка лицензии

### Через REST API

```bash
curl http://localhost:8080/api/v1/license/status
```

Ответ:
```json
{
    "valid": true,
    "expires_at": "2025-12-31T23:59:59Z",
    "features": ["basic", "pro", "enterprise"],
    "days_remaining": 365
}
```

### Через CLI

```bash
./imatic-license verify
```

### Получить HWID машины

```bash
./imatic-license export-hwid
```

---

## Диагностика проблем

### Ошибка: "license not found"

**Причина**: Файл `/app/license/license.dat` отсутствует.

**Решение**:
```bash
sudo cp license.dat /app/license/license.dat
sudo chmod 644 /app/license/license.dat
```

### Ошибка: "hardware mismatch"

**Причина**: HWID машины не совпадает с HWID в лицензии.

**Решение**:
1. Проверить HWID машины: `./imatic-license export-hwid`
2. Сравнить с HWID в файле лицензии
3. Если машина была изменена (новая материнская плата, сетевая карта) — нужна новая лицензия

### Ошибка: "license expired"

**Причина**: Срок действия лицензии истёк.

**Решение**:
1. Проверить системное время: `date`
2. Если время верное — обратиться к издателю за продлением
3. Для продления: издатель создаёт новую лицензию с новой датой

### Ошибка: "invalid signature"

**Причина**: Лицензия была изменена или создана неверным закрытым ключом.

**Решение**:
1. Проверить, что лицензия получена от официального издателя
2. Проверить, что публичный ключ в приложении не был изменён

### Ошибка: "json parse error"

**Причина**: Файл лицензии повреждён.

**Решение**:
```bash
# Проверить синтаксис JSON
cat /app/license/license.dat | python3 -m json.tool

# Переустановить лицензию
sudo cp license.dat /app/license/
```

---

## Обновление лицензии

### Продление срока

```bash
# Издатель создаёт новую лицензию с новой датой
./license-generator create \
    --hardware-hash A1B2C3D4E5F67890 \
    --expires 2026-12-31 \
    --features basic,pro \
    --output new_license.dat

# Клиент заменяет файл
sudo cp new_license.dat /app/license/license.dat
```

### Добавление функций

```bash
./license-generator create \
    --hardware-hash A1B2C3D4E5F67890 \
    --expires 2025-12-31 \
    --features basic,pro,enterprise,analytics \
    --output new_license.dat
```

---

## Удаление лицензии

```bash
# Удалить файл лицензии
sudo rm /app/license/license.dat

# Приложение перестанет работать (если лицензия обязательна)
```

---

## Резервное копирование

### Бэкап лицензии
```bash
sudo cp /app/license/license.dat /backup/license.dat
```

### Бэкап ключей (издатель)
```bash
tar -czf keys-backup.tar.gz private.pem public.pem
gpg -c keys-backup.tar.gz  # Зашифровать паролем
```

---

## FAQ

### Q: Что делать при смене оборудования?

A: Если HWID изменился, старая лицензия недействительна. Нужно:
1. Получить новый HWID на новой машине
2. Обратиться к издателю за новой лицензией
3. Предоставить доказательства покупки

### Q: Можно ли перенести лицензию на другую машину?

A: Нет, лицензия привязана к HWID. При переносе на другую машину нужна новая лицензия.

### Q: Что если приложение не может прочитать файл?

A: Проверить права доступа:
```bash
ls -la /app/license/license.dat
# Должно быть: -rw-r--r--
```

### Q: Как проверить целостность лицензии?

A: Использовать команду верификации:
```bash
./imatic-license verify --file /app/license/license.dat --verbose
```

### Q: Сколько стоит генерация лицензии?

A: Операция локальная, стоимость — только вычислительные ресурсы (~0.1 сек на CPU).

### Q: Можно ли использовать один закрытый ключ для многих клиентов?

A: Да, один закрытый ключ издателя может генерировать лицензии для всех клиентов.

---

## Контакты для поддержки

При ошибках, не описанных выше, обращаться к команде разработки с:
- Логом ошибок (`/app/logs/`)
- HWID машины
- Файлом license.dat (если не повреждён)