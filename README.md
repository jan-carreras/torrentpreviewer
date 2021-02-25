[![Go Report Card](https://goreportcard.com/badge/github.com/jan-carreras/torrentpreviewer)](https://goreportcard.com/report/github.com/jan-carreras/torrentpreviewer)
[![torrentPreview](https://circleci.com/gh/jan-carreras/torrentpreviewer.svg?style=shield)](https://app.circleci.com/pipelines/github/jan-carreras/torrentpreviewer)


# Torrent Preview

Is a simple BitTorrent client that generates images from the videos in Torrent files.

## The problem

To know the content of a torrent there are usually those approaches:

1. To include JPG images of the videos in the torrent inside the Torrent. You could download just those images to check
   if the content is what you're looking for
1. To include links in the description of the torrent as free text to an external service where the images are stored
1. When sending/posting the torrent to your favourite forum, share a screenshot of your screen of _some_ of the videos

The problem with images inside the torrent is that you cannot download the images if there are no seeders at the moment.
The images do not necessary need to match with the video content or quality. The problem with links to image of
screenshots is that those links can be broken and thus unusable. We can do better.

## Solution

> Generate the screenshot from the videos inside the Torrent using TorrentPreview. - Jan Carreras, bored, one random Tuesday

So, that is that. Torrent Preview receives a Magnet link, _unmagnetizes_ it converting it to a Torrent file, and downloads
the first 8MB of each video of the torrent to extract a Screenshot at the frame corresponding to the second 5 of the
video. It stores the screenshot and removes the video.

Then it exposes via an HTTP API all the information about the torrent itself, the files and it's corresponding images.

# Usage

This software is not meant to be used locally, but as a service. Simple example:

- Go to [torrentPreview.com](http://torrentpreview.com/)
-Add `magnet:?xt=urn:btih:3F8F219568B8B229581DDDD7BC5A5E889E906A9B&dn=Pulp%20Fiction%20%281994%29%201080p%20BrRip%20x264%20-%201.4GB%20-%20YIFY`
in the "Magnet / TorrentID" and hit Preview
- It will load [this](http://torrentpreview.com/?id=3f8f219568b8b229581dddd7bc5a5e889e906a9b) showing all the files of
  the torrent, and a blackish image extracted from the film. Right now it only extracts images from mp4 files (on a good
  day).
  
This website is in (at best) beta phase. Demo purposes only. Do not try to feed read magnets. It should identify them,
but the images are not going to be extracted.

## Developers

You want to start the project locally? You'll need to:

- Install some system dependencies
    - make, [go >= 1.15](https://golang.org/dl/), [ffmpeg](https://ffmpeg.org/download.html), [sqlite3](https://www.sqlite.org/download.html)
- Compile the program
- Configure and run the various components
- Consume the API or query the DB for results

### Build

We must generate 2 binaries:

- torrentprev: CLI client to register Magnet links and download torrents
- http: API to query torrent and image information

#### OSX Build

```bash
make build
```    

#### Linux Build

````bash
make build-linux
````

You'll find the binaries in `./bin` prefixed with your architecture like: `darwin-http` or `linux-torrentprev`

#### Database

Create an empty sqlite3 database on the root of the project:

```bash
touch prevtorrent.sqlite
```

Then create the schema by issuing:

```bash
sqlite3 prevtorrent.sqlite < infrastructure/database/sqlite.schema.sql

```
