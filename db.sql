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
INSERT INTO devtest (userID, ign, score, proof)
VALUES ('test', 'test', 2, 'some-uri');
INSERT INTO devtest (userID, ign, score, proof)
VALUES (
        'user2#2222',
        'user2',
        3,
        'https://cdn.discordapp.com/attachments/301032274860965890/847478722826207242/unknown.png'
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
select e.score,
    u.id as uid,
    u.ign
from (
        select sum(e.score) as score,
            e.uid
        from (
                select e.score,
                    p.user_id as uid
                from event_scores as e
                    left join participation as p on p.id = e.participation_id
                where p.event_id = '557d5c54-017d-44e9-a1f0-e6d48572e633'
                    AND p.participating = TRUE
                    AND e.verified = TRUE
            ) as e
        group by e.uid
    ) as e
    left join users as u on u.id = e.uid
order by e.score desc;
select count(
        case
            s.verified
            when TRUE then 1
            else null
        end
    ) as verified,
    count(
        case
            s.verified
            when FALSE then 1
            else null
        end
    ) as pending
from (
        select s.verified
        from event_scores as s
            left join participation as p on p.id = s.participation_id
        where p.participating = TRUE
            and p.event_id = '557d5c54-017d-44e9-a1f0-e6d48572e633'
    ) as s;
select s.score,
    u.id,
    u.ign
from (
        select s.uid,
            max(s.score) as score
        from (
                select s.participation_id as pid,
                    s.score as score,
                    p.user_id as uid
                from event_scores as s
                    left join participation as p on p.id = s.participation_id
                where p.participating = TRUE
                    and p.event_id = '557d5c54-017d-44e9-a1f0-e6d48572e633'
                    and s.verified = TRUE
            ) as s
        group by s.uid
    ) as s
    left join users as u on u.id = s.uid;