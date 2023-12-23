# Project Title

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
- [github.com/gin-gonic/gin](https://github.com/gin-gonic/gin)
- [github.com/jinzhu/gorm](https://github.com/jinzhu/gorm)
- [github.com/stripe/stripe-go](https://github.com/stripe/stripe-go)
- [gorm.io/gorm](https://gorm.io/gorm)
- [github.com/dgrijalva/jwt-go](https://github.com/dgrijalva/jwt-go)
- [github.com/go-playground/validator/v10](https://github.com/go-playground/validator/v10)
- [github.com/joho/godotenv](https://github.com/joho/godotenv)
- [github.com/lib/pq](https://github.com/lib/pq)

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
```

Run it

```
$ gin -a 8080 -i run main.go
```
