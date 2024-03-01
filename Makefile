network:
	docker network create -d bridge shared-network

user-service:
	docker-compose -f src/user_service/docker-compose.yml up -d

content-service:
	docker-compose -f src/content_service/docker-compose.yml up -d
