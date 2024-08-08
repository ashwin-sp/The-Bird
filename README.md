# Distributed Systems Final Project 

Sudhanshu Pandey (NetID: sp6370)

Ashwin Suresh Babu (NetID: as14091)

## Presentation

![Architecture Diagram](https://github.com/os3224/final-project-b0c9bd62-as14091-sp6370/blob/main/Architecture.jpeg)

## Getting Started

Open the project in Visual Studio Code and ...

To start the web server: 
```
cd cmd/web
go run web.go 
```

To start the auth server: 
```
cd cmd/auth 
go run auth.go
```

To start the social server: 
```
cd cmd/social 
go run social.go
```

To start the tests for auth server:
```
cd web/auth
go test -v
```

To start the tests for social server: 
```
cd web/social 
go test -v
```

To start raft:

```
export GOPATH=<directory>
cd <directory>/src/go.etcd.io/etcd/contrib/raftimpl
go install github.com/mattn/goreman@latest
go build -o raftimpl
$HOME/go/bin/goreman start
```

To Execute Raft Persistance Test:
```
cd <directory>/src/go.etcd.io/etcd/contrib/raftimpl
chmod +x test_raft.sh
./test_raft.sh
```

