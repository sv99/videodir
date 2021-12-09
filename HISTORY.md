videodir
========

Сервис для выгрузки видеоданных из видеорегистраторов.

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

лето 2020
---------

После обновления последнего регистратора с Windows XP в июле 2020,
можно обновить `golang` и перейти с `dep` сборки на `mod`.

ноябрь 2021
-----------

1. version 0.5
2. upgrade to fiber/v2
3. Тестирование работы на Synology DS7 and SurveillanceStation

Пока оставил zerolog, переходить на родной не стал.
