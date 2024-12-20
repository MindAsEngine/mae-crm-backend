# Базовый образ Python
FROM python:3.10-slim
# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app
# Копируем файлы в контейнер
COPY requirements.txt ./
COPY . ./
# Установка зависимостей
RUN pip install --no-cache-dir -r requirements.txt
# Указываем команду для запуска приложения
CMD ["python", "rabbit_integration.py"]
