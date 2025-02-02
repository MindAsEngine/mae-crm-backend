import logging

# Настройка логирования
logging.basicConfig(
    level=logging.INFO,  # Уровень логов: DEBUG, INFO, WARNING, ERROR, CRITICAL
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
    handlers=[
        logging.StreamHandler()  # Вывод логов в консоль (stdout)
    ]
)

# Пример использования
logger = logging.getLogger(__name__)