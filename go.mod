module legalbot

go 1.23.8

require github.com/jackc/pgx/v5 v5.5.0
require github.com/testcontainers/testcontainers-go v0.0.0

replace github.com/jackc/pgx/v5 => ./stub/pgxv5
replace github.com/testcontainers/testcontainers-go => ./stub/testcontainers-go

