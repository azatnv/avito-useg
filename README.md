# Сервис динамического сегментирования пользователей

Запуск сервиса из корня проекта
```bash
docker-compose up -d
```
Остановка
```bash
docker-compose down [-v]
```

## API

Работа с сервисом производится по адресу:

```bash
http://localhost/
```

### 1. Создание пользователя

#### POST /users
```json
{"id":  1000}
```
или несколько сразу 
```json
[
  {"id":  1000},
  {"id":  1001},
  {"id":  1002},
  {"id":  1003}
]
```

### 2.Создание сегмента

#### POST /segments
```json
{"name":  "AVITO_BLUE_BUTTON"}
```

### 3.Удаление сегмента
#### DELETE /segments
```json
{"name":  "AVITO_BLUE_BUTTON"}
```

### 4. Добавление сегментов пользователю
#### POST /users/segments
```json
{
  "id": 1000,
  "segments": [
    {"name": "AVITO_VOICE_MESSAGES"},
    {"name": "AVITO_PERFORMANCE_VAS"},
    {"name":  "AVITO_DISCOUNT_30"},
    {"name":  "AVITO_DISCOUNT_50"}
  ]
}
```
можно указать дату окончания действия сегмента:

```json
{
  "id": 1000,
  "segments": [
    {"name": "TEST_SEGMENT_DATE_END"}
  ],
  "date_end": "2023-09-21T10:00:00Z"
}
```
### 5. Получить сегменты пользователя
#### GET /users/segments
```json
{
  "id": 1000
}
```
### 6. Добавить новый сегмент и присвоить его определённому проценту пользователей
#### POST /segments?percent=40
```json
{
  "name": "AVITO_40_PERCENT_AUDITORY"
}
```
