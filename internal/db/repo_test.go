package db

import (
	"context"
	"fmt"
	"os"
	"testing"

	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestRepository_SaveAndGet(t *testing.T) {
	ctx := context.Background()
	container, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: tc.ContainerRequest{
			Image:        "postgres:16",
			Env:          map[string]string{"POSTGRES_PASSWORD": "pass"},
			ExposedPorts: []string{"5432/tcp"},
			WaitingFor:   wait.ForListeningPort("5432/tcp"),
		},
		Started: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer container.Terminate(ctx)

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatal(err)
	}
	port, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		t.Fatal(err)
	}

	os.Setenv("POSTGRES_DSN", fmt.Sprintf("postgres://postgres:pass@%s:%s/postgres?sslmode=disable", host, port.Port()))

	repo, err := New(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer repo.Close()

	if _, err := repo.pool.Exec(ctx, `CREATE TABLE bot_results (id bigserial primary key, chat_id bigint, data text, created_at timestamptz default now())`); err != nil {
		t.Fatal(err)
	}

	id, err := repo.SaveResult(ctx, 123, "hi")
	if err != nil {
		t.Fatal(err)
	}
	got, err := repo.GetResult(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if got.ChatID != 123 || got.Data != "hi" {
		t.Fatalf("unexpected result: %+v", got)
	}
}
