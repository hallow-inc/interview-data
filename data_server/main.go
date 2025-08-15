package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/google/uuid"
)

type Event struct {
	EventID    string         `json:"event_id"`
	UserID     *string        `json:"user_id"`
	EventType  string         `json:"event_type"`
	Source     string         `json:"source"`
	Timestamp  string         `json:"timestamp"`
	Properties map[string]any `json:"properties"`
}

type EventBatch struct {
	Events []Event `json:"events"`
}

type EventGenerator struct {
	webhookURL string
	running    bool
	mu         sync.Mutex
	httpClient *http.Client
}

var (
	eventTypes = []string{"click", "view", "purchase", "signup", "pray", "share", "like"}
	sources    = []string{"web", "apple", "android"}
	users      = make([]string, 10000)
	content    = make([]string, 1000)
)

func init() {
	for i := range 10000 {
		users[i] = fmt.Sprintf("user_%d", i+1)
	}
	for i := range 1000 {
		content[i] = fmt.Sprintf("content_%d", i+1)
	}
}

func NewEventGenerator(webhookURL string) *EventGenerator {
	return &EventGenerator{
		webhookURL: webhookURL,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

func (eg *EventGenerator) generateRandomEvent() []Event {
	event := Event{
		EventID:    uuid.New().String(),
		UserID:     &users[rand.Intn(len(users))],
		EventType:  eventTypes[rand.Intn(len(eventTypes))],
		Source:     sources[rand.Intn(len(sources))],
		Timestamp:  time.Now().Format(time.RFC3339),
		Properties: make(map[string]any),
	}

	// Add type-specific properties
	if event.EventType == "purchase" {
		event.Properties["amount"] = rand.Float64()*495 + 5
		event.Properties["product_id"] = fmt.Sprintf("prod_%d", rand.Intn(100)+1)
	}

	// Introduce data quality issues (~5% of the time)
	if rand.Float64() < 0.05 {
		if rand.Float64() < 0.5 {
			event.Timestamp = "invalid-timestamp" // Bad timestamp
		} else {
			event.UserID = nil // Missing user_id
		}
	}

	// Create duplicates (~3% of the time)
	if rand.Float64() < 0.03 {
		return []Event{event, event} // Duplicate event
	}

	return []Event{event}
}

func (eg *EventGenerator) pushEvents() {
	for {
		eg.mu.Lock()
		running := eg.running
		eg.mu.Unlock()

		if !running {
			break
		}

		// Generate 1-10 events per batch
		var allEvents []Event
		batchSize := rand.Intn(10) + 1
		for range batchSize {
			events := eg.generateRandomEvent()
			allEvents = append(allEvents, events...)
		}

		// Send webhook
		batch := EventBatch{Events: allEvents}
		jsonData, err := json.Marshal(batch)
		if err != nil {
			log.Printf("Error marshaling events: %v", err)
			continue
		}

		resp, err := eg.httpClient.Post(eg.webhookURL, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			log.Printf("Error sending webhook: %v", err)
		} else if resp.StatusCode != 200 {
			log.Printf("Webhook failed with status: %d", resp.StatusCode)
		}

		if resp != nil {
			resp.Body.Close()
		}

		// Wait 2-5 seconds between batches
		sleepDuration := time.Duration(rand.Intn(5)+2) * time.Second
		time.Sleep(sleepDuration)
	}
}

func (eg *EventGenerator) Start() {
	eg.mu.Lock()
	defer eg.mu.Unlock()

	if !eg.running {
		eg.running = true
		go eg.pushEvents()
		log.Println("Event generator started")
	}
}

func (eg *EventGenerator) Stop() {
	eg.mu.Lock()
	defer eg.mu.Unlock()

	eg.running = false
	log.Println("Event generator stopped")
}

func (eg *EventGenerator) IsRunning() bool {
	eg.mu.Lock()
	defer eg.mu.Unlock()
	return eg.running
}

func main() {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return ctx.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	app.Use(logger.New())
	app.Use(cors.New())

	webhookURL := os.Getenv("WEBHOOK_URL")
	if webhookURL == "" {
		webhookURL = "http://localhost:4000/webhook/events"
	}

	generator := NewEventGenerator(webhookURL)

	// Pull endpoints
	app.Get("/api/events", func(c *fiber.Ctx) error {
		count := min(c.QueryInt("count", 50), 200)

		var allEvents []Event
		for range count {
			events := generator.generateRandomEvent()
			allEvents = append(allEvents, events...)
		}

		// Simulate API issues
		if rand.Float64() < 0.1 { // 10% chance of error
			return c.Status(503).JSON(fiber.Map{
				"error": "Service temporarily unavailable",
			})
		}

		if rand.Float64() < 0.05 { // 5% chance of timeout
			time.Sleep(10 * time.Second)
		}

		return c.JSON(fiber.Map{
			"events":      allEvents,
			"total_count": len(allEvents),
			"has_more":    true,
		})
	})

	app.Get("/api/events/source/:source", func(c *fiber.Ctx) error {
		source := c.Params("source")
		count := min(c.QueryInt("count", 20), 100)

		var events []Event
		for range count {
			event := generator.generateRandomEvent()[0]
			event.Source = source // Override source
			events = append(events, event)
		}

		return c.JSON(fiber.Map{
			"events": events,
		})
	})

	app.Get("/api/content", func(c *fiber.Ctx) error {
		// Return all content records (1,000 total)
		return c.JSON(fiber.Map{
			"content":       content,
			"count":         len(content),
			"total_content": len(content),
		})
	})

	app.Get("/api/users", func(c *fiber.Ctx) error {
		offset := c.QueryInt("offset", 0)
		limit := min(c.QueryInt("limit", 200), 200) // Max 200 users per request

		// Validate offset
		if offset < 0 || offset >= len(users) {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid offset",
			})
		}

		// Calculate end index
		end := offset + limit
		if end > len(users) {
			end = len(users)
		}

		// Get slice of users
		userSlice := users[offset:end]

		// Create response with pagination info
		hasMore := end < len(users)
		nextOffset := -1
		if hasMore {
			nextOffset = end
		}

		return c.JSON(fiber.Map{
			"users":       userSlice,
			"count":       len(userSlice),
			"total_users": len(users),
			"offset":      offset,
			"limit":       limit,
			"has_more":    hasMore,
			"next_offset": nextOffset,
		})
	})

	// Control endpoints
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":            "healthy",
			"generator_running": generator.IsRunning(),
		})
	})

	app.Post("/start-generator", func(c *fiber.Ctx) error {
		generator.Start()
		return c.JSON(fiber.Map{
			"message": "Event generator started",
		})
	})

	app.Post("/stop-generator", func(c *fiber.Ctx) error {
		generator.Stop()
		return c.JSON(fiber.Map{
			"message": "Event generator stopped",
		})
	})

	// Auto-start generator after short delay
	go func() {
		time.Sleep(5 * time.Second)
		generator.Start()
	}()

	log.Println("Mock data service starting on :9090")
	log.Fatal(app.Listen(":9090"))
}
