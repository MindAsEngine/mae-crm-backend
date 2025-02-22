FROM postgres:latest

# Копируем файлы
COPY init.dump /docker-entrypoint-initdb.d/init.dump
COPY init-db.sh /docker-entrypoint-initdb.d/init-db.sh

# Делаем скрипт исполняемым
RUN chmod +x /docker-entrypoint-initdb.d/init-db.sh

CMD ["docker-entrypoint.sh", "postgres"]

# Используем кастомный entrypoint
#CMD ["/bin/bash", "/docker-entrypoint-initdb.d/init-db.sh"]
