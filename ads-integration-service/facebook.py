
from facebook_business.adobjects.adaccount import AdAccount
from facebook_business.adobjects.customaudience import CustomAudience
from facebook_business.api import FacebookAdsApi
import os
from logger import logger
from csv_prepare import format_users


def get_audiences(account):
    try:
        audiences = account.get_custom_audiences(
            fields=['id', 'name', 'operation_status', 'permission_for_actions'],
            params={}
        )
        logger.info("Facebook Ads get audiences success")
        # for audience in audiences:
        #     print(f"Facebook Ads get audiences success: {audience[CustomAudience.Field.name]} - {audience[CustomAudience.Field.permission_for_actions]}")
        return audiences
    except Exception as e:
        logger.error(f"Facebook Ads get audiences error: {str(e)}")


def create_audience(account, audience_name):
    try:
        fields = [
            CustomAudience.Field.id,
            CustomAudience.Field.name,
            CustomAudience.Field.operation_status,
            CustomAudience.Field.permission_for_actions,

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
        logger.info("Facebook Ads create audience success")
        print(f"Facebook Ads create audience success: {audience}")
        return audience
    except Exception as e:
        logger.error(f"Facebook Ads create audience error: {str(e)}")
        print(f"Facebook Ads create audience error: {str(e)}")

def add_users(audience_id, applications):
    try:
        users = format_users(applications)
        audience = CustomAudience(audience_id)
        print(f"Facebook add users: {users}")
        res = audience.add_users(
            users=users,
            schema=[
                CustomAudience.Schema.MultiKeySchema.phone,
                CustomAudience.Schema.MultiKeySchema.email,
                CustomAudience.Schema.MultiKeySchema.extern_id,
                CustomAudience.Schema.MultiKeySchema.gen,
                CustomAudience.Schema.MultiKeySchema.country,
                CustomAudience.Schema.MultiKeySchema.ct,
                CustomAudience.Schema.MultiKeySchema.fn,
                CustomAudience.Schema.MultiKeySchema.ln,
                CustomAudience.Schema.MultiKeySchema.doby,
                CustomAudience.Schema.MultiKeySchema.dobm,
                CustomAudience.Schema.MultiKeySchema.dobd
            ],
            is_raw=True
        )

        return {
            "audience_id": audience_id,
            "result": "success"
        }
    except Exception as e:
        print(f"Facebook add users error: {str(e)}")
        return {
            "result": "error",
            "message": str(e)
        }

def delete_users(audience_id, applications):
    try:
        users = format_users(applications)
        audience = CustomAudience(audience_id)
        print(f"Facebook delete users: {users}")
        res = audience.remove_users(
            users=users,
            schema=[
                    CustomAudience.Schema.MultiKeySchema.phone,
                CustomAudience.Schema.MultiKeySchema.email,
                CustomAudience.Schema.MultiKeySchema.extern_id,
                CustomAudience.Schema.MultiKeySchema.gen,
                CustomAudience.Schema.MultiKeySchema.country,
                CustomAudience.Schema.MultiKeySchema.ct,
                CustomAudience.Schema.MultiKeySchema.fn,
                CustomAudience.Schema.MultiKeySchema.ln,
                CustomAudience.Schema.MultiKeySchema.doby,
                CustomAudience.Schema.MultiKeySchema.dobm,
                CustomAudience.Schema.MultiKeySchema.dobd
                    ],
            is_raw=True
        )
        return {
            "audience_id": audience_id,
            "result": "success"
        }
    except Exception as e:
        print(f"Facebook delete users error: {str(e)}")
        return {
            "result": "error",
            "message": str(e)
        }


def send_audience(account, audience_name, applications_add, applications_remove):
    try:
        existing_audiences = get_audiences(account)
        if existing_audiences:
            audience = next((aud for aud in existing_audiences if aud["name"] == audience_name), None)
            if not audience:
                audience = create_audience(account, audience_name)
            if applications_remove and applications_remove.__len__() > 0:
                delete_users(audience["id"], applications_remove)
            if applications_add and applications_add.__len__() > 0:
                add_users(audience["id"], applications_add)

            return {
                "external_id": audience["id"],
                "status": "success",
                "name": audience_name,
                "result": "success"
            }
    except Exception as e:
        print(f"Facebook Ads error: {str(e)}")
        return {
            "result": "error",
            "message": str(e)
        }

def send_to_facebook_platform(audience_name, applications_add, applications_remove):
    FB_TOKEN = os.getenv("FB_ACCESS_TOKEN")
    FB_APP_ID = os.getenv("FB_APP_ID")
    FB_APP_SECRET = os.getenv("FB_APP_SECRET")
    FB_ACCOUNT_ID = os.getenv("FB_ACCOUNT_ID")
    FacebookAdsApi.init(
        access_token=FB_TOKEN,
        app_id=FB_APP_ID,
        app_secret=FB_APP_SECRET,
    )
    ad_account = AdAccount('act_' + FB_ACCOUNT_ID)
    print(f"Facebook Ads account: {ad_account}")
    return send_audience(
        account=ad_account,
        audience_name=audience_name,
        applications_add=applications_add,
        applications_remove=applications_remove
    )