# Подробная инструкция по запуску\обновлению приложения при помощи скрипта

## Что это такое?
Этот скрипт - это простая программа, которая поможет вам управлять Docker-приложением. Представьте, что это меню в кафе, где вместо блюд вы выбираете различные команды для работы с программой.

## Что нужно установить перед началом работы
1. **Docker** - это основная программа, без которой ничего не заработает
2. **Docker Compose** - помощник для Docker
3. **Git** - программа для загрузки обновлений

## Как запустить скрипт?
1. Откройте терминал (командную строку)
2. Перейдите в папку со скриптом (используйте команду `cd путь_к_папке`)
3. Выполните следующие команды:
```bash
# Сначала дайте скрипту права на запуск (при необходимости)
chmod +x menu.sh

# Затем запустите скрипт одним из способов:
./menu.sh           # для обычного режима
./menu.sh dev      # для режима разработчика (больше возможностей, не рекомендуется)
```

## Что умеет делать скрипт (подробное описание команд)

### 1. Проверка настроек (команда `envcheck` или `1`)
- Проверяет все настройки программы
- Если каких-то настроек не хватает, спросит их у вас
- Может показать все настройки, если вы захотите их посмотреть

### 2. Обновление программы (команда `update` или `2`)
- Может обновить программу двумя способами:
  1. Из интернета (GitHub) - выберите Y при вопросе про GitHub
  2. Из файла на компьютере (архива) - выберите Y при вопросе про архив
- При первом запуске выберите "Впервые загружаем из GitHub"

### 3. Управление образами (команда `images` или `3`)
- Образы - это как коробки с программами
- Можно удалить ненужные образы:
  - По одному (программа спросит про каждый)
  - Все сразу (будьте осторожны!)
- Отдельно спросит про удаление баз данных (лучше не удалять, если не уверены)

### 4. Управление хранилищами (команда `volumes` или `4`)
- Хранилища - это места, где хранятся данные программы
- ⚠️ БУДЬТЕ ОСТОРОЖНЫ: удаление хранилищ удалит все данные!
- Можно удалять:
  - Только хранилища программы (безопаснее)
  - Хранилища баз данных (опасно!)
  - Все хранилища (очень опасно!)

### 5. Управление сетью (команда `network` или `5`)
- Просто удаляет связь между частями программы
- Безопасная операция, ничего страшного не случится

### 6. Запуск программы (команда `start` или `6`)
- Запускает все части программы
- Подождите немного после запуска

### 7. Полная очистка (команда `purge` или `7`)
- ⚠️ ОЧЕНЬ ОПАСНАЯ КОМАНДА!
- Доступна только в режиме разработчика
- Удаляет вообще всё
- Используйте только если точно знаете, что делаете

### 8. Выход (команда `exit` или `8`)
- Просто закрывает программу

### Дополнительно:
- Команда `clear` - очищает экран, если стало слишком много текста
- Перед каждым опасным действием программа спросит подтверждение
- Всегда можно ответить "N" или нажать любую клавишу, чтобы отменить действие
- При любых сомнениях лучше остановиться и спросить помощи

## Важное предупреждение!
⚠️ Если вы не уверены в том, что делаете:
1. НЕ удаляйте базы данных
2. НЕ используйте команду purge
3. Спросите помощи у более опытного коллеги

## Пошаговая инструкция для первого запуска
1. Откройте терминал
2. Перейдите в нужную папку
3. Выполните `chmod +x menu.sh`
4. Запустите `./menu.sh`
5. Выберите команду `envcheck` (1)
6. Следуйте инструкциям на экране
7. Выберите команду `start` (6)
8. Готово!