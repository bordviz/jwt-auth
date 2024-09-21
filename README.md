# JWT-AUTH

## Запуск

Для запуска сервиса необходисмо выполнить команду:
`go run cmd/jwy-auth/main.go`

Или для запуцска через Docker выполнить команду:
`docker-compose up -d --build`

## Использование

> Токен передается в качестве Header (Пример: `Authorization: Bearer <token>`)

Для выполнения HTTP запросов вы можете воспользоваться [Postman коллекецией](./JWT.postman_collection.json)