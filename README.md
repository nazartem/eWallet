Приложение EWallet реализует систему обработки транзакций платёжной системы. Приложение реализовано в виде HTTP сервера,
реализующее следующий REST API:

* POST   /api/send                     :  отправляет средства с одного из кошельков на указанный кошелек (сумма,
отправитель и получатель передаются в виде JSON-объекта в теле запроса)
* GET    /api/transactions?count=N     :  возвращает информацию о N последних по времени переводах средств
* GET    /api/wallet/{address}/balance :  возвращает информацию о балансе кошелька в JSON-объекте

Сервер можно запустить локально командой:

```bash
    $ SERVERPORT=4112 go run ./cmd/.
```

В качестве SERVERPORT можно использовать любой порт, который будет прослушивать локальный сервер в ожидании подключений.