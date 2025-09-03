package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/netip"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mmcloughlin/geohash"
	"github.com/segmentio/ksuid"

	"github.com/radjathaher/alunalun/api/internal/config"
	"github.com/radjathaher/alunalun/api/internal/repository"
)

const (
	// Monas coordinates (National Monument, Jakarta)
	monasLat = -6.1754
	monasLng = 106.8272
	// 2km radius around Monas for realistic spread
	radiusKm = 2.0
)

var (
	yesFlag = flag.Bool("yes", false, "Skip confirmation prompts")
)

func main() {
	flag.Parse()
	
	log.Println("🌱 Seeding Development Database")
	log.Println("==============================")

	// Load configuration
	cfg := config.Load()
	
	// Build dev database URL
	devDBName := "dev_alunalun"
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.DB.User, cfg.DB.Password, cfg.DB.Host, cfg.DB.Port, devDBName, cfg.DB.SSLMode)

	// Show target database
	log.Printf("🎯 Target Database: %s", dbURL)
	
	// Confirmation prompt (unless -yes flag)
	if !*yesFlag {
		if !confirmAction() {
			log.Println("❌ Operation cancelled")
			os.Exit(0)
		}
	}

	// Seed data
	if err := seedData(dbURL); err != nil {
		log.Fatalf("❌ Seeding failed: %v", err)
	}

	log.Println("✅ Database seeding completed successfully!")
}

