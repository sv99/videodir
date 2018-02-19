videodir
========

Доступ к каталогу с видеофайлами на локальном компьютере.

Через REST список файлов и выгрузка отдельных файлов, как статики
или через REST запрос.

Структура каталогов ITV
-----------------------

    VIDEO
    ├── 24-01-18 01
    │   ├── 1._15
    │   ├── 1._18
    │   ├── 1._19
    │   ├── A._02
    ├── 24-01-18 02
    │   ├── 0._01
    │   ├── 0._02
    │   ├── 0._04

Каталог VIDEO в нем папки - имя **DD-MM-YY HH** и внутри
этих папок файлы видео.

На живом сервере на 2Тб разделе почти 2000 папок и больше 700 000
файлов.
 
Запрос на получение всех файлов раздела обрабатывается по
локальной сети больше 8 мин. _Нужно обрабатывать данные по каждой
папке отдельно._

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


web framework
-------------

Using [iris](https://iris-go.com) and REST with JWT auth. 

cross compilation
-----------------

Целевая система Windows.

    # compiling with additional environment variable
    GOOS=windows GOARCH=386 go build -o videodir.exe

Настроил дополнительную конфигурацию.

config
------

videodir.conf - TOML format

    # ServerAddr = ":8443"
    
    # HTTPS data``
    Cert = "localhost.crt"
    Key = "localhost.key"
    
    # array pathes for video data directories
    VideoDirs = [ "./video1/", "./video2/" ]

В windows нужно удваивать обратный слэш для VideoDirs

REST API
--------

API                    | Query Type | Result
---------------------- | ---------- | ------
/api/v1/version        | GET        | return {"version": "0.1"}
/api/v1/list           | GET        | list video dirs
/api/v1/list/{number}  | GET        | list files in video dirs
/api/v1/file           | POST       | get file {"volumeid": "0", "path": "/24-01-18 01/0._02"}

POST api tested in Postman

security
--------

HTTPS и логин

windows service
---------------

fff

todo
----

6. JWT auth

7. Test

9. Production - windows servise 