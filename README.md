# Файлы для итогового задания

В директории `tests` находятся тесты для проверки API, которое должно быть реализовано в веб-сервере.

Директория `web` содержит файлы фронтенда.

Суть проекта - реализовать функциональность планировщика задач.

Шаги, проделанные для реализации проекта:
1. Создан веб-сервер.
2. Спроектирована и создана БД, на основе указанных требований.
3. Написана функция для вычисления дат.
4. Реализованы различные обработчики для взаимодействия с задачами.
5. Подключена аутентификация.
6. Создан docker-образ.

Все задачи, включая те, что с повышенной сложностью, выполнены.

Main проверяет наличие БД и создает ее в случае отсутствия, название берется из env файла, 
если такой файл отсутствует или переменная окружения не задана, используется базовое значение. Также создается подключение к БД.
Инициализируются http обработчики для приложения и создается сервер на порте 7540.

api - содержит http обработчики для взаимодействия с задачами:
    nextDayHandler - получает следующую дату на основе правила повторения
    taskDoneHandler - удаляет выполненные задачи без правил повторения и переносит на новые даты те, у которых это правило есть.
    getSingleTaskHandler - получает задачу по id
    deleteTaskHandler - удаляет задачу
    authHandler - проверяет корректность введенного пароля и создает токен в случае успеха
    auth - проверяет авторизацию

http://localhost:7540/ - адрес сервера

переменные окружения хранятся в файле env и env_copy

Для локального запуска достаточно запустить функцию main

Для запуска тестов используются такие параметры, однако токен нужно будет поменять на свой после первой авторизации, он хранится в куках
var Port = 7540
var DBFile = "../scheduler.db"
var FullNextDate = true
var Search = true
var Token = `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJwYXNzd29yZF9oYXNoIjoiNTk5NDQ3MWFiYjAxMTEyYWZjYzE4MTU5ZjZjYzc0YjRmNTExYjk5ODA2ZGE1OWIzY2FmNWE5YzE3M2NhY2ZjNSJ9.KpzOXxQ3VFe8NXxxsYyIZQesk94p82sjQrNUSkV_T04`

Для создания и запуска docker-образа используются следующие команды:
$envVars = Get-Content .env | Where-Object { $_ -notmatch '^\s*#' -and $_ -match '=' } |
ConvertFrom-StringData

$port = $envVars.TODO_PORT
$dbfile = $envVars.TODO_DBFILE

docker build `
  --build-arg TODO_PORT=$port `
--build-arg TODO_DBFILE=$dbfile `
-t finapp .

docker run -d `
  -p "${port}:${port}" `
-e TODO_PORT=$port `
  -e TODO_DBFILE=$dbfile `
--name finapp_container `
finapp
