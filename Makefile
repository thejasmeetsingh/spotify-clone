create-network:
	docker network create -d bridge shared-network

start-services:
	docker-compose -f src/services/user/docker-compose.yml up -d

	docker-compose -f src/services/content/docker-compose.yml up -d

	docker-compose -f src/services/conversion/docker-compose.yml up -d

	docker run --name api_gateway -d -p 8000:8000 \
		-v ./src/api_gateway/nginx.conf:/etc/nginx/conf.d/default.conf \
		--network shared-network nginx:1.25.4-alpine

stop-services:
	docker-compose -f src/services/user/docker-compose.yml down

	docker-compose -f src/services/content/docker-compose.yml down

	docker-compose -f src/services/conversion/docker-compose.yml down

	docker container stop api_gateway

	docker container rm api_gateway

