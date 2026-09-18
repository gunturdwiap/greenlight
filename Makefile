DB_DSN ?= postgres://greenlight:pa55word@localhost/greenlight?sslmode=disable

.PHONY: migrate-up
migrate-up:
	migrate -path=./migrations -database="$(DB_DSN)" up
