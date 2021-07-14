DROP TABLE IF EXISTS users_tokens;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS history;

--                Таблица пользователей.

CREATE TABLE IF NOT EXISTS users (
    id      	 BIGSERIAL PRIMARY KEY,
    firstname    TEXT NOT NULL,
    lastname     TEXT NOT NULL,
    active       BOOLEAN NOT NULL DEFAULT TRUE,
    created      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    email        TEXT NOT NULL UNIQUE,
    password     TEXT NOT NULL,
    folder       TEXT NOT NULL,
    admin        BOOLEAN NOT NULL DEFAULT FAlSE,
	code         INTEGER NOT NULL
);

--                Таблица для ведения истории операция пользователья. 

CREATE TABLE IF NOT EXISTS history (
	id          BIGSERIAL PRIMARY KEY,
	user_id     TEXT NOT NULL,
	operation   TEXT NOT NULL,
	created     TEXT NOT NULL,
	file        TEXT NOT NULL,
	restore     TEXT NOT NULL,
	rev         TEXT NULL DEFAULT NULL,
    server      TEXT NOT NULL	
);

--                Таблица токенов для пользователя.

CREATE TABLE IF NOT EXISTS users_tokens (
    token 		TEXT NOT NULL UNIQUE,
    user_id     BIGINT NOT NULL REFERENCES users,
    expire 		TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP + INTERVAL '3 hour',
    created 	TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

