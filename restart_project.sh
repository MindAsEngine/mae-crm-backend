#!/bin/bash
docker compose down

chmod -R u+w .

ENV_FILE="environment.env"
TMP_FILE=$(mktemp)
backend_repo=$""
backend_branch=$""
frontend_repo=$""
frontend_branch=$""
APP_NAME=$(basename "$PWD")
NETWORK_NAME="${APP_NAME}_app-network"


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
        if [[ "$key" = "GIT_REPO_BACKEND_URL" ]]; then
            backend_repo="$value"
            echo "Найден репозиторий: $backend_repo"
        fi
        if [[ "$key" = "GIT_MAIN_BACKEND_BRANCH" ]]; then
            backend_branch="$value"
            echo "Найдена ветка: $backend_branch"
        fi
        if [[ "$key" = "GIT_REPO_FRONTEND_URL" ]]; then
            frontend_repo="$value"
            echo "Найден репозиторий: $frontend_repo"
        fi
        if [[ "$key" = "GIT_MAIN_FRONTEND_BRANCH" ]]; then
            frontend_branch="$value"
            echo "Найдена ветка: $frontend_branch"
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

echo "Обновить приложение из GitHub - [Y] из архива/отменить - [Any]"
read -r confirm_git
if [[ "$confirm_git" =~ ^[Yy]$ ]]; then
    echo "Найден репозиторий $backend_repo и ветка $backend_branch. Загрузить из GitHub?  - [Y] / [Any]"
    read -r confirm_git
    if [[ "$confirm_git" =~ ^[Yy]$ ]]; then
        echo "Впервые загружаем из GitHub - [Y] / [Any]"
        read -r confirm_git
        if [[ "$confirm_git" =~ ^[Yy]$ ]]; then
            echo "Клонируем из $backend_repo"
            mkdir -p macro-crm-temp
            cd macro-crm-temp || exit 1
            pwd
            git clone "$backend_repo" .
            git checkout "$backend_branch"
            cd ..
            mv macro-crm-temp/* .
            rm -rf macro-crm-temp
            cd reporting-service
            mkdir -p export
            cd ..
            echo "Клонируем из $frontend_repo"
            mkdir -p macro-crm-frontend
            cd mae-crm-frontend || exit 1
            git clone "$frontend_repo" .
            git checkout "$frontend_branch"
            cd ..
        else
            echo "Пуллим из $backend_repo $backend_branch"
            git fetch
            git pull "$backend_repo" "$backend_branch"
            git checkout "$backend_branch"
            echo "Пуллим из $frontend_repo $frontend_branch"
            cd macro-crm-frontend || exit 1
            git fetch
            git pull "$frontend_repo" "$frontend_branch"
            git checkout "$frontend_branch"
            cd ..
        fi
    fi
    echo "Обновлено при помощи git"
    sleep 1
else 
    echo "Загрузить из архива - [Y] отмена обновления - [Any])"
    read -r confirm_archive
    if [[ "$confirm_archive" =~ ^[Yy]$ ]]; then
        tar -xvvf GoldenHouseRepo.tar
        echo "Обновлено при помощи архива"
        sleep 1
    else
        echo "Обновление отменено"
    fi
fi


if [[ "$1" == "dev" ]]; then
    echo "Запущен в режиме разработки!"
    echo "Удалить всё и сразу из докера?  - [Y] / [Any]"
    read -r confirm_all
    if [[ "$confirm_all" =~ ^[Yy]$ ]]; then
        echo "УВЕРЕН?  - [Y] / [Any]"
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
echo "Удалить сеть приложения?  - [Y] / [Any]"
read -r confirm_net
if [[ "$confirm_net" =~ ^[Yy]$ ]]; then  
    docker network rm "$NETWORK_NAME"
    echo "Cеть приложения удалена."
fi
# Запрос на удаление образов контейнеров
echo "Удалять образы контейнеров (кроме бд)?  - [Y] / [Any]"
read -r confirm_images
if [[ "$confirm_images" =~ ^[Yy]$ ]]; then
    echo "Удалить ВСЕ образы контейнеров (кроме бд)  - [Y] / [Any]"
    read -r confirm_images
    if [[ "$confirm_images" =~ ^[Yy]$ ]]; then
        docker rmi "${APP_NAME}-ads-integration-service"
        docker rmi "${APP_NAME}-frontend"
        docker rmi "${APP_NAME}-reporting-service"
        docker rmi "${APP_NAME}-auth-service"
    elif [[ ! "$confirm_images" =~ ^[Yy]$ ]]; then
        echo "Удаление образов контейнеров (кроме бд) по очереди: "
        images=("${APP_NAME}-ads-integration-service" "${APP_NAME}-frontend" "${APP_NAME}-reporting-service" "${APP_NAME}-auth-service")
        for img in "${images[@]}"; do
            echo "Удалить образ $img?  - [Y] / [Any]"
            read -r confirm
            if [[ "$confirm" =~ ^[Yy]$ ]]; then
                docker rmi "$img"
            fi
        done
    fi
    echo "Текущие образы контейнеров приложения удалены."
    sleep 1
fi
echo "Удалять образы контейнеров баз данных (не рекомендуется)  - [Y] / [Any]"
read -r confirm_images
if [[ "$confirm_images" =~ ^[Yy]$ ]]; then
    echo "!!! Удалить ВСЕ образы контейнеров БАЗ ДАННЫХ (НЕ РЕКОМЕНДУЕТСЯ) !!!  - [Y] / [Any]"
    read -r confirm_images
    if [[ "$confirm_images" =~ ^[Yy]$ ]]; then
        docker rmi "${APP_NAME}-db"
        docker rmi "mongo"
        docker rmi "rabbitmq"
    elif [[ ! "$confirm_images" =~ ^[Yy]$ ]]; then
        echo "Удаление образов контейнеров баз данных по очереди: "
        images=("${APP_NAME}-db" "mongo" "rabbitmq")
        for img in "${images[@]}"; do
            echo "Удалить образ $img?  - [Y] / [Any]"
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
echo "Удалять Docker хранилища?  - [Y] / [Any]"
read -r confirm_volumes
if [[ "$confirm_volumes" =~ ^[Yy]$ ]]; then
    echo "!!! Удалить ВСЕ Docker хранилища (НЕ РЕКОМЕНДУЕТСЯ) !!!  - [Y] / [Any]"
    read -r confirm_volumes
    if [[ "$confirm_volumes" =~ ^[Yy]$ ]]; then
        echo "ВЫ УВЕРЕНЫ? ВСЕ ДАННЫЕ БУДУТ БЕЗВОЗВРАТНО УТЕРЯНЫ! !!!  - [Y] / [Any]"
        read -r confirm_volumes
        if [[ "$confirm_volumes" =~ ^[Yy]$ ]]; then
            docker volume rm $(docker volume ls -q)
            echo "Все хранилища и данные были удалены"
            sleep 1
        fi
    elif [[ ! "$confirm_volumes" =~ ^[Yy]$ ]]; then
        echo "Удалять Docker хранилища, кроме хранилищ БД?  - [Y] / [Any]"
        read -r confirm_volumes
        if [[ "$confirm_volumes" =~ ^[Yy]$ ]]; then
            volumes=("${APP_NAME}_ads_integration_volume" "${APP_NAME}_export_data")
            for vol in "${volumes[@]}"; do
                echo "Удалить хранилище $vol?  - [Y] / [Any]"
                read -r confirm
                if [[ "$confirm" =~ ^[Yy]$ ]]; then
                    docker volume rm "$vol"
                fi
            done
        echo "Хранилища приложения очищены"
        sleep 1
        elif [[ ! "$confirm_volumes" =~ ^[Yy]$ ]]; then
            echo "Удалять Docker хранилища баз данных? - [Y] / [Any]"
            read -r confirm_volumes
            if [[ "$confirm_volumes" =~ ^[Yy]$ ]]; then
                echo "ВЫ УВЕРЕНЫ? ВСЕ ДАННЫЕ БУДУТ БЕЗВОЗВРАТНО УТЕРЯНЫ! !!! - [Y] / [Any]"
                read -r confirm_volumes
                if [[ "$confirm_volumes" =~ ^[Yy]$ ]]; then
                    volumes=("${APP_NAME}_postgres_data" "${APP_NAME}_auth_data" "${APP_NAME}_rabbitmq_data")
                    for vol in "${volumes[@]}"; do
                        echo "Удалить хранилище $vol?  - [Y] / [Any]"
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
echo "Очистка произведена. Запускаю приложение..."
sleep 1
echo "Запуск приложения..."
sleep 1
docker compose up -d