import time

import mysql.connector
import os
from dotenv import load_dotenv
from logger import logger

def connect_to_database():
    load_dotenv()
    while True:
        try:
            return mysql.connector.connect(
                host=os.getenv("MYSQL_HOST"),
                user=os.getenv("MYSQL_USER"),
                password=os.getenv("MYSQL_PASSWORD"),
                database=os.getenv("MYSQL_DATABASE"),
                port=os.getenv("MYSQL_PORT")
            )
        except mysql.connector.Error as err:
            print(f"Соединение с базой данных не установлено: {err}. Попробую снова через 5 секунд.")
            logger.error(f"Соединение с базой данных не установлено: {err}. Попробую снова через 5 секунд.")
            time.sleep(5)
def get_applications_by_id(application_ids):
    try:
        conn = connect_to_database()
        logger.info("Соединение с базой данных установлено")
        cursor = conn.cursor(dictionary=True)
        
        query_parameterized = """
        SELECT deals.contacts_buy_sex, deals.contacts_buy_dob, deals.contacts_buy_name, 
         deals.name_first, deals.name_last, deals.name_middle,
        deals.contacts_buy_phones, deals.contacts_buy_emails, buys.contacts_buy_geo_country_name,
        buys.contacts_buy_geo_city_name, buys.contacts_id
        FROM macro_bi_cmp_528.estate_deals_contacts as deals 
        inner join macro_bi_cmp_528.estate_buys as buys 
        on buys.contacts_id = deals.id
        WHERE buys.id IN (%s) and deals.date_added > '2024-01-01'
        """
        # Format as: (%s,%s,%s) for the number of ids
        in_format = ','.join(['%s'] * len(application_ids))
        query_parameterized = query_parameterized % in_format
        
        cursor.execute(query_parameterized, tuple(application_ids))
        results = cursor.fetchall()

        logger.info("Данные из базы данных получены")
        conn.close()
        return results
    except mysql.connector.Error as err:
        logger.error(f"Ошибка при получении данных из базы данных: {err}")



if __name__ == "__main__":
    get_applications_by_id([4953867,4953925,4953927])