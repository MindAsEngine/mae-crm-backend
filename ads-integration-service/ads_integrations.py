import csv

from facebook_business.adobjects.adaccount import AdAccount
from facebook_business.api import FacebookAdsApi
from facebook_business.adobjects.customaudience import CustomAudience
# from google.ads.googleads.client import GoogleAdsClient
# from google.ads.googleads.errors import GoogleAdsException
import requests
import os
from logger import logger


import hashlib


def hash_data(data):
    return hashlib.sha256(data.encode('utf-8')).hexdigest()

def prepare_facebook_user_data(application):
    user = {}
    email = application.get("contacts_buy_emails", None)
    if email:
        user.update({"email": hash_data(email)})
    phone = application.get("contacts_buy_phones", None)
    if phone:
        user.update({"phone": hash_data(phone.replace("+.", ""))})
    gender = application.get("contacts_buy_sex", None)
    if gender:
        user.update({"gender": hash_data(gender)})
    country = application.get("contacts_buy_geo_country_name", "Узбекистан")
    if country:
        user.update({"country": hash_data(country)})
    city = application.get("contacts_buy_geo_city_name", "Ташкент")
    if city:
        user.update({"city": hash_data(city)})

    name_full = application.get("contacts_buy_name", None)
    if name_full:
        words = name_full.split()
        user.update({"first_name": hash_data(words[0])})
        if len(words) > 1:
            user.update({"last_name": hash_data(words[1])})
    name_first = application.get("name_first", None)
    if name_first:
        user.update({"first_name": hash_data(name_first)})
    name_last = application.get("name_last", None)
    if name_last:
        user.update({"last_name": hash_data(name_last)})
    dob = application.get("contacts_buy_dob", None)
    if dob:
        user.update({"dob": hash_data(dob.__str__())})
    return user

def get_audiences(account_id):
    try:
        ad_account = AdAccount(account_id)
        audiences = ad_account.get_custom_audiences()
        return [aud for aud in audiences]
    except Exception as e:
        print(f"Facebook Ads get audiences error: {str(e)}")
        raise

def update_audience(audience_id, application_ids):
    try:
        audience = CustomAudience(audience_id)
        users = [prepare_facebook_user_data(app) for app in application_ids]
        audience.add_users(CustomAudience.Schema.phone_hash, users)
        return True
    except Exception as e:
        print(f"Facebook Ads update error: {str(e)}")
        raise

def send_audience(account_id, audience_name, application_ids):
    try:
        ad_account = AdAccount(account_id)
        existing_audiences = get_audiences(account_id)
        if existing_audiences:
            audience = next((aud for aud in existing_audiences if aud["name"] == audience_name), None)
            if audience:
                return update_audience(audience["id"], application_ids)
        audience = ad_account.create_custom_audience(
            fields=[CustomAudience.Field.id],
            params={
                CustomAudience.Field.name: audience_name,
            }
        )
        users = [prepare_facebook_user_data(app) for app in application_ids]
        audience.add_users(CustomAudience.Schema.phone_hash, users)
        return {
            "audience_id": audience["id"],
            "result": "success"
        }
    except Exception as e:
        print(f"Facebook Ads error: {str(e)}")
        return {
            "result": "error",
            "message": str(e)
        }

def send_to_facebook_platform(audience_name, application_ids):
    FacebookAdsApi.init(
        access_token=os.getenv("FB_ACCESS_TOKEN"),
        app_id=os.getenv("FB_APP_ID"),
        app_secret=os.getenv("FB_APP_SECRET")
    )
    return send_audience(
        account_id=os.getenv("FB_ACCOUNT_ID"),
        audience_name=audience_name,
        application_ids=application_ids
    )


def prepare_google_user_data(application):
    user_data = {}
    # Хэширование email
    email = application.get("contacts_buy_emails")
    if email:
        user_data["hashed_email"] = hash_data(email.lower().strip())
    # Хэширование телефона
    phone = application.get("contacts_buy_phones")
    if phone:
        user_data["hashed_phone_number"] = hash_data(phone.strip())
    return user_data


