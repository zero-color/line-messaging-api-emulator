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

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(255) UNIQUE NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    picture_url TEXT,
    status_message TEXT,
    language VARCHAR(10),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create index on user_id for faster lookups
CREATE INDEX idx_users_user_id ON users(user_id);

-- Create bot_followers table for many-to-many relationship
CREATE TABLE IF NOT EXISTS bot_followers (
    id SERIAL PRIMARY KEY,
    bot_id INTEGER NOT NULL REFERENCES bots(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    followed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(bot_id, user_id)
);

-- Create indexes for bot_followers
CREATE INDEX idx_bot_followers_bot_id ON bot_followers(bot_id);
CREATE INDEX idx_bot_followers_user_id ON bot_followers(user_id);