-- Create staff_users table
CREATE TABLE IF NOT EXISTS staff_users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    branch VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_staff_users_email ON staff_users(email);
CREATE INDEX idx_staff_users_branch ON staff_users(branch);

-- Create members table
CREATE TABLE IF NOT EXISTS members (
    id UUID PRIMARY KEY,
    old_id UUID,
    name VARCHAR(255) NOT NULL,
    last4 CHAR(4),
    total_points INT NOT NULL DEFAULT 0,
    milestone_score INT NOT NULL DEFAULT 0,
    points_1_0_liter INT NOT NULL DEFAULT 0,
    points_1_5_liter INT NOT NULL DEFAULT 0,
    branch VARCHAR(255),
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    membership_number VARCHAR(50),
    registration_receipt_number VARCHAR(50),
    welcome_bonus_claimed BOOLEAN NOT NULL DEFAULT FALSE,
    registered_by_staff VARCHAR(50),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_members_branch ON members(branch);
CREATE INDEX idx_members_status ON members(status);
CREATE INDEX idx_members_membership_number ON members(membership_number);
CREATE INDEX idx_members_created_at ON members(created_at DESC);

-- Create member_point_transactions table
CREATE TABLE IF NOT EXISTS member_point_transactions (
    id UUID PRIMARY KEY,
    member_id UUID NOT NULL REFERENCES members(id) ON DELETE CASCADE,
    staff_user_id UUID NOT NULL REFERENCES staff_users(id),
    action VARCHAR(10) NOT NULL CHECK (action IN ('add', 'deduct', 'redeem', 'adjust')),
    product_type VARCHAR(20) NOT NULL,
    points INT NOT NULL,
    receipt_text TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_transactions_member_id ON member_point_transactions(member_id);
CREATE INDEX idx_transactions_staff_user_id ON member_point_transactions(staff_user_id);
CREATE INDEX idx_transactions_created_at ON member_point_transactions(created_at DESC);
CREATE INDEX idx_transactions_member_created ON member_point_transactions(member_id, created_at DESC);
