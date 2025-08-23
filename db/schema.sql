-- Create bots table
CREATE TABLE IF NOT EXISTS bots (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(255) UNIQUE NOT NULL,
    basic_id VARCHAR(255) UNIQUE NOT NULL,
    chat_mode VARCHAR(10) NOT NULL CHECK (chat_mode IN ('chat', 'bot')),
    display_name VARCHAR(255) NOT NULL,
    mark_as_read_mode VARCHAR(10) NOT NULL CHECK (mark_as_read_mode IN ('auto', 'manual')),
    picture_url TEXT,
    premium_id VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create index on user_id for faster lookups
CREATE INDEX idx_bots_user_id ON bots(user_id);

-- Create index on basic_id for faster lookups
CREATE INDEX idx_bots_basic_id ON bots(basic_id);