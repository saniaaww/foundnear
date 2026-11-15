CREATE OR REPLACE FUNCTION set_updated_at() RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE roles (
    id   BIGSERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE
);

CREATE TABLE users (
    id            BIGSERIAL PRIMARY KEY,
    role_id       BIGINT       NOT NULL REFERENCES roles(id) ON DELETE RESTRICT,
    name          VARCHAR(100) NOT NULL,
    email         VARCHAR(150) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    phone         VARCHAR(30),
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_users_role_id ON users(role_id);
CREATE TRIGGER trg_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE categories (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(100) NOT NULL UNIQUE,
    description TEXT
);

CREATE TABLE item_reports (
    id            BIGSERIAL PRIMARY KEY,
    user_id       BIGINT       NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    category_id   BIGINT       NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
    report_type   VARCHAR(10)  NOT NULL,
    item_name     VARCHAR(150) NOT NULL,
    description   TEXT         NOT NULL DEFAULT '',
    private_info  TEXT,
    location      VARCHAR(255) NOT NULL,
    incident_date DATE         NOT NULL,
    image_url     VARCHAR(500),
    status        VARCHAR(20)  NOT NULL DEFAULT 'active',
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_item_reports_type   CHECK (report_type IN ('lost', 'found')),
    CONSTRAINT chk_item_reports_status CHECK (status IN ('active', 'matched', 'returned', 'closed'))
);
CREATE INDEX idx_item_reports_user_id       ON item_reports(user_id);
CREATE INDEX idx_item_reports_category_id   ON item_reports(category_id);
CREATE INDEX idx_item_reports_report_type   ON item_reports(report_type);
CREATE INDEX idx_item_reports_status        ON item_reports(status);
CREATE INDEX idx_item_reports_incident_date ON item_reports(incident_date DESC);
CREATE TRIGGER trg_item_reports_updated_at BEFORE UPDATE ON item_reports
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE item_claims (
    id                BIGSERIAL PRIMARY KEY,
    report_id         BIGINT      NOT NULL REFERENCES item_reports(id) ON DELETE RESTRICT,
    claimant_id       BIGINT      NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    claim_description TEXT        NOT NULL,
    status            VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_item_claims_status CHECK (status IN ('pending', 'approved', 'rejected', 'cancelled'))
);
CREATE INDEX idx_item_claims_report_id   ON item_claims(report_id);
CREATE INDEX idx_item_claims_claimant_id ON item_claims(claimant_id);
CREATE INDEX idx_item_claims_status      ON item_claims(status);
CREATE UNIQUE INDEX uq_item_claims_pending
    ON item_claims(report_id, claimant_id) WHERE status = 'pending';
CREATE TRIGGER trg_item_claims_updated_at BEFORE UPDATE ON item_claims
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE OR REPLACE FUNCTION enforce_claim_on_found_report() RETURNS TRIGGER AS $$
DECLARE
    v_type VARCHAR(10);
BEGIN
    SELECT report_type INTO v_type FROM item_reports WHERE id = NEW.report_id;
    IF FOUND AND v_type <> 'found' THEN
        RAISE EXCEPTION 'only found reports can be claimed' USING ERRCODE = 'P0001';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_item_claims_found_only BEFORE INSERT ON item_claims
    FOR EACH ROW EXECUTE FUNCTION enforce_claim_on_found_report();

CREATE TABLE claim_verifications (
    id                BIGSERIAL PRIMARY KEY,
    claim_id          BIGINT      NOT NULL UNIQUE REFERENCES item_claims(id) ON DELETE RESTRICT,
    officer_id        BIGINT      NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    verification_note TEXT        NOT NULL DEFAULT '',
    result            VARCHAR(20) NOT NULL,
    verified_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_claim_verifications_result CHECK (result IN ('approved', 'rejected'))
);
CREATE INDEX idx_claim_verifications_officer_id ON claim_verifications(officer_id);

CREATE TABLE item_returns (
    id          BIGSERIAL PRIMARY KEY,
    claim_id    BIGINT      NOT NULL UNIQUE REFERENCES item_claims(id) ON DELETE RESTRICT,
    returned_by BIGINT      NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    received_by BIGINT      NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    return_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    return_note TEXT        NOT NULL DEFAULT ''
);
CREATE INDEX idx_item_returns_returned_by ON item_returns(returned_by);
CREATE INDEX idx_item_returns_received_by ON item_returns(received_by);
