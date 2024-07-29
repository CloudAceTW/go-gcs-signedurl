dep-up:
	docker compose -f docker-compose-dev.yml up -d

dep-down:
	docker compose -f docker-compose-dev.yml down

dev-build:
	docker build -t go-gcs-signedurl-dev -f Dockerfile-dev .

dev-console:
	docker run --network=go-gcs-signedurl_redis -p 8080:8080 --env-file .env -v $$(pwd):/opt/app -it --rm go-gcs-signedurl-dev sh
