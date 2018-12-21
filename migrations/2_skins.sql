-- +migrate Up
CREATE TABLE IF NOT EXISTS skin (
    skin_id serial PRIMARY KEY,
    skin_name varchar(32) NOT NULL,
    cost integer DEFAULT 0
);

INSERT INTO skin (skin_name, cost) VALUES
    ('Classic', 0), -- default skin
    ('Nature', 50),
    ('Home', 100),
    ('Pumpkin', 150),
    ('Freak', 200),
    ('Christmas', 0);

ALTER TABLE user_profile 
    ADD coins integer DEFAULT 0 CONSTRAINT nonnegative_coins CHECK (coins >= 0),
    ADD skin integer REFERENCES skin DEFAULT 1; -- default skin

CREATE TABLE IF NOT EXISTS user_purchased_skins (
    user_id integer REFERENCES user_profile NOT NULL,
    skin_id integer REFERENCES skin NOT NULL
);

-- +migrate Down
ALTER TABLE user_profile
    DROP coins,
    DROP skin;

DROP TABLE IF EXISTS skin;
DROP TABLE IF EXISTS user_purchased_skins;
