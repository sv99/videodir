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
    │   ├── 1._15
    │   ├── 1._18
    │   ├── 1._19
    │   └── A._02
    └── 24-01-18 02
        ├── 0._01
        ├── 0._02
        └── 0._04

Структура каталог VIDEO одноуровневая - папки **DD-MM-YY HH** для каждого часа
и внутри этих папок файлы видео. Номер камеры в расширении видеофайлов,
0._04 - видеофайл для 4 камеры. 

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
    │   │   ├── Event20180101102142001
    │   │   ├── Event20180101122042001
    │   │   ├── Event20180101182042001
    │   │   └── Event20180101222042001
    │   ├── 0102
    │   │   ├── Event20180102102142001
    │   │   ├── Event20180102122042001
    │   │   ├── Event20180102182042001
    │   │   └── Event20180102222042001
    │   ├── 1230
    │   │   ├── Event20171230102142001
    │   │   ├── Event20171230122042001
    │   │   └── Event20171230182042001
    │   └── 1231
    │       ├── Event20171231102142001
    │       ├── Event20171231122042001
    │       └── Event20171231182042001
    ├── aud01
    │   └── 0101
    │       ├── Event20180101102142001
    │       ├── Event20180101122042001
    │       ├── Event20180101182042001
    │       └── Event20180101222042001
    └── aud02
        └── 0102
            ├── Event20180101102142001
            ├── Event20180101122042001
            ├── Event20180101182042001
            └── Event20180101222042001

Структура каталога VIDEO двухуровневая сначала папки по камерам
и микрофонам и внутри них папки по месяцам и дням.
События - EventXXX непосредственно в папках для каждого дня.

Предположительно, при перекрытии года файлы попадут в ту же папку.
В имени файла присутствует полный год - конфликта имен не будет.

dependencies using dep
----------------------

Не хранит историю git - только актуалные файлы. В результате получаем очень
компактный размер папки vendor. На текущий момент меньше 8Мб.

```bash
brew install dep
dep init
# -v show extended log
dep ensure -v
```

Using:
 
[iris](https://iris-go.com) WEB framework for REST
 
[JWT auth](https://github.com/dgrijalva/jwt-go) + [JWT middleware for iris](github.com/iris-contrib/middleware/jwt) 

[TOML](https://github.com/BurntSushi/toml) for config

[htpasswd](https://github.com/foomo/htpasswd) Пришлось сделать форк - оригинальный
пакет не компилируется под 386 битную систему - ошибка переполнения int.
В авторском репозитарии уже 2 года висит patch request.

[CLI](https://github.com/teris-io/cli) for parsing command line

cross compilation
-----------------

Целевая система Windows.

    # compiling with additional environment variable
    GOOS=windows GOARCH=386 go build -o videodir.exe

Настроил дополнительную конфигурацию для генерации videodir.exe.

config
------

videodir.conf - TOML format

    # ServerAddr = ":8443"
    
    # HTTPS data``
    Cert = "localhost.crt"
    Key = "localhost.key"
    
    # array pathes for video data directories
    VideoDirs = [ "./video1/", "./video2/" ]

В windows нужно удваивать обратный слэш для VideoDirs.
Также двойной слэш возвращается и в результатах запросов с windows
сервера.

Handlers
--------

Handlers               | Query Type | Result
---------------------- | ---------- | ------
/                      | GET        | return index.html no auth
/login                 | POST       | post {"username: "some", "password": "pass"} return {"token": "JWT TOKEN"}
/api/v1/version        | GET        | return {"version": "0.1"}
/api/v1/volumes        | GET        | get array volumes with video dirs
/api/v1/list           | POST       | post { "path": [ "/24-01-18 01/" ] } get directory list scan all volumes, path may be empty for root directory
/api/v1/file           | POST       | post { "path": [ "/24-01-18 01/", "0._02" ] } get file scan all volumes and return file stream, path not may be empty

path передаем как массив элементов пути, в противном случае, когда 
передаем путь из windows система видит ескейп последовательности
вместо путей.

POST api tested in Postman

security
--------

Use HTTPS и JWT token

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

CLI для работы с htpasswd

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

Использовал для запуска сервиса утилиту [NSSM](http://nssm.cc/).
Результаты положительные. Можно не использовать свой лог файл.
NSSM позволяет перенаправить stdout и stderr в файл, в качестве бонуса
возможность настроить ротацию логов, правда она срабатывает только
при рестарте сервиса.

Первоначальное тестирование на предмет утечек памяти дало положительные
результаты. Память освобождается не очень быстро, но предсказуемо.
При пересылке файлов программа не занимала больше 20Мб.

    nssm install videodir C:/tools/videodir/videodir.exe
    # GUI edit params
    nssm edit videodir
    nssm remove videodir

    nssm start videodir
    nssm stop videodir
    nssm restart videodir
    
Альтернативный вариант - полноценный сервис Windows на базе
пакета [svc](https://github.com/golang/sys/tree/master/windows/svc)

todo
----

1.  Возможно использование дополнительных параметров командной строки для
работы с паролями в htpasswd  

2. Test

3. Тестирование на предмет утечек памяти в реальных условиях
Проблема не подтверждена. Вызов сборщика мусора не мгновенный но
предсказуемый. Нужно отследить долговременное выделение памяти.

4. Распределение проекта по отдлельным файлам.
