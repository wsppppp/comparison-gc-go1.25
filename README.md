# Go GC Stress Benchmark

Этот проект сравнивает:
- стандартный сборщик мусора Go
- экспериментальный Greentea GC

## Требования
- Go версии 1.25+

## Структура проекта

.
├── gc-hard-bench/.../stress_bench.go
├── logs/
├── profiles/
├── scripts/
│   └── analyze.py
└── README.md

## Сборка

### Обычный GC
go build -o oldgc stress_bench.go

### Greentea GC
GOEXPERIMENT=greenteagc go build -o greenteagc stress_bench.go

## Запуск тестов (пример: 8 прогонов)

mkdir -p logs profiles

### Classic GC
for i in {1..8}
do
GOMEMLIMIT=2GiB GODEBUG=gctrace=1 ./oldgc > logs/oldgc_$i.log
mv heap.prof profiles/heap_old_$i.prof
done

### Greentea GC
for i in {1..8}
do
GOMEMLIMIT=2GiB GODEBUG=gctrace=1 ./greenteagc > logs/greentea_$i.log
mv heap.prof profiles/heap_new_$i.prof
done

## Анализ логов

python3 scripts/analyze.py

## Анализ heap-профилей

go tool pprof -top ./oldgc profiles/heap_old_1.prof
go tool pprof -top ./greenteagc profiles/heap_new_1.prof
