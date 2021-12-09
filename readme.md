videodir
========

Доступ к каталогу с видеофайлами на видеорегистраторе.
Интересуют видеорегистраторы ITV Intellect и Geovision.
Использование для архивирования с удаленного клиента.

Через REST API список файлов и выгрузка отдельных файлов.

Нужные для работы файлы:

    VIDEODIR
    ├── videodir.exe
    ├── videodir.conf
    ├── *.crt
    ├── *.key
    ├── htpasswd
    ├── favicon.ico
    └── index.html

Структура каталогов ITV
-----------------------

    VIDEO
    ├── 24-01-18 01
    │   ├── 1._15
    │   ├── 1._18
    │   ├── 1._19
    │   └── A._02
    └── 24-01-18 02
        ├── 0._01
        ├── 0._02
        └── 0._04

Структура каталог `VIDEO` одноуровневая - папки **DD-MM-YY HH** для каждого часа
и внутри этих папок файлы видео. Номер камеры в расширении видеофайлов,
`0._04` - видеофайл для 4 камеры. 

На живом сервере на 2Тб разделе почти 2000 папок и больше 700 000
файлов.
 
Запрос на получение всех файлов раздела обрабатывается по
локальной сети больше 8 мин. _Нужно обрабатывать данные по каждой
папке отдельно._

Структура каталогов Geovision
-----------------------------

    VIDEO
    ├── cam01
    │   ├── 0101
    │   │   ├── Event20180101102142001
    │   │   ├── Event20180101122042001
    │   │   ├── Event20180101182042001
    │   │   └── Event20180101222042001
    │   ├── 0102
    │   │   ├── Event20180102102142001
    │   │   ├── Event20180102122042001
    │   │   ├── Event20180102182042001
    │   │   └── Event20180102222042001
    │   ├── 1230
    │   │   ├── Event20171230102142001
    │   │   ├── Event20171230122042001
    │   │   └── Event20171230182042001
    │   └── 1231
    │       ├── Event20171231102142001
    │       ├── Event20171231122042001
    │       └── Event20171231182042001
    ├── aud01
    │   └── 0101
    │       ├── Event20180101102142001
    │       ├── Event20180101122042001
    │       ├── Event20180101182042001
    │       └── Event20180101222042001
    └── aud02
        └── 0102
            ├── Event20180101102142001
            ├── Event20180101122042001
            ├── Event20180101182042001
            └── Event20180101222042001

Структура каталога `VIDEO` двухуровневая сначала папки по камерам
и микрофонам и внутри них папки по месяцам и дням.
События - `EventXXX` непосредственно в папках для каждого дня.

Предположительно, при перекрытии года файлы попадут в ту же папку.
В имени файла присутствует полный год - конфликта имен не будет.

Структура каталогов Synology SurveillanceStation
------------------------------------------------

    surveillance
    ├── 152
    │   ├── 20211120AM
    │   │   ├── 152-20211120-101913-1637392753.mp4
    │   │   └── 152-20211120-111623-1637396183.mp4
    │   └── 20211120PM
    │       └── 152-20211120-162424-1637414664.mp4
    ├── 153
    │   ├── 20211120AM
    │   │   ├── 153-20211120-101913-1637392753.mp4
    │   │   └── 153-20211120-111623-1637396183.mp4
    │   └── 20211120PM
    │       └── 153-20211120-162424-1637414664.mp4
    └── 154
        └── 20211120PM
            └── 154-20211120-162424-1637414664.mp4

Структура каталога `surveillance` двухуровневая сначала папки по камерам
и внутри них папки по дням. Записи за день разделены на две части.

Структура архива не позволяет данным за разные года перекрываться.
В имени как папки, так и файла присутствует полная дата,
представленная в виде, удобном для сортировки.

cross compilation
-----------------

```bash
# windows service
make service
# linux binary
make linux
```

golang packages
---------------

