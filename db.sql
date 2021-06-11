CREATE TABLE role_lookup (
    id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    guild_id text NOT NULL,
    action text NOT NULL,
    role_id text NOT NULL,
    UNIQUE (guild_id, action)
);
CREATE TABLE events (
    id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    guild_id text NOT NULL,
    name text NOT NULL,
    start_date timestamptz DEFAULT current_timestamp,
    end_date timestamptz NOT NULL,
    active boolean,
    event_type text
);
CREATE TABLE users (
    id text NOT NULL PRIMARY KEY,
    ign text NOT NULL
);
CREATE TABLE participation (
    id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id text,
    event_id uuid,
    participating boolean NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE,
    UNIQUE (user_id, event_id)
);
CREATE TABLE event_scores (
    id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    score int NOT NULL,
    proof text NOT NULL,
    verified boolean DEFAULT FALSE,
    participation_id uuid,
    FOREIGN KEY (participation_id) REFERENCES participation(id) ON DELETE CASCADE
);