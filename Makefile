include .env

dev-migrate-up:
	migrate \
		-database 'postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable' \
		-path infra/db/migrations \
		up
dev-migrate-down:
	migrate \
		-database 'postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable' \
		-path infra/db/migrations \
		down
dev-migrate-new:
	migrate \
		create -ext sql -dir infra/db/migrations $(NAME)
dev-migrate-fix:
	migrate \
		-database 'postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable' \
		-path infra/db/migrations \
		force $(VERSION)
