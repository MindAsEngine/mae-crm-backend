from facebook_business.adobjects.adaccount import AdAccount
from facebook_business.api import FacebookAdsApi
from facebook_business.adobjects.customaudience import CustomAudience
from google.ads.googleads.client import GoogleAdsClient
from google.ads.googleads.errors import GoogleAdsException
import requests
import os
import logging
import datetime
import json

logger = logging.getLogger()

class FacebookAdsIntegration:
    def __init__(self, access_token, app_id, app_secret):
        FacebookAdsApi.init(access_token=access_token, app_id=app_id, app_secret=app_secret)

    def get_audiences(self, account_id):
        try:
            account = AdAccount(f'act_{account_id}')
            audiences = account.get_custom_audiences(
                fields=[
                    CustomAudience.Field.id,
                    CustomAudience.Field.name,
                    CustomAudience.Field.approximate_count
                ]
            )
            return audiences
        except Exception as e:
            logger.error(f"Facebook get audiences error: {str(e)}")
            raise

    def update_audience(self, audience_id, application_ids):
        try:
            audience = CustomAudience(audience_id)
            audience.add_users(
                schema=CustomAudience.Schema.phone_number,
                data=application_ids
            )
            return True
        except Exception as e:
            logger.error(f"Facebook update audience error: {str(e)}")
            raise
    def send_audience(self, account_id, audience_name, application_ids):
        try:
            audience = CustomAudience(parent_id=account_id)
            audience.update({
                CustomAudience.Field.name: audience_name,
                CustomAudience.Field.description: f'Created at {datetime.datetime.now()}'
            })
            audience.create()

            # Upload user data
            audience.add_users(schema=CustomAudience.Schema.phone_number, data=application_ids)
            return audience.get_id()
        except Exception as e:
            logger.error(f"Facebook audience upload error: {str(e)}")
            raise

class GoogleAdsIntegration:
    def __init__(self, client_id, client_secret, developer_token, refresh_token):
        self.client = GoogleAdsClient.load_from_dict({
            'client_id': client_id,
            'client_secret': client_secret,
            'developer_token': developer_token,
            'refresh_token': refresh_token,
            'use_proto_plus': True,
        })

    def get_audiences(self, customer_id):
        try:
            ga_service = self.client.get_service('GoogleAdsService')
            query = """
                SELECT
                  audience.id,
                  audience.name,
                  audience.size_for_display,
                  audience.size_for_search
                FROM audience
                WHERE audience.type = 'CUSTOMER_MATCH'
            """
            response = ga_service.search(customer_id=customer_id, query=query)
            return [row.audience for row in response]
        except GoogleAdsException as e:
            logger.error(f"Google Ads get audiences error: {e.failure.errors}")
            raise

    def update_audience(self, customer_id, audience_id, application_ids):
        try:
            audience_service = self.client.get_service("CustomerMatchUploadService")
            operations = []

            for app_id in application_ids:
                operation = self.client.get_type("CustomerMatchUploadOperation")
                operation.create.user_data.phone_number = app_id
                operations.append(operation)

            response = audience_service.upload(
                resource_name=audience_id,
                customer_id=customer_id,
                operations=operations
            )
            return True
        except GoogleAdsException as e:
            logger.error(f"Google Ads update error: {e.failure.errors}")
            raise
    def send_audience(self, customer_id, audience_name, application_ids):
        try:
            audience_service = self.client.get_service("CustomerMatchUploadService")

            audience_operation = self.client.get_type("CustomerMatchUploadOperation")
            audience = audience_operation.create
            audience.audience_name = audience_name

            for app_id in application_ids:
                user_data = audience.user_data.add()
                user_data.phone_number = app_id

            response = audience_service.upload(
                customer_id=customer_id,
                operations=[audience_operation]
            )
            return response.results[0].resource_name
        except GoogleAdsException as e:
            logger.error(f"Google Ads error: {e.failure.errors}")
            raise

