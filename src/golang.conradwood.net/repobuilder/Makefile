.PHONY:	client server
all:	client server
client:
	cd client && go install ${LDFLAGS} repobuilder-client.go
server:
	cd server && go install ${LDFLAGS} repobuilder-server.go `ls -1 *.go|grep -v repobuilder-server.go`
