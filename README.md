# Задание 1. Анализ и планирование

Что есть:
* «Тёплый дом» — это небольшая компания, которая организует удалённое управление отоплением в доме. 
* Приложение компании позволяет только управлять отоплением в доме и проверять температуру.
* Для подклчения устройства требуется обязательный выезд специалиста компании на дом.
* Нынешнее приложение представляет собой монолит на Go с СУБД Postgres.
* Всё управление идёт от сервера к датчику, включая данные о температуре от датчика.

Что нужно:
* Мигрировать монолит таким образом, чтобы обеспечить возможность удобного управления и масштабирования в связи с резким увеличением числа пользователей.
* Реализовать работу приложения по модели SaaS. Пользователь должен подключать устрофства сам.
* Приложение должно позволять полбзователю управлять отоплением, включать и выключать свет, запирать и отпирать автоматические ворота, удалённо наблюдать за домом, возможно и будущее неуточнённое поведение.
* Компания обеспечивает инфраструктуру а устрйства изготавливают сторонние компании 
* Покупатели должны иметь возможность программировать систему для управления различными модулями в соответствии со своими потребностями.

### 1. Описание функциональности монолитного приложения

**Управление отоплением:**

- Пользователи могут удалённо включать и выключать обогревательные приборы.

**Мониторинг температуры:**

- Пользователи могут просматривать текущие значения датчиков температуры.

### 2. Анализ архитектуры монолитного приложения

* Язык программирования - Go
* База данных - PostgreSQL
* Взаимодействие между монолитом и БД синхронное
* Пользователь взаимодействует с монолитом через тонкий клиент по HTTP/REST

### 3. Определение доменов и границы контекстов (AS-IS)

Домен - Организация удалённого управления отоплением в доме

+ Поддомен - Управление отоплением (Core)
    * Контекст - Мониторинг и управление обогревательными приборами
      Единый язык:
        - Пользователь
        - Показание_датчика_температуры
        - Статус_работы_нагревательного_прибора
        - Команда_вкл_выкл_нагревательного_прибора

+ Поддомен - Подключение устройств (Supporting)
    * Контекст - Монтаж устройства
      Единый язык:
        - Специалист
        - Адрес_монтажа
        - Устройство

    * Контекст - Регистрация устройства
      Единый язык:
        - Устройство
        - Реестр_устройств

### **4. Проблемы монолитного решения**

- Сложность масштабирования. Невозможно эффективно масштабировать часть монолита. Приходится разворачивать его целиком.
**Это повышает стоимость инфраструктуры и, как следствие, конечную стоимость для пользователя** 
- Высокая связность частей ПО. **Добавление нового типа датчика/устройства увеличивает вероятность негативного влияния на текщий код и повышает вероятность ошибок.**
- Дительные циклы разработки и развертывания. **Добавление нового типа датчика/устройства потребует знаний всего монолита и потребует более тщательного тестирования.**
- Неэффективное управление командой. **Параллельная работа нескольких разработчиков приводит к измененю общего кода и много времени уходит на согласование и объединение изменений**

### 5. Визуализация контекста системы — диаграмма C4

