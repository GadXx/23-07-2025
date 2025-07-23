# ZipCollector

**ZipCollector** — сервис для скачивания файлов по ссылкам, их архивации и выдачи пользователю zip-архива.

## Запуск

1. Установите зависимости:
   ```sh
   go mod download
   ```
2. Запустите приложение:
   ```sh
   go run ./cmd/main.go
   ```
3. Swagger
  ```sh
   http://localhost:8080/docs/
   ```

## Основные URL-эндпоинты

### 1. `POST /start`
Создать новую задачу на скачивание и архивирование файлов.
- **Request JSON:** `{ "link": "<url>" }`
- **Response JSON:** `{ "taskID": "..." }`

### 2. `POST /add/{taskID}`
Добавить ссылку к существующей задаче.
- **Request JSON:** `{ "link": "<url>" }`
- **Response JSON:** `{ "taskID": "..." }`

### 3. `GET /status/{taskID}`
Получить статус загрузки файлов по задаче.
- **Response JSON:**
  ```json
  {
    "success": true,
    "data": {
      "<url1>": "ok",
      "<url2>": "failed to download file"
    }
  }
  ```

### 4. `GET /zip/{taskID}`
Получить ссылку для скачивания архива задачи.
- **Response JSON:** `{ "archiveUrl": "http://<host>/archive/{taskID}.zip" }`

### 5. `GET /archive/{taskID}.zip`
Скачать zip-архив напрямую.
- **Response:** zip-файл

### 6. `DELETE /remove/{taskID}`
Удалить задачу.
- **Response JSON:** `{ "taskID": "..." }`

---