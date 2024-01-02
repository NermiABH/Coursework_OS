build:
	sudo docker-compose build
run:
	sudo docker-compose up
restart:
	sudo docker-compose restart
rebuild:
	sudo docker-compose up -d --no-deps --build
down:
	sudo docker-compose down
stop:
	sudo docker-compose stop
server_1:
	go run server1/server1.go
server_2:
	go run server2/server2.go
server_3:
	go run server3/server3.go
client run:
	go run client/client.go