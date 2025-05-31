package testcontainers

import "context"

type GenericContainerRequest struct {
	ContainerRequest ContainerRequest
	Started          bool
}

type ContainerRequest struct {
	Image        string
	Env          map[string]string
	ExposedPorts []string
	WaitingFor   interface{}
}

type Container interface {
	Host(ctx context.Context) (string, error)
	MappedPort(ctx context.Context, port string) (natPort, error)
	Terminate(ctx context.Context) error
}

type natPort struct{ port string }

func (n natPort) Port() string { return n.port }

func GenericContainer(ctx context.Context, req GenericContainerRequest) (Container, error) {
	return stubContainer{}, nil
}

type stubContainer struct{}

func (stubContainer) Host(ctx context.Context) (string, error) { return "localhost", nil }
func (stubContainer) MappedPort(ctx context.Context, p string) (natPort, error) {
	return natPort{port: "5432"}, nil
}
func (stubContainer) Terminate(ctx context.Context) error { return nil }
