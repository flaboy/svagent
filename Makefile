OBJ:=svagent

all: ${OBJ}

${OBJ}: $(shell find . -name "*.go") proto/*.proto
	go build -o $@ .

proto/*.go: proto/*.proto
	protoc -I ./proto --go_out=plugins=grpc:proto ./proto/*.proto	

clean:
	rm ${OBJ}

start: all
	./${OBJ} start

win64:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o ${OBJ}.exe .

linux64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ${OBJ}.linux64 .

arm:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm go build -o ${OBJ}.arm .