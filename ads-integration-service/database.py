import mysql.connector
from datetime import datetime
import pandas as pd
from tabulate import tabulate


def connect_to_database():
    return mysql.connector.connect(
        host="localhost",
        user="user",
        password="password",
        database="macro_bi_cmp_528",
        port=3306
    )


def get_applications_by_id(application_ids):
    try:
        conn = connect_to_database()
        cursor = conn.cursor(dictionary=True)
        
        query_parameterized = """
        SELECT 
            eb.id,
            eb.date_added,
            eb.date_modified,
            eb.status_name,
            eb.status_reason_id,
            ebrs.name as reason_name
        FROM estate_buys eb
        LEFT JOIN estate_statuses_reasons ebrs 
            ON ebrs.status_reason_id = eb.status_reason_id
        WHERE eb.id IN (%s)
        """
        # Format as: (%s,%s,%s) for the number of ids
        in_format = ','.join(['%s'] * len(application_ids))
        query_parameterized = query_parameterized % in_format
        
        cursor.execute(query_parameterized, tuple(application_ids))
        results = cursor.fetchall()
        
        # Convert to pandas DataFrame for better output
        df = pd.DataFrame(results)
        
        # Format datetime columns
        date_columns = ['date_added', 'date_modified']
        for col in date_columns:
            df[col] = df[col].apply(lambda x: x.strftime('%Y-%m-%d %H:%M:%S'))
            
        print("\nQuery Results:")
        print(tabulate(df, headers='keys', tablefmt='psql'))
        
        print(f"\nTotal rows: {len(results)}")
        
    except mysql.connector.Error as err:
        print(f"Database error: {err}")
    finally:
        if 'conn' in locals():
            conn.close()
