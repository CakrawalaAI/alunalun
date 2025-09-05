package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mmcloughlin/geohash"

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
		userIDs[i] = uuidToString(createdUser.ID)
	}
	log.Printf("✅ Created %d users", len(users))

	// Create auth providers for users
	log.Println("🔐 Creating auth providers...")
	for i, userID := range userIDs {
		provider := "google"
		if i%3 == 1 {
			provider = "github"
		}
		
		providerID := pgtype.UUID{}
		providerID.Scan(generateUUID())
		userUUID := pgtype.UUID{}
		userUUID.Scan(userID)
		_, err := txQueries.CreateUserAuthProvider(ctx, &repository.CreateUserAuthProviderParams{
			ID:               providerID,
			UserID:           userUUID,
			Provider:         provider,
			ProviderUserID:   fmt.Sprintf("%s-%s-%d", provider, userID[:8], i),
			ProviderMetadata: []byte(fmt.Sprintf(`{"source":"dev_seed","provider":"%s"}`, provider)),
			CreatedAt:        timestamptz(time.Now()),
		})
		if err != nil {
			return fmt.Errorf("failed to create auth provider for user %s: %w", userID, err)
		}
	}
	log.Printf("✅ Created auth providers for %d users", len(userIDs))

	// Create posts and locations
	log.Println("📍 Creating posts with locations...")
	posts := createPosts(userIDs)
	pinCount := 0
	commentCount := 0
	
	for _, post := range posts {
		// Create post
		createdPost, err := txQueries.CreatePost(ctx, post.CreatePostParams)
		if err != nil {
			return fmt.Errorf("failed to create post: %w", err)
		}

		// Create location if this is a pin
		if post.Type == "pin" && post.Location != nil {
			_, err := txQueries.CreatePostLocation(ctx, &repository.CreatePostLocationParams{
				PostID:        createdPost.ID,
				StMakepoint:   post.Location.Lng, // longitude first for PostGIS
				StMakepoint_2: post.Location.Lat,
				Geohash:     post.Location.Geohash,
				CreatedAt:   post.CreatedAt,
			})
			if err != nil {
				return fmt.Errorf("failed to create location for post: %w", err)
			}
			pinCount++
		} else if post.Type == "comment" {
			commentCount++
		}
	}
	log.Printf("✅ Created %d pins and %d comments with locations", pinCount, commentCount)

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
		display  string
	}{
		{"jakarta_explorer", "explorer@jakarta.dev", "Jakarta Explorer"},
		{"monas_wanderer", "wanderer@jakarta.dev", "Monas Wanderer"},
		{"kota_tua_lover", "kotatua@jakarta.dev", "Kota Tua Lover"},
		{"ancol_vibes", "ancol@jakarta.dev", "Ancol Vibes"},
		{"thamrin_walker", "thamrin@jakarta.dev", "Thamrin Walker"},
		{"sudirman_jogger", "sudirman@jakarta.dev", "Sudirman Jogger"},
		{"kemang_foodie", "kemang@jakarta.dev", "Kemang Foodie"},
		{"menteng_resident", "menteng@jakarta.dev", "Menteng Resident"},
		{"pondok_indah", "pi@jakarta.dev", "Pondok Indah"},
		{"senayan_gym", "senayan@jakarta.dev", "Senayan Gym"},
		{"kelapa_gading", "kg@jakarta.dev", "Kelapa Gading"},
		{"bsd_commuter", "bsd@jakarta.dev", "BSD Commuter"},
		{"cipete_cafe", "cipete@jakarta.dev", "Cipete Cafe"},
		{"radio_dalam", "radal@jakarta.dev", "Radio Dalam"},
		{"blok_m_shopper", "blokm@jakarta.dev", "Blok M Shopper"},
		{"fatmawati_local", "fatmawati@jakarta.dev", "Fatmawati Local"},
		{"tebet_native", "tebet@jakarta.dev", "Tebet Native"},
		{"kuningan_worker", "kuningan@jakarta.dev", "Kuningan Worker"},
		{"setiabudi_one", "setiabudi@jakarta.dev", "Setiabudi One"},
		{"gandaria_city", "gandaria@jakarta.dev", "Gandaria City"},
	}

	result := make([]*repository.CreateUserParams, len(users))
	now := time.Now()
	
	for i, user := range users {
		id := pgtype.UUID{}
		id.Scan(generateUUID())
		createdAt := now.Add(-time.Duration(i*2) * time.Hour) // Spread creation times
		avatarURL := fmt.Sprintf("https://api.dicebear.com/7.x/initials/svg?seed=%s", user.username)
		
		result[i] = &repository.CreateUserParams{
			ID:          id,
			Username:    user.username,
			Email:       user.email,
			DisplayName: stringPtr(user.display),
			AvatarUrl:   stringPtr(avatarURL),
			CreatedAt:   timestamptz(createdAt),
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
		userID := pgtype.UUID{}
		userID.Scan(userIDs[rand.Intn(len(userIDs))])
		content := pinContents[rand.Intn(len(pinContents))]
		
		// Random time within the last 30 days
		randomHours := rand.Intn(30 * 24)
		createdAt := baseTime.Add(time.Duration(randomHours) * time.Hour)
		
		// Random location around Monas
		location := generateLocationAroundMonas()
		
		postID := pgtype.UUID{}
		postID.Scan(generateUUID())
		
		post := PostWithLocation{
			CreatePostParams: &repository.CreatePostParams{
				ID:         postID,
				UserID:     userID,
				Type:       "pin",
				Content:    content,
				ParentID:   pgtype.UUID{}, // empty UUID for pins
				Visibility: stringPtr("public"),
				Metadata:   []byte(`{"source":"dev_seed"}`),
				CreatedAt:  timestamptz(createdAt),
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
		userID := pgtype.UUID{}
		userID.Scan(userIDs[rand.Intn(len(userIDs))])
		content := commentContents[rand.Intn(len(commentContents))]
		
		// Comment created after the parent post
		commentTime := parentPost.CreatedAt.Time.Add(time.Duration(rand.Intn(48)) * time.Hour)
		parentID := parentPost.ID
		
		commentID := pgtype.UUID{}
		commentID.Scan(generateUUID())
		
		comment := PostWithLocation{
			CreatePostParams: &repository.CreatePostParams{
				ID:         commentID,
				UserID:     userID,
				Type:       "comment",
				Content:    content,
				ParentID:   parentID,
				Visibility: stringPtr("public"),
				Metadata:   []byte(`{"source":"dev_seed"}`),
				CreatedAt:  timestamptz(commentTime),
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
	latOffset := (distance * math.Cos(angle)) / 111.0
	lngOffset := (distance * math.Sin(angle)) / (111.0 * math.Cos(monasLat*math.Pi/180))
	
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

func generateUUID() string {
	// Generate a UUID v4
	b := make([]byte, 16)
	rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40 // Version 4
	b[8] = (b[8] & 0x3f) | 0x80 // Variant 10
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func uuidToString(u pgtype.UUID) string {
	return fmt.Sprintf("%x-%x-%x-%x-%x", u.Bytes[0:4], u.Bytes[4:6], u.Bytes[6:8], u.Bytes[8:10], u.Bytes[10:16])
}