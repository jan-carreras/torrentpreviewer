INSERT INTO torrents (id, name, length, pieceLength, raw)
VALUES ('cb84ccc10f296df72d6c40ba7a07c178a4323a14', 'Test Name', 1000, 100, '');

INSERT INTO files (torrent_id, id, name, length)
VALUES ('cb84ccc10f296df72d6c40ba7a07c178a4323a14', 0, 'film1.mp4', 600);
INSERT INTO files (torrent_id, id, name, length)
VALUES ('cb84ccc10f296df72d6c40ba7a07c178a4323a14', 1, 'img2.jpg', 400);

INSERT INTO media (torrent_id, file_id, name, length)
VALUES ('cb84ccc10f296df72d6c40ba7a07c178a4323a14', 0, 'fil1.mp4.pjg', 10);
