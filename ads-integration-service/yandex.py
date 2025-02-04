import datetime

import requests
from logger import logger
from csv_prepare import prepare_csv


class YandexIntegration:
    def __init__(self, oauth_token=None):
        self.base_url = "https://api-audience.yandex.ru/v1/management/"
        self.headers = {"Authorization": f"OAuth {oauth_token}"}

    def get_audiences(self):
        response = requests.get(
            url=self.base_url + 'segments',
            headers=self.headers
        )
        if response.status_code != 200:
            logger.error(f"Ошибка получения все аудиторий Яндекс.Аудитории: {response.text}")
            return None
        logger.debug(f"Получены аудитории Яндекс.Аудитории: {response.json()}")
        return response.json()

    def update_audience_name_by_id(self, audience_id, name):
        response = requests.put(
            url=self.base_url + f'segment/{audience_id}',
            headers=self.headers,
            json={"segment":
                {
                    "name": name,
                }}
        )
        if response.status_code != 200:
            logger.error(f"Ошибка обновления имени аудитории Яндекс.Аудитории: {response.text}")
            raise Exception(f"Ошибка обновления имени аудитории Яндекс.Аудитории: {response.text}")
        logger.debug(f"Имя аудитории успешно обновлено: {name}")
        return name

    def confirm_yandex_platform(self, audience_id, audience_name):
        response = requests.post(
            url=self.base_url + f'segment/{audience_id}/confirm',
            headers=self.headers,
            json={"segment":
                {
                    "name": audience_name,
                    "content_type": "crm",
                    "hashing_alg": "SHA256",
                    "hashed": True,
                }
            }
        )
        if response.status_code != 200:
            logger.error(f"Ошибка подтверждения аудитории Яндекс.Аудитории: {response.text}")
            raise Exception(f"Ошибка подтверждения аудитории Яндекс.Аудитории: {response.text}")
        logger.debug(f"Аудитория успешно подтверждена")
        return response.json()

    def upload_applications(self, applications, audience_name):
        url = self.base_url + 's/upload_csv_file'
        csv_file = prepare_csv(applications, audience_name)
        response = requests.post(
            url=url,
            headers=self.headers,
            files={"file": open(csv_file, "rb")}
        )
        if response.status_code != 200:
            logger.error(f"Ошибка загрузки данных в аудиторию Яндекс.Аудитории: {response.text}")
            return {"result": "error", "message": response.text}
        logger.debug(f"Данные успешно загружены в новую аудиторию")
        json_response = response.json()
        external_id = json_response["segment"]["id"]
        status = json_response["segment"]["status"]
        time_to_confirm = self.calculate_time_to_confirm()
        return {
            "result": "success",
            "external_id": external_id,
            "status": status,
            "name": audience_name,
            "time_to_confirm": time_to_confirm
        }

    def update_applications(self, application_ids, audience_id, audience_name, mode="replace"):
        csv_file = prepare_csv(application_ids, audience_name)
        url = self.base_url + f'{audience_id}/modify_data'
        response = requests.post(
            url=url,
            headers=self.headers,
            params={'modification_type': mode},
            files={"file": open(csv_file, "rb")}
        )
        if response.status_code != 200:
            logger.error(f"Ошибка загрузки данных в аудиторию Яндекс.Аудитории: {response.text}")
            return {"result": "error", "message": response.text}
        logger.debug(f"Данные успешно загружены в существующую аудиторию")
        json_response = response.json()
        status = json_response["segment"]["status"]
        time_to_confirm = self.calculate_time_to_confirm()
        return {
            "result": "success",
            "external_id": audience_id,
            "status": status,
            "name": audience_name,
            "time_to_confirm": time_to_confirm
        }

    def calculate_time_to_confirm(self, seconds=5, minutes=0, hours=0):
        return datetime.datetime.now() + datetime.timedelta(seconds=seconds, minutes=minutes, hours=hours)

    def send_audience(self, audience_name,
                      applications_for_delete,
                      application_for_add,
                      external_id=None):
        try:
            if external_id:
                return self.update_applications(
                    application_ids=application_for_add,
                    mode='add',
                    audience_id=external_id,
                    audience_name=audience_name)
            audiences = self.get_audiences()
            if audiences:
                segments = audiences.get("segments", [])
                audience = next((aud for aud in segments if aud["name"] == audience_name), None)
                if audience:
                    return self.update_applications(
                        application_ids=application_for_add,
                        audience_id=audience["id"],
                        audience_name=audience_name)
                else:
                    return self.upload_applications(application_for_add, audience_name)
        except Exception as ex:
            logger.error(f"Ошибка отправки данных в Яндекс.Аудитории: {str(ex)}")
            return {"result": "error", "message": str(ex)}
