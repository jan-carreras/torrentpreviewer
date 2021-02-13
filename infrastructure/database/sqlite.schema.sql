CREATE TABLE IF NOT EXISTS torrents
(
    id          VARCHAR(40) NOT NULL,
    name        TEXT        NOT NULL,
    length      INT         NOT NULL,
    pieceLength INT         NOT NULL,
    raw         BLOB        NOT NULL,

    PRIMARY KEY (id)
);