class GoogleAdsIntegration:
    def __init__(self, client=None, customer_id=None):
        self.client = client
        self.customer_id = customer_id

    # def create_audience(self, audience_name):
    #     try:
    #         user_list_service = self.client.get_service("UserListService")
    #         user_list_operation = self.client.get_type("UserListOperation")
    #
    #         user_list = user_list_operation.create
    #         user_list.name = audience_name
    #         user_list.description = "Customer Match audience created via API"
    #         user_list.membership_life_span = 30
    #
    #         response = user_list_service.mutate_user_lists(
    #             customer_id=self.customer_id,
    #             operations=[user_list_operation]
    #         )
    #         audience_resource_name = response.results[0].resource_name
    #         print(f"Создана аудитория: {audience_resource_name}")
    #         return audience_resource_name
    #     except GoogleAdsException as ex:
    #         print(f"Ошибка создания аудитории Google Ads: {str(ex)}")
    #         raise

#     def update_audience(self, audience_resource_name, applications):
#         try:
#             offline_user_data_job_service = self.client.get_service("OfflineUserDataJobService")
#             offline_user_data_job_operation = self.client.get_type("OfflineUserDataJobOperation")
#             offline_user_data_job = offline_user_data_job_service.create_offline_user_data_job(
#                 customer_id=self.customer_id,
#                 job_type="CUSTOMER_MATCH_USER_LIST"
#             )
#             job_resource_name = offline_user_data_job.resource_name
#             print(f"Создана Offline User Data Job: {job_resource_name}")
#
#             # Создание операций добавления данных
#             operations = []
#             for app in applications:
#                 user_data = prepare_google_user_data(app)
#                 user_data_operation = offline_user_data_job_operation.create
#                 if "hashed_phone_number" in user_data:
#                     self.client.get_type("UserIdentifier").hashed_phone_number = user_data["hashed_phone_number"]
#                     user_data_operation.user_identifiers.append(
#                     self.client.get_type("UserIdentifier").hashed_phone_number
#                     )
#                 operations.append(user_data_operation)
#
#             # Добавление пользователей в аудиторию
#             offline_user_data_job_service.add_offline_user_data_job_operations(
#                 resource_name=job_resource_name,
#                 operations=operations
#             )
#             offline_user_data_job_service.run_offline_user_data_job(
#                 resource_name=job_resource_name
#             )
#             print(f"Пользователи успешно добавлены в аудиторию: {audience_resource_name}")
#             return True
#         except Exception as ex:
#             print(f"Ошибка обновления аудитории Google Ads: {str(ex)}")
#             raise
#
#
#     def send_audience(self, audience_name, applications):
#         try:
#             # Проверяем существование аудитории
#             query = f"""
#                 SELECT user_list.resource_name, user_list.name
#                 FROM user_list
#                 WHERE user_list.name = '{audience_name}'
#             """
#             response = self.client.get_service("GoogleAdsService").search_stream(
#                 customer_id=self.customer_id, query=query
#             )
#             existing_audience = None
#             for batch in response:
#                 for row in batch.results:
#                     existing_audience = row.user_list.resource_name
#
#             # Если аудитория уже существует, обновляем её
#             if existing_audience:
#                 return self.update_audience(existing_audience, applications)
#
#             # Создаём новую аудиторию
#             audience_resource_name = self.create_audience(audience_name)
#             self.update_audience(audience_resource_name, applications)
#             return {"result": "success", "audience_id": audience_resource_name}
#         except GoogleAdsException as ex:
#             print(f"Ошибка Google Ads API: {str(ex)}")
#             return {"result": "error", "message": str(ex)}
# def send_to_google_platform(audience_name, application_ids):
#     google_integration = GoogleAdsIntegration(
#         client=GoogleAdsClient.load_from_storage("GOOGLE_ADS_YAML_PATH"),
#         customer_id=os.getenv("GOOGLE_CUSTOMER_ID")
#     )
#     return google_integration.send_audience(audience_name, application_ids)

