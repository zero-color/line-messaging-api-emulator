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

-- Create webhooks table for bot webhook configurations
CREATE TABLE IF NOT EXISTS webhooks (
    id SERIAL PRIMARY KEY,
    bot_id INTEGER NOT NULL REFERENCES bots(id) ON DELETE CASCADE,
    endpoint TEXT NOT NULL,
    active BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(bot_id)
);

-- Create index on bot_id for faster lookups
CREATE INDEX idx_webhooks_bot_id ON webhooks(bot_id);

-- Create messages table for tracking sent messages
CREATE TABLE IF NOT EXISTS messages (
    id SERIAL PRIMARY KEY,
    bot_id INTEGER NOT NULL REFERENCES bots(id) ON DELETE CASCADE,
    message_type VARCHAR(50) NOT NULL, -- push, broadcast, multicast, narrowcast, reply
    recipient_type VARCHAR(50), -- user, group, room, all, multiple
    recipient_id TEXT, -- user_id, group_id, room_id, or comma-separated list for multicast
    content JSONB NOT NULL, -- Store the actual message content as JSON
    retry_key UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for messages
CREATE INDEX idx_messages_bot_id ON messages(bot_id);
CREATE INDEX idx_messages_retry_key ON messages(retry_key) WHERE retry_key IS NOT NULL;
CREATE INDEX idx_messages_created_at ON messages(created_at);