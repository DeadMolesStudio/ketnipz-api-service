-- +migrate Up
CREATE TABLE IF NOT EXISTS skin (
    skin_id serial PRIMARY KEY,
    skin_name varchar(32) NOT NULL,
    cost integer DEFAULT 0
);

ALTER TABLE user_profile 
    ADD coins integer DEFAULT 0 CONSTRAINT nonnegative_coins CHECK (coins >= 0),
    ADD skin integer REFERENCES skin DEFAULT NULL;

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
