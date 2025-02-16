# Магазин мерча

## Описание реализации
- В качестве библиотеки для HTTP сервера был выбран [echo](https://github.com/labstack/echo). Данная библиотека имеет встроенный logger и middleware.
- Для оптимизации запросов в базу данных были созданы индексы для полей, по которым часто извлекаются данные.
- В коде присутствуют кастомные ошибки, пользователь видит только одну из них. В случае, если возникла какая то проблемы, пользователь не будет видеть детали проблемы, а увидит лишь `Internal server error` или другую ошибку, связанную с данными.

## Запуск приложения
Для приложения написан `Dockerfile`, а также `docker-compose.yml`, в котором дополнительно поднимается контейнер с PostgreSQL и выполняется скрипт инициализации базы со всеми необходимыми таблицами.
1) Создайте конфиг в папке `config` в формате yaml
```yaml
port: 8080

storage:
  db_address: "localhost:5432"
  db_name: "db_market"
  db_user: "postgres"
  db_password: "postgres"
  db_sslmode: "disable"
```

2) Добавьте .env файл в корень проекта и укажите там значение `JWT_SECRET` (секретный ключ для генерации JWT токена).
3) Запустите сборку контейнера `docker-compose up -d --build`.


## Тестирование
Были написаны unit-тесты для бизнес-логики, [тестовое покрытие](https://github.com/ArtemSarafannikov/AvitoTestTask/blob/master/cover.html) составляет 97.7% пакета `service`.
```shell
go test ./internal/service
```
Также были написаны [интеграционные тесты](https://github.com/ArtemSarafannikov/AvitoTestTask/tree/master/internal/tests) для сценариев покупки мерча и передачи монеток другим сотрудникам. Для них созданы отдельные `Dockerfile.test` и `docker-compose-test.yml`.
Для запуска используйте
```shell
docker-compose -f docker-compose-test.yml up --build --abort-on-container-exit
```
Данная команда поднимет контейнер с PostgreSQL, запустит тесты и завершит выполнение контейнеров