[Диаграмма котекста C4 AS-IS](https://github.com/vmftov/architecture-pro-warmhouse/blob/warmhouse/c4/as-is-context.png)

# Задание 2. Проектирование микросервисной архитектуры

Все диаграммы лежат в папке c4.

**Диаграмма контекста (Context)**

[Диаграмма котекста C4](https://github.com/vmftov/architecture-pro-warmhouse/blob/warmhouse/c4/to-be-context.png)

**Диаграмма контейнеров (Containers)**

[Диаграмма контейнеров C4](https://github.com/vmftov/architecture-pro-warmhouse/blob/warmhouse/c4/to-be-container.png)

**Диаграмма компонентов (Components)**

[Диаграмма компонентов C4 сервиса автоматизации устройств](https://github.com/vmftov/architecture-pro-warmhouse/blob/warmhouse/c4/to-be-component-device_automation_service.png)

**Диаграмма кода (Code)**

[Диаграмма классов движка автоматизации устройств](https://github.com/vmftov/architecture-pro-warmhouse/blob/warmhouse/c4/to-be-code-automation_engine.png)

# Задание 3. Разработка ER-диаграммы

[Логическая обобщенная схема БД](https://github.com/vmftov/architecture-pro-warmhouse/blob/warmhouse/c4/to-be-er.png)

# Задание 4. Создание и документирование API

### 1. Тип API

* Для запросов на получение данных в web и мобильном приложении (например, списка датчиков) используется синхронный REST API. Это позволяет получать необходимые данные в момент запроса.
* Для обновления часто меняющихся данных в web и мобильном приложении (например, телеметрии) используется асинхронный API через Kafka. Общение с клиентом осуществляется посредством WebSocket. Это позволяет разгрузить сервер от polling запросов и оптимизировать выдачу.
* Для изменения данных, требующих немедленного подтверждения (изменение описания датчика) используется синхронный REST API. Такие команды обычно выполняются быстро, а пользователь сразу видит результат что повышает usability и упрощает кодирование.
* Для запросов на получение данных между микросервисами используется синхронный REST API. Это позволяет получать необходимые данные в момент запроса и упростить кодирование.
* Для изменения данных при взаимодействии между микросервисами используется асинхронный API через Kafka. Это позволяет упростить масштабирование сервисов. Такж, это позволяет уменьшить число кросс-сервисных связей и упростить мониторинг и отладку. Использование Kafka с персистентным хранилищем позволяет повысить надежность доставки сообщений.
* При генерации отчёта взаимодействие комбинированное. Запрос на создание отчёта синхронный, но быстрый. Возвращает идентификатор отчёта. Событие о готовности отчёта поступает через kafka. По идентификатору сервис узнает его ли это отчёт. Затем отчёт считывается через синхронный REST API.

### 2. Документация API

#### Пример синхронного API сервиса управления устройствами (сокращённый)

```swagger/openapi

openapi: 3.0.3

info:
  title: WarmHouse - API сервиса управления устройствами
  version: 1.0.0

tags:
  - name: PhysicalDevices
  - name: LogicalSensors
  - name: LogicalRelays

paths:
  /devices:
    post:
      tags: [PhysicalDevices]
      summary: Создать устройство
      requestBody:
        required: true
        content:
          application/json:
            schema: { $ref: '#/components/schemas/CreateDeviceRequest' }
      responses:
        '201':
          description: Created
          content:
            application/json:
              schema: { $ref: '#/components/schemas/PhysicalDevice' }
        '400': { $ref: '#/components/responses/BadRequest' }
        '500': { $ref: '#/components/responses/ServerError' }

  /devices/{physicalDeviceId}:
    get:
      tags: [PhysicalDevices]
      parameters: [ { $ref: '#/components/parameters/PhysicalDeviceId' } ]
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema: { $ref: '#/components/schemas/PhysicalDeviceDetailed' }
        '404': { $ref: '#/components/responses/NotFound' }
        '500': { $ref: '#/components/responses/ServerError' }

    patch:
      tags: [PhysicalDevices]
      summary: Обновить свойства устройства
      parameters: [ { $ref: '#/components/parameters/PhysicalDeviceId' } ]
      requestBody:
        required: true
        content:
          application/json:
            schema: { $ref: '#/components/schemas/UpdateDeviceRequest' }
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema: { $ref: '#/components/schemas/PhysicalDevice' }
        '400': { $ref: '#/components/responses/BadRequest' }
        '404': { $ref: '#/components/responses/NotFound' }
        '500': { $ref: '#/components/responses/ServerError' }

    delete:
      tags: [PhysicalDevices]
      summary: Удалить устройство
      parameters: [ { $ref: '#/components/parameters/PhysicalDeviceId' } ]
      responses:
        '204': { description: Удалено }
        '500': { $ref: '#/components/responses/ServerError' }

  /devices/{physicalDeviceId}/sensors:
    get:
      tags: [LogicalSensors]
      parameters: [ { $ref: '#/components/parameters/PhysicalDeviceId' } ]
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: array
                items: { $ref: '#/components/schemas/LogicalSensor' }
        '404': { $ref: '#/components/responses/NotFound' }
        '500': { $ref: '#/components/responses/ServerError' }

  /devices/{physicalDeviceId}/relays:
    get:
      tags: [LogicalRelays]
      parameters: [ { $ref: '#/components/parameters/PhysicalDeviceId' } ]
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: array
                items: { $ref: '#/components/schemas/LogicalRelay' }
        '404': { $ref: '#/components/responses/NotFound' }
        '500': { $ref: '#/components/responses/ServerError' }

components:
  parameters:
    PhysicalDeviceId:
      name: physicalDeviceId
      in: path
      required: true
      schema: { type: string, format: uuid }

  responses:
    BadRequest:
      description: Данные запроса в непавильном формате или содержат ошибки
      content:
        application/json:
          schema: { $ref: '#/components/schemas/Error' }
    NotFound:
      description: Ресурс не найден
      content:
        application/json:
          schema: { $ref: '#/components/schemas/Error' }
    ServerError:
      description: Внутренняя ошибка сервера
      content:
        application/json:
          schema: { $ref: '#/components/schemas/Error' }

  schemas:
    CreateDeviceRequest:
      type: object
      required: [physicalDeviceId, deviceMetadataId, serialNumber, physicalId]
      properties:
        physicalDeviceId: { type: string, format: uuid }
        deviceMetadataId: { type: string, format: uuid }
        serialNumber: { type: string }
        physicalId: { type: string }
        parentGatewayDeviceId: { type: string, format: uuid, nullable: true }
        displayName: { type: string }
        areaId: { type: string, format: uuid, nullable: true }

    UpdateDeviceRequest:
      type: object
      properties:
        displayName: { type: string, maxLength: 128 }
        areaId: { type: string, format: uuid, nullable: true }

    PhysicalDevice:
      type: object
      required: [physicalDeviceId, deviceMetadataId, serialNumber, displayName, status]
      properties:
        physicalDeviceId: { type: string, format: uuid }
        deviceMetadataId: { type: string, format: uuid }
        serialNumber: { type: string }
        physicalId: { type: string }
        parentGatewayDeviceId: { type: string, format: uuid, nullable: true }
        displayName: { type: string }
        status: { type: string }
        areaId: { type: string, format: uuid, nullable: true }

    PhysicalDeviceDetailed:
      allOf:
        - $ref: '#/components/schemas/PhysicalDevice'
        - type: object
          properties:
            metadata: { $ref: '#/components/schemas/DeviceMetadata' }
            sensors:
              type: array
              items: { $ref: '#/components/schemas/LogicalSensor' }
            relays:
              type: array
              items: { $ref: '#/components/schemas/LogicalRelay' }

    DeviceMetadata:
      type: object
      properties:
        deviceMetadataId: { type: string, format: uuid }
        deviceType: { type: string, example: "thermostat" }
        modelNumber: { type: string }
        protocol: { type: string, example: "zigbee" }
        displayName: { type: string }

    LogicalSensor:
      type: object
      properties:
        logicalSensorId: { type: string, format: uuid }
        physicalDeviceId: { type: string, format: uuid }
        logicalSensorMetadataId: { type: string, format: uuid }
        displayName: { type: string }
        endpointName: { type: string }
        telemetryUnit: { type: string }
        currentValue: { type: number, format: float, nullable: true }

    LogicalRelay:
      type: object
      properties:
        logicalRelayId: { type: string, format: uuid }
        physicalDeviceId: { type: string, format: uuid }
        logicalRelayMetadataId: { type: string, format: uuid }
        displayName: { type: string }
        endpointName: { type: string }
        minValue: { type: number, format: float, nullable: true }
        maxValue: { type: number, format: float, nullable: true }
        allowedValues:
          type: array
          items: { type: number, format: float }
        currentValue: { type: number, format: float, nullable: true }

    Error:
      type: object
      required: [code, message]
      properties:
        code: { type: string }
        message: { type: string }
        traceId: { type: string }

```

#### Пример асинхронного API сервиса телеметрии (сокращённый)

```asyncapi

asyncapi: '2.6.0'

info:
  title: WarmHouse - Асинхронное API сервиса телеметрии
  version: '1.0.0'

servers:
  kafka:
    url: kafka.warmhouse.internal:9092
    protocol: kafka

defaultContentType: application/json

channels:
  smarthome.sensor.telemetry:
    description: Поток телеметрии от сервиса соединения с устройствами
    subscribe:
      operationId: consumeSensorTelemetry
      message:
        $ref: '#/components/messages/SensorTelemetryMessage'

  smarthome.relay.value-changed:
    description: События изменения состояния реле.
    subscribe:
      operationId: consumeRelayValueChanged
      message:
        $ref: '#/components/messages/RelayValueChangedMessage'

components:
  messages:
    SensorTelemetryMessage:
      name: SensorTelemetryMessage
      contentType: application/json
      payload: { $ref: '#/components/schemas/SensorTelemetryPayload' }
    RelayValueChangedMessage:
      name: RelayValueChangedMessage
      contentType: application/json
      payload: { $ref: '#/components/schemas/RelayValueChangedPayload' }

  schemas:
    EventEnvelope:
      type: object
      required: [eventId, eventType, occurredAt, producer]
      properties:
        eventId: { type: string, format: uuid }
        eventType: { type: string }
        producerId: { type: string }

    SensorTelemetryPayload:
      allOf:
        - $ref: '#/components/schemas/EventEnvelope'
        - type: object
          required: [physicalDeviceId, logicalSensorId, value, timestamp]
          properties:
            physicalDeviceId: { type: string, format: uuid }
            logicalSensorId: { type: string, format: uuid }
            value: { type: number, format: float }
            timestamp: { type: string, format: date-time }

    RelayValueChangedPayload:
      allOf:
        - $ref: '#/components/schemas/EventEnvelope'
        - type: object
          required: [physicalDeviceId, logicalRelayId, value, timestamp]
          properties:
            physicalDeviceId: { type: string, format: uuid }
            logicalRelayId: { type: string, format: uuid }
            value: { type: number, format: float }
            timestamp: { type: string, format: date-time }

```

# Задание 5. Работа с docker и docker-compose

**У вас есть расхождени между сайтом и этим readme, где написано, что порт должен быть 8081 и readme в app с 8080 портом. Используется 8080.**

Сервис написан на python и располагается в папке

[apps/task_5/temperature-api/docker-compose.yaml](https://github.com/vmftov/architecture-pro-warmhouse/blob/warmhouse/apps/task_5/temperature-api/docker-compose.yaml)

Приложение доступно по адресу 

[http://localhost:8080](http://localhost:8080)

# **Задание 6. Разработка MVP**

Приложение располагается в папке

[apps/task_6/mvp/docker-compose.yaml](https://github.com/vmftov/architecture-pro-warmhouse/blob/warmhouse/apps/task_6/mvp/docker-compose.yaml)

Схема взаимодействия контейнеров

[Задание 6 - Взаимодействие сервисов](https://github.com/vmftov/architecture-pro-warmhouse/blob/warmhouse/apps/task_6/mvp/containers.png)

DISCLAIMER: Я профессиональный .NET разработчик и не знаю go. Поэтому монолит на go мне честно переделал ИИ. Остальные .NET, python сервисы и docker файлы мои.

Список публичных сервисов:
* Kafka UI          [http://localhost:8090/](http://localhost:8090/)
* Монолит на go     [http://localhost:8080/](http://localhost:8080/)
* Сервис устрйоств  [http://localhost:5002](http://localhost:5002)
* Сервис телеметрии [http://localhost:5001](http://localhost:5001)
* Postgres          localhost:5433

Описание:
* Добавлено 2 сервиса. Сервис телеметрии написан на Python, сервис управления датчиками написан на .NET 8.
* Исправлен монолит. Убрана база данных. Логика переделана на работу с микросервисами.
* Монолит общается с сервисами по REST.
* Сервис устройств общается с сервисом телеметрии через Kafka, отправляя значение телеметрии. В реальной системе, данные бы приходили от другого микросервиса, взаимодействующего с датчиками.
* Сервис телеметрии и сервис управления датчиками имеют собственные БД Postgres для хранения данных.

Как проверить:
1) Выполнить запросы согласно postman collection. Запросы должны быть выполнены успешно. **В процессе участвует kafka с автогенерацией топиков, и первый запрос на создание датчика происходит дольше обычного (несколько секунд).** Следующие быстро. Значение value (температура) будет меняться при каждом обращении и возвращать случайное значение.

1) Чтобы проверить, что сервисы взаимодействуют через Kafka, необходимо перейти по адресу [http://localhost:8090/](http://localhost:8090/) , открыть вкладку Topics. В таблице должен присутствовать топик smarthome.sensor.telemetry. А в нём сообщение о получении телеметрии от датчика. В сообщении будет водно что значение температуры равно 1_000_000. Это специальное некоррректное значение, встретив которое сервис телеметрии будет генерировать случайные значения, чтобы удовлетворить заданию. Его формирует сервис устройств при создании. В реальной системе, данные бы приходили от другого микросервиса, взаимодействующего с датчиками.

1) На Вкладке Consumers, в таблице должен находиться smarthome.sensor.telemetry.consumer. Это подписчик python сервиса телеметрии. Он читает smarthome.sensor.telemetry. Убедиться, что сообщения прочитаны можно сравнив Current Offset и End offset. Они должны быть равны.

1) Также, убедиться в работоспособности можно подключившись к базам данных  devices_db и telemetry_db. Порт 5433, пользователь postgres, пароль password. 
