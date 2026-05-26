ALL:
	go run cmd/api/main.go

parse:
	go run cmd/import/main.go -file=pkg/questions/questions.json -drop=true

env:
	export DB_USER="atlas"
	export DB_PASSWORD="123"
	export DB_HOST="localhost"
	export DB_PORT="5432"
	export DB_NAME="med_book"
