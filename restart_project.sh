#!/bin/bash
docker compose down
ENV_FILE="environment.env"
TMP_FILE=$(mktemp)
repo=$""
branch=$""

echo "Проверка файла $ENV_FILE..."
sleep 1
if [ ! -f "$ENV_FILE" ]; then
    echo "Ошибка: Файл $ENV_FILE не найден!"
    exit 1
fi
# Если файл в формате DOS, преобразуем его в Unix-формат
if file "$ENV_FILE" | grep -q "CRLF"; then
    echo "Обнаружен DOS-формат, преобразуем..."
    sed -i 's/\r$//' "$ENV_FILE"
    sleep 1
fi
# Если файл не заканчивается переводом строки, добавляем его
tail -n1 "$ENV_FILE" | read -r _ || echo "" >> "$ENV_FILE"
# Читаем файл в массив, гарантируя разбиение по строкам
mapfile -t lines < "$ENV_FILE"
# Обрабатываем каждую строку отдельно
for line in "${lines[@]}"; do
    # Если строка пустая или комментарий — записываем без изменений
    if [[ "$line" =~ ^[[:space:]]*$ || "$line" =~ ^[[:space:]]*# ]]; then
        printf "%s\n" "$line" >> "$TMP_FILE"
        continue
    fi
    # Если строка содержит символ "="
    if [[ "$line" == *"="* ]]; then
        key="${line%%=*}"       # Всё до первого '='
        value="${line#*=}"      # Всё после первого '='
        # Обрезаем пробелы
        value="$(echo "$value" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')"
        # Если значение пустое – запрашиваем ввод
        if [[ "$key" = "GIT_REPO_URL" ]]; then
            repo="$value"
            echo "Найден репозиторий: $repo"
        fi
        if [[ "$key" = "GIT_BRANCH" ]]; then
            branch="$value"
            echo "Найдена ветка: $branch"
        fi
        if [[ -z "$value" ]]; then
            while true; do
                read -rp "Введите значение для $key: " input
                input="$(echo "$input" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')"
                if [[ -n "$input" ]]; then
                    value="$input"
                    break
                fi
            done
        fi
        # Записываем строку в формате KEY=VALUE с переводом строки
        printf "%s=%s\n" "$key" "$value" >> "$TMP_FILE"
    else
        printf "%s\n" "$line" >> "$TMP_FILE"
    fi
done
mv "$TMP_FILE" "$ENV_FILE"
chmod 777 "$ENV_FILE"
echo "Просмотреть файл переменных?"
read -r confirm_cat_env
if [[ "$confirm_cat_env" =~ ^[Yy]$ ]]; then
    cat cat "$ENV_FILE"
fi
echo "Все переменные установлены."
sleep 1
if [[ "$1" == "dev" ]]; then
    echo "Запущен в режиме разработки!"
    echo "Удалить всё и сразу? Y/Any"
    read -r confirm_all
    if [[ "$confirm_all" =~ ^[Yy]$ ]]; then
        echo "УВЕРЕН? Y/Any"
        read -r confirm_all
        if [[ "$confirm_all" =~ ^[Yy]$ ]]; then
            docker system prune -a 
            docker volume rm $(docker volume ls -q)
            echo "Удалено всё в т.ч. все хранилища, образы, контейнеры, сети."
        fi
    elif [[ ! "$confirm_all" =~ ^[Yy]$ ]]; then
        echo "Окей, пойдём долгим путём"
    fi
fi
echo "Удалить сеть приложения? Y/Any"
read -r confirm_net
if [[ "$confirm_net" =~ ^[Yy]$ ]]; then
    docker network rm mae-crm_app-network
    echo "Cеть приложения удалена."
fi
# Запрос на удаление образов контейнеров
echo "Удалять образы контейнеров (кроме бд)? (Y/Any)"
read -r confirm_images
if [[ "$confirm_images" =~ ^[Yy]$ ]]; then
    echo "Удалить ВСЕ образы контейнеров (кроме бд) (Y/Any)"
    read -r confirm_images
    if [[ "$confirm_images" =~ ^[Yy]$ ]]; then
        docker rmi mae-crm-ads-integration-service
        docker rmi mae-crm-frontend
        docker rmi mae-crm-reporting-service
        docker rmi mae-crm-auth-service
    elif [[ ! "$confirm_images" =~ ^[Yy]$ ]]; then
        echo "Удаление образов контейнеров (кроме бд) по очереди: "
        images=("mae-crm-ads-integration-service" "mae-crm-frontend" "mae-crm-reporting-service" "mae-crm-auth-service")
        for img in "${images[@]}"; do
            echo "Удалить образ $img? (Y/Any)"
            read -r confirm
            if [[ "$confirm" =~ ^[Yy]$ ]]; then
                docker rmi "$img"
            fi
        done
    fi
    echo "Текущие образы контейнеров приложения удалены."
    sleep 1
fi
echo "Удалять образы контейнеров баз данных (не рекомендуется) (Y/Any)"
read -r confirm_images
if [[ "$confirm_images" =~ ^[Yy]$ ]]; then
    echo "!!! Удалить ВСЕ образы контейнеров БАЗ ДАННЫХ (НЕ РЕКОМЕНДУЕТСЯ) !!! (Y/Any)"
    read -r confirm_images
    if [[ "$confirm_images" =~ ^[Yy]$ ]]; then
        docker rmi mae-crm-db
        docker rmi mae-crm-mongo
        docker rmi mae-crm-rabbitmq
    elif [[ ! "$confirm_images" =~ ^[Yy]$ ]]; then
        echo "Удаление образов контейнеров баз данных по очереди: "
        images=("mae-crm-db" "mongo" "rabbitmq")
        for img in "${images[@]}"; do
            echo "Удалить образ $img? (Y/Any)"
            read -r confirm
            if [[ "$confirm" =~ ^[Yy]$ ]]; then
                docker rmi "$img"
            fi
        done
    fi
    echo "Текущие образы баз данных удалены"
    sleep 1
fi
# Запрос на удаление volumes
echo "Удалять Docker хранилища? (Y/Any)"
read -r confirm_volumes
if [[ "$confirm_volumes" =~ ^[Yy]$ ]]; then
    echo "!!! Удалить ВСЕ Docker хранилища (НЕ РЕКОМЕНДУЕТСЯ) !!! (Y/Any)"
    read -r confirm_volumes
    if [[ "$confirm_volumes" =~ ^[Yy]$ ]]; then
        echo "ВЫ УВЕРЕНЫ? ВСЕ ДАННЫЕ БУДУТ БЕЗВОЗВРАТНО УТЕРЯНЫ! !!! (Y/Any)"
        read -r confirm_volumes
        if [[ "$confirm_volumes" =~ ^[Yy]$ ]]; then
            docker volume rm $(docker volume ls -q)
            echo "Все хранилища и данные были удалены"
            sleep 1
        fi
    elif [[ ! "$confirm_volumes" =~ ^[Yy]$ ]]; then
        echo "Удалять Docker хранилища, кроме хранилищ БД? (Y/Any)"
        read -r confirm_volumes
        if [[ "$confirm_volumes" =~ ^[Yy]$ ]]; then
            volumes=("mae-crm_ads_integration_volume" "mae-crm_export_data")
            for vol in "${volumes[@]}"; do
                echo "Удалить хранилище $vol? (Y/N)"
                read -r confirm
                if [[ "$confirm" =~ ^[Yy]$ ]]; then
                    docker volume rm "$vol"
                fi
            done
        echo "Хранилища приложения очищены"
        sleep 1
        elif [[ ! "$confirm_volumes" =~ ^[Yy]$ ]]; then
            echo "Удалять Docker хранилища баз данных? (Y/N)"
            read -r confirm_volumes
            if [[ "$confirm_volumes" =~ ^[Yy]$ ]]; then
                echo "ВЫ УВЕРЕНЫ? ВСЕ ДАННЫЕ БУДУТ БЕЗВОЗВРАТНО УТЕРЯНЫ! !!! (Y/Any)"
                read -r confirm_volumes
                if [[ "$confirm_volumes" =~ ^[Yy]$ ]]; then
                    volumes=("mae-crm_postgres_data" "mae-crm_auth_data" "mae-crm_rabbitmq_data")
                    for vol in "${volumes[@]}"; do
                        echo "Удалить хранилище $vol? (Y/Any)"
                        read -r confirm
                        if [[ "$confirm" =~ ^[Yy]$ ]]; then
                            docker volume rm "$vol"
                        fi
                    done
                    echo "Хранилища баз данных удалены"
                    sleep 1
                fi
            fi
        fi
    fi
fi
echo "Очистка произведена. Обновляю приложение..."
sleep 1
echo "Обновить приложение из GitHUB - [Y] из архива - [Any]"
read -r confirm_git
if [[ "$confirm_git" =~ ^[Yy]$ ]]; then
    echo "Найден репозиторий $repo и ветка $branch загрузить оттуда? (Y/Any)"
    read -r confirm_git
    if [[ "$confirm_git" =~ ^[Yy]$ ]]; then
        echo "Впервые загружаем оттуда? (Y/Any)"
        read -r confirm_git
        if [[ "$confirm_git" =~ ^[Yy]$ ]]; then
            echo "Клонируем из $repo"
            git clone "$repo" "$(dirname "$0")"
            git checkout $branch
        else
            echo "Пуллим из $repo $branch"
            git fetch
            git pull $repo $branch
            git checkout $branch
        fi
    fi
    echo "Обновлено при помощи git"
    sleep 1
else 
    echo "Загрузить из архива? (Y/Any)"
    read -r confirm_archive
    if [[ "$confirm_archive" =~ ^[Yy]$ ]]; then
        unzip GoldenHouseRepo.zip -d "$(dirname "$0")"
        echo "Обновлено при помощи архива"
        sleep 1
    fi
fi

echo "Запуск приложения..."
docker compose up -d