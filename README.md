# rufs-base-go
Restful Utilities for Full Stack - Base Modules and Microservices to Golang

Offer base package to rufs compliance microservices

You need Golang installed and PostgreSql server already running with your database.

## First Step

Open terminal and clone this repository with `git clone https://github.com/alexsandrostefenon/rufs-base-go`.

Requires Golang version >= 1.18

## Run Ecosystem

## PostgreSql setup

create database :

sudo su postgres;

or

su -c "su postgres";

export PGDATABASE=postgres;
psql -c "CREATE USER development WITH CREATEDB LOGIN PASSWORD '123456'";
psql -c 'CREATE DATABASE rufs_base_development WITH OWNER development';
exit;

Note, database "rufs_base_development" is only for testing purposes.

### Run Ecosystem

#Only to clean already existent configuration :

rm *openapi-base.json;

#Only to clean already existent testing data :

export PGHOST=localhost;
export PGPORT=5432;
export PGUSER=development;
export PGPASSWORD=123456;

psql rufs_base_development -c "DROP DATABASE IF EXISTS rufs_base;" &&
psql rufs_base_development -c "CREATE DATABASE rufs_base;" &&

#Execute rufs-proxy to load and start microservices :

PGHOST=localhost PGPORT=5432 PGUSER=development PGPASSWORD=123456 PGDATABASE=rufs_base go test -timeout 3600s -run ^TestExternal$ ./rufs-base-go

## Web application

In EcmaScript2017 compliance browser open url

`
curl -X 'GET' http://localhost:8080/rest/login -d '{}' -H 'Connection: close' -H 'content-type: undefined';
`

For custom service configuration or user edition, use user 'admin' with password 'admin'.
rufs-base-es6/README.