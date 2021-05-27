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