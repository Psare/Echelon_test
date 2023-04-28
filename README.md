# Echelon_test

Приветствую! Чтобы воспользоваться программой ее нужно скачать, git clone git@github.com:Psare/Echelon_test.git
Перейти в папку cd Echelon_test
И запустить go build или go run main.go
Чтобы поменять файл вывода, нужно добавить флаг --output <filename> , например, go run main.go --output filename.sqlite
Задание
Написать сервис-скрапер дерева OID с сайта https://oidref.com/. Требования к скраперу:
• CLI, приложение терминала Linux/Windows
• парсит данные сайта https://oidref.com/
• охраняет данные файл в формате SQLite, схема таблиц - на ваш выбор
• имеет возможность сохранения прогресса парсинга, сохраняет отметку прогресса
парсинга в таблице, в случае рестарта и наличия отметки - возобновляет парсинг с
момента остановки
• отображает в консоле информацию о статусе загрузки (выполнения или просто число
скачанных описаний MIB)
• имеет входной параметр --output <filename>, по умолчанию mibs.sqlite
• вежливо парсит сайт, не создаёт излишней нагрузки https://www.scraperapi.com/blog/5-
tips-for-web-scraping/
• в случае отказов делает повторы (retries), при превышении порога делает таймаут, после
возобновляет попытки – дать возможность запустить утилиту на сутки, и она будет
парсить сайт даже если происходили длительные отказы;
