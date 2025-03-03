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

user_menu="Введите команду из списка:
    1 | envcheck - проверка файла переменных
    2 | update - обновление приложения из GitHub
    3 | images - удаление образов контейнеров
    4 | volumes - удаление хранилищ
    5 | network - удаление сети
    6 | start - запуск приложения
    7 | purge - полная очистка (dev only)
    8 | exit - выход
    clear - очистить терминал"

set_git() {
    echo "Проверка файла $ENV_FILE на наличие репозитория"

    # Проверяем, существует ли файл
    if [ ! -f "$ENV_FILE" ]; then
        echo "Ошибка: Файл $ENV_FILE не найден!"
        return 1
    fi

    # Читаем значения из файла и записываем в глобальные переменные
    declare -g backend_repo=$(grep "^GIT_REPO_BACKEND_URL=" "$ENV_FILE" | cut -d '=' -f2-)
    declare -g frontend_repo=$(grep "^GIT_REPO_FRONTEND_URL=" "$ENV_FILE" | cut -d '=' -f2-)
    declare -g backend_branch=$(grep "^GIT_MAIN_BACKEND_BRANCH=" "$ENV_FILE" | cut -d '=' -f2-)
    declare -g frontend_branch=$(grep "^GIT_MAIN_FRONTEND_BRANCH=" "$ENV_FILE" | cut -d '=' -f2-)

    # Проверяем, что переменные не пустые
    if [[ -z "$backend_repo" || -z "$frontend_repo" || -z "$backend_branch" || -z "$frontend_branch" ]]; then
        echo "Ошибка: Один или несколько параметров не найдены в $ENV_FILE"
        return 1
    fi

    # Выводим значения (для проверки)
    echo "Backend Repo: $backend_repo"
    echo "Backend Branch: $backend_branch"
    echo "Frontend Repo: $frontend_repo"
    echo "Frontend Branch: $frontend_branch"
}

mode="user"
if [[ "$1" == "dev" ]]; then
    mode="dev"
    echo "Запущен в режиме разработки!"
fi

while true; do
    echo -e "$user_menu"
    read -r command
    case $command in 
        0 | stop)
            echo "Остановка приложения..."
            docker compose down
            sleep 2
            ;;
        1 | envcheck)
            echo "Проверка файла $ENV_FILE..."
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
            echo "Просмотреть файл переменных? - [Y] / [Any]"
            read -r confirm_cat_env
            if [[ "$confirm_cat_env" =~ ^[Yy]$ ]]; then
                cat cat "$ENV_FILE"
            fi
            echo "Все переменные установлены."
            sleep 2
            ;;
        2 | update)
            echo "Обновить приложение из GitHub - [Y] из архива/отменить - [Any]"
            read -r confirm_git
            if [[ "$confirm_git" =~ ^[Yy]$ ]]; then
                set_git
                echo "Найден репозиторий $backend_repo и ветка $backend_branch. Загрузить из GitHub? - [Y] / [Any]"
                read -r confirm_git
                if [[ "$confirm_git" =~ ^[Yy]$ ]]; then
                    echo "Настраиваем репозиторий в текущей папке..."

                    # Проверяем, что мы в нужной директории
                    if [[ ! -d "." ]]; then
                        echo "Ошибка: текущая директория недоступна"
                        continue
                    fi

                    # Проверяем, есть ли .git, если нет — инициализируем
                    if [[ ! -d ".git" ]]; then
                        echo "Создаём git-репозиторий..."
                        git init -b "$backend_branch" || { echo "Ошибка при git init"; continue; }
                    fi

                    # Проверяем, есть ли текущий origin
                    current_origin=$(git remote get-url origin 2>/dev/null)
                    if [[ -z "$current_origin" || "$current_origin" != "$backend_repo" ]]; then
                        echo "Меняем origin на $backend_repo"
                        git remote remove origin 2>/dev/null
                        git remote add origin "$backend_repo" || { echo "Ошибка при добавлении origin"; continue; }
                    fi

                    # Загружаем изменения
                    git fetch origin || { echo "Ошибка при git fetch"; continue; }
                    git reset --hard "origin/$backend_branch" || { echo "Ошибка при git reset"; continue; }
                    git pull origin "$backend_branch" || { echo "Ошибка при git pull"; continue; }

                    echo "Настраиваем фронтенд..."

                    # Проверяем, существует ли папка с фронтендом, иначе создаем
                    if [[ ! -d "macro-crm-frontend" ]]; then
                        mkdir -p macro-crm-frontend
                    fi

                    cd macro-crm-frontend || { echo "Ошибка: не удалось зайти в macro-crm-frontend"; continue; }

                    # Аналогично настраиваем репозиторий для фронтенда
                    if [[ ! -d ".git" ]]; then
                        git init -b "$frontend_branch" || { echo "Ошибка при git init"; continue; }
                    fi

                    current_origin=$(git remote get-url origin 2>/dev/null)
                    if [[ -z "$current_origin" || "$current_origin" != "$frontend_repo" ]]; then
                        echo "Меняем origin на $frontend_repo"
                        git remote remove origin 2>/dev/null
                        git remote add origin "$frontend_repo" || { echo "Ошибка при добавлении origin"; continue; }
                    fi

                    git fetch origin || { echo "Ошибка при git fetch"; continue; }
                    git reset --hard "origin/$frontend_branch" || { echo "Ошибка при git reset"; continue; }
                    git pull origin "$frontend_branch" || { echo "Ошибка при git pull"; continue; }
                    cd ..

                    echo "Обновление завершено!"
                fi
            else
                echo "Загрузить из архива - [Y] отмена обновления - [Any]"
                read -r confirm_archive
                if [[ "$confirm_archive" =~ ^[Yy]$ ]]; then
                    tar -xvvf GoldenHouseRepo.tar
                    echo "Обновлено при помощи архива"
                else
                    echo "Обновление отменено"
                fi
            fi
            sleep 2
            ;;

        3 | images)
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
            sleep 2
            ;;
        4 | volumes)
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
                    echo "Удалять Docker хранилища, данными сервиса рекламных кабинетов и с данными экспорта из системы - [Y] / [Any]"
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
            sleep 2
            ;;
        5 | network)
            echo "Удалить сеть приложения?  - [Y] / [Any]"
            read -r confirm_net
            if [[ "$confirm_net" =~ ^[Yy]$ ]]; then  
                docker network rm "$NETWORK_NAME"
                echo "Cеть приложения удалена."
            fi
            sleep 2
            ;;
        6 | start)
            echo "Запуск приложения..."
            sleep 1
            docker compose up -d
            sleep 2
            ;;
        7 | purge)
            if [[ "$mode" == "dev" ]]; then
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
                fi
            else 
                echo "Недостаточно прав для выполнения команды"
            fi
            sleep 2
            ;;
        8 | exit)
            echo "Завершение работы..."
            exit 0
            return
            ;;
        clear)
            clear
            echo -e "Терминал очищен."
            ;;
        *)
            echo "Такой команды нет, попробуйте еще раз."
            sleep 2
            ;;
    esac
done