func confirmAction() bool {
	fmt.Print("📝 Proceed with database seeding? (Y/n): ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	response := strings.ToLower(strings.TrimSpace(scanner.Text()))
	return response == "" || response == "y" || response == "yes"
}

func seedData(dbURL string) error {
	// Connect using pgxpool for better performance
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		return fmt.Errorf("failed to create pool: %w", err)
	}
	defer pool.Close()

	queries := repository.New(pool)
	ctx := context.Background()

	// Check if data already exists (idempotent)
	userCount, err := queries.CountUsers(ctx)
	if err != nil {
		return fmt.Errorf("failed to check existing data: %w", err)
	}

	if userCount > 0 {
		log.Printf("📄 Database already has %d users - skipping seeding", userCount)
		return nil
	}

	// Begin transaction
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	txQueries := queries.WithTx(tx)

	// Create users
	log.Println("👥 Creating users...")
	users := createUsers()
	userIDs := make([]string, len(users))
	
	for i, user := range users {
		createdUser, err := txQueries.CreateUser(ctx, user)
		if err != nil {
			return fmt.Errorf("failed to create user %s: %w", user.Username, err)
		}
		userIDs[i] = createdUser.ID
	}
	log.Printf("✅ Created %d users", len(users))

	// Create posts and locations
	log.Println("📍 Creating posts with locations...")
	posts := createPosts(userIDs)
	
	for _, post := range posts {
		// Create post
		createdPost, err := txQueries.CreatePost(ctx, post.CreatePostParams)
		if err != nil {
			return fmt.Errorf("failed to create post %s: %w", post.ID, err)
		}

		// Create location if this is a pin
		if post.Type == "pin" && post.Location != nil {
			_, err := txQueries.CreatePostLocation(ctx, &repository.CreatePostLocationParams{
				PostID:        createdPost.ID,
				StMakepoint:   post.Location.Lng,
				StMakepoint_2: post.Location.Lat,
				Geohash:       &post.Location.Geohash,
				CreatedAt:     post.CreatedAt,
			})
			if err != nil {
				return fmt.Errorf("failed to create location for post %s: %w", post.ID, err)
			}
		}

		// Create user event
		eventID := ksuid.New().String()
		eventType := "create_pin"
		if post.Type == "comment" {
			eventType = "create_comment"
		}
		
		_, err = txQueries.CreateUserEvent(ctx, &repository.CreateUserEventParams{
			ID:          eventID,
			UserID:      &post.UserID,
			SessionID:   ksuid.New().String(),
			EventType:   eventType,
			IpAddress:   mustParseIP("127.0.0.1"),
			Fingerprint: stringPtr("dev_seed"),
			CreatedAt:   post.CreatedAt,
		})
		if err != nil {
			return fmt.Errorf("failed to create user event for post %s: %w", post.ID, err)
		}
	}
	log.Printf("✅ Created %d posts with locations and events", len(posts))

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

type PostWithLocation struct {
	*repository.CreatePostParams
	Location *Location
}

type Location struct {
	Lat     float64
	Lng     float64
	Geohash string
}

func createUsers() []*repository.CreateUserParams {
	// Indonesian names and Jakarta-themed usernames
	users := []struct {
		username string
		email    string
	}{
		{"jakarta_explorer", "explorer@jakarta.dev"},
		{"monas_wanderer", "wanderer@jakarta.dev"},
		{"kota_tua_lover", "kotatua@jakarta.dev"},
		{"ancol_vibes", "ancol@jakarta.dev"},
		{"thamrin_walker", "thamrin@jakarta.dev"},
		{"sudirman_jogger", "sudirman@jakarta.dev"},
		{"kemang_foodie", "kemang@jakarta.dev"},
		{"menteng_resident", "menteng@jakarta.dev"},
		{"pondok_indah", "pi@jakarta.dev"},
		{"senayan_gym", "senayan@jakarta.dev"},
		{"kelapa_gading", "kg@jakarta.dev"},
		{"bsd_commuter", "bsd@jakarta.dev"},
		{"cipete_cafe", "cipete@jakarta.dev"},
		{"radio_dalam", "radal@jakarta.dev"},
		{"blok_m_shopper", "blokm@jakarta.dev"},
		{"fatmawati_local", "fatmawati@jakarta.dev"},
		{"tebet_native", "tebet@jakarta.dev"},
		{"kuningan_worker", "kuningan@jakarta.dev"},
		{"setiabudi_one", "setiabudi@jakarta.dev"},
		{"gandaria_city", "gandaria@jakarta.dev"},
	}

	result := make([]*repository.CreateUserParams, len(users))
	now := time.Now()
	
	for i, user := range users {
		id := ksuid.New().String()
		createdAt := now.Add(-time.Duration(i*2) * time.Hour) // Spread creation times
		
		result[i] = &repository.CreateUserParams{
			ID:        id,
			Email:     &user.email,
			Username:  user.username,
			Metadata:  []byte(`{"source":"dev_seed"}`),
			CreatedAt: timestamptz(createdAt),
			UpdatedAt: timestamptz(createdAt),
		}
	}
	
	return result
}

func createPosts(userIDs []string) []PostWithLocation {
	// Jakarta-themed content around Monas
	pinContents := []string{
		"Amazing view of Monas from here! 🏛️",
		"Great coffee spot near National Monument ☕",
		"Street food heaven in this area! 🍜",
		"Beautiful sunset behind Monas today 🌅",
		"Historic building with interesting architecture 🏢",
		"Perfect spot for morning jog around Merdeka Square 🏃",
		"Local warung with authentic Jakarta flavors 🍛",
		"Green space in the middle of busy Jakarta 🌳",
		"Traditional market with fresh produce 🥬",
		"Modern mall contrasting with old Jakarta 🏪",
		"Friendly locals always ready to help 👥",
		"Jakarta's traffic but worth the view! 🚗",
		"Cultural event happening here today 🎭",
		"Religious site with peaceful atmosphere 🕌",
		"Educational museum about Indonesian history 📚",
		"Art installation in urban setting 🎨",
		"Public transport hub connecting the city 🚇",
		"River view in central Jakarta 🌊",
		"Night market with delicious street snacks 🌙",
		"Modern office district during rush hour 🏙️",
	}

	commentContents := []string{
		"Totally agree! Love this place too",
		"Thanks for the recommendation 👍",
		"I was here yesterday, amazing!",
		"Great photo! What time did you take this?",
		"This brings back memories of my childhood",
		"Anyone know the opening hours?",
		"Perfect for weekend family trips",
		"The food here is incredible!",
		"Jakarta keeps surprising me ❤️",
		"Classic Jakarta experience right here",
	}

	var posts []PostWithLocation
	now := time.Now()
	baseTime := now.Add(-30 * 24 * time.Hour) // Start from 30 days ago
	
	// Create ~200 posts (180 pins + ~20 comments)
	for i := 0; i < 180; i++ {
		userID := userIDs[rand.Intn(len(userIDs))]
		content := pinContents[rand.Intn(len(pinContents))]
		
		// Random time within the last 30 days
		randomHours := rand.Intn(30 * 24)
		createdAt := baseTime.Add(time.Duration(randomHours) * time.Hour)
		
		// Random location around Monas
		location := generateLocationAroundMonas()
		
		post := PostWithLocation{
			CreatePostParams: &repository.CreatePostParams{
				ID:        ksuid.New().String(),
				UserID:    userID,
				Type:      "pin",
				Content:   &content,
				ParentID:  nil,
				Metadata:  []byte(`{"source":"dev_seed"}`),
				CreatedAt: timestamptz(createdAt),
				UpdatedAt: timestamptz(createdAt),
			},
			Location: &location,
		}
		
		posts = append(posts, post)
	}

	// Add some comments to random pins (about 1 in 9 pins gets a comment)
	for i := 0; i < len(posts); i += 9 {
		if i >= len(posts) {
			break
		}
		
		parentPost := posts[i]
		userID := userIDs[rand.Intn(len(userIDs))]
		content := commentContents[rand.Intn(len(commentContents))]
		
		// Comment created after the parent post
		commentTime := parentPost.CreatedAt.Time.Add(time.Duration(rand.Intn(48)) * time.Hour)
		
		comment := PostWithLocation{
			CreatePostParams: &repository.CreatePostParams{
				ID:        ksuid.New().String(),
				UserID:    userID,
				Type:      "comment",
				Content:   &content,
				ParentID:  &parentPost.ID,
				Metadata:  []byte(`{"source":"dev_seed"}`),
				CreatedAt: timestamptz(commentTime),
				UpdatedAt: timestamptz(commentTime),
			},
			Location: nil, // Comments don't have locations
		}
		
		posts = append(posts, comment)
	}

	return posts
}

func generateLocationAroundMonas() Location {
	// Generate random point within 2km radius of Monas
	// Using simple random distribution (not perfectly uniform, but good enough for dev data)
	
	angle := rand.Float64() * 2 * math.Pi // Random angle in radians
	distance := rand.Float64() * radiusKm // Random distance in km
	
	// Convert to lat/lng offset (rough approximation for Jakarta)
	// 1 degree lat ≈ 111km, 1 degree lng ≈ 111km * cos(lat)
	latOffset := (distance * 1000) / 111000 * 57.2958 * math.Cos(angle) // Convert to degrees
	lngOffset := (distance * 1000) / (111000 * math.Cos(monasLat*0.0174533)) * math.Sin(angle)
	
	lat := monasLat + latOffset
	lng := monasLng + lngOffset
	
	return Location{
		Lat:     lat,
		Lng:     lng,
		Geohash: geohash.Encode(lat, lng),
	}
}

// Helper functions
func timestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func stringPtr(s string) *string {
	return &s
}

func mustParseIP(ip string) netip.Addr {
	addr, err := netip.ParseAddr(ip)
	if err != nil {
		panic(fmt.Sprintf("invalid IP address %s: %v", ip, err))
	}
	return addr
}