# Базовый образ Python
FROM python:3.10-slim

# Установка системных зависимостей
RUN apt-get update && apt-get install -y \
    gcc \
    libmariadb-dev \
    libpq-dev \
    pkg-config \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем файлы в контейнер
COPY requirements.txt ./
COPY . ./

# Установка pip и других необходимых инструментов
RUN pip install --upgrade pip setuptools wheel

# Установка Python-зависимостей
RUN pip install --no-cache-dir -r requirements.txt

# Указываем команду для запуска приложения
CMD ["python", "rabbit_integration.py"]
