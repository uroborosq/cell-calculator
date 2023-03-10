# Калькулятор CSV-файлов

## Фунциональное описание
Программа принимает в качестве первого аргумента путь к файлу в формате CSV, раздленного запятыми. Файл может содержать 
числа и формулы со ссылками на другие ячейки. Размер файла неограничен. 
## Техническое описание
Реализовано без использования сторонних библиотек, а так же "encodings/csv". Парсер имеет два поведения в зависимости от величины файла:
1. Пройтись по файлу, собрать все адреса ячеек на которые ссылаются формулы, пройтись еще раз, собрать значения и загрузить в память, пройтись в третий раз и выводить значения в консоль, одновременно считая формулы, используя заготовленные значения ячеек. 
2.  Идти по файлу, если встречается формула - прочесать весь файл в поисках необходимых значений, посчитать, закешировать значения в ограниченного размера буфер. Использует фиксированное количество памяти, но работает, очевидно, очень долго.

При возникновении ошибки, программа останавливается и выводит сообщение в консоль.
## Инструкция по сборке, запуску и тестированию

### Сборка

Сборка осуществляется стандартными средствами Go для всех поддерживаемых операционных систем (Linux, Windows, Mac, etc.)

```shell
go build
```

### Запуск

```shell
./cell-calulator /path/to/file
```

### Тестирование

```shell
go test -test.v
```