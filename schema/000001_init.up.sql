CREATE TABLE songs (
    id SERIAL PRIMARY KEY,
    "group" VARCHAR(100),
    title VARCHAR(100),
    release_date DATE,
    text TEXT,
    link TEXT,
    UNIQUE ("group", title)
);