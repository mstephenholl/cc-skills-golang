# Manual Constructor Injection

## Complete Application Example

```go
func main() {
    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
    defer stop()

    // Layer 1: Configuration
    cfg := LoadConfig()
    logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

    // Layer 2: Infrastructure
    db, err := postgres.Connect(cfg.DatabaseURL)
    if err != nil {
        logger.Error("database connection failed", "error", err)
        os.Exit(1)
    }
    defer db.Close()

    cache := redis.NewClient(cfg.RedisURL)
    defer cache.Close()

    mailer := smtp.NewMailer(cfg.SMTPAddr)

    // Layer 3: Repositories
    userRepo := postgres.NewUserRepository(db)
    orderRepo := postgres.NewOrderRepository(db)

    // Layer 4: Services
    userSvc := service.NewUserService(userRepo, cache, mailer, logger)
    orderSvc := service.NewOrderService(orderRepo, userSvc, logger)
    paymentSvc := service.NewPaymentService(orderRepo, cfg.StripeKey, logger)

    // Layer 5: Transport
    handler := http.NewHandler(userSvc, orderSvc, paymentSvc, logger)
    server := http.NewServer(cfg.Port, handler)

    // Run
    go server.ListenAndServe()
    <-ctx.Done()
    server.Shutdown(context.Background())
}
```

Initialize in this order — infrastructure first, then repositories, then services, then transport — so each constructor receives dependencies that already exist; `defer` each `Close` right after its constructor so shutdown runs in reverse order.
