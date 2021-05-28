CREATE TABLE relic_campaign_may_2021 (
    submissionID uuid DEFAULT gen_random_uuid (),
    userID text NOT NULL,
    ign text NOT NULL,
    score int NOT NULL,
    proof text NOT NULL,
    verified boolean DEFAULT FALSE
);

CREATE TABLE devtest (
    submissionID uuid DEFAULT gen_random_uuid (),
    userID text NOT NULL,
    ign text NOT NULL,
    score int NOT NULL,
    proof text NOT NULL,
    verified boolean DEFAULT FALSE
);

INSERT INTO devtest (
    userID, ign, score, proof
) VALUES (
    'test', 'test', 2, 'some-uri'
);

INSERT INTO devtest (
    userID, ign, score, proof
) VALUES (
    'user2#2222', 'user2', 3, 'https://cdn.discordapp.com/attachments/301032274860965890/847478722826207242/unknown.png'
);

-- CREATE TABLE ign_lookup (
--     id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
--     user_id text NOT NULL,
--     guild_id text NOT NULL,
--     ign text NOT NULL,
--     UNIQUE (user_id, guild_id)
-- );

-- CREATE TABLE event_info (
--     id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
--     guild_id text NOT NULL,
--     name text NOT NULL
-- )

CREATE TABLE role_lookup (
    id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    guild_id text NOT NULL,
    action text NOT NULL,
    role_id text NOT NULL,
    UNIQUE (guild_id, action)
);