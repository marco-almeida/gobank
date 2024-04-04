CREATE TABLE IF NOT EXISTS accounts (
		id BIGSERIAL PRIMARY KEY NOT NULL,
		user_id BIGINT NOT NULL,
		balance BIGINT NOT NULL DEFAULT 0,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);