## Реализация CRUD-методов для базы данных пользователей на PostgreSQL

### Реализованные методы:
`GET /service`

`GET /service/:id`

`POST /service`

`PUT /service/:id`

`DELETE /service/:id`

`GET /logs`

### Структура записи в базе данных пользователей:
```
{
  "name": "Alice Johnson",
  "email": "alice.johnson@gmail.com"
}
```

### Связи:
![335916395-76ef7dcc-ecf9-472e-b3f0-3edf8579746d](https://github.com/LonoGit/test_microservice/assets/168482657/eefcab8b-dfcc-4afc-bb68-f6f6f00232a1)


### Примеры запросов:

Создание пользователя:
```
curl -X POST http://localhost:8080/service \
-H "Content-Type: application/json" \
-d '{
  "name": " Alice Johnson ",
  "email": " alice.johnson@gmail.com"
}'
```

Изменение пользователя:
```
curl -X PUT http://localhost:8080/service/1 \
-H "Content-Type: application/json" \
-d '{
  "name": "Emily Smith",
  "email": "emily.smith@gmail.com"
}'
```

Удаление пользователя:
```
curl -X DELETE http://localhost:8080/service/1
```

Получение списка пользователей:
```
curl -X GET http://localhost:8080/service
```

Получение пользователя по индексу:
```
curl -X GET http://localhost:8080/service/1
```

Получение списка логов:
```
curl -X GET http://localhost:8080/logs
```

### Запуск:
Для локального запуска без использования `Docker`, необходимо наличие баз данных в `PostgreSQL` с именами:
- `users_db`
- `log_db`
