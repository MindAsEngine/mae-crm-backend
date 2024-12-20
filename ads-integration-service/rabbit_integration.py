import datetime
import os
import time

import pika
import json
from dotenv import load_dotenv


# def push_test_message(_channel, queue_name, exchange_name):
#     message = {
#         "audience_id": 61,
#         "applications": [
#             {
#                 "id": 4953867,
#                 "created_at": "2024-01-10T00:00:00Z",
#                 "updated_at": "2024-01-10T10:13:59Z",
#                 "status_name": "Нецелевой",
#                 "manager_id": 77394,
#                 "client_id": 4095583,
#                 "status_id": 3,
#                 "reason_name": "Конкурент",
#                 "status_reason_id": 10602,
#                 "client_data": {
#                     "fio": "Иванов Иван Иванович",
#                     "phone": "+79999999999",
#                     "birth_place": "Москва"
#                 }
#             }
#         ]
#     }


def callback(ch, method, properties, body):
    try:
        # Распаковка сообщения
        data = json.loads(body)
        print(f"Получены данные: {data}")
        # Обработка данных
        process_message(data)
    except Exception as e:
        print(f"Ошибка обработки сообщения: {e}")


def process_message(data):
    # Пример обработки данных
    audience_id = data.get('audience_id', 'Неизвестно')
    application_ids = data.get('application_ids', [])
    integration_names = data.get('integration_names', [])
    print(f"Обрабатываем аудиторию '{audience_id}' с {len(application_ids)} записями и интеграциями {integration_names}")

    if "facebook" in integration_names:{

    }
    if "google" in integration_names:{

    }
    if "yandex" in integration_names:{

    }
    # Если аудитория есть в 
    # Отправка данных в рекламный кабинет
    send_to_ad_platform(audience_id, application_ids)


def send_to_facebook_platform(audience_name, application_ids):
    # Здесь будет код для отправки данных в API рекламных кабинетов
    pass

def get_from_facebook_platform():
    # Здесь будет код для получения данных из API рекламных кабинетов
    pass

def send_to_google_platform(audience_name, application_ids):
    # Здесь будет код для отправки данных в API рекламных кабинетов
    pass

def get_from_google_platform():
    # Здесь будет код для получения данных из API рекламных кабинетов
    pass

def send_to_yandex_platform(audience_name, application_ids):
    # Здесь будет код для отправки данных в API рекламных кабинетов
    pass

def get_from_yandex_platform():
    # Здесь будет код для получения данных из API рекламных кабинетов
    pass

if __name__ == '__main__':
    # Загрузка переменных окружения
    load_dotenv()

    RABBITMQ_HOST = os.getenv("RABBITMQ_HOST")
    RABBITMQ_PORT = os.getenv("RABBITMQ_PORT")  # Ensure the port is an integer
    RABBITMQ_USER = os.getenv("RABBITMQ_USER")
    RABBITMQ_PASSWORD = os.getenv("RABBITMQ_PASSWORD")
    RABBITMQ_QUEUE = os.getenv("RABBITMQ_QUEUE")
    RABBITMQ_EXCHANGE = os.getenv("RABBITMQ_EXCHANGE")

    # Подключение к RabbitMQ
    credentials = pika.PlainCredentials(RABBITMQ_USER, RABBITMQ_PASSWORD)

    connection_params = pika.ConnectionParameters(
        host=RABBITMQ_HOST,
        port=RABBITMQ_PORT,
        credentials=credentials,
    )

    try:
        while True:
            try:
                connection = pika.BlockingConnection(connection_params)
                channel = connection.channel()
                break  # Если подключение удалось, выходим из цикла
            except pika.exceptions.AMQPConnectionError as e:
                print(f"Ошибка соединения: {e}. Попробую снова через 5 секунд.")
                time.sleep(5)  # Пауза перед повторной попыткой подключения
        channel.queue_declare(queue=RABBITMQ_QUEUE)
        channel.exchange_declare(exchange=RABBITMQ_EXCHANGE, exchange_type='direct')
        channel.queue_bind(exchange=RABBITMQ_EXCHANGE, queue=RABBITMQ_QUEUE)

        # Создаем канал с подтверждением доставки
        channel.basic_consume(queue=RABBITMQ_QUEUE, on_message_callback=callback, auto_ack=True)

        # Отправка тестового сообщения
        push_test_message(channel, RABBITMQ_QUEUE, RABBITMQ_EXCHANGE)
        print('Ожидание сообщений...')
        channel.start_consuming()
    except Exception as e:
        print(f"Ошибка подключения: {e}")
