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
	UserID     *User          `json:"user_id"`
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

type User struct {
	ID        string    `json:"user_id"`
	Age       int       `json:"age"`
	Status    string    `json:"status"`
	Country   string    `json:"country"`
	CreatedAt time.Time `json:"created_at"`
}

type Content struct {
	ID         string    `json:"content_id"`
	MediaType  string    `json:"media_type"`
	PrayerType string    `json:"prayer_type"`
	CreatedAt  time.Time `json:"created_at"`
}

var (
	eventTypes = []string{"click", "view", "purchase", "signup", "pray", "share", "like"}
	sources    = []string{"web", "apple", "android"}
	users      = make([]User, 10000)
	content    = make([]Content, 1000)

	countries   = []string{"US", "BR", "IT", "FR"}
	statuses    = []string{"free", "paid", "trial"}
	mediaTypes  = []string{"video", "audio", "text"}
	prayerTypes = []string{"academic", "podcast", "reflection", "lectio_divina", "rosary", "meditation"}
)

func randomTimestampInYear(year int) time.Time {
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)

	daysInYear := 365
	if (year%4 == 0 && year%100 != 0) || (year%400 == 0) {
		daysInYear = 366
	}
	days := rand.Intn(daysInYear)
	return start.AddDate(0, 0, days)
}

func init() {
	for i := range 10000 {
		user := User{
			ID:        fmt.Sprintf("user_%d", i+1),
			Age:       rand.Intn(65) + 18,
			Status:    statuses[rand.Intn(len(statuses))],
			Country:   countries[rand.Intn(len(countries))],
			CreatedAt: randomTimestampInYear(2023),
		}
		users[i] = user
	}
	for i := range 1000 {
		asset := Content{
			ID:         fmt.Sprintf("content_%d", i+1),
			MediaType:  mediaTypes[rand.Intn(len(mediaTypes))],
			PrayerType: prayerTypes[rand.Intn(len(prayerTypes))],
			CreatedAt:  randomTimestampInYear(2024),
		}
		content[i] = asset
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

	if event.EventType == "purchase" {
		event.Properties["amount"] = rand.Float64()*495 + 5
		event.Properties["product_id"] = fmt.Sprintf("prod_%d", rand.Intn(100)+1)
	}

	if event.EventType == "view" || event.EventType == "like" || event.EventType == "share" {
		randomContent := content[rand.Intn(len(content))]
		event.Properties["content_id"] = randomContent.ID
		event.Properties["media_type"] = randomContent.MediaType
		event.Properties["prayer_type"] = randomContent.PrayerType
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

	app.Get("/api/events", func(c *fiber.Ctx) error {
		count := min(c.QueryInt("count", 50), 200)

		var allEvents []Event
		for range count {
			events := generator.generateRandomEvent()
			allEvents = append(allEvents, events...)
		}

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
			event.Source = source
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
		end := min(offset+limit, len(users))

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
