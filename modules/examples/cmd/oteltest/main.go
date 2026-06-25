package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

func main() {

	// Handle SIGINT (CTRL+C) gracefully.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Set up OpenTelemetry.
	otelShutdown, err := setupOTelSDK(ctx)
	if err != nil {
		return
	}
	// Handle shutdown properly so nothing leaks.
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	mainMeter := otel.GetMeterProvider().Meter("localtest/other/myTestMeter")

	mainCounter, _ := mainMeter.Int64Counter(
		"temporary.call",
		metric.WithUnit("{call}"),
		metric.WithDescription("car counter"),
	)
	mainCounter.Add(context.Background(), 1, metric.WithAttributes(
		attribute.String("model", "Camaro"),
		attribute.Bool("IsCar", true),
	))

	// _, span := otel.Tracer("khanakiatrace").Start(ctx, "test")
	// span.AddEvent("test event")
	// span.End()
	// time.Sleep(time.Second * 10)

	// Run tracing tests
	// runTracingTests(ctx)

	// // Run tests in a loop for better visualization
	// go func() {
	// 	for i := 0; i < 5; i++ {
	// 		time.Sleep(2 * time.Second)
	// 		log.Printf("Running tracing test iteration %d", i+1)
	// 		runTracingTests(ctx)
	// 	}
	// }()

	logger := otelslog.NewLogger("test")

	// You can use the logger directly anywhere in your app now
	logger.Debug("Something interesting happened")

	fmt.Println("Press Ctrl+C to exit...")
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Received interrupt signal, shutting down...")
			return
		default:
			time.Sleep(time.Second)
		}
	}

}
