# Todolist
This repository provides a base project to implemet Todolist application.

`docker-compose.yml` provides Go 1.17 build tool, MySQL server and phpMyAdmin.

## Dependencies
- [Gin Web Framework](https://pkg.go.dev/github.com/gin-gonic/gin)
- [Sqlx](https://pkg.go.dev/github.com/jmoiron/sqlx)

## How to run the application
First, you need to start Docker containers.
```sh
$ docker-compose up -d
```
This command will take time to download and build the containers.

Now you can start the application with the following command.
```sh
$ docker-compose exec app go run main.go
```
You can also execute `go run main.go` directly if you have Go development tools on your machine, but you need to setup the configuration to connect the application with MySQL server.

When you finish exercise, please don't forget to stop the containers.
```sh
$ docker-compose down
```

## Advanced: How to initialize database
When you modify the database schema, you will need to discard the current DB volumes for creating a new one.
It will be easier to rebuild everything than to rebuild only DB container.
Following command helps you to do it.
```sh
$ docker-compose down --rmi all --volumes --remove-orphans
$ docker-compose up -d
```
