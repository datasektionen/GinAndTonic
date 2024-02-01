# GinAndTonic

## Description

GinAndTonic is a robust backend application designed to work in conjunction with [Tessera](https://github.com/DowLucas/tessera). This project is built with Go and uses the Gin framework to provide a fast and flexible API for the Tessera frontend.

The application includes a variety of features such as authentication, ticket allocation, event management, and more. It is designed with a modular structure, making it easy to understand and extend. The codebase includes several packages such as controllers, services, models, and utils that encapsulate different functionalities of the application.

GinAndTonic uses a PostgreSQL database for data persistence and GORM as an ORM for data manipulation. It also integrates with Stripe for payment processing.

This project is suitable for anyone looking to understand how to build a comprehensive backend system using Go, Gin, and other modern technologies.

## Getting Started

### Dependencies

Here are the dependencies used in this project:

- [PostgreSQL](https://www.postgresql.org/): Used as the primary database for data persistence.
- [Stripe](https://stripe.com/): Used for handling payment processing.
- [Gin Web Framework](https://github.com/gin-gonic/gin): An HTTP web framework written in Go (golang) to build web applications.
- [GORM](https://gorm.io/gorm): The fantastic ORM library for Golang, aims to be developer friendly.
- [Stripe Go](https://github.com/stripe/stripe-go): Go library for the Stripe API.
- [JWT Go](https://github.com/golang-jwt/jwt): Community maintained clone of `github.com/dgrijalva/jwt-go`.
- [GoDotEnv](https://github.com/joho/godotenv): A Go port of Ruby's dotenv library (Loads environment variables from `.env`).
- [Validator](https://github.com/go-playground/validator): Go Struct and Field validation, including Cross Field, Cross Struct, Map, Slice and Array diving.
- [lib/pq](https://github.com/lib/pq): Pure Go Postgres driver for database/sql.
- [Gin CORS Middleware](https://github.com/gin-contrib/cors): Official CORS gin's middleware.
- [Golang-JWT](https://github.com/golang-jwt/jwt): Community maintained clone of `github.com/dgrijalva/jwt-go`.
- [Cron](https://github.com/robfig/cron): A cron library for Go.
- [Testify](https://github.com/stretchr/testify): A sacred extension to the standard go testing package.
- [Stripe Go v72](https://github.com/stripe/stripe-go): Go library for the Stripe API, version 72.
- [Go Time](https://golang.org/x/time): Supplementary time packages for Go.
- [GORM PostgreSQL Driver](https://gorm.io/docs/connecting_to_the_database.html#PostgreSQL): GORM official driver for PostgreSQL.
- [GORM SQLite Driver](https://gorm.io/docs/connecting_to_the_database.html#SQLite): GORM official driver for SQLite.

Please note that the actual dependencies and their versions are specified in the [go.mod](go.mod) file.

### Installing

```bash
$ git clone git@github.com:DowLucas/tessera.git
$ cd tessera
$ go mod download
```

#### Installing gin

```bash
$ go get github.com/codegangsta/gin
gin -h
```

Set up your environment variables. You can do this by creating a .env file in the root directory of the project and populating it with the necessary variables. Here's an example of what your .env file might look like:

```
DB_USER=<INSERT USER HERE>
DB_PASSWORD=<INSERT PASSWORD HERE>
DB_NAME=<INSERT DB NAME HERE>
DB_PORT=<INSERT DB PORT HERE>
PORT=8080
SECRET_KEY=<INSERT KEY HERE>
LOGIN_BASE_URL=http://localhost:1337
FRONTEND_BASE_URL=<INSERT URL HERE>
LOGIN_API_KEY=<INSERT LOGIN API KEY>
JWT_KEY=<INSERT JWT KEY>
STRIPE_SECRET_KEY=<INSERT STRIPE SECRET KEY>
SPAM_API_KEY=<INSERT SPAM API KEY>
SPAM_TEST_EMAIL=<YOUR EMAIL>
```

Run it

```
$ gin -a 8080 -i run main.go
```

#### Using Docker

You can also use docker to run the application. The benefit with running docker is that it uses [nyckeln-under-dorrmattan](https://github.com/datasektionen/nyckeln-under-dorrmattan) to mimic login, so you dont have to clone the repo. It also creates and manages the PostgresSQL database for you.

    $ docker-compose up --build

docker-compose uses `Dockerfile.dev`. `Dockerfile.prod` is only used for deploying to dokku.
