CREATE EXTENSION IF NOT EXISTS citext;

CREATE TYPE unit_system AS ENUM ('metric', 'imperial');

CREATE TYPE exercise_type AS ENUM ('Bodyweight', 'Weighted', 'Assisted');

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username CITEXT UNIQUE NOT NULL CHECK (
        LENGTH(username) BETWEEN 3 AND 24
        AND POSITION(' ' IN username) = 0
        AND username ~ '^[A-Za-z0-9]([A-Za-z0-9._]{0,22}[A-Za-z0-9])?$'
        AND username !~ '[-_.]{2,}'
    ),
    email VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(50),
    sex VARCHAR(6) CHECK (sex IN ('male', 'female')),
    preferred_units unit_system NOT NULL DEFAULT 'metric',
    country_code CHAR(2) CHECK (country_code ~ '^[A-Z]{2}$'), -- ISO 3166-1 alpha-2 country codes
    avatar_url VARCHAR(255),
    bio TEXT CHECK (LENGTH(bio) <= 160), -- Limit bio to 160 characters
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE user_providers (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,
    provider_user_id VARCHAR(100) NOT NULL,
    first_name VARCHAR(50),
    last_name VARCHAR(50),
    nickname VARCHAR(50),
    avatar_url VARCHAR(255),
    location VARCHAR(100),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, provider, provider_user_id)
);

CREATE TABLE initial_user_providers (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,
    provider_user_id VARCHAR(100) NOT NULL,
    first_name VARCHAR(50),
    last_name VARCHAR(50),
    nickname VARCHAR(50),
    avatar_url VARCHAR(255),
    location VARCHAR(100),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, provider)
);

CREATE TABLE refresh_tokens (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id),
    UNIQUE (token)
);

CREATE TABLE exercises (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL
);

CREATE TABLE bodyweight_logs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    bodyweight DECIMAL(10, 2) NOT NULL, -- Store in kilograms
    log_date TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, log_date)
);

CREATE TABLE exercise_logs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    exercise_id INTEGER NOT NULL REFERENCES exercises(id) ON DELETE CASCADE,
    reps INTEGER NOT NULL,
    weight DECIMAL(10, 2) NOT NULL, -- Store in kilograms
    additional_weight DECIMAL(10, 2), -- Store in kilograms, only for bodyweight exercises
    exercise_type exercise_type, -- Only for bodyweight exercises
    bodyweight_id INTEGER NOT NULL REFERENCES bodyweight_logs(id),
    log_date TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, exercise_id, log_date)
);

CREATE TABLE trophies (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE user_trophies (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    trophy_id INTEGER NOT NULL REFERENCES trophies(id) ON DELETE CASCADE,
    display_order INT DEFAULT 0 CHECK (display_order IN (0, 1, 2)),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, trophy_id)
);

INSERT INTO exercises (name) VALUES
    ('Bench Press'),
    ('Deadlift'),
    ('Overhead Press'),
    ('Power Clean'),
    ('Power Snatch'),
    ('Back Squat'),
    ('Front Squat'),
    ('Dip'),
    ('Pull Up');

INSERT INTO trophies (name, description, created_at, updated_at) VALUES
    ('pull-up-king', 'Achieved by lifting twice your body weight in a pull-up.', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('pull-up-pro', 'Perform 10 consecutive pull-ups.', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('2-ez-plates', 'Lift 100 kg in the bench press.', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('deadlift-dynamo', 'Pull three times your body weight in a single deadlift.', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('squat-sovereign', 'Squat 2.5 times your body weight in one go to reign supreme.', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('powerlifting-prodigy', 'Total a lift of five times your body weight across bench press, squat, and deadlift.', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('olympic-overachiever', 'Snatch or Clean & Jerk 1.5 times your body weight.', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('shoulder-mount', 'Press 1.5 times your body weight overhead.', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('dip-master', 'Perform 20 consecutive dips.', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('quad-king', 'Achieve a 200% bodyweight front squat.', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);
