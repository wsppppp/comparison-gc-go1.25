# 🟦 Go GC Benchmark Suite (Сравнительный анализ старого и нового сборщика мусора Go)

В этом проекте — набор синтетических и стрессовых бенчмарков для анализа поведения сборщиков мусора Go (старый и greenteagc). Цель: сравнение времени работы, эксплуатационных пауз и структуры живой памяти.

---

## 📦 Подготовка
**Убедитесь, что установлен Go версии 1.25+, поддерживающий greenteagc:**
   ```sh
   go version
   ```

---

## ⚙️ Компиляция

Соберите два исполняемых файла — с разными режимами GC:

```sh
go build -o oldgc main.go
GOEXPERIMENT=greenteagc go build -o greenteagc main.go
```
_`main.go` — замените на соответствующий бенчмарк-файл, если используете нестандартное имя_

---

## 🚀 Запуск экспериментов

### Базовые тесты (несколько прогонов):

```sh
mkdir -p "logs"
for i in {1..10}
do
    GODEBUG=gctrace=1 ./oldgc > logs/oldgc_run${i}.log
    mv heap.prof heap_old_${i}.prof
done

for i in {1..10}
do
    GODEBUG=gctrace=1 ./greenteagc > logs/greenteagc_run${i}.log
    mv heap.prof heap_new_${i}.prof
done
```
_Все логи и heap-дампы будут в папке `logs/` или рабочей директории_

### Для стрессового бенчмарка — аналогично, но увеличьте количество прогонов (30+):

```sh
for i in {1..30}
do
    GODEBUG=gctrace=1 ./oldgc > logs/oldgc_stress_${i}.log
    mv heap.prof heap_old_stress_${i}.prof
done

for i in {1..30}
do
    GODEBUG=gctrace=1 ./greenteagc > logs/greenteagc_stress_${i}.log
    mv heap.prof heap_new_stress_${i}.prof
done
```

---

## 🧮 Сбор и обработка результатов

1. **В логах запуска (например, `logs/oldgc_run1.log`) будут:**
    - `Elapsed (s)`
    - `GC Pause total (ms)`
    - `NumGC`
    - `HeapAlloc (MiB)`
    - `TotalAlloc (MiB)`
---

## 🔬 Анализ heap-слепков

1. **Откройте heap-профиль одного из прогонов:**
    ```sh
    go tool pprof -top ./oldgc heap_old_1.prof
    go tool pprof -top ./greenteagc heap_new_1.prof
    # ...или любой другой файл heap_*.prof
    ```
2. **Можно изучить распределение памяти по основным функциям и типам (`main.worker`, служебные runtime)**
    - Основная доля должна быть в пользовательских структурах.
    - Можно поискать аномалии, утечки.

3. _Для визуализации:_
    ```sh
    go tool pprof -http=:8080 ./oldgc heap_old_1.prof
    ```
   Откроется Web-интерфейс с графиками heap (но там нужно кое-что предустановить).

---