def create_csv_file(applications):
    results = []
    for app in applications:
        user_data = prepare_facebook_user_data(app)
        results.append(user_data)
    with open('yandex.csv', 'w', newline='') as file:
        writer = csv.writer(file)
        writer.writerow(results[0].keys())
        for result in results:
            writer.writerow(result.values())
        return file.name


class YandexIntegration:
    def __init__(self, oauth_token=None):
        self.base_url = "https://api-audience.yandex.ru/v1/management/segment"
        self.headers = {"Authorization": f"OAuth {oauth_token}"}

    def get_audiences(self):
        response = requests.get(
            url=self.base_url+'s',
            headers=self.headers
        )
        if response.status_code != 200:
            logger.error(f"Ошибка получения все аудиторий Яндекс.Аудитории: {response.text}")
            raise Exception(f"Ошибка получения все аудиторий Яндекс.Аудитории: {response.text}")
        logger.debug(f"Получены аудитории Яндекс.Аудитории: {response.json()}")
        return response.json()

    def update_name(self, audience_id, name):
        response = requests.post(
            url=self.base_url + f'/{audience_id}/confirm',
            headers=self.headers,
            json={"segment":
                      {
                          "name": name,
                          "content_type": "crm",
                          "id": audience_id,
                          "hashing_alg": "SHA256",
                          "hashed": True,
                              }}
        )
        if response.status_code != 200:
            logger.error(f"Ошибка обновления имени аудитории Яндекс.Аудитории: {response.text}")
            raise Exception(f"Ошибка обновления имени аудитории Яндекс.Аудитории: {response.text}")
        print(f"Имя аудитории успешно обновлено: {name} {response.text}")
        logger.debug(f"Имя аудитории успешно обновлено: {name}")
        return name




    def upload_applications(self, applications, audience_id=None, audience_name=None):
        if not audience_id:
            url = self.base_url + 's/upload_csv_file'
        else:
            url = self.base_url + f'/{audience_id}/modify_data'
        csv_file = create_csv_file(applications)
        response = requests.post(
            url= url,
            headers=self.headers,
            files={"file": open(csv_file, "rb")}
        )
        if response.status_code != 200:
            logger.error(f"Ошибка загрузки данных в аудиторию Яндекс.Аудитории: {response.text}")
            raise Exception(f"Ошибка загрузки данных в аудиторию Яндекс.Аудитории: {response.text}")
        if audience_id:
           logger.debug(f"Данные успешно загружены в существующую аудиторию")
        else:
            logger.debug(f"Данные успешно загружены в новую аудиторию")
        json_response = response.json()
        external_id = json_response["segment"]["id"]
        status = json_response["segment"]["status"]
        if not audience_id and audience_name:
            name = self.update_name(external_id, audience_name)
        else:
            name = json_response["segment"]["name"]
        return {
            "result": "success",
            "external_id": external_id,
            "status": status,
            "name": name
        }

    def send_audience(self,  audience_name, application_ids, external_id=None):
        try:
            audiences = self.get_audiences()
            if audiences:
                segments = audiences.get("segments", [])
                if external_id and external_id != -1:
                    return self.upload_applications(application_ids, audience_id=external_id)
                audience = next((aud for aud in segments if aud["name"] == audience_name), None)
                if audience:
                    return self.upload_applications(application_ids, audience_id=audience["id"])
                return self.upload_applications(application_ids, audience_name=audience_name)
        except Exception as ex:
            print(ex)
            return {"result": "error", "message": str(ex)}



def send_to_yandex_platform(audience_name, application_ids, external_id=None):
    yandex_integration = YandexIntegration(
        oauth_token=os.getenv("YANDEX_OAUTH_TOKEN")
    )
    return yandex_integration.send_audience(
        audience_name=audience_name,
        application_ids=application_ids,
        external_id=external_id
    )