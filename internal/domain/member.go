package domain

import (
	"time"

	"github.com/google/uuid"
)

type Member struct {
	ID                        uuid.UUID  `db:"id" json:"id"`
	OldID                     *uuid.UUID `db:"old_id" json:"old_id,omitempty"`
	Name                      string     `db:"name" json:"name"`
	Last4                     *string    `db:"last4" json:"last4,omitempty"`
	TotalPoints               int        `db:"total_points" json:"total_points"`
	MilestoneScore            int        `db:"milestone_score" json:"milestone_score"`
	Points1_0Liter            int        `db:"points_1_0_liter" json:"points_1_0_liter"`
	Points1_5Liter            int        `db:"points_1_5_liter" json:"points_1_5_liter"`
	Branch                    *string    `db:"branch" json:"branch,omitempty"`
	Status                    string     `db:"status" json:"status"`
	MembershipNumber          *string    `db:"membership_number" json:"membership_number,omitempty"`
	RegistrationReceiptNumber *string    `db:"registration_receipt_number" json:"registration_receipt_number,omitempty"`
	WelcomeBonusClaimed       bool       `db:"welcome_bonus_claimed" json:"welcome_bonus_claimed"`
	RegisteredByStaff         *string    `db:"registered_by_staff" json:"registered_by_staff,omitempty"`
	CreatedAt                 time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt                 time.Time  `db:"updated_at" json:"updated_at"`
}

type MemberCreateRequest struct {
	Name                      string  `json:"name" binding:"required"`
	Last4                     *string `json:"last4,omitempty"`
	Branch                    *string `json:"branch,omitempty"`
	RegistrationReceiptNumber *string `json:"registration_receipt_number,omitempty"`
}

type MemberUpdateRequest struct {
	Name   *string `json:"name,omitempty"`
	Last4  *string `json:"last4,omitempty"`
	Branch *string `json:"branch,omitempty"`
	Status *string `json:"status,omitempty"`
}

type MemberListResponse struct {
	Members []Member `json:"members"`
	Total   int      `json:"total"`
	Page    int      `json:"page"`
	Limit   int      `json:"limit"`
}
