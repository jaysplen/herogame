-- +goose Up
-- Repositions world-expansion nodes for watercolor map layout (1200x820 stage).

UPDATE map_nodes SET x = 180,  y = 540 WHERE id = 1;  -- Ironkeep Castle (SW)
UPDATE map_nodes SET x = 300,  y = 460 WHERE id = 2;  -- Moss Crossing
UPDATE map_nodes SET x = 260,  y = 230 WHERE id = 3;  -- North Forest
UPDATE map_nodes SET x = 180,  y = 700 WHERE id = 4;  -- South Quarry
UPDATE map_nodes SET x = 440,  y = 480 WHERE id = 5;  -- Bandit Camp
UPDATE map_nodes SET x = 1040, y = 280 WHERE id = 6;  -- Sunspire Castle (NE)
UPDATE map_nodes SET x = 380,  y = 200 WHERE id = 7;  -- Gem Caves
UPDATE map_nodes SET x = 340,  y = 700 WHERE id = 8;  -- Coal Pit
UPDATE map_nodes SET x = 560,  y = 280 WHERE id = 9;  -- Ruined Watch
UPDATE map_nodes SET x = 520,  y = 620 WHERE id = 10; -- Old Lumber
UPDATE map_nodes SET x = 700,  y = 300 WHERE id = 11; -- Stone Ridge
UPDATE map_nodes SET x = 680,  y = 600 WHERE id = 12; -- Golden Field
UPDATE map_nodes SET x = 880,  y = 200 WHERE id = 13; -- East Pass
UPDATE map_nodes SET x = 900,  y = 540 WHERE id = 14; -- South Gate
UPDATE map_nodes SET x = 620,  y = 440 WHERE id = 15; -- Wolf Den
UPDATE map_nodes SET x = 760,  y = 460 WHERE id = 16; -- Mercury Marsh

-- +goose Down
UPDATE map_nodes SET x = 120,  y = 360 WHERE id = 1;
UPDATE map_nodes SET x = 250,  y = 360 WHERE id = 2;
UPDATE map_nodes SET x = 250,  y = 220 WHERE id = 3;
UPDATE map_nodes SET x = 250,  y = 500 WHERE id = 4;
UPDATE map_nodes SET x = 390,  y = 360 WHERE id = 5;
UPDATE map_nodes SET x = 760,  y = 360 WHERE id = 6;
UPDATE map_nodes SET x = 390,  y = 220 WHERE id = 7;
UPDATE map_nodes SET x = 390,  y = 500 WHERE id = 8;
UPDATE map_nodes SET x = 520,  y = 260 WHERE id = 9;
UPDATE map_nodes SET x = 520,  y = 460 WHERE id = 10;
UPDATE map_nodes SET x = 640,  y = 260 WHERE id = 11;
UPDATE map_nodes SET x = 640,  y = 460 WHERE id = 12;
UPDATE map_nodes SET x = 760,  y = 240 WHERE id = 13;
UPDATE map_nodes SET x = 760,  y = 480 WHERE id = 14;
UPDATE map_nodes SET x = 520,  y = 360 WHERE id = 15;
UPDATE map_nodes SET x = 640,  y = 360 WHERE id = 16;
