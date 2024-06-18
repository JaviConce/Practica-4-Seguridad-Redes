.PHONY: file auth broker
all: build network containers

test:
	python3 Test/Testing.py
go: auth broker file

auth:
	cd docker/auth/AuthCode && go build auth.go && mv auth ..
broker:
	cd docker/broker/BrokerCode && go build broker.go && mv broker ..
file:
	cd docker/file/Filecode && go build file.go && mv file ..


containers:
	docker run --privileged -p 5000:5000 --rm -ti -d --name router --hostname router midebian-router
	docker network connect dmz router 
	docker network connect srv router
	docker network connect dev router

	docker run --privileged --rm -ti -d --name work --hostname work --ip 10.0.3.3 --network dev midebian-work

	docker run --privileged --rm -ti -d --name jump --hostname jump --ip 10.0.1.3 --network dmz midebian-jump

	docker run --privileged --rm -ti -d --name broker --hostname broker --ip 10.0.1.4 --network dmz midebian-broker

	docker run --privileged --rm -ti -d --name auth --hostname auth --ip 10.0.2.3 --network srv	midebian-auth

	docker run --privileged --rm -ti -d --name file --hostname file --ip 10.0.2.4 --network srv midebian-file 

build:
	#docker build --rm -f docker/Dockerfile --tag midebian docker/
	docker build --rm -f docker/router/Dockerfile --tag midebian-router docker/router
	docker build --rm -f docker/jump/Dockerfile --tag midebian-jump docker/jump
	docker build --rm -f docker/work/Dockerfile --tag midebian-work docker/work
	docker build --rm -f docker/broker/Dockerfile --tag midebian-broker docker/broker
	docker build --rm -f docker/auth/Dockerfile --tag midebian-auth docker/auth
	docker build --rm -f docker/file/Dockerfile --tag midebian-file docker/file

network:
	docker network create -d bridge --subnet 10.0.1.0/24 dmz
	docker network create -d bridge --subnet 10.0.2.0/24 srv
	docker network create -d bridge --subnet 10.0.3.0/24 dev

remove:
	docker stop router work jump broker file auth
	docker network prune -f
