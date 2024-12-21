from facebook_business.adobjects.adaccount import AdAccount
from facebook_business.api import FacebookAdsApi
from facebook_business.adobjects.customaudience import CustomAudience
from google.ads.googleads.client import GoogleAdsClient
from google.ads.googleads.errors import GoogleAdsException
import requests
import os
import datetime

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
        user.update({"phone": hash_data(phone)})
    gender = application.get("contacts_buy_sex", None)
    if gender:
        user.update({"gender": hash_data(gender)})
    country = application.get("contacts_buy_geo_country_name", "Узбекистан")
    if country:
        user.update({"country": hash_data(country)})
    city = application.get("contacts_buy_geo_city_name", "Ташкент")
    if city:
        user.update({"city": hash_data(city)})
    name = application.get("contacts_buy_name", None)
    first_name = name.split()[0] if name else None
    if first_name:
        user.update({"first_name": hash_data(first_name)})
    last_name = name.split()[1] if name else None
    if last_name:
        user.update({"last_name": hash_data(last_name)})
    dob = application.get("contacts_buy_dob", None)
    if dob:
        user.update({"dob": hash_data(dob)})
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

    def create_audience(self, audience_name):
        try:
            user_list_service = self.client.get_service("UserListService")
            user_list_operation = self.client.get_type("UserListOperation")

            user_list = user_list_operation.create
            user_list.name = audience_name
            user_list.description = "Customer Match audience created via API"
            user_list.membership_life_span = 30

            response = user_list_service.mutate_user_lists(
                customer_id=self.customer_id,
                operations=[user_list_operation]
            )
            audience_resource_name = response.results[0].resource_name
            print(f"Создана аудитория: {audience_resource_name}")
            return audience_resource_name
        except GoogleAdsException as ex:
            print(f"Ошибка создания аудитории Google Ads: {str(ex)}")
            raise

    def update_audience(self, audience_resource_name, applications):
        try:
            offline_user_data_job_service = self.client.get_service("OfflineUserDataJobService")
            offline_user_data_job_operation = self.client.get_type("OfflineUserDataJobOperation")
            offline_user_data_job = offline_user_data_job_service.create_offline_user_data_job(
                customer_id=self.customer_id,
                job_type="CUSTOMER_MATCH_USER_LIST"
            )
            job_resource_name = offline_user_data_job.resource_name
            print(f"Создана Offline User Data Job: {job_resource_name}")

            # Создание операций добавления данных
            operations = []
            for app in applications:
                user_data = prepare_google_user_data(app)
                user_data_operation = offline_user_data_job_operation.create
                if "hashed_phone_number" in user_data:
                    self.client.get_type("UserIdentifier").hashed_phone_number = user_data["hashed_phone_number"]
                    user_data_operation.user_identifiers.append(
                    self.client.get_type("UserIdentifier").hashed_phone_number
                    )
                operations.append(user_data_operation)

            # Добавление пользователей в аудиторию
            offline_user_data_job_service.add_offline_user_data_job_operations(
                resource_name=job_resource_name,
                operations=operations
            )
            offline_user_data_job_service.run_offline_user_data_job(
                resource_name=job_resource_name
            )
            print(f"Пользователи успешно добавлены в аудиторию: {audience_resource_name}")
            return True
        except Exception as ex:
            print(f"Ошибка обновления аудитории Google Ads: {str(ex)}")
            raise


    def send_audience(self, audience_name, applications):
        try:
            # Проверяем существование аудитории
            query = f"""
                SELECT user_list.resource_name, user_list.name
                FROM user_list
                WHERE user_list.name = '{audience_name}'
            """
            response = self.client.get_service("GoogleAdsService").search_stream(
                customer_id=self.customer_id, query=query
            )
            existing_audience = None
            for batch in response:
                for row in batch.results:
                    existing_audience = row.user_list.resource_name

            # Если аудитория уже существует, обновляем её
            if existing_audience:
                return self.update_audience(existing_audience, applications)

            # Создаём новую аудиторию
            audience_resource_name = self.create_audience(audience_name)
            self.update_audience(audience_resource_name, applications)
            return {"result": "success", "audience_id": audience_resource_name}
        except GoogleAdsException as ex:
            print(f"Ошибка Google Ads API: {str(ex)}")
            return {"result": "error", "message": str(ex)}

