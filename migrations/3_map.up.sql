CREATE TABLE temple (
    martial_order martialorder,
    location integer REFERENCES location (id)
);

INSERT INTO location (id, name) VALUES
    (0, 'Throne'),
    (1, 'Sulfer Point'),
    (2, 'Here Be'),
    (3, 'No Name'),
    (4, 'St. Cecil''s Bridge'),
    (5, 'Eye of Gideon'),
    (6, 'New Delphia'),
    (7, 'Fog'),
    (8, 'Worm Land'),
    (9, 'Passage of Smoke'),
    (10, 'The Ash Sea'),
    (11, 'Asteria'),
    (12, 'Yerk'),
    (13, 'Hideous Marsh'),
    (14, 'Necropolis'),
    (15, 'Crawler Pits'),
    (16, 'Obsidian Lake'),
    (17, 'Grisag'),
    (18, 'Terre Ignot'),
    (19, 'Camp Gray'),
    (20, 'Camp Watkins'),
    (21, 'Gallows'),
    (22, 'Mercy Cove'),
    (23, 'Giant''s Bluff'),
    (24, 'H. Beach'),
    (25, 'Duncan Talley'),
    (26, 'Mangrove'),
    (27, 'Lighthouse'),
    (28, 'Apostle Valley'),
    (29, 'Poppy Fields'),
    (30, 'Agathinias'),
    (31, 'Ithmont'),
    (32, 'Whitecrypt'),
    (33, 'Visygi'),
    (34, 'Outer Realm'),
    (35, 'Memoria'),
    (36, 'Hem Wood'),
    (37, 'River Crossing'),
    (38, 'Fool''s Way'),
    (39, 'Landfall'),
    (40, 'Bouchard''s Island');

INSERT INTO adjacent_location (location, adjacent) VALUES
    (0, 14), (0, 15), (0, 16), (0, 27),
    (1, 4),
    (2, 7), (2, 8), (2, 9), (2, 10),
    (3, 10),
    (4, 1), (4, 5),
    (5, 4), (5, 6), (5, 12), (5, 13), (5, 0),
    (6, 5), (6, 7), (6, 13), (6, 14),
    (7, 2), (7, 6), (7, 8), (7, 14), (7, 15),
    (8, 2), (8, 7), (8, 9), (8, 15), (8, 16), (8, 17),
    (9, 2), (9, 8), (9, 10), (9, 17), (9, 18),
    (10, 2), (10, 3), (10, 9), (10, 18),
    (11, 12), (11, 19),
    (12, 5), (12, 11), (12, 13),
    (13, 5), (13, 6), (13, 12), (13, 25),
    (14, 0), (14, 6), (14, 7), (14, 15),
    (15, 0), (15, 7), (15, 8), (15, 14), (15, 16),
    (16, 0), (16, 8), (16, 15), (16, 17),
    (17, 8), (17, 9), (17, 16), (17, 18),
    (18, 9), (18, 10), (18, 17), (18, 29), (18, 30),
    (19, 11), (19, 20),
    (20, 19), (20, 21),
    (21, 20), (21, 22),
    (22, 21), (22, 23),
    (23, 22), (23, 24),
    (24, 23), (24, 32), (24, 36),
    (25, 13), (25, 26), (25, 31),
    (26, 25), (26, 27), (26, 31),
    (27, 0), (27, 26), (27, 28), (27, 31), (27, 33),
    (28, 27), (28, 29), (28, 33), (28, 34),
    (29, 18), (29, 28), (29, 30), (29, 34), (29, 35),
    (30, 18), (30, 29), (30, 35),
    (31, 25), (31, 26), (31, 27), (31, 32), (31, 33),
    (32, 24), (32, 31), (32, 33), (32, 36), (32, 37),
    (33, 27), (33, 28), (33, 31), (33, 32), (33, 37),
    (34, 28), (34, 29), (34, 35), (34, 37),
    (35, 29), (35, 30), (35, 34), (35, 40),
    (36, 24), (36, 32), (36, 37), (36, 38), (36, 39),
    (37, 32), (37, 33), (37, 34), (37, 36), (37, 38),
    (38, 36), (38, 37), (38, 39),
    (39, 36), (39, 38),
    (40, 35);

INSERT INTO temple (martial_order, location) VALUES ('Staghorn Sect', 39);
INSERT INTO temple (martial_order, location) VALUES ('Order Gorgana', 11);
INSERT INTO temple (martial_order, location) VALUES ('The Baaturate', 3);

