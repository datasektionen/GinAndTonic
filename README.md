# GinAndTonic

## Description

GinAndTonic is a robust backend application designed to work in conjunction with [Tessera](https://github.com/DowLucas/tessera). This project is built with Go and uses the Gin framework to provide a fast and flexible API for the Tessera frontend.

The application includes a variety of features such as authentication, ticket allocation, event management, and more. It is designed with a modular structure, making it easy to understand and extend. The codebase includes several packages such as controllers, services, models, and utils that encapsulate different functionalities of the application.

GinAndTonic uses a PostgreSQL database for data persistence and GORM as an ORM for data manipulation. It also integrates with Stripe for payment processing.

This project is suitable for anyone looking to understand how to build a comprehensive backend system using Go, Gin, and other modern technologies.

## Getting Started

## Dependencies

The project uses a variety of Go modules to handle different aspects of the application, from web server management to database interactions, payment processing, and more. Here's a list of the key dependencies:

### Web Framework and Middleware

- **Gin Web Framework**: A high-performance HTTP web framework that provides a robust set of features for building web applications. ([github.com/gin-gonic/gin](https://github.com/gin-gonic/gin))
- **Gin CORS Middleware**: Official CORS middleware for Gin, allowing for flexible cross-origin resource sharing policies. ([github.com/gin-contrib/cors](https://github.com/gin-contrib/cors))

### Database and ORM

- **GORM**: The fantastic ORM library for Golang, aiming to be developer-friendly. It simplifies CRUD operations and database interactions. ([gorm.io/gorm](https://gorm.io/gorm))
- **lib/pq**: Pure Go Postgres driver, supports basic features of PostgreSQL. ([github.com/lib/pq](https://github.com/lib/pq))
- **GORM PostgreSQL Driver**: Official GORM driver for PostgreSQL. ([gorm.io/driver/postgres](https://gorm.io/docs/connecting_to_the_database.html#PostgreSQL))
- **GORM SQLite Driver**: Official GORM driver for SQLite. ([gorm.io/driver/sqlite](https://gorm.io/docs/connecting_to_the_database.html#SQLite))

### Payment Processing

- **Stripe Go**: Official Go library for the Stripe API, for integrating payment processing. ([github.com/stripe/stripe-go](https://github.com/stripe/stripe-go) and [github.com/stripe/stripe-go/v72](https://github.com/stripe/stripe-go))

### Authentication and Security

- **JWT Go**: Community maintained clone of `github.com/dgrijalva/jwt-go`, used for creating and validating JSON Web Tokens. ([github.com/golang-jwt/jwt](https://github.com/golang-jwt/jwt))

### Environment Configuration

- **GoDotEnv**: Loads environment variables from `.env` files, making it easier to manage configuration in different environments. ([github.com/joho/godotenv](https://github.com/joho/godotenv))

### Validation

- **Validator**: Provides struct and field validation, including cross-field, cross-struct, and collection validations. ([github.com/go-playground/validator](https://github.com/go-playground/validator))

### Scheduling

- **Cron**: A cron library for Go that allows scheduling recurring tasks using a cron-spec syntax. ([github.com/robfig/cron](https://github.com/robfig/cron))

### Testing

- **Testify**: An extension to the standard Go testing package, offering more powerful assertions and testing utilities. ([github.com/stretchr/testify](https://github.com/stretchr/testify))

### Supplementary Packages

- **Go Time**: Supplementary time packages for Go, useful for more complex time manipulation beyond the standard library. ([golang.org/x/time](https://golang.org/x/time))

Please refer to the individual project pages for more detailed information on each dependency, including how to use them in your Go projects.


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
SPAM_TEST_EMAIL=<YOUR TEST EMAIL>
STRIPE_WEBHOOK_SECRET=<INSERT STRIPE WEBHOOK SECRET>
SPAM_TEST_EMAIL=<YOUR_EMAIL>
AWS_ACCESS_KEY_ID=<AWS_ACCESS_KEY_ID>
AWS_SECRET_ACCESS_KEY=<AWS_SECRET_ACCESS_KEY>
AWS_REGION=<AWS_REGION>
```

Run it

```
$ gin -a 8080 -i run main.go
```

#### Using Docker

You can also use docker to run the application. The benefit with running docker is that it uses [nyckeln-under-dorrmattan](https://github.com/datasektionen/nyckeln-under-dorrmattan) to mimic login, so you dont have to clone the repo. It also creates and manages the PostgresSQL database for you.

    $ docker-compose up --build

docker-compose uses `Dockerfile.dev`. `Dockerfile.prod` is only used for deploying to dokku.
