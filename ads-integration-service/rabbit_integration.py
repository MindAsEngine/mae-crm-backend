from database import get_applications_by_id
import os
import time
import ads_integrations
import pika
import json
from dotenv import load_dotenv
import logging

logger = logging.getLogger()

def callback(ch, method, properties, body):
    try:
        process_message(ch, method, body)
    except Exception as e:
        print(f"Ошибка обработки сообщения: {e}")


def process_message(ch, method, body):
    try:
        data = json.loads(body)
        audience_id = data.get('audience_id')
        audience_name = data.get('name', f'Audience_{audience_id}')
        application_ids = data.get('application_ids', [])
        integration_names = data.get('integration_names', [])
        print(f"Processing audience '{audience_id}' with {len(application_ids)} records and integrations {integration_names}")
        # print(body)
        logger.info(f"Processing audience '{audience_id}' with {len(application_ids)} records and integrations {integration_names}")
        applications = get_applications_by_id(application_ids)
        results = {}

        if "facebook" in integration_names:
            results["facebook"] = ads_integrations.send_to_facebook_platform(audience_name, applications)

        if "google" in integration_names:
            results["google"] = ads_integrations.send_to_google_platform(audience_name, applications)

    except Exception as e:
        print(f"Message processing error: {str(e)}")
        logger.error(f"Message processing error: {str(e)}")


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
            except Exception as e:
            #except pika.exceptions.AMQPConnectionError as e:
                print(f"Ошибка соединения: {e}. Попробую снова через 5 секунд.")
                time.sleep(5)  # Пауза перед повторной попыткой подключения

        # Создаем канал с подтверждением доставки
        channel.basic_consume(queue=RABBITMQ_QUEUE, on_message_callback=callback, auto_ack=True)

        print('Ожидание сообщений...')
        channel.start_consuming()
    except Exception as e:
        print(f"Ошибка подключения: {e}")
