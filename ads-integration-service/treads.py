import datetime
import threading
from collections import defaultdict
from apscheduler.schedulers.background import BackgroundScheduler
import os
import time
import pika
import json

from apscheduler.triggers.date import DateTrigger
from dotenv import load_dotenv
from logger.logger import logger
from mysql.database import get_applications_by_id

from yandex.yandex import YandexIntegration




def connect_to_rabbit():
    credentials = pika.PlainCredentials(RABBITMQ_USER, RABBITMQ_PASSWORD)
    connection_params = pika.ConnectionParameters(
        host=RABBITMQ_HOST,
        port=RABBITMQ_PORT,
        credentials=credentials,
    )
    while True:
        try:
            connection = pika.BlockingConnection(connection_params)
            logger.info("Подключение к RabbitMQ установлено")
            ch = connection.channel()
            ch.queue_declare(queue=RABBITMQ_STATUS_QUEUE, durable=True)
            return ch
        except Exception as ex:
            print(f"Ошибка при подключении к RabbitMQ: {ex}. Попробую снова через 5 секунд.")
            logger.error(f"Ошибка при подключении к RabbitMQ: {ex}. Попробую снова через 5 секунд.")
            time.sleep(5)

def process_queue(ch, sch):
    ya_integration = YandexIntegration(oauth_token=os.getenv("YANDEX_OAUTH_TOKEN"))
    def process_message(data):
        try:
            audience_id = data.get('audience_id')
            audience_name = data.get('audience_name', f'Audience_{audience_id}')
            application_ids_to_delete = data.get('application_ids_to_delete', [])
            application_ids_to_add = data.get('application_ids_to_add', [])

            integrations = data.get('integrations', [])
            external_id = data.get('external_id', -1)

            applications_to_delete = get_applications_by_id(application_ids_to_delete)
            applications_to_add = get_applications_by_id(application_ids_to_add)

            results = {"audience_id": audience_id}
            integrations_statuses = []

            for integration in integrations:
                cabinet = integration.get("cabinet_name")
                if cabinet == "yandex":
                    send_result = ya_integration.send_audience(
                        audience_name=audience_name,
                        applications_for_delete=applications_to_delete,
                        application_for_add=applications_to_add,
                        external_id=external_id
                    )

                    integrations_statuses.append({"cabinet": "yandex",
                                                  "status": send_result,
                                                  "timestamp": datetime.datetime.now().isoformat()})
                    if send_result.result == "success":
                        def execute_confirm_yandex_platform(res):
                            ya_integration.confirm_yandex_platform(audience_name=res.name,
                                                                   audience_id=res.external_id)

                        task_name = f"confirm_yandex_{audience_id}"
                        run_at = send_result.time_to_confirm
                        scheduler.add_job(
                            execute_confirm_yandex_platform,
                            args=[send_result],
                            trigger=DateTrigger(run_date=run_at),
                            id=f"task_{task_name}_{run_at.timestamp()}"  # Уникальный ID задачи
                        )

                if cabinet == "facebook":
                    pass
                    # integrations_statuses.append({"cabinet": "facebook",
                    #                               "status": send_to_facebook_platform(audience_name,
                    #                                                                                    applications),
                    #                               "timestamp": datetime.datetime.now().isoformat()})
                if cabinet == "google":
                    pass
                    # integrations_statuses.append({"cabinet": "google",
                    #                               "status": send_to_google_platform(audience_name,
                    #                                                                                  applications),
                    #                               "timestamp": datetime.datetime.now().isoformat()})
            if integrations_statuses.__len__() == 0:
                results.update({"error": "No integrations specified",
                                "timestamp": datetime.datetime.now().isoformat()})
            else:
                results.update({"integrations": integrations_statuses})
            return results
        except Exception as ex:
            logger.error(f"Ошибка обработки сообщения: {str(ex)}")
            return {"error": str(ex), "timestamp": datetime.datetime.now().isoformat()}

    def callback(ch, method, properties, body):
        try:
            result = None
            message = json.loads(body)
            audience_id = message.get("audience_id")
            current_chunk = message.get('current_chunk')
            total_chunks = message.get('total_chunks')

            if audience_id is None:
                logger.error("audience_id не указан в сообщении")
                result = {"error": "audience_id не указан в сообщении", "timestamp": datetime.datetime.now().isoformat()}
            else:
                if current_chunk is None or total_chunks is None:
                    logger.error("current_chunk или total_chunks не указаны в сообщении")
                    result = {"error": "current_chunk или total_chunks не указаны в сообщении", "timestamp": datetime.datetime.now().isoformat()}
                else:
                    if current_chunk == 1:
                        message_storage[audience_id] = []
                    message_storage[audience_id].append(message)
                    if current_chunk == total_chunks:
                        with message_lock:
                            logger.info(f"Получены все части сообщения для audience_id={audience_id}")
                            application_ids_to_delete = []
                            application_ids_to_add = []
                            for message in message_storage[audience_id]:
                                application_ids_to_delete.extend(message.get('application_ids_to_delete', []))
                                application_ids_to_add.extend(message.get('application_ids_to_add', []))
                            processed_data = {
                                "audience_id": audience_id,
                                "external_id": message_storage[audience_id][0].get('external_id', -1),
                                "audience_name": message_storage[audience_id][0].get('audience_name',
                                                                                     f'Audience_{audience_id}'),
                                "application_ids_to_delete": application_ids_to_delete,
                                "application_ids_to_add": application_ids_to_add,

                                "integrations": message_storage[audience_id][0].get('integrations', [])
                            }
                            result = process_message(processed_data)
                        del message_storage[audience_id]
            if result:
                ch.basic_publish(exchange='', routing_key=RABBITMQ_STATUS_QUEUE,
                                 body=json.dumps(result))
                logger.info(f"Сообщение обработано: {result}")
        except Exception as ex:
            logger.error(f"Ошибка обработчика сообщения: {ex}")

    ch.basic_consume(queue=RABBITMQ_QUEUE, on_message_callback=callback, auto_ack=True)



if __name__ == '__main__':
    load_dotenv()
    RABBITMQ_HOST = os.getenv("RABBITMQ_HOST")
    RABBITMQ_PORT = os.getenv("RABBITMQ_PORT")  # Ensure the port is an integer
    RABBITMQ_USER = os.getenv("RABBITMQ_USER")
    RABBITMQ_PASSWORD = os.getenv("RABBITMQ_PASSWORD")
    RABBITMQ_QUEUE = os.getenv("RABBITMQ_QUEUE")
    RABBITMQ_EXCHANGE = os.getenv("RABBITMQ_EXCHANGE")
    RABBITMQ_STATUS_QUEUE = os.getenv("RABBITMQ_STATUS_QUEUE")

    scheduler = BackgroundScheduler()
    scheduler.start()

    message_storage = defaultdict(list)  # Ключ: audience_id, Значение: список сообщений
    message_lock = threading.Lock()
    channel = connect_to_rabbit()

    process_queue(channel, scheduler)
    try:
        channel.start_consuming()
    except Exception as ex:
        print(f"Ошибка при запуске consumer: {ex}")
        logger.error(f"Ошибка при запуске consumer: {ex}")
    finally:
        channel.close()
        logger.info("Подключение к RabbitMQ закрыто")
        print("Подключение к RabbitMQ закрыто")
