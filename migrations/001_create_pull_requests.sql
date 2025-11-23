CREATE TABLE IF NOT EXISTS pull_requests (
    id BIGSERIAL PRIMARY KEY,
    pull_request_id VARCHAR(255) UNIQUE NOT NULL,
    pull_request_name VARCHAR(255) NOT NULL,
    author_id VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'OPEN',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    merged_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS pr_reviewers (
    pull_request_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    PRIMARY KEY (pull_request_id, user_id),
    FOREIGN KEY (pull_request_id) REFERENCES pull_requests(pull_request_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_pull_requests_author_id ON pull_requests(author_id);
CREATE INDEX IF NOT EXISTS idx_pull_requests_status ON pull_requests(status);
CREATE INDEX IF NOT EXISTS idx_pr_reviewers_user_id ON pr_reviewers(user_id);