* Первый вариант сделал на [github.com/kataras/iris](https://iris-go.com)
  Проблемы вылезли с назойливым предложением обновиться.
  В результате ушел на [github.com/gofiber/fiber](https://github.com/gofiber/fiber).
  Единственный минус - пока нет httptest клиента, аналогичного имеющемся в `iris`.
 
* TOML config file parsing with [github.com/BurntSushi/toml](https://github.com/BurntSushi/toml)

* [htpasswd github.com/foomo/htpasswd](https://github.com/foomo/htpasswd)

* Embedding index.html and favicon.ico with [go-bindata](https://github.com/go-bindata/go-bindata) package.

* Windows service based on `golang.org/x/sys/windows/svc` и [github.com/billgraziano/go-windows-svc](https://github.com/billgraziano/go-windows-svc)

* [CLI github.com/teris-io/cli](https://github.com/teris-io/cli) - для windows дополнительные команды
  для управления сервисом.

* [Logger github.com/rs/zerolog](https://github.com/rs/zerolog)

* [Еще один вариант CLI github.com/spf13/cobra](https://github.com/spf13/cobra)

go-bindata
----------

Embedding index.html and favicon.ico

```bash
# install go-bindata
go get -u github.com/go-bindata/go-bindata/...
# generate assets.go
go-bindata -pkg videodir -o assets.go -nocompress -nocompress -prefix static static/
```

config
------

videodir.conf - TOML format

    LogLevel = "info"
    # ServerAddr = ":8443"
    
    # HTTPS data``
    Cert = "localhost.crt"
    Key = "localhost.key"
    
    # array pathes for video data directories
    VideoDirs = [ "./video1/", "./video2/" ]

В windows нужно удваивать обратный слэш для VideoDirs.
Также двойной слэш возвращается и в результатах запросов с windows
сервера.

**iris.yml** config file for iris framework

Handlers
--------

Handlers               | Query Type | Result
---------------------- | ---------- | ------
/                      | GET        | return index.html no auth
/login                 | POST       | post {"username: "some", "password": "pass"} return {"token": "JWT TOKEN"}
/api/v1/version        | GET        | return {"version": "0.1"}
/api/v1/volumes        | GET        | get array volumes with video dirs
/api/v1/list           | POST       | post { "path": [ "/24-01-18 01/" ] } get directory list, scan all volumes, path may be empty for root directory
/api/v1/file           | POST       | post { "path": [ "/24-01-18 01/", "0._02" ] } get file, scan all volumes and return file stream, path not may be empty
/api/v1/filesize       | POST       | post { "path": [ "/24-01-18 01/", "0._02" ] } get filesize, scan all volumes and return file size
/api/v1/remove         | POST       | post { "path": [ "/24-01-18 01/", "0._02" ] } remove path (directory o single file) return {"result": "OK"} or {"result": err.Error()}, search path for remove on all volumes

path передаем как массив элементов пути, в противном случае, когда 
передаем путь из windows система видит ескейп последовательности
вместо путей.

POST api tested in Postman

security
--------

Use HTTPS и JWT token (SigningMethodHS384)

Для HTTPS использовал RSA ключи, эти же ключи использовал для
подписи и проверки JWT. RSA используется в JWT библиотеке, 
менять ничего не хотелось. 

    openssl req \
        -x509 \
        -nodes \
        -newkey rsa:2048 \
        -keyout server.key \
        -out server.crt \
        -days 3650 \
        -subj "/C=RU/ST=SanktPetersburg/L=SanktPetersburg/O=Security/OU=IT Department/CN=*"

Использовал сгенерированные на основании rost.cert RSA ключи.

Пароли храним в htpasswd - точку перед именем очень не любит Windows.

    # create htpasswd with bcrypt hashes
    htpasswd -cbB htpasswd admin admin
    # add or update bcrypt hash
    htpasswd -bB htpasswd dima dima

CLI для работы с htpasswd. Для работы достаточно htpasswd нулевого размера. 

```bash
>videodir -h

Description:
    videodir tool

Sub-commands:
    videodir list     list users from htpasswd
    videodir add      add or update user in the htpasswd
    videodir remove   remove user from htpasswd

>videodir add --help
videodir add <name> <password>

Description:
    add or update user in the htpasswd

Arguments:
    name       user name
    password   password
    
>videodir remove --help
 videodir remove <name>
 
 Description:
     remove user from htpasswd
 
 Arguments:
     name   user name
```

windows service
---------------

Словил проблему с инициализацией приложения. Для правильного конфигурирования
логгера нужно, чтобы приложение было создано в режиме сервиса (nonInteractive mode).
Поэтому приложение нельзя инициализировать сразу - есть два режима
запуска из командной строки и в режиме сервиса. Режим сервиса тоже два варианта
старт с ключом `start` или из апплета Сервис. 

synology service
----------------

Возникла необходимость забирать архив из Synology SurveillanceStation.

DSM7.0 использует `systemd`

Копируем нужные файлы в домашний каталог пользователя `admin`

```
ls ~/videodir
log/
htpasswd
videodir_linux_amd64
videodir.config
videodir.service
server.crt
server.key
```

Теперь можно попытаться запустить сервис.

```bash
systemctl start videodir.service
systemctl stop videodir.service
systemctl status videodir.service

systemctl enable videodir.service
```

todo
----

1. Выявлена проблема связанная со сканированием портов и попытками взлома.
   Сервер упал и не поднялся самостоятельно, хотя вроде бы должен был.
   Проблему решил радикально фильтрацией по IP на Mikrotik но осадочек остался.
   Повторно не подключал.
