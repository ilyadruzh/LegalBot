package db

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestRepository_SaveAndGet(t *testing.T) {
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker not installed")
	}
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

func TestRepository_SaveAndGet_Memory(t *testing.T) {
	pool, _ := pgxpool.NewWithConfig(context.Background(), &pgxpool.Config{})
	repo := &Repository{pool: pool}
	id, err := repo.SaveResult(context.Background(), 1, "data")
	if err != nil {
		t.Fatal(err)
	}
	got, err := repo.GetResult(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	if got.ChatID != 1 || got.Data != "data" {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestRepository_Delete(t *testing.T) {
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker not installed")
	}
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

	id, err := repo.SaveResult(ctx, 1, "foo")
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.DeleteResult(ctx, id); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.GetResult(ctx, id); err == nil {
		t.Fatalf("expected error after delete")
	}
}

func TestRepository_Delete_Memory(t *testing.T) {
	pool, _ := pgxpool.NewWithConfig(context.Background(), &pgxpool.Config{})
	repo := &Repository{pool: pool}
	id, err := repo.SaveResult(context.Background(), 2, "foo")
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.DeleteResult(context.Background(), id); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.GetResult(context.Background(), id); err == nil {
		t.Fatalf("expected error after delete")
	}
}

func TestRepository_RecentResults_Memory(t *testing.T) {
	pool, _ := pgxpool.NewWithConfig(context.Background(), &pgxpool.Config{})
	repo := &Repository{pool: pool}
	for i := 0; i < 3; i++ {
		if _, err := repo.SaveResult(context.Background(), 5, fmt.Sprintf("r%v", i)); err != nil {
			t.Fatal(err)
		}
	}
	res, err := repo.RecentResults(context.Background(), 5, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(res) != 2 {
		t.Fatalf("expected 2 results, got %d", len(res))
	}
	if res[0].Data == res[1].Data {
		t.Fatalf("results not ordered")
	}
}

func TestRepository_DeleteHistory_Memory(t *testing.T) {
	pool, _ := pgxpool.NewWithConfig(context.Background(), &pgxpool.Config{})
	repo := &Repository{pool: pool}
	if _, err := repo.SaveResult(context.Background(), 7, "x"); err != nil {
		t.Fatal(err)
	}
	if err := repo.DeleteHistory(context.Background(), 7); err != nil {
		t.Fatal(err)
	}
	res, err := repo.RecentResults(context.Background(), 7, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(res) != 0 {
		t.Fatalf("expected no results after delete")
	}
}

func TestRepository_WithLogger(t *testing.T) {
	pool, _ := pgxpool.NewWithConfig(context.Background(), &pgxpool.Config{})
	repo := &Repository{pool: pool}
	l := slog.New(slog.NewTextHandler(io.Discard, nil))
	WithLogger(l)(repo)
	if repo.Logger != l {
		t.Fatal("logger not set")
	}
	repo.Close()
}

func TestRepository_New_Error(t *testing.T) {
	os.Unsetenv("POSTGRES_DSN")
	if _, err := New(context.Background()); err == nil {
		t.Fatal("expected error")
	}
}
