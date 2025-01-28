import hashlib
import csv

def hash_data(data):
    return hashlib.sha256(data.encode('utf-8')).hexdigest()

def prepare_user(application):
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

def prepare_users(applications):
    users = []
    for application in applications:
        user = prepare_user(application)
        users.append(user)
    return users

def prepare_csv(applications, filename):
    users = prepare_users(applications)
    with open(filename, mode='w', newline='') as file:
        writer = csv.writer(file)
        writer.writerow(users[0].keys())
        for user in users:
            writer.writerow(user.values())
        return filename