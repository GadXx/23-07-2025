# ZipCollector

**ZipCollector** — сервис для скачивания файлов по ссылкам, их архивации и выдачи пользователю zip-архива.

## Запуск

1. Установите Go 1.24.4 или новее.
2. Клонируйте репозиторий и перейдите в папку проекта.
3. Установите зависимости:
   ```sh
   go mod download
   ```
4. Создайте файл `.env` (или задайте переменные окружения):
   - `SESSION_DIR` — директория для временных файлов задач (например, `tmp`)
   - `ARCHIVE_DIR` — директория для хранения архивов (например, `archive`)
   - `QUEUE_SIZE` — размер очереди загрузки (по умолчанию 100)
   - `DOWNLOADER_WORKERS` — количество воркеров загрузки (по умолчанию 10)
5. Запустите приложение:
   ```sh
   go run ./cmd/main.go
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