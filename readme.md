### Инструкция по запуску
1) Из Docker: Находясь в папке Crypto_Rest_test необходимо при запущенном Docker написать команду в терминал `docker-compose up --build`
2) Без Docker: Находясь в папке Crypto_Rest_test/app написать команду в терминал `go run /cmd/main.go`

### Инструкция по использованию
1) Через Swagger по адресу http://localhost:8080/swagger/index.html 
2) Через postman (json с настройками) в самом низу

### Особенности
1) В ТЗ не требовалось создание хендлера выдающий список всех отслеживаемых монет, но я его сделал на всякий случай по адресу `/currency/watchlist`
2) `/currency/add` поддерживает ввод сразу нескольких криптовалют через запятую, к примеру: `btc,usdt,eth`
3) `/currency/remove` поддерживает ввод сразу нескольких криптовалют через запятую, к примеру: `btc,usdt,eth`

### Тестовое задание
Микросервис, собирающий, хранящий и отображающий стоимости криптовалют.

`/currency/add` - Добавление криптовалюты в список наблюдения
`/currency/remove` - Удаление криптовалюты из списка наблюдения

Добавление криптовалюты в список наблюдения подразумевает что мы собираем и записываем цену криптовалюты в локальную базу раз в N секунд. Удаление из этого списка останавливает сбор цены. Для получения актуальной цены можно воспользоваться открытыми API.

`/currency/price` - Получение цены криптовалюты

Должна быть возможность получить цену конкретной криптовалюты в конкретный момент времени из локальной базы. Например: на запрос {"coin": "BTC", "timestamp": 1736500490}, сервис должен вернуть стоимость BTC в момент времени 1736500490. Если не удалось найти стоимость в конкретный момент времени, возвращаем стоимость в ближайший к запрошенному моменту времени.

Приложение должно быть упаковано в Docker Compose. В качестве базы необходимо использовать Postgres.

В Readme.md описать инструкцию по запуску. Плюсом будет наличие swagger либо какой-то базовой документации.

### Файл конфигурации Postman
```json
{
	"info": {
		"_postman_id": "4311585a-c616-4cff-ad36-64bed32d76ce",
		"name": "Crypto_rest_test",
		"schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json",
		"_exporter_id": "37801451"
	},
	"item": [
		{
			"name": "add coins",
			"request": {
				"auth": {
					"type": "noauth"
				},
				"method": "POST",
				"header": [],
				"body": {
					"mode": "raw",
					"raw": "{\r\n    \"coins\": \"btc,usdt,eth\"\r\n}",
					"options": {
						"raw": {
							"language": "json"
						}
					}
				},
				"url": {
					"raw": "localhost:8080/currency/add",
					"host": [
						"localhost"
					],
					"port": "8080",
					"path": [
						"currency",
						"add"
					]
				},
				"description": "you may add more than 1 coin once, just put \",\" between coins name"
			},
			"response": []
		},
		{
			"name": "watch list",
			"request": {
				"auth": {
					"type": "noauth"
				},
				"method": "GET",
				"header": [],
				"url": {
					"raw": "localhost:8080/currency/watchlist",
					"host": [
						"localhost"
					],
					"port": "8080",
					"path": [
						"currency",
						"watchlist"
					]
				},
				"description": "Retrieves all watched at this moment coins"
			},
			"response": []
		},
		{
			"name": "get price",
			"protocolProfileBehavior": {
				"disableBodyPruning": true
			},
			"request": {
				"auth": {
					"type": "noauth"
				},
				"method": "GET",
				"header": [],
				"body": {
					"mode": "raw",
					"raw": "{\r\n    \"coin\": \"eth\",\r\n    \"timestamp\": \"1737122365\"\r\n}",
					"options": {
						"raw": {
							"language": "json"
						}
					}
				},
				"url": {
					"raw": "localhost:8080/currency/price?coin=eth&timestamp=1737119353",
					"host": [
						"localhost"
					],
					"port": "8080",
					"path": [
						"currency",
						"price"
					],
					"query": [
						{
							"key": "coin",
							"value": "eth"
						},
						{
							"key": "timestamp",
							"value": "1737119353"
						}
					]
				}
			},
			"response": []
		},
		{
			"name": "remove coins from watchlist",
			"request": {
				"auth": {
					"type": "noauth"
				},
				"method": "DELETE",
				"header": [],
				"body": {
					"mode": "raw",
					"raw": "{\r\n    \"coins\": \"usdt,btc\"\r\n}",
					"options": {
						"raw": {
							"language": "json"
						}
					}
				},
				"url": {
					"raw": "localhost:8080/currency/remove",
					"host": [
						"localhost"
					],
					"port": "8080",
					"path": [
						"currency",
						"remove"
					]
				}
			},
			"response": []
		}
	]
}
```