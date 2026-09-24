.PHONY: test

test:
	@echo "=== 1. Limpiando contenedores y volúmenes previos ==="
	docker compose down -v --remove-orphans

	@echo "=== 2. Generando código Go con sqlc ==="
	~/go/bin/sqlc generate
	
	@echo "=== 3. Levantando contenedor de PostgreSQL ==="
	docker compose up -d

	@echo "=== 4. Esperando a que PostgreSQL acepte conexiones ==="
	@until docker exec mi_postgres pg_isready -U root > /dev/null 2>&1; do \
		echo "Postgres se está inicializando... esperando 1s"; \
		sleep 1; \
	done

	@echo "=== 5. Ejecutando la suite de pruebas ==="
	go test -v ./tests/...

	@echo "=== 6. Limpieza posterior (destruyendo contenedor) ==="
	docker compose down -v