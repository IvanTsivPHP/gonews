Все конфиги вынесены в config.json
стандарный запуск запускает на memdb
флаг -db меняет запускаемую базу
первый запуск:
go run cmd/server/server.go -db=mongodb --seed
или
go run cmd/server/server.go -db=postgres --migrate
повторные
go run cmd/server/server.go -db=mongodb
или
go run cmd/server/server.go -db=postgres
