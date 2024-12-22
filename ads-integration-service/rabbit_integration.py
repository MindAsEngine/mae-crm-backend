import datetime

from database import get_applications_by_id
import os
import time
import ads_integrations
import pika
import json
from dotenv import load_dotenv
import csv



def callback(ch, method, properties, body):
    try:
        status_queue = os.getenv("RABBITMQ_STATUS_QUEUE")
        result = process_message(ch, method, body)
        ch.basic_publish(exchange='',
                         routing_key=status_queue,
                         properties=pika.BasicProperties(correlation_id=properties.correlation_id),
                         body=json.dumps(result))
        print(f"Сообщение обработано: {result}")

    except Exception as e:
        print(f"Ошибка обработки сообщения: {e}")


def process_message(ch, method, body):
    try:
        data = json.loads(body)
        audience_id = data.get('audience_id')
        audience_name = data.get('audience_name', f'Audience_{audience_id}')
        application_ids = data.get('application_ids', [])
        integrations = data.get('integrations', [])
        print(f"Processing audience '{audience_id}' with {len(application_ids)} records and integrations {integrations}")

        applications = get_applications_by_id(application_ids)
        results = {"audience_id": audience_id}
        integrations_statuses = []
        for integration in integrations:
            cabinet = integration.get("cabinet_name")
            if cabinet == "yandex":
                integrations_statuses.append({"cabinet": "yandex",
                                              "status": ads_integrations.send_to_yandex_platform(audience_name, applications),
                                              "timestamp": datetime.datetime.now().isoformat()})
            if cabinet == "facebook":
                integrations_statuses.append({"cabinet": "facebook",
                                              "status": "Not implemented",
                                              "timestamp": datetime.datetime.now().isoformat()})
            if cabinet == "google":
                integrations_statuses.append({"cabinet": "google",
                                              "status": "Not implemented",
                                              "timestamp": datetime.datetime.now().isoformat()})
        if integrations_statuses.__len__() == 0:
            results.update({"error": "No integrations specified",
                                          "timestamp": datetime.datetime.now().isoformat()})
        else:
            results.update({"integrations": integrations_statuses})
        return results

    except Exception as ex:
        print(f"Message processing error: {str(ex)}")
        return {"error": str(ex), "timestamp": datetime.datetime.now().isoformat()}



if __name__ == '__main__':
    # Загрузка переменных окружения
    load_dotenv()

    RABBITMQ_HOST = os.getenv("RABBITMQ_HOST")
    RABBITMQ_PORT = os.getenv("RABBITMQ_PORT")  # Ensure the port is an integer
    RABBITMQ_USER = os.getenv("RABBITMQ_USER")
    RABBITMQ_PASSWORD = os.getenv("RABBITMQ_PASSWORD")
    RABBITMQ_QUEUE = os.getenv("RABBITMQ_QUEUE")
    RABBITMQ_EXCHANGE = os.getenv("RABBITMQ_EXCHANGE")
    RABBITMQ_STATUS_QUEUE = os.getenv("RABBITMQ_STATUS_QUEUE")

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
        channel.queue_declare(queue=RABBITMQ_STATUS_QUEUE, durable=True)
        # Создаем канал с подтверждением доставки

        channel.basic_consume(queue=RABBITMQ_QUEUE, on_message_callback=callback, auto_ack=True)
        print('Ожидание сообщений...')
        channel.start_consuming()
    except Exception as e:
        print(f"Ошибка подключения: {e}")