class YandexDirectIntegration:
    def __init__(self, oauth_token):
        self.oauth_token = oauth_token
        self.api_url = "https://api.direct.yandex.com/json/v5/"

    def get_audiences(self):
        try:
            headers = {
                "Authorization": f"Bearer {self.oauth_token}",
                "Accept-Language": "ru"
            }

            request_data = {
                "method": "get",
                "params": {
                    "SelectionCriteria": {
                        "Types": ["CUSTOMER_MATCH"]
                    },
                    "FieldNames": ["Id", "Name", "Size"]
                }
            }

            response = requests.post(
                f"{self.api_url}audiences",
                headers=headers,
                json=request_data
            )
            response.raise_for_status()
            return response.json()["result"]["Audiences"]
        except requests.exceptions.RequestException as e:
            logger.error(f"Yandex Direct get audiences error: {str(e)}")
            raise

    def update_audience(self, audience_id, application_ids):
        try:
            headers = {
                "Authorization": f"Bearer {self.oauth_token}",
                "Accept-Language": "ru",
                "Content-Type": "application/json"
            }

            upload_data = {
                "method": "upload",
                "params": {
                    "AudienceId": audience_id,
                    "Users": [{"Phone": str(id)} for id in application_ids]
                }
            }

            response = requests.post(
                f"{self.api_url}audiences/upload",
                headers=headers,
                json=upload_data
            )
            response.raise_for_status()
            return True
        except requests.exceptions.RequestException as e:
            logger.error(f"Yandex Direct update error: {str(e)}")
            raise

    def send_audience(self, client_id, audience_name, application_ids):
        try:
            headers = {
                "Authorization": f"Bearer {self.oauth_token}",
                "Accept-Language": "ru",
                "Content-Type": "application/json"
            }

            # Create audience
            audience_data = {
                "method": "create",
                "params": {
                    "Audiences": [{
                        "Name": audience_name,
                        "Type": "CUSTOMER_MATCH",
                        "Description": f"Created at {datetime.datetime.now()}"
                    }]
                }
            }

            response = requests.post(
                f"{self.api_url}audiences",
                headers=headers,
                json=audience_data
            )
            response.raise_for_status()
            audience_id = response.json()["result"]["AudienceId"]

            # Upload users
            upload_data = {
                "method": "upload",
                "params": {
                    "AudienceId": audience_id,
                    "Users": [{"Phone": str(id)} for id in application_ids]
                }
            }

            response = requests.post(
                f"{self.api_url}audiences/upload",
                headers=headers,
                json=upload_data
            )
            response.raise_for_status()
            return audience_id
        except requests.exceptions.RequestException as e:
            logger.error(f"Yandex Direct error: {str(e)}")
            raise


# Update existing functions
def send_to_facebook_platform(audience_name, application_ids):
    facebook_integration = FacebookAdsIntegration(
        access_token=os.getenv("FB_ACCESS_TOKEN"),
        app_id=os.getenv("FB_APP_ID"),
        app_secret=os.getenv("FB_APP_SECRET")
    )
    return facebook_integration.send_audience(
        account_id=os.getenv("FB_ACCOUNT_ID"),
        audience_name=audience_name,
        application_ids=application_ids
    )


def send_to_google_platform(audience_name, application_ids):
    google_integration = GoogleAdsIntegration(
        client_id=os.getenv("GOOGLE_CLIENT_ID"),
        client_secret=os.getenv("GOOGLE_CLIENT_SECRET"),
        developer_token=os.getenv("GOOGLE_DEVELOPER_TOKEN"),
        refresh_token=os.getenv("GOOGLE_REFRESH_TOKEN")
    )
    return google_integration.send_audience(
        customer_id=os.getenv("GOOGLE_CUSTOMER_ID"),
        audience_name=audience_name,
        application_ids=application_ids
    )


def send_to_yandex_platform(audience_name, application_ids):
    yandex_integration = YandexDirectIntegration(
        oauth_token=os.getenv("YANDEX_OAUTH_TOKEN")
    )
    return yandex_integration.send_audience(
        client_id=os.getenv("YANDEX_CLIENT_ID"),
        audience_name=audience_name,
        application_ids=application_ids
    )