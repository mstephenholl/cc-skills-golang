# Testing with samber/do

## Container Cloning

Clone containers for isolated tests:

```go
func TestUserService(t *testing.T) {
    // Clone the container built by the same wiring as main() — the original stays untouched
    testInjector := newAppInjector().Clone()

    // Override with mocks — name the interface explicitly: without [Database], T is inferred
    // as *MockDatabase and registers a new service instead of replacing Database
    mockDB := &MockDatabase{}
    do.OverrideValue[Database](testInjector, mockDB)

    // Test with mocked dependencies
    service := do.MustInvoke[UserService](testInjector)
    // ... test code
}
```

## Reusable Test Helpers

```go
func SetupTestContainer(t *testing.T) do.Injector {
    injector := do.New()

    do.Provide(injector, func(i do.Injector) (Database, error) {
        return &MockDatabase{}, nil
    })

    return injector
}
```

## Quick Reference

### Testing & Overrides

| Function                         | Purpose                         |
| -------------------------------- | ------------------------------- |
| `injector.Clone()`               | Clone container for testing     |
| `injector.CloneWithOpts()`       | Clone with custom options       |
| `do.Override[T]()`               | Replace service (use in tests)  |
| `do.OverrideNamed[T]()`          | Replace named service           |
| `do.OverrideValue[T]()`          | Replace value service           |
| `do.OverrideNamedValue[T]()`     | Replace named value             |
| `do.OverrideTransient[T]()`      | Replace transient factory       |
| `do.OverrideNamedTransient[T]()` | Replace named transient factory |
