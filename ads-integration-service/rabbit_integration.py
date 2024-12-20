import datetime
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
        # Распаковка сообщения
        # data = json.loads(body)
        # print(f"Получены данные: {data}")
        # Обработка данных
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
        print(body)
        logger.info(f"Processing audience '{audience_id}' with {len(application_ids)} records and integrations {integration_names}")

        results = {}

        if "facebook" in integration_names:
            try:
                facebook_id = ads_integrations.send_to_facebook_platform(audience_name, application_ids)
                results["facebook"] = {
                    "status": "success",
                    "audience_id": facebook_id
                }
            except Exception as e:
                print(f"Facebook integration error: {str(e)}")
                logger.error(f"Facebook integration error: {str(e)}")
                results["facebook"] = {
                    "status": "error",
                    "message": str(e)
                }

        if "google" in integration_names:
            try:
                google_id = ads_integrations.send_to_google_platform(audience_name, application_ids)
                results["google"] = {
                    "status": "success",
                    "audience_id": google_id
                }
            except Exception as e:
                print(f"Google integration error: {str(e)}")
                logger.error(f"Google integration error: {str(e)}")
                results["google"] = {
                    "status": "error",
                    "message": str(e)
                }

        if "yandex" in integration_names:
            try:
                yandex_id = ads_integrations.send_to_yandex_platform(audience_name, application_ids)
                results["yandex"] = {
                    "status": "success",
                    "audience_id": yandex_id
                }
            except Exception as e:
                print(f"Yandex integration error: {str(e)}")
                logger.error(f"Yandex integration error: {str(e)}")
                results["yandex"] = {
                    "status": "error",
                    "message": str(e)
                }

        # # Send results back to reporting service
        # send_status_update(audience_id, results)
        #
        # ch.basic_ack(delivery_tag=method.delivery_tag)
        
    except Exception as e:
        print(f"Message processing error: {str(e)}")
        logger.error(f"Message processing error: {str(e)}")
        # ch.basic_nack(delivery_tag=method.delivery_tag, requeue=False)

# def send_status_update(audience_id, results):
#     try:
#         channel.basic_publish(
#             exchange='audiences',
#             routing_key='audience.status',
#             body=json.dumps({
#                 'audience_id': audience_id,
#                 'results': results,
#                 'timestamp': datetime.datetime.now().isoformat()
#             })
#         )
#     except Exception as e:
#         logger.error(f"Failed to send status update: {str(e)}")


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
        #channel.queue_declare(queue=RABBITMQ_QUEUE)
        #channel.exchange_declare(exchange=RABBITMQ_EXCHANGE, exchange_type='direct')
        #channel.queue_bind(exchange=RABBITMQ_EXCHANGE, queue=RABBITMQ_QUEUE)

        # Создаем канал с подтверждением доставки
        channel.basic_consume(queue=RABBITMQ_QUEUE, on_message_callback=callback, auto_ack=True)

        print('Ожидание сообщений...')
        channel.start_consuming()
    except Exception as e:
        print(f"Ошибка подключения: {e}")
