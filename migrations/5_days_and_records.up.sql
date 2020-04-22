CREATE TYPE combattype AS ENUM (
    'attack', 'counterattack', 'revive'
);

CREATE TYPE combatresult AS ENUM (
    'success', 'failure', 'notarget'
);

CREATE TYPE ownershipevent AS ENUM (
    'capture', 'occupy'
);

CREATE TABLE calendar (
    count smallint NOT NULL
);

INSERT INTO calendar (count) VALUES (0);

CREATE TABLE move_record (
    day smallint NOT NULL,
    location integer REFERENCES location (id),
    player integer REFERENCES player (id),
    timestamp timestamptz DEFAULT now()
);

CREATE TABLE combat_record (
    day smallint NOT NULL,
    location integer REFERENCES location (id),
    type combattype NOT NULL,
    attacker integer REFERENCES player (id),
    defender integer REFERENCES player (id),
    attacker_class playerclass NOT NULL,
    defender_class playerclass NOT NULL,
    result combatresult NOT NULL
);

CREATE TABLE ownership_record (
    day smallint NOT NULL,
    location integer REFERENCES location (id),
    event ownershipevent NOT NULL,
    martial_order martialorder NOT NULL
);
