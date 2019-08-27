CREATE TYPE martialorder AS ENUM ('staghorn', 'gorgona', 'baaturate');
CREATE TYPE playerclass AS ENUM (
    'recruit',
    'infantry', 'spear', 'glaivemaster', 'sword', 'legionary',
    'cavalry', 'heavycavalry', 'monsterknight', 'lightcavalry', 'horsearcher',
    'ranger', 'archer', 'mage', 'medic', 'healer'
);

CREATE TABLE location
(
    id    serial PRIMARY KEY,
    name  text NOT NULL UNIQUE,
    owner martialorder
);

CREATE TABLE player
(
    id              serial PRIMARY KEY,
    twitter_id      text         NOT NULL UNIQUE,
    receive_updates boolean      NOT NULL DEFAULT TRUE,
    active          boolean      NOT NULL DEFAULT TRUE,
    dead            boolean      NOT NULL DEFAULT FALSE,
    martial_order   martialorder NOT NULL,
    location        integer REFERENCES location (id),
    class           playerclass  NOT NULL DEFAULT 'recruit',
    experience      smallint     NOT NULL DEFAULT 0,
    rank            smallint     NOT NULL DEFAULT 1
);

CREATE TABLE adjacent_location
(
    location integer REFERENCES location (id) ON DELETE CASCADE,
    adjacent integer REFERENCES location (id) ON DELETE CASCADE
);