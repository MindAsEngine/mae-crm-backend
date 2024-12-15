-- Server version	5.7.43-47

DROP TABLE IF EXISTS `advertising_expenses`;
CREATE TABLE `advertising_expenses` (
  `id` int(11) NOT NULL DEFAULT '0' COMMENT 'Id',
  `company_id` int(11) unsigned NOT NULL COMMENT 'Company_id',
  `updated_at` timestamp NULL DEFAULT NULL COMMENT 'Timestamp модификации',
  `expenses_date` date NOT NULL COMMENT 'Дата начисления затрат',
  `expenses_summa` decimal(38,2) DEFAULT NULL COMMENT 'Сумма затрат',
  `complex` varchar(64) CHARACTER SET utf8 DEFAULT NULL COMMENT 'ЖК из справочника',
  `house` varchar(64) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Номер дома',
  `utm_source` varchar(64) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Utm_source',
  `utm_campaign` varchar(64) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Utm_campaign',
  `utm_medium` varchar(64) CHARACTER SET utf8 DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
DROP TABLE IF EXISTS `calls`;
CREATE TABLE `calls` (
  `id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT 'Id',
  `calls_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT 'Id звонка',
  `updated_at` timestamp NULL DEFAULT NULL COMMENT 'Timestamp модификации',
  `company_id` int(11) unsigned NOT NULL COMMENT 'Company_id',
  `call_date` timestamp NULL DEFAULT NULL COMMENT 'Дата/время звонка',
  `calls_status` varchar(16) CHARACTER SET utf8 NOT NULL COMMENT 'Статус звонка',
  `direction` varchar(16) CHARACTER SET utf8 NOT NULL COMMENT 'Направление звонка',
  `phone` varchar(16) CHARACTER SET utf8 NOT NULL COMMENT 'Телефон звонящего',
  `contacts_id` int(10) unsigned DEFAULT NULL COMMENT 'Контакт звонящего',
  `manager_id` int(11) unsigned NOT NULL COMMENT 'Менеджер звонка (ответил/позвонил)',
  `manager_ext` varchar(64) CHARACTER SET utf8 NOT NULL COMMENT 'Расширение телефона менеджера',
  `estate_id` int(11) unsigned DEFAULT NULL COMMENT 'Заявка звонка',
  `duration` int(4) unsigned NOT NULL COMMENT 'Длительность звонка, сек.',
  `vendor` varchar(32) CHARACTER SET utf8 NOT NULL COMMENT 'Вендор телефонии',
  `gateway_phone` varchar(64) CHARACTER SET utf8 NOT NULL COMMENT 'Шлюз звонка',
  `is_first_unique` tinyint(1) unsigned DEFAULT NULL COMMENT 'Признак первого уникального звонка',
  `is_group_call` tinyint(1) unsigned NOT NULL COMMENT 'Признак группового звонка',
  `is_no_target` int(1) unsigned NOT NULL COMMENT 'Нецелевой звонок',
  `callback_id` int(11) unsigned DEFAULT NULL COMMENT 'Ссылка на звонок перезвонивший по пропущенному',
  `callback_date` timestamp NULL DEFAULT NULL COMMENT 'Ссылка перезвона по пропущенному',
  `callback_users_id` int(11) unsigned DEFAULT NULL COMMENT 'Перезвонивший менеджер',
  PRIMARY KEY (`id`),
  KEY `call_date` (`call_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
DROP TABLE IF EXISTS `company_departments`;
CREATE TABLE `company_departments` (
  `id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT 'Id',
  `departments_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT 'Id отдела',
  `company_id` int(11) unsigned NOT NULL COMMENT 'Company_id',
  `updated_at` timestamp NULL DEFAULT NULL COMMENT 'Timestamp модификации',
  `department_name` varchar(64) CHARACTER SET utf8 NOT NULL COMMENT 'Отдел',
  `department_type` varchar(32) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Тип отдела',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
DROP TABLE IF EXISTS `estate_advertising_channels`;
CREATE TABLE `estate_advertising_channels` (
  `id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT 'Id',
  `company_id` int(11) unsigned NOT NULL COMMENT 'Company_id',
  `name` varchar(255) CHARACTER SET utf8 NOT NULL COMMENT 'Название канала',
  `is_archived` tinyint(1) NOT NULL COMMENT 'В архиве',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
DROP TABLE IF EXISTS `estate_attributes`;
CREATE TABLE `estate_attributes` (
  `id` varchar(160) CHARACTER SET utf8 NOT NULL DEFAULT '' COMMENT 'Id набора данных',
  `company_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT 'Company_id',
  `updated_at` timestamp NULL DEFAULT NULL COMMENT 'Timestamp модификации',
  `entity` varchar(128) CHARACTER SET utf8 NOT NULL DEFAULT '' COMMENT 'Сущность (contacts, estate_buy, estate_sell, estate_deal)',
  `entity_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT 'Id сущности',
  `attr_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT 'Id атрибута',
  `attr_value` varchar(255) CHARACTER SET utf8mb4 NOT NULL DEFAULT '' COMMENT 'Значение атрибута',
  PRIMARY KEY (`id`),
  KEY `entity` (`entity`),
  KEY `entity_id` (`entity_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
DROP TABLE IF EXISTS `estate_attributes_names`;
CREATE TABLE `estate_attributes_names` (
  `id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT 'Id атрибута',
  `attr_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT 'Id атрибута',
  `company_id` int(11) unsigned NOT NULL COMMENT 'Company_id',
  `updated_at` timestamp NULL DEFAULT NULL COMMENT 'Timestamp модификации',
  `attr_title` varchar(128) CHARACTER SET utf8 NOT NULL COMMENT 'Заголовок атрибута',
  `attr_type` enum('varchar','text','bool','int','decimal') CHARACTER SET utf8 NOT NULL COMMENT 'Тип атрибута',
  `attr_values` text CHARACTER SET utf8 NOT NULL COMMENT 'Список возможных значений',
  `is_multiple` int(1) unsigned NOT NULL COMMENT 'Признак возможности множества значений',
  `entity` varchar(128) CHARACTER SET utf8 NOT NULL COMMENT 'Сущность к которой принадлежит атрибут',
  PRIMARY KEY (`id`),
  KEY `attr_id` (`attr_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
DROP TABLE IF EXISTS `estate_buys`;
CREATE TABLE `estate_buys` (
  `id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT 'Id заявки',
  `estate_buy_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT 'Id заявки',
  `company_id` int(10) unsigned NOT NULL COMMENT 'Company_id',
  `date_added` date DEFAULT NULL COMMENT 'Дата (Y-m-d) добавления заявки',
  `date_modified` int(11) unsigned NOT NULL COMMENT 'Дата изменения',
  `updated_at` timestamp NULL DEFAULT NULL COMMENT 'Timestamp модификации',
  `contacts_id` int(10) unsigned DEFAULT NULL COMMENT 'Id контакта',
  `contacts_buy_type` tinyint(3) unsigned NOT NULL COMMENT 'Тип контакта, 0 - ФЛ, 1- ЮЛ',
  `contacts_buy_sex` varchar(1) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'Пол',
  `contacts_buy_marital_status` varchar(16) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Семейное положение',
  `contacts_buy_dob` date DEFAULT NULL COMMENT 'Дата рождения',
  `contacts_buy_geo_country_id` int(11) unsigned NOT NULL COMMENT 'Id страны покупателя',
  `contacts_buy_geo_country_name` varchar(128) CHARACTER SET utf8 DEFAULT '' COMMENT 'Название страны покупателя',
  `contacts_buy_geo_region_id` int(11) unsigned NOT NULL COMMENT 'Id региона покупателя',
  `contacts_buy_geo_region_name` varchar(64) CHARACTER SET utf8 DEFAULT '' COMMENT 'Название региона покупателя',
  `contacts_buy_geo_city_id` int(11) NOT NULL COMMENT 'Id города покупателя',
  `contacts_buy_geo_city_name` varchar(128) CHARACTER SET utf8 DEFAULT '' COMMENT 'Название города покупателя',
  `contacts_buy_geo_city_short_name` varchar(255) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Обозначение названия города покупателя',
  `type` varchar(10) CHARACTER SET utf8 NOT NULL DEFAULT 'living' COMMENT 'Метка типа  (buy|rent)',
  `category` varchar(16) CHARACTER SET utf8 NOT NULL COMMENT 'Метка категории  (flat|garage|storageroom|house|comm)',
  `status` tinyint(3) unsigned NOT NULL COMMENT 'Id статуса/этапа',
  `status_custom` int(11) unsigned DEFAULT NULL COMMENT 'Id подстатуса',
  `status_name` varchar(32) CHARACTER SET utf8mb4 DEFAULT NULL COMMENT 'Имя статуса',
  `custom_status_name` varchar(255) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Имя кастомного подстатуса',
  `status_reason_id` bigint(11) DEFAULT NULL COMMENT 'Тип причины перевода в неактивный статус',
  `is_primary_request` tinyint(1) unsigned NOT NULL DEFAULT '0' COMMENT 'Первичная заявка',
  `manager_id` int(10) unsigned NOT NULL COMMENT 'Менеджер заявки',
  `departments_id` int(11) unsigned DEFAULT NULL COMMENT 'Id отдела заявки',
  `geo_country_name` varchar(128) CHARACTER SET utf8 DEFAULT '' COMMENT 'Страна заявки',
  `geo_region_name` varchar(64) CHARACTER SET utf8 DEFAULT '' COMMENT 'Регион заявки',
  `geo_city_name` varchar(128) CHARACTER SET utf8 DEFAULT '' COMMENT 'Город заявки',
  `estate_sell_id` int(11) unsigned DEFAULT '0' COMMENT 'Id объекта в сделке',
  `house_id` int(11) unsigned DEFAULT NULL COMMENT 'Id дома объекта',
  `first_house_interest` int(11) DEFAULT NULL COMMENT 'Id дома - первого интереса заявки',
  `first_complex_interest` int(11) DEFAULT NULL COMMENT 'Id ЖК (справочное) - первого интереса заявки',
  `first_meetings_id` int(11) unsigned DEFAULT NULL COMMENT 'Id первой встречи по заявке',
  `first_meetings_house_id` int(11) unsigned DEFAULT NULL COMMENT 'Id дома первой встречи-показа',
  `first_meetings_office_id` int(11) unsigned DEFAULT NULL COMMENT 'Id первой встречи в офисе',
  `channel_type` varchar(32) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Тип источника заявки (www|office|agent|call)',
  `channel_name` varchar(255) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Имя источника (сайт, номер телефона/название канала)',
  `channel_medium` varchar(255) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Channel_medium',
  `utm_source` varchar(64) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Utm_source',
  `utm_medium` varchar(64) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Utm_medium',
  `utm_campaign` varchar(64) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Utm_campaign',
  `utm_content` varchar(255) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Utm_content',
  `deal_id` int(10) unsigned DEFAULT '0' COMMENT 'Id сделки',
  `is_payed_reserve` int(1) unsigned DEFAULT NULL COMMENT 'Признак платной брони',
  `deal_sum` decimal(16,2) unsigned DEFAULT NULL COMMENT 'Сумма сделки',
  `deal_price` decimal(16,2) unsigned DEFAULT NULL COMMENT 'Цена объекта на момент начала оформления сделки',
  `deal_area` decimal(10,4) DEFAULT NULL COMMENT 'Площадь в сделке',
  `deal_sum_addons` decimal(16,2) unsigned DEFAULT NULL COMMENT 'Сумма допов в сделке',
  `deal_date` date DEFAULT NULL COMMENT 'Дата проведения сделки',
  `agreement_type` varchar(16) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Тип сделки (ДДУ, ДУСТ и тд)',
  `is_concession` int(1) unsigned DEFAULT NULL COMMENT 'Признак договора уступки',
  `deal_mediator_comission` decimal(16,4) unsigned DEFAULT NULL COMMENT 'Сумма комиссии агенту в сделке',
  `deal_program_name` varchar(255) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Программа покупки',
  `ipoteka_bank_name` varchar(255) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Имя ипотечного банка в сделке',
  `ipoteka_rate` decimal(5,3) unsigned DEFAULT NULL COMMENT 'Ставка по ипотеке',
  `agent_name` varchar(255) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Агент заявки/сделки',
  `agency_name` varchar(255) CHARACTER SET utf8 DEFAULT NULL,
  `advertising_channel_id` int(11) DEFAULT NULL COMMENT 'Id рекламного канала',
  PRIMARY KEY (`id`),
  KEY `date_added` (`date_added`),
  KEY `status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
DROP TABLE IF EXISTS `estate_buys_attributes`;
CREATE TABLE `estate_buys_attributes` (
  `id` varchar(160) CHARACTER SET utf8 NOT NULL DEFAULT '' COMMENT 'Id записи',
  `company_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT 'Company_id',
  `updated_at` timestamp NULL DEFAULT NULL COMMENT 'Timestamp модификации',
  `entity` varchar(128) CHARACTER SET utf8 NOT NULL DEFAULT '' COMMENT 'Сущность (contacts, estate_buy, estate_deal)',
  `entity_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT 'Id сущности',
  `attr_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT 'Id атрибута',
  `attr_value` varchar(255) CHARACTER SET utf8mb4 NOT NULL DEFAULT '' COMMENT 'Значение атрибута',
  PRIMARY KEY (`id`),
  KEY `entity` (`entity`),
  KEY `entity_id` (`entity_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
DROP TABLE IF EXISTS `estate_buys_attributes_names`;
CREATE TABLE `estate_buys_attributes_names` (
  `id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT 'Id атрибута',
  `attr_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT 'Id атрибута',
  `company_id` int(11) unsigned NOT NULL COMMENT 'Company_id',
  `updated_at` timestamp NULL DEFAULT NULL COMMENT 'Timestamp модификации',
  `attr_title` varchar(128) CHARACTER SET utf8 NOT NULL COMMENT 'Заголово атрибута',
  `attr_type` enum('varchar','text','bool','int','decimal') CHARACTER SET utf8 NOT NULL COMMENT 'Тип атрибута',
  `attr_values` text CHARACTER SET utf8 NOT NULL COMMENT 'Список возможных значений',
  `is_multiple` int(1) unsigned NOT NULL COMMENT 'Признак возможности множества значений',
  `entity` varchar(128) CHARACTER SET utf8 NOT NULL COMMENT 'Сущность к которой принадлежит атрибут',
  PRIMARY KEY (`id`),
  KEY `attr_id` (`attr_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
DROP TABLE IF EXISTS `estate_buys_statuses_log`;
CREATE TABLE `estate_buys_statuses_log` (
  `id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT 'Id',
  `company_id` int(10) unsigned NOT NULL COMMENT 'Company_id',
  `log_date` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00' COMMENT 'Дата события',
  `estate_buy_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT 'Id заявки',
  `deal_id` int(11) unsigned DEFAULT NULL COMMENT 'Id сделки в момент события',
  `deal_sum` decimal(16,2) unsigned DEFAULT NULL COMMENT 'Сумма сделки в момент события',
  `is_payed_reserve` tinyint(1) unsigned NOT NULL DEFAULT '0' COMMENT 'Признак включения платной брони',
  `status_from` tinyint(3) unsigned NOT NULL COMMENT 'Исходный статус',
  `status_from_name` varchar(32) CHARACTER SET utf8mb4 DEFAULT NULL COMMENT 'Название исходного статуса',
  `status_to` tinyint(3) unsigned NOT NULL COMMENT 'Новый статус',
  `status_to_name` varchar(32) CHARACTER SET utf8mb4 DEFAULT NULL COMMENT 'Название нового статуса',
  `status_custom_from` int(11) unsigned DEFAULT NULL COMMENT 'Исходный подстатус',
  `status_custom_from_name` varchar(255) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Название исходного подстатуса',
  `status_custom_to` int(11) unsigned DEFAULT NULL COMMENT 'Новый кастомный подстатус',
  `status_custom_to_name` varchar(255) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Название нового подстатуса',
  PRIMARY KEY (`id`),
  KEY `estate_buy_id` (`estate_buy_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
DROP TABLE IF EXISTS `estate_deals`;
CREATE TABLE `estate_deals` (
  `id` int(10) unsigned NOT NULL DEFAULT '0' COMMENT 'Id',
  `company_id` int(10) unsigned NOT NULL COMMENT 'Company_id',
  `deal_id` int(10) unsigned NOT NULL DEFAULT '0' COMMENT 'Id сделки',
  `deal_status` tinyint(3) unsigned NOT NULL COMMENT 'Статус сделки',
  `updated_at` timestamp NULL DEFAULT NULL COMMENT 'Timestamp модификации',
  `deal_status_name` varchar(32) CHARACTER SET utf8mb4 DEFAULT NULL COMMENT 'Имя статуса сделки',
  `estate_buy_id` int(10) unsigned NOT NULL COMMENT 'Id заявки в сделке',
  `estate_sell_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT 'Id объекта в сделке',
  `buy_deal_shows_id` int(11) NOT NULL COMMENT 'Id сделки в заявке (для проверки)',
  `sell_deal_shows_id` int(11) NOT NULL COMMENT 'Id сделки в объекте (для проверки)',
  `house_id` int(11) unsigned NOT NULL COMMENT 'Id дома объекта',
  `seller_contacts_id` int(11) unsigned DEFAULT '0' COMMENT 'Id продавца',
  `seller_contacts_name` varchar(255) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Продавец',
  `date_finished_plan` date DEFAULT NULL COMMENT 'Плановая дата сделки',
  `date_modified` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00' COMMENT 'Дата изменения',
  `deal_date` date DEFAULT NULL COMMENT 'Дата проведения сделки',
  `deal_date_start` date DEFAULT NULL COMMENT 'Дата начала оформления сделки',
  `deal_date_cancelled` date DEFAULT NULL COMMENT 'Дата расторжения заключенной сделки',
  `deal_date_combined` date DEFAULT NULL COMMENT 'Служебное поле',
  `deal_manager_id` int(10) unsigned NOT NULL COMMENT 'Менеджер сделки',
  `estate_buy_date_added` date DEFAULT NULL COMMENT 'Дата (Y-m-d) добавления заявки',
  `reserve_date` date DEFAULT NULL COMMENT 'Дата (Y-m-d) окончания брони',
  `reserve_date_start` date DEFAULT NULL COMMENT 'Дата (Y-m-d) постановки брони',
  `is_payed_reserve` int(1) unsigned NOT NULL COMMENT 'Признак платной брони',
  `deal_sum` decimal(16,2) unsigned NOT NULL COMMENT 'Сумма сделки',
  `deal_price` decimal(16,2) unsigned NOT NULL COMMENT 'Цена объекта на момент начала оформления сделки',
  `deal_area` decimal(10,4) NOT NULL COMMENT 'Площадь в сделке',
  `deal_sum_addons` decimal(16,2) unsigned NOT NULL COMMENT 'Сумма допов в сделке',
  `agreement_type` varchar(16) CHARACTER SET utf8 NOT NULL COMMENT 'Тип сделки (ДДУ, ДУСТ и тд)',
  `agreement_number` varchar(64) CHARACTER SET utf8 NOT NULL COMMENT 'Номер договора (в печатной форме)',
  `agreement_date` date DEFAULT NULL COMMENT 'Дата договора (в печатной форме)',
  `preliminary_date` date DEFAULT NULL COMMENT 'Дата предварительного договора',
  `is_preliminary` int(1) unsigned NOT NULL DEFAULT '0' COMMENT 'Признак наличия предварительного договора',
  `signed_date` date DEFAULT NULL COMMENT 'Дата подписания клиентом',
  `signed_by_company_date` date DEFAULT NULL COMMENT 'Дата подписания компанией',
  `arles_agreement_date` date DEFAULT NULL COMMENT 'Дата договора бронирования',
  `arles_agreement_num` varchar(64) CHARACTER SET utf8 NOT NULL COMMENT 'Номер договора бронирования',
  `agreement_osnova_date` date DEFAULT NULL COMMENT 'Дата договора основания',
  `agreement_verified_date` timestamp NULL DEFAULT NULL COMMENT 'Дата проверки договора',
  `terms_approved_send` timestamp NULL DEFAULT NULL COMMENT 'Дата отправки на согласование',
  `terms_approved_date` timestamp NULL DEFAULT NULL COMMENT 'Дата согласования договора',
  `justice_registration_method` int(11) unsigned DEFAULT NULL COMMENT 'Способ передачи на регистрацию',
  `justice_date_send_plan` date DEFAULT NULL COMMENT 'Плановая дата отправки на регистрацию',
  `justice_date_send` date DEFAULT NULL COMMENT 'Дата отправки на регистрацию',
  `justice_date_received_plan` date DEFAULT NULL COMMENT 'Плановая дата возврата с регистрации',
  `justice_date_received` date DEFAULT NULL COMMENT 'Фактическая дата возврата с регистрации',
  `justice_date` date DEFAULT NULL COMMENT 'Дата регистрации',
  `justice_number` varchar(256) CHARACTER SET utf8 NOT NULL COMMENT 'Номер регистрации',
  `registration_users_id` int(11) unsigned DEFAULT NULL COMMENT 'Ответственный за регистрацию сотрудник',
  `is_concession` int(1) unsigned NOT NULL COMMENT 'Признак договора уступки',
  `bulk_deal_id` int(10) unsigned DEFAULT NULL COMMENT 'Id оптовой сделки (для главной = deal_id)',
  `is_bulk` int(1) NOT NULL DEFAULT '0' COMMENT 'Признак оптовой сделки',
  `bulk_deal_sum` decimal(16,2) unsigned DEFAULT NULL COMMENT 'Стоимость оптовой сделки',
  `bulk_deal_sum_m2` decimal(16,2) unsigned DEFAULT NULL COMMENT 'Стоимость за м2 оптовой сделки',
  `bulk_deal_area` decimal(10,4) unsigned DEFAULT NULL COMMENT 'Площадь оптовой сделки',
  `agreement_owner_date` date DEFAULT NULL COMMENT 'Дата подписания акта п/п',
  `deal_program_name` varchar(255) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Программа покупки',
  `has_ipoteka` int(1) NOT NULL DEFAULT '0' COMMENT 'Признак ипотечной сделки',
  `ipoteka_bank_name` varchar(255) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Имя ипотечного банка в сделке',
  `ipoteka_rate` decimal(5,3) unsigned DEFAULT NULL COMMENT 'Ставка по ипотеке',
  `agreement_city_name` varchar(32) CHARACTER SET utf8mb4 DEFAULT NULL COMMENT 'Город ипотечного банка',
  `bank_first_income` decimal(16,2) DEFAULT NULL COMMENT 'Сумма первоначального взноса',
  `bank_commission` decimal(16,2) DEFAULT NULL COMMENT 'Комиссия банка',
  `bank_agreement_term` int(3) unsigned NOT NULL COMMENT 'Срок кредита, мес.',
  `has_agent` int(1) NOT NULL DEFAULT '0' COMMENT 'Признак агентской сделки',
  `deal_mediator_comission` decimal(16,4) unsigned NOT NULL COMMENT 'Сумма комиссии агенту в сделке',
  `agent_name` varchar(255) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Агент заявки/сделки',
  `agency_name` varchar(255) CHARACTER SET utf8 DEFAULT NULL,
  `contacts_buy_id` int(10) unsigned DEFAULT NULL COMMENT 'Id главного покупателя',
  `estate_client_aim` varchar(32) CHARACTER SET utf8mb4 DEFAULT NULL COMMENT 'Цель приобретения',
  `mother_capital_cert_sum` decimal(16,4) DEFAULT NULL COMMENT 'Сумма материнского капитала',
  `contacts_buy_type` tinyint(3) unsigned DEFAULT NULL COMMENT 'Тип контакта, 0 - ФЛ, 1- ЮЛ',
  `contacts_buy_sex` varchar(1) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'Пол',
  `contacts_buy_marital_status` varchar(16) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Семейное положение',
  `contacts_buy_dob` date DEFAULT NULL COMMENT 'Дата рождения',
  `deal_contacts_count` int(1) DEFAULT NULL COMMENT 'Участников в сделке',
  `status` tinyint(3) unsigned NOT NULL COMMENT 'Id статуса заявки',
  `status_custom` int(11) unsigned DEFAULT NULL COMMENT 'Id подстатуса заявки',
  `custom_status_name` varchar(255) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Имя кастомного подстатуса заявки',
  `is_primary_request` tinyint(1) unsigned NOT NULL DEFAULT '0' COMMENT 'Первичная заявка',
  `manager_id` int(10) unsigned NOT NULL COMMENT 'Менеджер заявки',
  `departments_id` int(11) unsigned DEFAULT NULL COMMENT 'Id отдела заявки',
  `finances_income` decimal(16,2) unsigned DEFAULT NULL COMMENT 'Поступления по графику сделки',
  `finances_income_mortgage` decimal(16,2) unsigned DEFAULT NULL COMMENT 'Поступления ипотечных платежей по графику сделки',
  `finances_income_reserved` decimal(16,2) unsigned DEFAULT NULL COMMENT 'Ожидаемые поступления по графику сделки',
  `finances_income_reserved_mortgage` decimal(16,2) unsigned DEFAULT NULL COMMENT 'Ожидаемые поступления по графику сделки ипотечных платежей',
  `finances_other_income` decimal(16,2) unsigned DEFAULT NULL COMMENT 'Другие поступления по сделке',
  `finances_other_income_reserved` decimal(16,2) unsigned DEFAULT NULL COMMENT 'Другие ожидаемые поступления по сделке',
  `finances_over_deal_sum` decimal(16,2) unsigned DEFAULT NULL COMMENT 'Поступления по сделке сверх суммы договора',
  `finances_over_deal_sum_reserved` decimal(16,2) unsigned DEFAULT NULL COMMENT 'Ожидаемые поступления по сделке сверх суммы договора',
  `finances_income_date_first` timestamp NULL DEFAULT NULL COMMENT 'Дата первого пришедшего поступления',
  `finances_income_date_last` timestamp NULL DEFAULT NULL COMMENT 'Дата последнего пришедшего поступления',
  `first_meetings_id` int(11) unsigned DEFAULT NULL COMMENT 'Id первой встречи по заявке',
  `first_meetings_house_id` int(11) unsigned DEFAULT NULL COMMENT 'Id первой встречи-показа',
  `first_meetings_office_id` int(11) unsigned DEFAULT NULL COMMENT 'Id первой встречи в офисе',
  `channel_type` varchar(32) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Тип источника заявки (www|office|agent|call)',
  `channel_name` varchar(255) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Имя источника (сайт, номер телефона/название канала)',
  `channel_medium` varchar(255) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Channel_medium',
  `utm_source` varchar(64) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Utm_source',
  `utm_medium` varchar(64) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Utm_medium',
  `utm_campaign` varchar(64) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Utm_campaign',
  `utm_content` varchar(255) CHARACTER SET utf8 DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `deal_date` (`deal_date`),
  KEY `deal_status` (`deal_status`),
  KEY `house_id` (`house_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
DROP TABLE IF EXISTS `estate_deals_addons`;
CREATE TABLE `estate_deals_addons` (
  `id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT 'Id наценки в сделке',
  `company_id` int(10) unsigned NOT NULL COMMENT 'Company_id',
  `deal_id` int(10) unsigned NOT NULL DEFAULT '0' COMMENT 'Id сделки',
  `deal_date_combined` date DEFAULT NULL COMMENT 'Служебное поле',
  `updated_at` timestamp NULL DEFAULT NULL COMMENT 'Timestamp модификации',
  `addon_name` varchar(255) CHARACTER SET utf8 NOT NULL COMMENT 'Имя наценки',
  `addon_price_default` decimal(16,2) unsigned NOT NULL COMMENT 'Величина наценки по-умолчанию',
  `addon_price` decimal(16,2) NOT NULL COMMENT 'Величина наценки в сделке',
  PRIMARY KEY (`id`),
  KEY `deal_id` (`deal_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
DROP TABLE IF EXISTS `estate_deals_contacts`;
CREATE TABLE `estate_deals_contacts` (
  `id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT 'Id',
  `company_id` int(10) unsigned DEFAULT NULL COMMENT 'Company_id',
  `date_added` date DEFAULT NULL COMMENT 'Дата (Y-m-d) добавления контакта',
  `date_modified` datetime DEFAULT NULL,
  `contacts_buy_type` tinyint(3) unsigned NOT NULL COMMENT 'Тип контакта, 0 - ФЛ, 1- ЮЛ',
  `contacts_buy_sex` varchar(1) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'Пол',
  `contacts_buy_marital_status` varchar(16) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Семейное положение',
  `contacts_buy_dob` date DEFAULT NULL COMMENT 'Дата рождения',
  `contacts_buy_name` varchar(255) CHARACTER SET utf8 NOT NULL DEFAULT '' COMMENT 'ФИО/название ЮЛ',
  `contacts_buy_phones` varchar(255) CHARACTER SET utf8 NOT NULL DEFAULT '' COMMENT 'Contacts_buy_phones',
  `contacts_buy_emails` varchar(255) CHARACTER SET utf8 NOT NULL DEFAULT '' COMMENT 'Contacts_buy_emails',
  `passport_bithplace` varchar(100) CHARACTER SET utf8 NOT NULL DEFAULT '' COMMENT 'Место рождения',
  `passport_address` varchar(255) CHARACTER SET utf8 NOT NULL DEFAULT '',
  `comm_inn` varchar(12) CHARACTER SET utf8 NOT NULL DEFAULT '' COMMENT 'ИНН ЮЛ',
  `comm_kpp` varchar(9) CHARACTER SET utf8 NOT NULL DEFAULT '' COMMENT 'КПП ЮЛ',
  `fl_inn` varchar(14) CHARACTER SET utf8 NOT NULL DEFAULT '' COMMENT 'ИНН ФЛ',
  `roles_set` varchar(255) CHARACTER SET utf8 NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
DROP TABLE IF EXISTS `estate_deals_discounts`;
CREATE TABLE `estate_deals_discounts` (
  `id` int(11) NOT NULL DEFAULT '0' COMMENT 'Id корректировки',
  `company_id` int(10) unsigned NOT NULL COMMENT 'Company_id',
  `deal_id` int(10) unsigned NOT NULL DEFAULT '0' COMMENT 'Id сделки',
  `updated_at` timestamp NULL DEFAULT NULL COMMENT 'Timestamp модификации',
  `deal_date_combined` varchar(11) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'Служебное поле',
  `promo_id` int(11) unsigned DEFAULT NULL COMMENT 'Id акции',
  `type` enum('discount','drop','increase','restoration') CHARACTER SET utf8 NOT NULL DEFAULT 'discount' COMMENT 'Discount | increase',
  `amount` decimal(16,2) NOT NULL COMMENT 'Сумма корректировки',
  `rule` enum('discount','discount_m2','discount_none') CHARACTER SET utf8 NOT NULL DEFAULT 'discount' COMMENT 'Правило корректировки',
  `rule_type` enum('cash','percent') CHARACTER SET utf8 NOT NULL DEFAULT 'cash' COMMENT 'Способ корректировки',
  `rule_value` decimal(16,2) NOT NULL,
  `comment` varchar(255) CHARACTER SET utf8mb4 DEFAULT NULL COMMENT 'Комментарий к корректировке',
  PRIMARY KEY (`id`),
  KEY `deal_id` (`deal_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
DROP TABLE IF EXISTS `estate_deals_docs`;
CREATE TABLE `estate_deals_docs` (
  `id` int(11) NOT NULL DEFAULT '0' COMMENT 'Id',
  `company_id` int(10) unsigned NOT NULL COMMENT 'Company_id',
  `updated_at` timestamp NULL DEFAULT NULL COMMENT 'Дата обновления записи',
  `deal_id` int(11) unsigned NOT NULL COMMENT 'Id сделки',
  `document_type` varchar(32) CHARACTER SET utf8mb4 NOT NULL DEFAULT '' COMMENT 'Тип документа',
  `document_type_name` varchar(64) CHARACTER SET utf8mb4 DEFAULT NULL COMMENT 'Наименование типа документа',
  `users_id` int(11) unsigned NOT NULL COMMENT 'Пользователь, добавивший документ',
  `document_date` date NOT NULL COMMENT 'Дата документа',
  `document_number` varchar(64) CHARACTER SET utf8mb4 NOT NULL DEFAULT '' COMMENT 'Номер документа',
  `registration_number` varchar(64) CHARACTER SET utf8mb4 NOT NULL DEFAULT '' COMMENT 'Рег.номер документа',
  `date_registration` date NOT NULL COMMENT 'Дата регистрации документа',
  `prev_area` decimal(16,2) unsigned DEFAULT NULL COMMENT 'Предыдущая площадь сделки',
  `prev_summa` decimal(16,2) NOT NULL COMMENT 'Предыдущая сумма сделки',
  `document_summa` decimal(16,2) NOT NULL COMMENT 'Сумма по документу',
  `document_area` decimal(16,2) unsigned DEFAULT NULL COMMENT 'Площадь по документу',
  `has_file` int(1) NOT NULL DEFAULT '0' COMMENT 'Признак наличия подгруженного файла',
  PRIMARY KEY (`id`),
  KEY `document_date` (`document_date`),
  KEY `deal_id` (`deal_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
DROP TABLE IF EXISTS `estate_deals_statuses`;
CREATE TABLE `estate_deals_statuses` (
  `status_id` bigint(20) NOT NULL DEFAULT '0',
  `status_name` varchar(19) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  PRIMARY KEY (`status_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
DROP TABLE IF EXISTS `estate_houses`;
CREATE TABLE `estate_houses` (
  `id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT 'Id дома',
  `house_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT 'Id дома',
  `company_id` int(10) unsigned NOT NULL COMMENT 'Company_id',
  `updated_at` timestamp NULL DEFAULT NULL COMMENT 'Timestamp модификации',
  `status` tinyint(3) unsigned NOT NULL COMMENT 'Статус дома',
  `complex_id` int(11) unsigned DEFAULT '0' COMMENT 'Id группы домов',
  `complex_name` varchar(255) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Имя группы домов',
  `house_category` varchar(16) CHARACTER SET utf8 NOT NULL COMMENT 'Категория дома',
  `geo_city_complex_id` int(11) DEFAULT NULL COMMENT 'Id ЖК (справочник)',
  `buildState` varchar(1000) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Состояние проекта',
  `inServiceState` int(11) DEFAULT NULL COMMENT 'Признак ввода в эксплуатацию',
  `inServiceDate` int(11) DEFAULT NULL COMMENT 'Дата ввода в эксплуатацию',
  `inServiceMonth` int(11) DEFAULT NULL COMMENT 'Месяц ввода в эксплуатацию',
  `inServiceQuartal` int(11) DEFAULT NULL COMMENT 'Квартал ввода в эксплуатацию',
  `inServiceYear` int(11) DEFAULT NULL COMMENT 'Год ввода в эксплуатацию',
  `group_sellStart` date DEFAULT NULL COMMENT 'Дата начала продаж',
  `public_house_name` varchar(1000) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Публичное имя дома',
  `group_code` varchar(1000) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Код группы',
  `geo_country_name` varchar(128) CHARACTER SET utf8 DEFAULT '' COMMENT 'Адрес дома: страна',
  `geo_region_name` varchar(64) CHARACTER SET utf8 DEFAULT '' COMMENT 'Адрес дома: регион',
  `geo_city_name` varchar(128) CHARACTER SET utf8 DEFAULT '' COMMENT 'Адрес дома: город',
  `geo_city_short_name` varchar(255) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Адрес дома: обозначение города',
  `geo_street_name` varchar(255) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Адрес дома: улица',
  `geo_street_short_name` varchar(32) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Адрес дома: обозначение улицы',
  `geo_house` varchar(1000) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Адрес дома: номер',
  `geo_building` varchar(1000) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Адрес дома: строение',
  `geo_korpus` varchar(1000) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Адрес дома: корпус',
  `geo_block` varchar(1000) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Адрес дома: секция',
  `geo_quarter` varchar(1000) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Адрес дома: квартал',
  `estate_buildingQueue` varchar(1000) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Очередь стр-ва',
  `seller_name` varchar(255) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Продавец',
  `name` varchar(255) CHARACTER SET utf8mb4 NOT NULL DEFAULT '' COMMENT 'Имя дома',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
DROP TABLE IF EXISTS `estate_houses_price_stat`;
CREATE TABLE `estate_houses_price_stat` (
  `id` int(11) NOT NULL DEFAULT '0' COMMENT 'Id',
  `company_id` int(10) unsigned NOT NULL COMMENT 'Company_id',
  `house_id` int(11) unsigned NOT NULL COMMENT 'Id дома',
  `month_stat_date` date NOT NULL COMMENT 'Дата фиксации',
  `category` varchar(16) CHARACTER SET utf8 NOT NULL COMMENT 'Категория дома',
  `flat_class` varchar(32) CHARACTER SET utf8 NOT NULL COMMENT 'Класс объектов',
  `avg_price` int(11) unsigned NOT NULL COMMENT 'Средняя цена объектов',
  `avg_price_m2` int(11) unsigned NOT NULL COMMENT 'Средняя цена за м² объектов',
  PRIMARY KEY (`id`),
  KEY `house_id` (`house_id`),
  KEY `month_stat_date` (`month_stat_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
DROP TABLE IF EXISTS `estate_meetings`;
CREATE TABLE `estate_meetings` (
  `id` int(11) NOT NULL DEFAULT '0' COMMENT 'Id',
  `meetings_id` int(11) NOT NULL DEFAULT '0' COMMENT 'Id встречи',
  `company_id` int(10) unsigned NOT NULL COMMENT 'Company_id',
  `updated_at` timestamp NULL DEFAULT NULL COMMENT 'Timestamp модификации',
  `estate_buy_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT 'Id заявки',
  `contacts_id` int(11) unsigned DEFAULT NULL COMMENT 'Id контакта с которым проводилась встреча',
  `users_id` int(11) unsigned NOT NULL COMMENT 'Id менеджера, проводившего встречу',
  `date_added` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00' COMMENT 'Дата добавления отчета по встрече',
  `meeting_date` date DEFAULT NULL COMMENT 'Дата учета встречи',
  `meeting_type` varchar(16) CHARACTER SET utf8 NOT NULL DEFAULT 'office' COMMENT 'Тип встречи: офис/объект',
  `complex_id` int(11) unsigned DEFAULT NULL COMMENT 'Id группы домов',
  `house_id` int(11) NOT NULL COMMENT 'Id дома встречи',
  `no_meeting` int(1) unsigned DEFAULT '0' COMMENT 'Признак несостоявшейся встречи',
  `is_first_meeting` int(1) NOT NULL DEFAULT '0' COMMENT 'Признак первой встречи',
  `is_last_meeting` int(1) NOT NULL DEFAULT '0' COMMENT 'Признак последней встречи',
  PRIMARY KEY (`id`),
  KEY `date_added` (`date_added`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
DROP TABLE IF EXISTS `estate_promos`;
CREATE TABLE `estate_promos` (
  `id` int(11) NOT NULL DEFAULT '0' COMMENT 'Id',
  `company_id` int(10) unsigned NOT NULL COMMENT 'Company_id',
  `promo_id` int(11) unsigned NOT NULL COMMENT 'Акция',
  `estate_sell_id` int(11) unsigned NOT NULL COMMENT 'Объект',
  `price` decimal(16,2) unsigned DEFAULT NULL COMMENT 'Цена объекта по данной акции',
  PRIMARY KEY (`id`),
  KEY `estate_sell_id` (`estate_sell_id`),
  KEY `promo_id` (`promo_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
DROP TABLE IF EXISTS `estate_sales_plans`;
CREATE TABLE `estate_sales_plans` (
  `id` int(11) NOT NULL DEFAULT '0' COMMENT 'Id',
  `company_id` int(11) unsigned NOT NULL COMMENT 'Company_id',
  `updated_at` timestamp NULL DEFAULT NULL COMMENT 'Timestamp модификации',
  `title` varchar(255) CHARACTER SET utf8 NOT NULL COMMENT 'Название плана',
  `levels` text CHARACTER SET utf8 NOT NULL COMMENT 'Используемые уровни (в порядке иерархии)',
  `indicators` text CHARACTER SET utf8 NOT NULL COMMENT 'Используемые метрики',
  `period` varchar(255) CHARACTER SET utf8 NOT NULL COMMENT 'Период планирования',
  `is_independent` tinyint(1) unsigned NOT NULL DEFAULT '0' COMMENT 'Признак независимости плана',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
DROP TABLE IF EXISTS `estate_sales_plans_metrics`;
CREATE TABLE `estate_sales_plans_metrics` (
  `id` varchar(23) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'Id набора данных',
  `company_id` int(11) unsigned NOT NULL COMMENT 'Company_id',
  `updated_at` timestamp NULL DEFAULT NULL COMMENT 'Timestamp модификации',
  `plan_id` int(11) NOT NULL DEFAULT '0' COMMENT 'Id плана',
  `metrics_id` int(11) NOT NULL DEFAULT '0' COMMENT 'Id записи метрики',
  `plan_date` date DEFAULT NULL COMMENT 'Дата плана %Y-%m-15',
  `year` int(4) unsigned NOT NULL COMMENT 'Год',
  `quarter` int(2) unsigned DEFAULT NULL COMMENT 'Квартал',
  `month` int(2) unsigned DEFAULT NULL COMMENT 'Месяц',
  `complex_id` int(11) unsigned DEFAULT NULL COMMENT 'Группа домов',
  `house_id` int(11) unsigned DEFAULT NULL COMMENT 'Дом',
  `manager_id` int(11) unsigned DEFAULT NULL COMMENT 'Менеджер',
  `category` varchar(30) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Категория квартир',
  `rooms` int(2) DEFAULT NULL COMMENT 'Комнатность',
  `is_studio` tinyint(1) DEFAULT NULL COMMENT 'Признак студии',
  `departments_id` int(11) unsigned DEFAULT NULL COMMENT 'Отдел',
  `estate_class` varchar(30) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Класс квартиры',
  `deal_programs` int(11) unsigned DEFAULT NULL COMMENT 'Программа покупки',
  `finances_income` decimal(16,2) unsigned DEFAULT NULL COMMENT 'Сумма привлеченных денег',
  `price_m2` decimal(16,2) unsigned DEFAULT NULL COMMENT 'Стоимость за м2',
  `quantity` decimal(16,2) unsigned DEFAULT NULL COMMENT 'Сделок, шт.',
  `sum` decimal(16,2) unsigned DEFAULT NULL COMMENT 'Сумма продаж',
  `area` decimal(16,2) unsigned DEFAULT NULL COMMENT 'Объем продаж, м2',
  `leads` int(16) unsigned DEFAULT NULL COMMENT 'Количество заявок',
  `deal_price` decimal(16,2) unsigned DEFAULT NULL COMMENT 'Плановая цена сделки',
  `provision_method` varchar(30) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Способ обеспечения',
  PRIMARY KEY (`id`),
  KEY `plan_id` (`plan_id`),
  KEY `year` (`year`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
DROP TABLE IF EXISTS `estate_sells`;
CREATE TABLE `estate_sells` (
  `id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT 'Id',
  `estate_sell_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT 'Estate_sell_id',
  `company_id` int(10) unsigned NOT NULL COMMENT 'Company_id',
  `date_modified` int(11) unsigned NOT NULL COMMENT 'Дата изменения',
  `updated_at` timestamp NULL DEFAULT NULL COMMENT 'Timestamp модификации',
  `estate_sell_type` varchar(10) CHARACTER SET utf8 NOT NULL DEFAULT 'living' COMMENT 'Метка типа  (sell)',
  `estate_sell_category` varchar(16) CHARACTER SET utf8 NOT NULL COMMENT 'Метка категория  (flat|garage|storageroom|house|comm)',
  `estate_sell_status` tinyint(3) unsigned NOT NULL COMMENT 'Id статуса/этапа',
  `estate_sell_status_name` varchar(32) CHARACTER SET utf8mb4 DEFAULT NULL COMMENT 'Имя статуса',
  `house_id` int(11) unsigned NOT NULL COMMENT 'Id дома',
  `seller_contacts_id` int(11) unsigned DEFAULT '0' COMMENT 'Id продавца',
  `seller_contacts_name` varchar(255) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Продавец',
  `plans_name` varchar(64) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Планировка',
  `plans_group` varchar(32) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Группа планировок',
  `flatClass` varchar(32) CHARACTER SET utf8mb4 DEFAULT NULL COMMENT 'Класс квартиры',
  `estate_studia` int(11) DEFAULT NULL COMMENT 'Признак студии',
  `estate_apartments` int(11) DEFAULT NULL COMMENT 'Признак апартаментов',
  `estate_rooms` int(11) DEFAULT NULL COMMENT 'Комнат',
  `geo_house_entrance` int(11) DEFAULT NULL COMMENT 'Подъезд/секция',
  `estate_floor` int(11) DEFAULT NULL COMMENT 'Этаж',
  `estate_riser` int(11) DEFAULT NULL COMMENT 'Номер на площадке (стояк)',
  `geo_flatnum` varchar(32) CHARACTER SET utf8mb4 DEFAULT NULL COMMENT 'Номер объекта',
  `geo_flatnum_postoffice` varchar(32) CHARACTER SET utf8mb4 DEFAULT NULL COMMENT 'Почтовый номер объекта',
  `estate_area` decimal(16,4) DEFAULT NULL COMMENT 'Площадь объекта',
  `estate_price` decimal(16,4) DEFAULT NULL COMMENT 'Цена объекта',
  `estate_price_action` decimal(16,4) DEFAULT NULL COMMENT 'Цена по спецпредложению',
  `estate_price_m2` decimal(16,4) DEFAULT NULL COMMENT 'Стоимость за м2',
  `estate_areaBti` decimal(16,4) DEFAULT NULL COMMENT 'Площадь по БТИ',
  `estate_areaBti_koef` decimal(16,4) DEFAULT NULL COMMENT 'Площадь по БТИ (коэф.)',
  `estate_restoration_price` decimal(16,4) DEFAULT NULL COMMENT 'Стоимость отделки',
  `estate_area_inside` decimal(16,4) DEFAULT NULL COMMENT 'Площадь без ЛП',
  `estate_areaBti_inside` decimal(16,4) DEFAULT NULL COMMENT 'Площадь БТИ без ЛП',
  `estate_restoration` varchar(255) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Вид отделки',
  `estate_sale_type` varchar(1000) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Тип продажи',
  `estate_dealAreaBeforeBtiRecalc` decimal(16,4) DEFAULT NULL COMMENT 'Площадь до перерасчета по обмерам БТИ',
  `special_notes` varchar(255) CHARACTER SET utf8mb4 NOT NULL DEFAULT '' COMMENT 'Служебные отметки',
  `deal_id` int(10) unsigned DEFAULT '0' COMMENT 'Id сделки',
  `deal_manager_id` int(10) unsigned DEFAULT NULL COMMENT 'Менеджер сделки',
  `estate_buy_id` int(10) unsigned DEFAULT NULL COMMENT 'Id заявки по сделке',
  `agreement_type` varchar(16) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Тип договора',
  `agreement_date` date DEFAULT NULL COMMENT 'Дата договора',
  `deal_sum` decimal(16,2) unsigned DEFAULT NULL COMMENT 'Сумма сделки',
  `deal_price` decimal(16,2) unsigned DEFAULT NULL COMMENT 'Цена объекта на момент начала оформления сделки',
  `deal_area` decimal(10,4) DEFAULT NULL COMMENT 'Площадь в сделке',
  `deal_sum_addons` decimal(16,2) unsigned DEFAULT NULL COMMENT 'Сумма допов в сделке',
  `deal_date` date DEFAULT NULL COMMENT 'Дата проведения сделки',
  `deal_date_start` date DEFAULT NULL COMMENT 'Дата начала оформления сделки',
  PRIMARY KEY (`id`),
  KEY `house_id` (`house_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
DROP TABLE IF EXISTS `estate_sells_attr`;
CREATE TABLE `estate_sells_attr` (
  `id` varchar(19) CHARACTER SET utf8 NOT NULL DEFAULT '' COMMENT 'Id',
  `company_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT 'Company_id',
  `estate_sell_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT 'Id объекта',
  `updated_at` timestamp NULL DEFAULT NULL COMMENT 'Timestamp модификации',
  `attr_table` varchar(8) CHARACTER SET utf8 NOT NULL DEFAULT '' COMMENT 'Тип данных (int|decimal|varchar)',
  `attr_name` varchar(64) CHARACTER SET utf8 NOT NULL DEFAULT '' COMMENT 'Имя атрибута',
  `attr_value` varchar(32) CHARACTER SET utf8mb4 NOT NULL DEFAULT '' COMMENT 'Значение атрибута',
  PRIMARY KEY (`id`),
  KEY `estate_sell_id` (`estate_sell_id`),
  KEY `attr_name` (`attr_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
DROP TABLE IF EXISTS `estate_sells_price_min_stat`;
CREATE TABLE `estate_sells_price_min_stat` (
  `id` int(11) NOT NULL DEFAULT '0' COMMENT 'Id',
  `company_id` int(10) unsigned NOT NULL COMMENT 'Company_id',
  `estate_sell_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT 'Id объекта',
  `calculation_date` date NOT NULL COMMENT 'Дата за которую расчитана минимальная цена',
  `price` decimal(16,2) unsigned DEFAULT NULL COMMENT 'Минимально возможная цена объекта с учетом актуальных акций',
  `area` decimal(10,4) unsigned DEFAULT NULL COMMENT 'Площадь объекта на момент фактировки цены',
  PRIMARY KEY (`id`),
  KEY `estate_sell_id` (`estate_sell_id`),
  KEY `calculation_date` (`calculation_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
DROP TABLE IF EXISTS `estate_sells_price_stat`;
CREATE TABLE `estate_sells_price_stat` (
  `id` int(16) unsigned NOT NULL DEFAULT '0' COMMENT 'Id',
  `company_id` int(10) unsigned NOT NULL COMMENT 'Company_id',
  `estate_sell_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT 'Id объекта',
  `date_stat` date NOT NULL COMMENT 'Дата замера',
  `price` int(11) unsigned NOT NULL COMMENT 'Цена общая',
  `price_m2` int(11) unsigned NOT NULL COMMENT 'Цена за м²',
  PRIMARY KEY (`id`),
  KEY `estate_sell_id` (`estate_sell_id`),
  KEY `date_stat` (`date_stat`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
DROP TABLE IF EXISTS `estate_sells_statuses_log`;
CREATE TABLE `estate_sells_statuses_log` (
  `id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT 'Id',
  `company_id` int(10) unsigned NOT NULL COMMENT 'Company_id',
  `log_date` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00' COMMENT 'Дата события',
  `estate_sell_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT 'Id объекта',
  `deal_id` int(11) unsigned DEFAULT NULL COMMENT 'Id сделки в момент события',
  `deal_sum` decimal(16,2) unsigned DEFAULT NULL COMMENT 'Сумма сделки в момент события',
  `is_payed_reserve` tinyint(1) unsigned NOT NULL DEFAULT '0' COMMENT 'Признак включения платной брони',
  `status_from` tinyint(3) unsigned NOT NULL COMMENT 'Исходный статус',
  `status_from_name` varchar(32) CHARACTER SET utf8mb4 DEFAULT NULL COMMENT 'Название исходного статуса',
  `status_to` tinyint(3) unsigned NOT NULL COMMENT 'Новый статус',
  `status_to_name` varchar(32) CHARACTER SET utf8mb4 DEFAULT NULL COMMENT 'Название нового статуса',
  PRIMARY KEY (`id`),
  KEY `estate_sell_id` (`estate_sell_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
DROP TABLE IF EXISTS `estate_statuses`;
CREATE TABLE `estate_statuses` (
  `status_id` bigint(20) NOT NULL DEFAULT '0',
  `status_name` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  PRIMARY KEY (`status_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
DROP TABLE IF EXISTS `estate_statuses_reasons`;
CREATE TABLE `estate_statuses_reasons` (
  `status_reason_id` int(11) NOT NULL DEFAULT '0' COMMENT 'Тип причины перевода в неактивный статус',
  `company_id` int(11) unsigned NOT NULL COMMENT 'Company_id',
  `updated_at` timestamp NULL DEFAULT NULL COMMENT 'Timestamp модификации',
  `type` enum('giveup','wait','inactive') CHARACTER SET utf8 DEFAULT NULL COMMENT 'Тип причины',
  `name` varchar(64) CHARACTER SET utf8 NOT NULL COMMENT 'Причина',
  `is_archived` tinyint(1) unsigned NOT NULL DEFAULT '0' COMMENT 'Перемещено в архив',
  PRIMARY KEY (`status_reason_id`),
  KEY `type` (`type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
DROP TABLE IF EXISTS `estate_tags`;
CREATE TABLE `estate_tags` (
  `id` varchar(30) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'Id связи',
  `company_id` int(11) unsigned NOT NULL COMMENT 'Company_id',
  `updated_at` timestamp NULL DEFAULT NULL COMMENT 'Timestamp модификации',
  `estate_id` int(11) unsigned NOT NULL COMMENT 'Id объекта/заявки/дома/группы домов',
  `tags_id` int(11) unsigned NOT NULL COMMENT 'Id тега',
  PRIMARY KEY (`id`),
  KEY `tags_id` (`tags_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
DROP TABLE IF EXISTS `estate_transfer`;
CREATE TABLE `estate_transfer` (
  `id` int(11) NOT NULL DEFAULT '0' COMMENT 'Id',
  `company_id` int(10) unsigned NOT NULL COMMENT 'Company_id',
  `estate_sell_id` int(11) unsigned NOT NULL COMMENT 'Id объекта',
  `updated_at` timestamp NULL DEFAULT NULL COMMENT 'Timestamp модификации',
  `transfer_type` varchar(3) CHARACTER SET utf8 NOT NULL COMMENT 'Признак передачи (out), приемки (in)',
  `transfer_status` varchar(16) CHARACTER SET utf8mb4 NOT NULL DEFAULT '' COMMENT 'Статус передачи (finish - "Нет замечаний"|notices - "Есть замечания"|declined - "Клиент уклонился от осмотра"|"" - "Не передано")',
  `house_id` int(11) unsigned NOT NULL COMMENT 'Id дома',
  `plan_date` timestamp NULL DEFAULT NULL COMMENT 'Плановая дата передачи',
  `finish_date` timestamp NULL DEFAULT NULL COMMENT 'Фактическая дата передачи',
  `formal_signed_date` timestamp NULL DEFAULT NULL COMMENT 'Дата формальной передачи',
  `attempts_count` mediumint(8) unsigned NOT NULL COMMENT 'Количество осмотров с покупателем',
  `out_responsible_id` int(11) unsigned DEFAULT NULL COMMENT 'Ответственный за передачу сотрудник',
  PRIMARY KEY (`id`),
  KEY `estate_sell_id` (`estate_sell_id`),
  KEY `house_id` (`house_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
DROP TABLE IF EXISTS `estate_transfer_attempts`;
CREATE TABLE `estate_transfer_attempts` (
  `id` int(11) NOT NULL DEFAULT '0' COMMENT 'Id',
  `company_id` int(10) unsigned NOT NULL COMMENT 'Company_id',
  `transfer_id` int(11) unsigned NOT NULL COMMENT 'Передача ключей',
  `attempt_user_id` int(11) unsigned NOT NULL,
  `date_added` timestamp NULL DEFAULT NULL COMMENT 'Дата осмотра',
  `estate_sell_id` int(11) unsigned NOT NULL COMMENT 'Id объекта',
  `updated_at` timestamp NULL DEFAULT NULL COMMENT 'Timestamp модификации',
  `is_success` int(1) unsigned NOT NULL COMMENT 'Признак удачной передачи',
  PRIMARY KEY (`id`),
  KEY `estate_sell_id` (`estate_sell_id`),
  KEY `transfer_id` (`transfer_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
DROP TABLE IF EXISTS `finances`;
CREATE TABLE `finances` (
  `id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT 'Id',
  `company_id` int(10) unsigned NOT NULL COMMENT 'Company_id',
  `updated_at` timestamp NULL DEFAULT NULL COMMENT 'Timestamp модификации',
  `status` tinyint(3) unsigned NOT NULL COMMENT 'Status',
  `types_id` int(10) unsigned NOT NULL COMMENT ' тип платежа',
  `subtypes_id` int(11) unsigned NOT NULL COMMENT ' подтип платежа',
  `users_id` int(10) unsigned NOT NULL COMMENT ' инициатор платежа',
  `manager_id` int(10) unsigned NOT NULL COMMENT ' менеджер платежа',
  `date_added` datetime DEFAULT NULL COMMENT ' дата добавления операции',
  `date_to` datetime DEFAULT NULL COMMENT ' планируемая дата оплаты',
  `summa` decimal(16,2) unsigned NOT NULL COMMENT ' сумма платежа',
  `estate_sell_id` int(11) unsigned DEFAULT NULL COMMENT ' объект платежа (дом или объект в сделке)',
  `deal_id` int(11) unsigned DEFAULT NULL COMMENT ' id сделки платежа',
  `contacts_id` int(10) unsigned DEFAULT NULL COMMENT ' контрагент платежа',
  `contacts_agreements_id` int(11) unsigned NOT NULL COMMENT ' контракт (документ)',
  `inventory_demands_id` int(11) unsigned DEFAULT NULL COMMENT 'Id заявки',
  `approved_by` int(11) unsigned DEFAULT NULL COMMENT ' согласовавший сотрудник',
  `approved_date` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00' COMMENT ' дата согласования',
  `accepted_for_payment` int(1) unsigned NOT NULL COMMENT ' признак акцептования',
  `accepted_by` int(11) unsigned DEFAULT NULL COMMENT ' акцептовавший сотрудник',
  `accepted_date` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00' COMMENT ' дата акцептования',
  `accepted_summa` decimal(16,2) unsigned DEFAULT NULL COMMENT ' акцептованная сумма',
  `is_burning` int(1) unsigned NOT NULL DEFAULT '0' COMMENT ' признак горящего платежа',
  `is_first_payment` tinyint(1) unsigned NOT NULL COMMENT ' признак первого платежа в графике платежей',
  `is_over_deal_sum` tinyint(1) unsigned NOT NULL COMMENT ' признак платежа в графике платежей сверх суммы сделки',
  `status_name` varchar(16) CHARACTER SET utf8mb4 DEFAULT NULL COMMENT 'Status_name',
  `types_name` varchar(64) CHARACTER SET utf8mb4 NOT NULL DEFAULT '' COMMENT 'Types_name',
  `account_in_id` int(10) unsigned DEFAULT NULL COMMENT ' счет зачисления',
  `account_out_id` int(10) unsigned DEFAULT NULL COMMENT ' счет списания',
  `contact_in_id` int(10) unsigned DEFAULT NULL COMMENT ' контрагент получатель',
  `contact_out_id` int(10) unsigned DEFAULT NULL COMMENT ' контрагент плательщик',
  PRIMARY KEY (`id`),
  KEY `date_added` (`date_added`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
DROP TABLE IF EXISTS `finances_accounts`;
CREATE TABLE `finances_accounts` (
  `id` int(10) unsigned NOT NULL DEFAULT '0' COMMENT 'Id',
  `account_id` int(10) unsigned NOT NULL DEFAULT '0' COMMENT 'Account_id',
  `company_id` int(10) unsigned NOT NULL COMMENT 'Company_id',
  `updated_at` timestamp NULL DEFAULT NULL COMMENT 'Timestamp модификации',
  `organization_id` int(11) unsigned DEFAULT NULL COMMENT ' контакт организации счета',
  `account_name` varchar(255) CHARACTER SET utf8 NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
DROP TABLE IF EXISTS `geo_city_complex`;
CREATE TABLE `geo_city_complex` (
  `id` int(10) unsigned NOT NULL DEFAULT '0' COMMENT 'Id ЖК (справочник)',
  `geo_complex_id` int(10) unsigned NOT NULL DEFAULT '0' COMMENT 'Id ЖК (справочник)',
  `company_id` int(10) unsigned NOT NULL DEFAULT '0' COMMENT 'Company_id',
  `geo_complex_name` varchar(255) CHARACTER SET utf8 NOT NULL COMMENT 'Название ЖК (справочно)',
  `city_name` varchar(128) CHARACTER SET utf8 NOT NULL DEFAULT '' COMMENT 'Город',
  `sort_order` tinyint(4) unsigned NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
DROP TABLE IF EXISTS `inventory_demands`;
CREATE TABLE `inventory_demands` (
  `id` int(11) NOT NULL DEFAULT '0' COMMENT 'Id позиции в заказе',
  `company_id` int(11) unsigned NOT NULL COMMENT 'Company_id',
  `updated_at` timestamp NULL DEFAULT NULL COMMENT 'Timestamp модификации',
  `demand_item_id` int(11) NOT NULL DEFAULT '0' COMMENT 'Id позиции в заказе',
  `demand_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT 'Id заказе',
  `date_added` date DEFAULT NULL COMMENT 'Дата добавления заказа',
  `date_status_changed` date DEFAULT NULL COMMENT 'Дата последнего изменения статуса заказа',
  `days_status_changed` int(7) DEFAULT NULL COMMENT 'Дней с момента последнего изменения статуса заказа',
  `status` int(1) unsigned NOT NULL COMMENT 'Статус заказа',
  `status_name` varchar(32) CHARACTER SET utf8mb4 DEFAULT NULL COMMENT 'Имя статуса заказа',
  `projects_id` int(11) unsigned NOT NULL COMMENT 'Id строительного проекта (ГПР)',
  `projects_tasks_id` int(11) unsigned DEFAULT NULL COMMENT 'Id работы в ГПР',
  `parent_id` int(11) unsigned DEFAULT NULL COMMENT 'Id родительского заказа (в случае разделения заказа)',
  `contacts_id` int(11) unsigned DEFAULT NULL COMMENT 'Id поставщика (контакт)',
  `delivery_type` varchar(16) CHARACTER SET utf8 NOT NULL COMMENT 'Тип поставки',
  `demander_id` int(11) unsigned NOT NULL COMMENT 'Id инициатора заказа (пользователь)',
  `supplier_id` int(11) unsigned NOT NULL COMMENT 'Id снабженца (пользователь)',
  `supplier_contact_id` int(11) unsigned DEFAULT '0' COMMENT 'Id поставщика',
  `demander_user_name` varchar(255) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Имя инициатора',
  `supplier_user_name` varchar(255) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Имя снабженца',
  `supplier_contact_name` varchar(255) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Имя поставщика',
  `warehouse_id` int(11) DEFAULT NULL COMMENT 'Id склада на который везут ТМЦ',
  `item_date_demand` date DEFAULT NULL COMMENT 'Запрошенная дата поставки позиции',
  `item_date_fact` date DEFAULT NULL COMMENT 'Фактическая дата поставки позиции',
  `noms_id` int(11) NOT NULL COMMENT 'Id номенклатуры из справочника',
  `item_measure` varchar(32) CHARACTER SET utf8 NOT NULL COMMENT 'Ед.измерения в позиции заказа',
  `item_price` decimal(21,9) unsigned NOT NULL COMMENT 'Цена ТМЦ в позиции заказа',
  `item_quantity` decimal(21,9) NOT NULL COMMENT 'Количество ТМЦ в позиции заказа',
  `item_summa` decimal(21,9) unsigned NOT NULL COMMENT 'Стоимость позиции заказа',
  `item_quantity_income` decimal(43,9) NOT NULL DEFAULT '0.000000000',
  `item_price_income` decimal(13,0) DEFAULT NULL,
  `item_quantity_part` decimal(43,9) NOT NULL DEFAULT '0.000000000',
  `item_summa_part` decimal(43,9) NOT NULL DEFAULT '0.000000000',
  `item_quantity_outcome` decimal(43,9) NOT NULL DEFAULT '0.000000000',
  `item_max_demand_days` int(7) DEFAULT NULL,
  `item_overdue_days` int(7) DEFAULT NULL,
  `item_overdue_interval` varchar(10) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `demand_item_payed_summa` decimal(42,6) NOT NULL DEFAULT '0.000000',
  `is_expired_approvement` int(1) NOT NULL DEFAULT '0',
  `is_burning` tinyint(1) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  KEY `date_added` (`date_added`),
  KEY `noms_id` (`noms_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

DROP TABLE IF EXISTS `inventory_noms_top`;
CREATE TABLE `inventory_noms_top` (
  `id` int(11) NOT NULL COMMENT 'Id',
  `company_id` int(11) unsigned NOT NULL COMMENT 'Company_id',
  `noms_id` int(11) NOT NULL COMMENT 'Noms_id',
  `noms_name` varchar(1000) CHARACTER SET utf8 NOT NULL COMMENT 'Имя часто заказываемой номенклатуры',
  `demands_count` bigint(21) NOT NULL DEFAULT '0' COMMENT 'Количество заказов данной номенклатуры',
  `item_avg_price` decimal(25,13) DEFAULT NULL COMMENT 'Средняя цена заказа данной номенклатуры',
  PRIMARY KEY (`id`),
  KEY `noms_id` (`noms_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
DROP TABLE IF EXISTS `inventory_warehouse`;
CREATE TABLE `inventory_warehouse` (
  `id` int(11) NOT NULL DEFAULT '0' COMMENT 'Id',
  `company_id` int(11) unsigned NOT NULL COMMENT 'Company_id',
  `updated_at` timestamp NULL DEFAULT NULL COMMENT 'Timestamp модификации',
  `warehouse_id` int(11) NOT NULL DEFAULT '0' COMMENT 'Warehouse_id',
  `warehouse_name` varchar(32) CHARACTER SET utf8 NOT NULL COMMENT 'Warehouse_name',
  `projects_id` int(11) unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `projects_id` (`projects_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
DROP TABLE IF EXISTS `inventory_warehouse_stocks`;
CREATE TABLE `inventory_warehouse_stocks` (
  `id` int(11) NOT NULL DEFAULT '0' COMMENT 'Позиция заказа, сформировавшего остаток на складе',
  `company_id` int(11) unsigned NOT NULL COMMENT 'Company_id',
  `demand_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT 'Заказ, сформировавший остаток на складе',
  `demand_date_added` date DEFAULT NULL COMMENT 'Дата появления заказа',
  `warehouse_id` int(11) NOT NULL DEFAULT '0' COMMENT 'Id склада',
  `projects_id` int(11) unsigned DEFAULT NULL COMMENT 'Id проекта к которому относится склад',
  `summa_left` decimal(64,18) DEFAULT NULL COMMENT 'Стоимость остатка номенклатуры',
  `task_date_finish_fact` date DEFAULT NULL COMMENT 'Фактическая дата закрытия заказа',
  `task_id` int(11) unsigned DEFAULT '0' COMMENT 'Id работы из ГПР',
  `noms_id` int(11) NOT NULL COMMENT 'Id номенклатуры в остатках',
  `date_received` date DEFAULT NULL COMMENT 'Дата передачи номенклатуры на склад',
  `stocks_days_interval` varchar(11) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'Интервал дней нахождения номенклатуры на складе',
  PRIMARY KEY (`id`),
  KEY `warehouse_id` (`warehouse_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
DROP TABLE IF EXISTS `noms`;
CREATE TABLE `noms` (
  `id` int(11) NOT NULL DEFAULT '0' COMMENT 'Id',
  `company_id` int(11) unsigned NOT NULL COMMENT 'Company_id',
  `updated_at` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00' COMMENT 'Timestamp модификации',
  `noms_name` varchar(1000) CHARACTER SET utf8 NOT NULL COMMENT 'Noms_name',
  `noms_parent_id` int(11) unsigned NOT NULL COMMENT 'Noms_parent_id',
  `type` enum('inventory','service','work','machine','equipment','hold') CHARACTER SET utf8 NOT NULL DEFAULT 'inventory' COMMENT 'Type',
  `code` varchar(32) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Code',
  `measure` varchar(32) CHARACTER SET utf8 NOT NULL COMMENT 'Measure',
  `category_full_name` char(0) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  KEY `noms_parent_id` (`noms_parent_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
DROP TABLE IF EXISTS `noms_category`;
CREATE TABLE `noms_category` (
  `id` int(11) NOT NULL DEFAULT '0' COMMENT 'Id',
  `company_id` int(11) unsigned NOT NULL COMMENT 'Company_id',
  `updated_at` timestamp NULL DEFAULT NULL COMMENT 'Timestamp модификации',
  `category_name` varchar(255) CHARACTER SET utf8 NOT NULL COMMENT 'Category_name',
  `category_parent_id` int(11) DEFAULT NULL COMMENT 'Category_parent_id',
  `category_type` varchar(16) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Category_type',
  `code` varchar(32) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Code',
  `category_full_name` varchar(341) CHARACTER SET utf8 DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `category_parent_id` (`category_parent_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
DROP TABLE IF EXISTS `projects`;
CREATE TABLE `projects` (
  `id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT 'Id',
  `company_id` int(11) unsigned NOT NULL COMMENT 'Company_id',
  `updated_at` timestamp NULL DEFAULT NULL COMMENT 'Timestamp модификации',
  `group_id` int(11) unsigned DEFAULT NULL COMMENT 'Group_id',
  `projects_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT 'Id проекта',
  `projects_name` varchar(255) CHARACTER SET utf8 NOT NULL COMMENT 'Наименование проекта',
  `projects_name_full` varchar(290) CHARACTER SET utf8 NOT NULL DEFAULT '' COMMENT 'Полное наименование проекта',
  `projects_sort_order` int(3) unsigned NOT NULL COMMENT 'Порядок проекта в группе проектов',
  `project_date_finish_plan` date DEFAULT NULL COMMENT 'Плановая дата завершения проекта',
  `project_date_finish` date DEFAULT NULL COMMENT 'Фактическая дата завершения проекта',
  `project_date_start` date DEFAULT NULL COMMENT 'Фактическая дата начала проекта',
  `completeness` decimal(5,2) unsigned NOT NULL COMMENT '% завершения проекта',
  `duration_days` int(7) DEFAULT NULL COMMENT 'Продолжительность проекта',
  `duration_gone_days` int(7) DEFAULT NULL COMMENT 'Текущая длительность проекта',
  `duration_left_days` int(7) DEFAULT NULL COMMENT 'Дней до окончания проекта по плану',
  `duration_overdue_days` int(7) DEFAULT NULL COMMENT 'Дней просрочки окончания проекта',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
DROP TABLE IF EXISTS `projects_tasks`;
CREATE TABLE `projects_tasks` (
  `id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT 'Id',
  `task_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT 'Id работы/группы работ',
  `company_id` int(11) unsigned NOT NULL COMMENT 'Company_id',
  `updated_at` timestamp NULL DEFAULT NULL COMMENT 'Timestamp модификации',
  `projects_id` int(11) unsigned NOT NULL COMMENT 'Id проекта',
  `is_group` tinyint(1) unsigned NOT NULL COMMENT 'Признак группы работ',
  `date_start` date DEFAULT NULL COMMENT 'Плановая дата начала работы',
  `date_finish` date DEFAULT NULL COMMENT 'Date_finish',
  `date_start_fact` date DEFAULT NULL COMMENT 'Фактическая дата начала работы',
  `date_finish_fact` date DEFAULT NULL COMMENT 'Фактическая дата окончания работы',
  `task_start_status` varchar(9) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'Статус начала работы',
  `task_finish_status` varchar(9) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'Статус окончания работы',
  `task_status_extended` varchar(24) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'Расширенный статус работы',
  `failed_start_days` int(7) DEFAULT NULL COMMENT 'Дней просрочки начала',
  `failed_finish_days` int(7) DEFAULT NULL COMMENT 'Дней просрочки окончания',
  `failed_start_interval` varchar(10) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'Интервал просрочки начала',
  `failed_finish_interval` varchar(10) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'Интервал просрочки окончания',
  `finish_delay_interval` varchar(10) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'Интервалы фактического окончания работ относительно плановых',
  `is_finish_delay` int(1) NOT NULL DEFAULT '0' COMMENT 'Признак просрочки окончания для завершенной работы',
  `task_name` varchar(255) CHARACTER SET utf8 NOT NULL COMMENT 'Название работы/группы работ',
  `prefix` varchar(32) CHARACTER SET utf8 NOT NULL COMMENT 'Префикс',
  `subname` varchar(255) CHARACTER SET utf8 NOT NULL COMMENT 'Комментарий',
  `sort_order` int(5) unsigned NOT NULL COMMENT 'Порядок сортировки работы внутри группы',
  `progress` int(3) unsigned NOT NULL COMMENT 'Прогресс выполнения работы',
  `level` int(11) unsigned NOT NULL DEFAULT '1' COMMENT 'Уровень вложенности',
  `left_key` int(11) NOT NULL COMMENT 'Left_key',
  `right_key` int(11) unsigned NOT NULL COMMENT 'Right_key',
  `users_inspected` int(11) unsigned NOT NULL COMMENT 'Пользователь, проверивший работы',
  `date_inspected` date DEFAULT NULL COMMENT 'Дата проверки работы',
  `group_name` varchar(255) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Имя непосредственно вышележащей группы работ',
  `full_group_name` varchar(511) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Путь до работы',
  `task_full_group_name` varchar(255) CHARACTER SET utf8mb4 DEFAULT NULL COMMENT 'Путь до работы включая название работы',
  `date_start_requests_count` bigint(21) NOT NULL DEFAULT '0',
  `date_finish_requests_count` bigint(21) NOT NULL DEFAULT '0',
  `task_quality_accepted_date` date DEFAULT NULL,
  `task_quality_accepted_user` varchar(255) CHARACTER SET utf8 DEFAULT NULL,
  `task_finished_action_date` date DEFAULT NULL,
  `task_finished_action_user` varchar(255) CHARACTER SET utf8 DEFAULT NULL,
  `is_task_finish_back_action` int(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `projects_id` (`projects_id`),
  KEY `is_group` (`is_group`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
DROP TABLE IF EXISTS `projects_tasks_checklists`;
CREATE TABLE `projects_tasks_checklists` (
  `id` int(11) NOT NULL DEFAULT '0' COMMENT 'Id',
  `company_id` int(11) unsigned NOT NULL COMMENT 'Company_id',
  `updated_at` timestamp NULL DEFAULT NULL COMMENT 'Timestamp модификации',
  `created_at` date DEFAULT NULL COMMENT 'Дата добавления элемента',
  `projects_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT 'Проект',
  `task_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT 'Работа',
  `type` enum('notices','defects','checklist') CHARACTER SET utf8 NOT NULL DEFAULT 'checklist' COMMENT 'Тип элемента (notices - Предписания, defects - Дефектовка)',
  `type_name` varchar(11) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'Название типа элемента',
  `type_status_name` varchar(22) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'Текущий статус записи',
  `user_initiator_name` varchar(255) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Инициатор добавления записи',
  `user_checked_name` varchar(255) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Инициатор закрытия записи',
  `checked_date_to` date DEFAULT NULL COMMENT 'Дата действия до',
  `checked_date` date DEFAULT NULL COMMENT 'Дата закрытия записи',
  `is_notices` int(1) NOT NULL DEFAULT '0' COMMENT 'Запись - предписание',
  `is_defects` int(1) NOT NULL DEFAULT '0' COMMENT 'Запись - дефектовка',
  `name` varchar(255) CHARACTER SET utf8 NOT NULL COMMENT 'Наименование записи',
  PRIMARY KEY (`id`),
  KEY `projects_id` (`projects_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
DROP TABLE IF EXISTS `projects_tasks_requests`;
CREATE TABLE `projects_tasks_requests` (
  `id` int(11) NOT NULL DEFAULT '0' COMMENT 'Id',
  `company_id` int(11) unsigned NOT NULL COMMENT 'Company_id',
  `updated_at` timestamp NULL DEFAULT NULL COMMENT 'Timestamp модификации',
  `projects_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT 'Id проекта',
  `status` tinyint(3) unsigned NOT NULL COMMENT 'Статус запроса',
  `status_name` varchar(32) CHARACTER SET utf8mb4 DEFAULT NULL COMMENT 'Наименование статуса запроса',
  `reason` varchar(255) CHARACTER SET utf8 NOT NULL COMMENT 'Причина отклонения запроса',
  `date_start` date NOT NULL COMMENT 'Плановая дата начала работ на момент запроса',
  `date_finish` date NOT NULL COMMENT 'Плановая дата окончания работ на момент запроса',
  `date_start_fact` date DEFAULT NULL COMMENT 'Фактическая дата начала работ на момент запроса',
  `date_finish_fact` date DEFAULT NULL COMMENT 'Фактическая дата окончания работ на момент запроса',
  `request_date_start` date NOT NULL COMMENT 'Новая запрошенная плановая дата начала',
  `request_date_finish` date NOT NULL COMMENT 'Новая запрошенная плановая дата окончания',
  `date_approved` date DEFAULT NULL COMMENT 'Дата одобрения запроса',
  `user_approved` varchar(255) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Пользователь, одобривший запрос',
  `user_requested` varchar(255) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Пользователь, открывший запрос',
  `request_date_created` date DEFAULT NULL COMMENT 'Дата создания запроса',
  `request_days` int(7) DEFAULT NULL COMMENT 'Длительность текущего открытого запроса',
  `open_request_days_interval` varchar(5) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'Интервал продолжительности текущего открытого запроса',
  `closed_request_days_interval` varchar(7) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'Интервал продолжительности одобрения запроса',
  `task_name` varchar(255) CHARACTER SET utf8 NOT NULL COMMENT 'Наименование работы',
  `task_group_name` varchar(255) CHARACTER SET utf8 NOT NULL COMMENT 'Наименование группы работ',
  PRIMARY KEY (`id`),
  KEY `request_date_created` (`request_date_created`),
  KEY `projects_id` (`projects_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
DROP TABLE IF EXISTS `promos`;
CREATE TABLE `promos` (
  `id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT 'Id',
  `promo_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT ' id акции',
  `company_id` int(10) unsigned NOT NULL COMMENT 'Company_id',
  `updated_at` timestamp NULL DEFAULT NULL COMMENT 'Timestamp модификации',
  `promo_name` varchar(64) CHARACTER SET utf8 NOT NULL COMMENT ' имя пользователя',
  `promo_discount` decimal(16,2) NOT NULL COMMENT ' величина скидки',
  `promo_rule` varchar(16) CHARACTER SET utf8 NOT NULL COMMENT ' правило изменения цены (без изменения, цена, цена за м²)',
  `promo_type` varchar(16) CHARACTER SET utf8 NOT NULL COMMENT ' скидка в валюте или процентах',
  `promo_date_from` date DEFAULT NULL,
  `promo_date_to` date DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
DROP TABLE IF EXISTS `stat`;
CREATE TABLE `stat` (
  `param_name` varchar(17) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `param_value` varchar(19) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT ''
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
DROP TABLE IF EXISTS `tags`;
CREATE TABLE `tags` (
  `id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT 'Id',
  `company_id` int(11) unsigned NOT NULL COMMENT 'Company_id',
  `updated_at` timestamp NULL DEFAULT NULL COMMENT 'Timestamp модификации',
  `tags_name` varchar(64) CHARACTER SET utf8 NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
DROP TABLE IF EXISTS `tasks`;
CREATE TABLE `tasks` (
  `id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT 'Id',
  `company_id` int(10) unsigned NOT NULL COMMENT 'Company_id',
  `updated_at` timestamp NULL DEFAULT NULL COMMENT 'Timestamp модификации',
  `category_id` int(11) unsigned DEFAULT NULL COMMENT 'Category_id',
  `date_modified` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00' COMMENT 'Дата изменения',
  `estate_id` int(11) unsigned DEFAULT NULL COMMENT 'Id объекта (дом или квартира) либо id заявки',
  `contacts_id` int(11) unsigned DEFAULT NULL COMMENT 'Id контакта',
  `category_name` varchar(64) CHARACTER SET utf8 DEFAULT NULL COMMENT 'Имя категории',
  `status` int(6) unsigned NOT NULL COMMENT 'Id статуса',
  `is_closed` int(1) unsigned NOT NULL COMMENT ' задача закрыта',
  `priority` int(1) NOT NULL COMMENT ' приоритет задачи',
  `progress` tinyint(3) unsigned NOT NULL COMMENT ' прогресс выполнения',
  `date_added` date DEFAULT NULL COMMENT ' дата добавления задачи',
  `date_finish` date DEFAULT NULL COMMENT ' плановая дата завершения',
  `date_finish_time` time DEFAULT NULL COMMENT ' плановое время завершения',
  `date_finish_fact` date DEFAULT NULL COMMENT ' фактическая дата завершения',
  `date_finish_fact_time` time DEFAULT NULL COMMENT ' фактическое время завершения',
  `date_combined` date DEFAULT NULL COMMENT 'Date_combined',
  `hours_plan` decimal(10,2) NOT NULL COMMENT ' часов запланировано',
  `hours_fact` decimal(10,2) NOT NULL COMMENT ' часов затрачено',
  `type` varchar(64) CHARACTER SET utf8 NOT NULL COMMENT ' тип задачи',
  `type_name` varchar(32) CHARACTER SET utf8mb4 DEFAULT NULL COMMENT 'Type_name',
  `status_name` varchar(16) CHARACTER SET utf8mb4 DEFAULT NULL COMMENT 'Status_name',
  `assigner_name` varchar(255) CHARACTER SET utf8 DEFAULT NULL COMMENT ' постановщик',
  `manager_name` varchar(255) CHARACTER SET utf8 DEFAULT NULL COMMENT ' исполнитель',
  PRIMARY KEY (`id`),
  KEY `category_id` (`category_id`),
  KEY `status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
DROP TABLE IF EXISTS `tasks_tags`;
CREATE TABLE `tasks_tags` (
  `id` varchar(29) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'Id связи',
  `company_id` int(11) unsigned NOT NULL COMMENT 'Company_id',
  `updated_at` timestamp NULL DEFAULT NULL COMMENT 'Timestamp модификации',
  `tasks_id` int(11) unsigned NOT NULL COMMENT 'Id задачи',
  `tags_id` int(11) unsigned NOT NULL COMMENT 'Id тега',
  PRIMARY KEY (`id`),
  KEY `tags_id` (`tags_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
DROP TABLE IF EXISTS `users`;
CREATE TABLE `users` (
  `id` int(10) unsigned NOT NULL DEFAULT '0' COMMENT 'Id',
  `company_id` int(10) unsigned DEFAULT NULL COMMENT 'Company_id',
  `updated_at` timestamp NULL DEFAULT NULL COMMENT 'Timestamp модификации',
  `users_name` varchar(255) CHARACTER SET utf8 NOT NULL COMMENT 'Имя пользователя',
  `departments_id` int(11) unsigned DEFAULT NULL COMMENT 'Id отдела пользователя',
  `post_title` varchar(255) CHARACTER SET utf8 NOT NULL COMMENT 'Должность',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
