up:
	docker compose up -d

down:
	docker compose down --volumes --remove-orphans

psql:
	docker compose exec pg psql -U postgres -d postgres -p 5432