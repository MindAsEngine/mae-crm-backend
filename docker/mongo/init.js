db = db.getSiblingDB('authdb');

db.createCollection('refresh_tokens');
db.createCollection('users');

const dateValue = new Date(-62169984000000);

db.users.insertMany([
    {
        "login": "user",
        "role": "user",
        "name": "",
        "surname": "",
        "patronymic": "",
        "refresh_token": "",
        "rt_token_expiry": dateValue,
        "at_token_expiry": dateValue,
        "password_hash": "$2a$10$REpR0LHbyBpA4jBdsipphO4FhFWDAkQbFIwkwCljmjiwUjv1QvbWa"
      }
]);

print("Database and collections initialized.");