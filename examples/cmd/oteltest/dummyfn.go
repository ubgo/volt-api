package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// dummyFunction1 makes HTTP requests to test tracing
func dummyFunction1(ctx context.Context) error {
	ctx, span := otel.Tracer("test-tracer").Start(ctx, "dummyFunction1")
	defer span.End()

	span.SetAttributes(attribute.String("function.name", "dummyFunction1"))
	span.AddEvent("Starting dummy function 1")

	// Make HTTP request to add some time
	client := &http.Client{Timeout: 10 * time.Second}

	// First HTTP call
	req1, err := http.NewRequestWithContext(ctx, "GET", "https://httpbin.org/delay/2", nil)
	if err != nil {
		span.RecordError(err)
		return err
	}

	resp1, err := client.Do(req1)
	if err != nil {
		span.RecordError(err)
		return err
	}
	defer resp1.Body.Close()

	span.AddEvent("First HTTP request completed", trace.WithAttributes(
		attribute.Int("status_code", resp1.StatusCode),
	))

	// Call nested function
	err = dummyFunction1Nested(ctx)
	if err != nil {
		span.RecordError(err)
		return err
	}

	span.AddEvent("dummyFunction1 completed successfully")
	return nil
}

// dummyFunction1Nested is a nested function to show span hierarchy
func dummyFunction1Nested(ctx context.Context) error {
	ctx, span := otel.Tracer("test-tracer").Start(ctx, "dummyFunction1Nested")
	defer span.End()

	span.SetAttributes(attribute.String("function.name", "dummyFunction1Nested"))
	span.AddEvent("Starting nested function")

	// Simulate some work
	time.Sleep(500 * time.Millisecond)

	// Make another HTTP request
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "GET", "https://httpbin.org/status/200", nil)
	if err != nil {
		span.RecordError(err)
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		span.RecordError(err)
		return err
	}
	defer resp.Body.Close()

	span.AddEvent("Nested HTTP request completed", trace.WithAttributes(
		attribute.Int("status_code", resp.StatusCode),
	))

	return nil
}

// dummyFunction2 tests database operations with tracing
func dummyFunction2(ctx context.Context) error {
	ctx, span := otel.Tracer("test-tracer").Start(ctx, "dummyFunction2")
	defer span.End()

	span.SetAttributes(
		attribute.String("function.name", "dummyFunction2"),
		attribute.String("operation.type", "database"),
	)

	span.AddEvent("Starting database operation simulation")

	// Simulate database query time
	time.Sleep(1 * time.Second)

	// Make HTTP request to external API
	clientHTTP := &http.Client{Timeout: 8 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "GET", "https://jsonplaceholder.typicode.com/posts/1", nil)
	if err != nil {
		span.RecordError(err)
		return err
	}

	resp, err := clientHTTP.Do(req)
	if err != nil {
		span.RecordError(err)
		return err
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		span.RecordError(err)
		return err
	}

	span.SetAttributes(
		attribute.Int("response.size", len(body)),
		attribute.Int("status_code", resp.StatusCode),
	)

	span.AddEvent("Database operation and HTTP request completed")
	return nil
}

// dummyFunction3 tests error scenarios and complex tracing
func dummyFunction3(ctx context.Context) error {
	ctx, span := otel.Tracer("test-tracer").Start(ctx, "dummyFunction3")
	defer span.End()

	span.SetAttributes(
		attribute.String("function.name", "dummyFunction3"),
		attribute.String("operation.type", "complex_processing"),
	)

	span.AddEvent("Starting complex processing")

	// Simulate processing time
	time.Sleep(800 * time.Millisecond)

	// Make multiple HTTP requests in parallel
	urls := []string{
		"https://httpbin.org/delay/1",
		"https://httpbin.org/delay/1",
		"https://httpbin.org/delay/1",
	}

	client := &http.Client{Timeout: 5 * time.Second}

	for i, url := range urls {
		ctx, childSpan := otel.Tracer("test-tracer").Start(ctx, fmt.Sprintf("parallel_request_%d", i))

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			childSpan.RecordError(err)
			childSpan.End()
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			childSpan.RecordError(err)
			childSpan.End()
			continue
		}

		childSpan.SetAttributes(
			attribute.String("url", url),
			attribute.Int("status_code", resp.StatusCode),
		)

		resp.Body.Close()
		childSpan.End()
	}

	// Simulate occasional error
	if time.Now().Unix()%10 == 0 {
		err := fmt.Errorf("simulated error in dummyFunction3")
		span.RecordError(err)
		span.SetAttributes(attribute.Bool("error.occurred", true))
		return err
	}

	span.AddEvent("Complex processing completed successfully")
	return nil
}

// runTracingTests executes all dummy functions to test tracing
func runTracingTests(ctx context.Context) {
	fmt.Println("runTracingTests")
	ctx, span := otel.Tracer("test-tracer").Start(ctx, "runTracingTests")
	defer span.End()

	span.SetAttributes(attribute.String("test.suite", "tracing_tests"))
	span.AddEvent("Starting tracing test suite")

	// Run function 1
	if err := dummyFunction1(ctx); err != nil {
		log.Printf("Error in dummyFunction1: %v", err)
	}

	// Run function 2
	if err := dummyFunction2(ctx); err != nil {
		log.Printf("Error in dummyFunction2: %v", err)
	}

	// Run function 3
	if err := dummyFunction3(ctx); err != nil {
		log.Printf("Error in dummyFunction3: %v", err)
	}

	span.AddEvent("Tracing test suite completed")
}
