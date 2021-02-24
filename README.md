[![torrentPreview](https://circleci.com/gh/jan-carreras/torrentpreviewer.svg?style=shield)](https://app.circleci.com/pipelines/github/jan-carreras/torrentpreviewer)

## Message for hiring companies engineers

The project is going to be a Torrent Previewer. Definition of the problem:

- Given Torrent (or Magnet link),
- Download the meta info of the torrent and files that contains,
- download the first 8MB from all the .mp4 files,
- and using ffmpeg, extract an image from the file so that we can see a preview,
- serve in a webpage the list of files found in the torrent with their respective previews.

Before starting writing this application I had 0 knowledge about how Torrent protocol works, DHT, torrents parts,
magnets, etc. And how to use ffmpeg, either. Now I have a pretty solid understanding on how those pieces fit together
and this version of the code can generate thumbnail already.

I've tested with torrents of more than 100GB of data containing more than 400 videos, and it works pretty smoothly, with
minimal memory footprint.

There are a lot of pending tasks and TODOs in the code since (a) it's not yet ready, and (b) it was not meant to be
shared.

I'm willing to discuss any implementation details if interested :) All my contact details and CV can be found
here: https://jcarreras.es

## Database

### Sqlite

Initialize by running:

```bash
sqlite3 prevtorrent.sqlite < infrastructure/database/sqlite.schema.sql
```