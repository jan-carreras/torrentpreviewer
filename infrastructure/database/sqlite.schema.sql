CREATE TABLE IF NOT EXISTS torrents
(
    id          VARCHAR(40) NOT NULL,
    name        TEXT        NOT NULL,
    length      INT         NOT NULL,
    pieceLength INT         NOT NULL,
    raw         BLOB        NOT NULL,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (id)
);


CREATE TABLE IF NOT EXISTS files
(
    torrent_id varchar(40) NOT NULL,
    id         int         NOT NULL,
    name       TEXT        NOT NULL,
    length     INT         NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (torrent_id, id),
    FOREIGN KEY (torrent_id) REFERENCES torrents (id)
);

CREATE TABLE IF NOT EXISTS media
(
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    torrent_id varchar(40) NOT NULL,
    file_id    int         NOT NULL,
    name       TEXT        NOT NULL,
    length     INT         NOT NULL,
    source     TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (torrent_id) REFERENCES torrents (id),
    FOREIGN KEY (torrent_id, file_id) REFERENCES files (torrent_id, id)
);
CREATE INDEX media_torrent_id_file_id ON media (torrent_id, file_id);
