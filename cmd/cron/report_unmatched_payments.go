package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"

	"github.com/joho/godotenv"
	_ "modernc.org/sqlite"

	"github.com/base48/member-portal/internal/config"
	"github.com/base48/member-portal/internal/db"
)

// Report payments that have a variable symbol but are not matched to any user
//
// Usage:
//   go run cmd/cron/report_unmatched_payments.go

type UnmatchedPayment struct {
	Payment       db.Payment
	VSAsUserID    int64
	UserExists    bool
	Reason        string
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	database, err := sql.Open("sqlite", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	queries := db.New(database)
	ctx := context.Background()

	// Get all unassigned payments
	unassignedPayments, err := queries.ListUnassignedPayments(ctx)
	if err != nil {
		log.Fatalf("Failed to list unassigned payments: %v", err)
	}

	log.Printf("Analyzing %d unassigned payments...\n", len(unassignedPayments))

	var problematic []UnmatchedPayment
	totalAmount := 0.0

	for _, payment := range unassignedPayments {
		// Skip if no identification (VS)
		if payment.Identification == "" {
			problematic = append(problematic, UnmatchedPayment{
				Payment:    payment,
				Reason:     "Empty variable symbol",
			})

			// Parse amount
			if amount, err := strconv.ParseFloat(payment.Amount, 64); err == nil {
				totalAmount += amount
			}
			continue
		}

		// Try to parse VS as user ID
		vsAsID, err := strconv.ParseInt(payment.Identification, 10, 64)
		if err != nil {
			problematic = append(problematic, UnmatchedPayment{
				Payment:    payment,
				Reason:     fmt.Sprintf("VS '%s' is not a valid user ID", payment.Identification),
			})

			if amount, err := strconv.ParseFloat(payment.Amount, 64); err == nil {
				totalAmount += amount
			}
			continue
		}

		// Check if user with this ID exists
		_, err = queries.GetUserByID(ctx, vsAsID)
		if err == sql.ErrNoRows {
			problematic = append(problematic, UnmatchedPayment{
				Payment:    payment,
				VSAsUserID: vsAsID,
				UserExists: false,
				Reason:     fmt.Sprintf("User ID %d does not exist", vsAsID),
			})

			if amount, err := strconv.ParseFloat(payment.Amount, 64); err == nil {
				totalAmount += amount
			}
			continue
		} else if err != nil {
			log.Printf("⚠ Error checking user %d: %v", vsAsID, err)
			continue
		}

		// If we get here, user exists but payment is not assigned - this is suspicious
		problematic = append(problematic, UnmatchedPayment{
			Payment:    payment,
			VSAsUserID: vsAsID,
			UserExists: true,
			Reason:     fmt.Sprintf("User ID %d EXISTS but payment not assigned (sync bug?)", vsAsID),
		})

		if amount, err := strconv.ParseFloat(payment.Amount, 64); err == nil {
			totalAmount += amount
		}
	}

	// Print report
	fmt.Println("\n" + repeat("=", 120))
	fmt.Println("UNMATCHED PAYMENTS REPORT")
	fmt.Println(repeat("=", 120))
	fmt.Printf("\nTotal unassigned payments: %d\n", len(unassignedPayments))
	fmt.Printf("Problematic payments: %d\n", len(problematic))
	fmt.Printf("Total unmatched amount: %.2f CZK\n\n", totalAmount)

	if len(problematic) == 0 {
		fmt.Println("✓ No problematic payments found!")
		return
	}

	// Group by reason
	emptyVS := []UnmatchedPayment{}
	invalidVS := []UnmatchedPayment{}
	userNotFound := []UnmatchedPayment{}
	syncBug := []UnmatchedPayment{}

	for _, p := range problematic {
		if p.Reason == "Empty variable symbol" {
			emptyVS = append(emptyVS, p)
		} else if p.Reason[:2] == "VS" {
			invalidVS = append(invalidVS, p)
		} else if !p.UserExists {
			userNotFound = append(userNotFound, p)
		} else {
			syncBug = append(syncBug, p)
		}
	}

	// Print each category
	if len(emptyVS) > 0 {
		fmt.Println("\n📝 PAYMENTS WITH EMPTY VARIABLE SYMBOL:")
		fmt.Println(repeat("-", 120))
		printPaymentTable(emptyVS)
	}

	if len(invalidVS) > 0 {
		fmt.Println("\n⚠ PAYMENTS WITH INVALID VARIABLE SYMBOL (not a number):")
		fmt.Println(repeat("-", 120))
		printPaymentTable(invalidVS)
	}

	if len(userNotFound) > 0 {
		fmt.Println("\n❌ PAYMENTS WITH VS POINTING TO NON-EXISTENT USER:")
		fmt.Println(repeat("-", 120))
		printPaymentTable(userNotFound)
	}

	if len(syncBug) > 0 {
		fmt.Println("\n🐛 PAYMENTS WITH VALID USER BUT NOT ASSIGNED (POTENTIAL SYNC BUG):")
		fmt.Println(repeat("-", 120))
		printPaymentTable(syncBug)
		fmt.Println("\n⚠️  These payments should be automatically assigned! Run sync again or investigate.")
	}

	fmt.Println("\n" + repeat("=", 120))
	fmt.Println("\n💡 Next steps:")
	fmt.Println("  1. For payments with valid VS but non-existent users: Check if user should be imported")
	fmt.Println("  2. For payments with empty/invalid VS: Manually assign via admin interface")
	fmt.Println("  3. For sync bugs: Re-run FIO sync or investigate the matching logic")
	fmt.Println()
}

func printPaymentTable(payments []UnmatchedPayment) {
	fmt.Printf("%-8s %-12s %-12s %-12s %-30s %-10s %s\n",
		"ID", "Date", "Amount", "Kind", "Remote Account", "VS", "Reason")
	fmt.Println(repeat("-", 120))

	for _, p := range payments {
		dateStr := p.Payment.Date.Format("2006-01-02")
		remoteAcc := p.Payment.RemoteAccount
		if len(remoteAcc) > 28 {
			remoteAcc = remoteAcc[:28] + ".."
		}

		reason := p.Reason
		if len(reason) > 40 {
			reason = reason[:40] + "..."
		}

		fmt.Printf("%-8d %-12s %10s CZK %-12s %-30s %-10s %s\n",
			p.Payment.ID,
			dateStr,
			p.Payment.Amount,
			p.Payment.Kind,
			remoteAcc,
			p.Payment.Identification,
			reason,
		)
	}
}

func repeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}
