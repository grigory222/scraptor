CREATE TABLE chats (
    id INTEGER PRIMARY KEY,
    type VARCHAR(10) CHECK (type IN ('personal', 'group'))
);

CREATE TABLE tokens (
    id SERIAL PRIMARY KEY,
    token TEXT NOT NULL
);

CREATE TABLE links (
    id SERIAL PRIMARY KEY,
    link TEXT NOT NULL,
    tag VARCHAR(10) CHECK (tag IN ('work', 'hobby', 'family')),
    token_id INTEGER REFERENCES tokens(id) ON DELETE SET NULL
);

CREATE TABLE chats_links (
    chat_id INTEGER REFERENCES chats(id) ON DELETE CASCADE,
    link_id INTEGER REFERENCES links(id) ON DELETE CASCADE,
    status VARCHAR(10) CHECK (status IN ('active', 'archive')),
    PRIMARY KEY (chat_id, link_id)
);
