# Go Webapp Example

This repo contains an example structure for a monolithic Go Web Application.

 You can read more details about this project on our blog:
 https://offline.ch/blog/go-applikations-architektur (german article)

## Architecture

This project loosely follows [Uncle Bob's Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html).

Other inspirations:

* https://github.com/golang-standards/project-layout
* https://pace.dev/blog/2018/05/09/how-I-write-http-services-after-eight-years
* https://github.com/MichaelMure/git-bug/tree/master/graphql/schema

## Features

It includes the following features:

* [Magefiles](https://magefile.org/)
* [Database migrations](https://github.com/golang-migrate/migrate)
* [Configuration](https://github.com/spf13/viper)
* [Data Seeder](https://github.com/romanyx/polluter) (Prod and Test)
* [Live reloading](https://github.com/cosmtrek/air) 
* [Linting](https://github.com/golangci/golangci-lint) 
* [GraphQL Server](https://gqlgen.com/)
* [User sessions](https://github.com/OFFLINE-GmbH/go-webapp-example/blob/master/internal/pkg/auth/auth.go#L114)
* [Role-based Access control](https://github.com/casbin/casbin)
* [One-time update routines](https://github.com/OFFLINE-GmbH/go-webapp-example/blob/master/internal/app/update.go#L58)
* [I18N](https://github.com/OFFLINE-GmbH/go-webapp-example/blob/master/pkg/i18n/i18n.go)
* [Background processes (Daemons)](https://github.com/OFFLINE-GmbH/go-webapp-example/blob/master/internal/daemon/example.go)
* [Unit](https://github.com/OFFLINE-GmbH/go-webapp-example/blob/master/internal/pkg/test/db.go) and [Integration](https://github.com/OFFLINE-GmbH/go-webapp-example/blob/master/internal/graphql/gqlresolvers/quote_test.go) tests

## Disclaimer

The whole project is tailored to our very specific needs for our [CareSuite software product](https://caresuite.ch/).
CareSuite is deployed via Docker as a monolithic application. It is not optimized for cloud deployments. 
While it might not fit your exact needs it can be a good inspiration when building a new application.   

## Authentication

All authentication has been disabled, so you can test the server without having to log in.

Remove the `if true {}` block in [internal/pkg/auth/auth.go:116](https://github.com/OFFLINE-GmbH/go-webapp-example/blob/master/internal/pkg/auth/auth.go#L116)
to require a session to access the backend server.

You can login with a `POST` request to `/backend/login`. You need to send a `username` and `password` value (by default both are set to `admin`).

## Get up and running

Use [Mage](https://magefile.org/) to run common tasks:

### Start the Docker stack

A MySQL server can be started using

```
mage -v run:docker
```

### Start the server in live reloading mode

To start the backend server, run

```
mage -v run:backend
```

If you make changes to the code, the binary will be rebuilt.

### Visit the GraphQL Playground

Once the Docker stack and the backend server are running, visit [`http://localhost:8888/backend/graphql-playground`](http://localhost:8888/backend/graphql-playground) in your browser.

You can run the following query to test your installation:

```graphql
query {
  quotes {
    id
    content
    author
  }
  
  quote (id: 1) {
    id
    content
    author
  }
  
  users {
    id
    name
    roles {
      id
      name
      permissions {
        code
        level
      }
    }
  }
}
```

Or create some data and check the `auditlogs` table afterwards:

```graphql
mutation {
  createQuote (input: {
    author: "Me"
    content: "Something nice"
  }) {
    id
    author
    content
  }
}
```

### Run the linter

To lint all backend code, run

```
mage -v lint:backend
```

### Run tests

To run tests, use 

```bash
# Run only unit tests
mage -v test:backend
# Run unit and integration tests
mage -v test:integration
```

### Run code generator

To rebuild the GraphQL server and all dataloaders, run  

```bash
mage -v run:generate
```


### Other mage tasks

Run `mage` without any arguments to get a list of all available tasks. These tasks are stored in the [magefile.go](./magefile.go).

### Change the configuration

The whole application is configured using the [`config.toml`](config.toml) file. Change it so it fits your needs.

### Build the docker image

Run the following command from the project's root directory:

```
docker build -t go-webapp-example -f build/docker/Dockerfile .
```

## Command line interface

You can use the following commands afters building the binary using `go build`.

### Run database migrations

Use the `migrate` command to manage the database migrations.

```bash
# Run missing migrations
./go-webapp-example migrate up
# Destroy database and start new
./go-webapp-example migrate fresh
# Show current version
./go-webapp-example migrate version
```

### Seed data

Use the `seed` command to populate the database with initial seed data.

```
./go-webapp-example seed
```

### Start the server

Use the `serve` command to start the backend server without live reloading.

```
./go-webapp-example serve
```

### Show version information

Use the `version` command to show version information.

```
./go-webapp-example version
```

