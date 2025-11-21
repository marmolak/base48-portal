package handler

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/base48/member-portal/internal/db"
)

// UnmatchedPaymentInfo contains payment with analysis
type UnmatchedPaymentInfo struct {
	Payment     db.Payment
	UserExists  bool
	Category    string // "empty_vs", "user_not_found", "sync_bug"
	Reason      string
	IsIncoming  bool
	AmountFloat float64
}

// AdminUnmatchedPaymentsHandler shows all payments that couldn't be automatically matched to users
// GET /admin/payments/unmatched
func (h *Handler) AdminUnmatchedPaymentsHandler(w http.ResponseWriter, r *http.Request) {
	user := h.auth.GetUser(r)
	if user == nil {
		http.Redirect(w, r, "/auth/login", http.StatusTemporaryRedirect)
		return
	}

	if !user.IsAdmin() {
		http.Error(w, "Forbidden - admin access required", http.StatusForbidden)
		return
	}

	ctx := r.Context()

	// Get database user
	dbUser, err := h.queries.GetUserByKeycloakID(ctx, sql.NullString{
		String: user.ID,
		Valid:  true,
	})
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Get all unassigned payments
	unassignedPayments, err := h.queries.ListUnassignedPayments(ctx)
	if err != nil {
		http.Error(w, "Failed to fetch unassigned payments", http.StatusInternalServerError)
		return
	}

	// Analyze each payment - ONLY INCOMING PAYMENTS
	var unmatchedList []UnmatchedPaymentInfo
	totalAmount := 0.0
	countPayments := 0

	// Count by category
	countEmptyVS := 0
	countUserNotFound := 0
	countSyncBug := 0

	for _, payment := range unassignedPayments {
		// Parse amount for totals and direction
		amountFloat := 0.0
		if amount, err := strconv.ParseFloat(payment.Amount, 64); err == nil {
			amountFloat = amount
		}

		// SKIP ALL OUTGOING PAYMENTS (negative or zero amounts)
		if amountFloat <= 0 {
			continue
		}

		// Only process incoming payments from here
		totalAmount += amountFloat
		countPayments++

		info := UnmatchedPaymentInfo{
			Payment:     payment,
			AmountFloat: amountFloat,
			IsIncoming:  true, // Always true now
		}

		// Skip if no identification (VS)
		if payment.Identification == "" {
			info.Category = "empty_vs"
			info.Reason = "Empty variable symbol"
			unmatchedList = append(unmatchedList, info)
			countEmptyVS++
			continue
		}

		// Check if user with this payments_id exists
		_, err = h.queries.GetUserByPaymentsID(ctx, sql.NullString{String: payment.Identification, Valid: true})
		if err == sql.ErrNoRows {
			info.Category = "user_not_found"
			info.UserExists = false
			info.Reason = "No user with this payments_id exists"
			unmatchedList = append(unmatchedList, info)
			countUserNotFound++
			continue
		} else if err != nil {
			http.Error(w, "Database error checking user", http.StatusInternalServerError)
			return
		}

		// User exists but payment not assigned - potential sync bug
		info.Category = "sync_bug"
		info.UserExists = true
		info.Reason = "User with this payments_id exists but payment not assigned (sync issue)"
		unmatchedList = append(unmatchedList, info)
		countSyncBug++
	}

	// Prepare template data
	data := map[string]interface{}{
		"User":              user,
		"DBUser":            dbUser,
		"UnmatchedList":     unmatchedList,
		"TotalCount":        len(unmatchedList),
		"TotalAmount":       totalAmount,
		"CountPayments":     countPayments,
		"CountEmptyVS":      countEmptyVS,
		"CountUserNotFound": countUserNotFound,
		"CountSyncBug":      countSyncBug,
	}

	h.render(w, "admin_payments_unmatched.html", data)
}
