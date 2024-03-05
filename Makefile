create-network:
	docker network create -d bridge shared-network

start-nginx:
	docker run --name api_gateway -d -p 8000:8000 \
		-v ./nginx.conf:/etc/nginx/conf.d/default.conf \
		--network shared-network nginx:1.25.4-alpine

user-service-up:
	docker-compose -f src/user_service/docker-compose.yml up -d

user-service-down:
	docker-compose -f src/user_service/docker-compose.yml down

content-service-up:
	docker-compose -f src/content_service/docker-compose.yml up -d

content-service-down:
	docker-compose -f src/content_service/docker-compose.yml down