# class YandexDirectIntegration:
#     def __init__(self, oauth_token):
#         self.oauth_token = oauth_token
#         self.api_url = "https://api.direct.yandex.com/json/v5/"
#
#     def get_audiences(self):
#         try:
#             headers = {
#                 "Authorization": f"Bearer {self.oauth_token}",
#                 "Accept-Language": "ru"
#             }
#
#             request_data = {
#                 "method": "get",
#                 "params": {
#                     "SelectionCriteria": {
#                         "Types": ["CUSTOMER_MATCH"]
#                     },
#                     "FieldNames": ["Id", "Name", "Size"]
#                 }
#             }
#
#             response = requests.post(
#                 f"{self.api_url}audiences",
#                 headers=headers,
#                 json=request_data
#             )
#             response.raise_for_status()
#             return response.json()["result"]["Audiences"]
#         except requests.exceptions.RequestException as e:
#             logger.error(f"Yandex Direct get audiences error: {str(e)}")
#             raise
#
#     def update_audience(self, audience_id, application_ids):
#         try:
#             headers = {
#                 "Authorization": f"Bearer {self.oauth_token}",
#                 "Accept-Language": "ru",
#                 "Content-Type": "application/json"
#             }
#
#             upload_data = {
#                 "method": "upload",
#                 "params": {
#                     "AudienceId": audience_id,
#                     "Users": [{"Phone": str(id)} for id in application_ids]
#                 }
#             }
#
#             response = requests.post(
#                 f"{self.api_url}audiences/upload",
#                 headers=headers,
#                 json=upload_data
#             )
#             response.raise_for_status()
#             return True
#         except requests.exceptions.RequestException as e:
#             logger.error(f"Yandex Direct update error: {str(e)}")
#             raise
#
#     def send_audience(self, audience_name, application_ids):
#         try:
#             headers = {
#                 "Authorization": f"Bearer {self.oauth_token}",
#                 "Accept-Language": "ru",
#                 "Content-Type": "application/json"
#             }
#
#             # Create audience
#             audience_data = {
#                 "method": "create",
#                 "params": {
#                     "Audiences": [{
#                         "Name": audience_name,
#                         "Type": "CUSTOMER_MATCH",
#                         "Description": f"Created at {datetime.datetime.now()}"
#                     }]
#                 }
#             }
#
#             response = requests.post(
#                 f"{self.api_url}audiences",
#                 headers=headers,
#                 json=audience_data
#             )
#             response.raise_for_status()
#             audience_id = response.json()["result"]["AudienceId"]
#
#             # Upload users
#             upload_data = {
#                 "method": "upload",
#                 "params": {
#                     "AudienceId": audience_id,
#                     "Users": [{"Phone": str(id)} for id in application_ids]
#                 }
#             }
#
#             response = requests.post(
#                 f"{self.api_url}audiences/upload",
#                 headers=headers,
#                 json=upload_data
#             )
#             response.raise_for_status()
#             return audience_id
#         except requests.exceptions.RequestException as e:
#             logger.error(f"Yandex Direct error: {str(e)}")
#             raise
#

# Update existing functions



def send_to_google_platform(audience_name, application_ids):
    google_integration = GoogleAdsIntegration(
        client=GoogleAdsClient.load_from_storage("GOOGLE_ADS_YAML_PATH"),
        customer_id=os.getenv("GOOGLE_CUSTOMER_ID")
    )
    return google_integration.send_audience(audience_name, application_ids)

#
# def send_to_yandex_platform(audience_name, application_ids):
#     yandex_integration = YandexDirectIntegration(
#         oauth_token=os.getenv("YANDEX_OAUTH_TOKEN")
#     )
#     return yandex_integration.send_audience(
#         client_id=os.getenv("YANDEX_CLIENT_ID"),
#         audience_name=audience_name,
#         application_ids=application_ids
#     )