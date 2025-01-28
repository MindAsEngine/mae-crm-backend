




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

def send_audience(account, audience_name, application_ids):
    try:

        fields = [
            'id',
        ]
        params = {
        }
        existing_audiences = account.get_custom_audiences(
            fields=fields,
            params=params,
        )

        if existing_audiences:
            audience = next((aud for aud in existing_audiences if aud["name"] == audience_name), None)
            if audience:
                return update_audience(audience["id"], application_ids)

        fields = [
        ]
        params = {
            'name': audience_name,
            'subtype': 'CUSTOM',
            'description': '',
            'customer_file_source': 'USER_PROVIDED_ONLY',
        }
        audience = account.create_custom_audience(
            fields=fields,
            params=params,
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
    ad_account = AdAccount('act_' + os.getenv("FB_AD_ACCOUNT_ID"))

    return send_audience(
        account=ad_account,
        audience_name=audience_name,
        application_ids=application_ids
    )



#
# def create_csv_file(applications):
#     results = []
#     for app in applications:
#         user_data = prepare_facebook_user_data(app)
#         results.append(user_data)
#     with open('yandex.csv', 'w', newline='') as file:
#         writer = csv.writer(file)
#         writer.writerow(results[0].keys())
#         for result in results:
#             writer.writerow(result.values())
#         return file.name





#
# def send_to_yandex_platform(audience_name, application_ids, external_id=None):
#     yandex_integration = YandexIntegration(
#         oauth_token=os.getenv("YANDEX_OAUTH_TOKEN")
#     )
#     return yandex_integration.send_audience(
#         audience_name=audience_name,
#         application_ids=application_ids,
#         external_id=external_id
#     )

#
# def prepare_google_user_data(application):
#     user_data = {}
#     # Хэширование email
#     email = application.get("contacts_buy_emails")
#     if email:
#         user_data["hashed_email"] = hash_data(email.lower().strip())
#     # Хэширование телефона
#     phone = application.get("contacts_buy_phones")
#     if phone:
#         user_data["hashed_phone_number"] = hash_data(phone.strip())
#     return user_data


# class GoogleAdsIntegration:
#     def __init__(self, client=None, customer_id=None):
#         self.client = client
#         self.customer_id = customer_id
#
#     def create_audience(self, audience_name):
#         try:
#             user_list_service = self.client.get_service("UserListService")
#             user_list_operation = self.client.get_type("UserListOperation")
#
#             user_list = user_list_operation.create
#             user_list.name = audience_name
#             user_list.description = "Customer Match audience created via API"
#             user_list.membership_life_span = 30
#
#             response = user_list_service.mutate_user_lists(
#                 customer_id=self.customer_id,
#                 operations=[user_list_operation]
#             )
#             audience_resource_name = response.results[0].resource_name
#             print(f"Создана аудитория: {audience_resource_name}")
#             return audience_resource_name
#         except GoogleAdsException as ex:
#             print(f"Ошибка создания аудитории Google Ads: {str(ex)}")
#             raise
#
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
#             logger.debug(f"Поиск аудитории: {response}")
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
#
# def send_to_google_platform(audience_name, application_ids):
#     os.environ["GOOGLE_ADS_CONFIGURATION_FILE_PATH"] = os.getenv("GOOGLE_ADS_YAML_PATH")
#     google_integration = GoogleAdsIntegration(
#         client=GoogleAdsClient.load_from_storage(),
#         customer_id=os.getenv("GOOGLE_CUSTOMER_ID")
#     )
#     return google_integration.send_audience(audience_name, application_ids)