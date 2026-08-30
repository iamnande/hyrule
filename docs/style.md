# go style

for everything general, defer to the
[Google Go Style Guide](https://google.github.io/styleguide/go/) - don't
re-litigate it here. this doc is only for the one rule we're treating as
harder than the guide does.

## interfaces: consumer defines, producer returns concrete types

the guide already says this
([decisions#interfaces](https://google.github.io/styleguide/go/decisions#interfaces));
here it's not a default, it's a lint-me-eventually hard rule (see
docs/conventions.md#known-gaps).

```go
// bad - go/internal/svc/pings/repository/pings.go defining its own interface
// next to the type that implements it. every caller pays for methods it
// doesn't use, and "mock the interface" means implementing all of them.
package repository

type PingRepository interface {
	Create(ctx context.Context, p Ping) error
	Get(ctx context.Context, id string) (Ping, error)
	List(ctx context.Context, filter Filter) ([]Ping, error)
	Delete(ctx context.Context, id string) error
}

type pingRepository struct{ pool *pgxpool.Pool }

func (r *pingRepository) Create(...) error { ... }
// ...
```

```go
// good - repository/pings.go exports a concrete type, nothing else.
package repository

type PingRepository struct{ pool *pgxpool.Pool }

func (r *PingRepository) Create(ctx context.Context, p Ping) error { ... }
func (r *PingRepository) Get(ctx context.Context, id string) (Ping, error) { ... }

// domain/pings.go declares exactly what it calls - one method, its own
// interface, sized to this one caller.
package domain

type pingCreator interface {
	Create(ctx context.Context, p repository.Ping) error
}

func NewService(repo pingCreator) *Service { ... }
```

a PR adding an interface in the same package as its implementation should
be read as a mistake unless the type wraps a genuinely external, swappable
dependency.
