package db

// Repository provides access to Postgres.
type Repository struct {
    // TODO: add connection pool
}

func New(conn string) (*Repository, error) {
    // TODO: connect to database
    return &Repository{}, nil
}
