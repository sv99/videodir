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
    ├── favicon.ico
    ├── *.crt
    ├── *.key
    ├── .htpasswd
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
В имени файла присутствует полынй год - конфликта имен не будет.

dependencies using dep
----------------------

    brew install dep
    dep init
    dep ensure

https
-----

    # Key considerations for algorithm "ECDSA" ≥ secp384r1
    # List ECDSA the supported curves (openssl ecparam -list_curves)
    openssl ecparam -genkey -name secp384r1 -out server.key

    # Generation of self-signed(x509) public key (PEM-encodings .pem|.crt) based on the private (.key)
    openssl req -new -x509 -sha256 -key server.key -out server.crt -days 3650

Использовал сгенерированные на основании rost.cert RSA ключи.

Этот же ключ используется для подписи и проверки JWT Token.

web framework
-------------

Using [iris](https://iris-go.com) and REST with JWT auth. 

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

REST API
--------

API                    | Query Type | Result
---------------------- | ---------- | ------
/api/v1/version        | GET        | return {"version": "0.1"}
/api/v1/volumes        | GET        | get array volumes with video dirs
/api/v1/list           | POST       | get directory list { "path": [ "/24-01-18 01/" ] } for all volumes, path may be empty for root directory
/api/v1/file           | POST       | get file { "path": [ "/24-01-18 01/", "0._02" ] } find path on all volumes and return file stream, path not may be empty


path передаем как массив строк, в противном случае, когда 
передаем путь из windows система видит ескейп последовательности
вместо путей

POST api tested in Postman

security
--------

    # create .htpasswd with bcrypt hashes
    htpasswd -cbB .htpasswd admin admin
    # add or update bcrypt hash
    htpasswd -bB .htpasswd dima dima
    
Use HTTPS и JWT token

JWT Token подписываем и проверяем тем же ключом, который использем
для HTTPS.

windows service
---------------

Никаких особых действий с сервисом не нужно, можно попробовать
утилиту [NSSM](http://nssm.cc/) и посмотреть что из этого выйдет.

Альтернативный вариант - сервис Windows на базе пакета [svc](https://github.com/golang/sys/tree/master/windows/svc)

todo
----

1. лог в файл

1.  Возможно использование дополнительных параметров командной строки для
работы с паролями в .htpasswd  

1. Test

2. Production - windows service

3. Тестирование на предмет утечек памяти в реальных условиях
Проблема не подтверждена. Вызов сборщика мусора не мгновенный но
предсказуемый. Нужно отследить долговременное выделение памяти.

4. Распределение проекта по отдлельным файлам.
