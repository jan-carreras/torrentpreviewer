[![Go Report Card](https://goreportcard.com/badge/github.com/jan-carreras/torrentpreviewer)](https://goreportcard.com/report/github.com/jan-carreras/torrentpreviewer)
[![torrentPreview](https://circleci.com/gh/jan-carreras/torrentpreviewer.svg?style=shield)](https://app.circleci.com/pipelines/github/jan-carreras/torrentpreviewer)
[![codecov](https://codecov.io/gh/jan-carreras/torrentpreviewer/branch/master/graph/badge.svg?token=4T0T6TSK6N)](https://codecov.io/gh/jan-carreras/torrentpreviewer)
[![Maintainability](https://api.codeclimate.com/v1/badges/9f57f2a367707783a156/maintainability)](https://codeclimate.com/github/jan-carreras/torrentpreviewer/maintainability)
![Version Beta](https://img.shields.io/badge/version-beta-yellow)
![Production Ready](https://img.shields.io/badge/production%20ready-nope-red)

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
- Add `magnet:?xt=urn:btih:3F8F219568B8B229581DDDD7BC5A5E889E906A9B&dn=Pulp%20Fiction%20%281994%29%201080p%20BrRip%20x264%20-%201.4GB%20-%20YIFY`
in the "Magnet / TorrentID" and hit Preview
- It will load [this](http://torrentpreview.com/?id=3f8f219568b8b229581dddd7bc5a5e889e906a9b) showing all the files of
  the torrent, and a blackish image extracted from the film. Right now it only extracts images from mp4 files (on a good
  day).
  
This website is in (at best) beta phase. Demo purposes only. Do not try to feed read magnets. It should identify them,
but the images are not going to be extracted.

## Developers

Two paths:

- Docker (super slow on OSX for FS well know problems) 
- Native go, everything installed in your system

**If you just want to try the project and see how it works, go to the Docker Section since it's easier for you.**


## Docker usage

- Execute `docker-compose build`.
  - Are you on OSX? It might take a while. Drink a coffee. Maybe two. Shower. Call mum.
- Create the database (see section below) on the root of the project
- Execute `docker-compose up` and you'll the HTTP listening on :8080

### Import magnet

```bash
docker-compose run cli /go/bin/torrentprev magnet "magnet:?xt=urn:btih:c92f656155d0d8e87d21471d7ea43e3ad0d42723"

DEBU[0000] not found in db. about to resolve the magnet using network  magnet="magnet:?xt=urn:btih:C92F656155D0D8E87D21471D7EA43E3AD0D42723" magnetID=c92f656155d0d8e87d21471d7ea43e3ad0d42723
2021-02-25T10:45:27+0000 NONE  portfwd.go:30: discovered 0 upnp devices
```

### Download torrent and generate image

From the previous command, get the ID from "magnetID=c92f656155d0d8e87d21471d7ea43e3ad0d42723" and use it to download it. Like this:

```bash
docker-compose run cli /go/bin/torrentprev download c92f656155d0d8e87d21471d7ea43e3ad0d42723 

magnet:?xt=urn:btih:c92f656155d0d8e87d21471d7ea43e3ad0d42723
root@86637674da0f:/go/src/github.com/jan-carreras/torrentpreviewer# /go/bin/torrentprev download c92f656155d0d8e87d21471d7ea43e3ad0d42723
DEBU[0000] torrent to be processed                       name="----REDACTED---" torrentID=c92f656155d0d8e87d21471d7ea43e3ad0d42723
DEBU[0000] images that we already have for the torrent   imagesTorrentCount=0 name="----REDACTED---" torrentID=c92f656155d0d8e87d21471d7ea43e3ad0d42723
DEBU[0000] pieces to download                            downloadPlanSize=8388608 name="----REDACTED---" pieceCount=4 pieceLength=2097152 torrentID=c92f656155d0d8e87d21471d7ea43e3ad0d42723
2021-02-25T10:46:47+0000 NONE  portfwd.go:30: discovered 0 upnp devices
INFO[0012] piece download completed                      complete=true pieceIdx=1 torrent="----REDACTED---" waitingFor=3
DEBU[0012] part added to registry                        piece=1 torrentID=c92f656155d0d8e87d21471d7ea43e3ad0d42723
INFO[0012] piece download completed                      complete=true pieceIdx=0 torrent="----REDACTED---" waitingFor=2
DEBU[0012] part added to registry                        piece=0 torrentID=c92f656155d0d8e87d21471d7ea43e3ad0d42723
INFO[0013] piece download completed                      complete=true pieceIdx=3 torrent="----REDACTED---" waitingFor=1
DEBU[0013] part added to registry                        piece=3 torrentID=c92f656155d0d8e87d21471d7ea43e3ad0d42723
INFO[0015] piece download completed                      complete=true pieceIdx=2 torrent="----REDACTED---" waitingFor=0
DEBU[0015] part added to registry                        piece=2 torrentID=c92f656155d0d8e87d21471d7ea43e3ad0d42723
DEBU[0015] download completed                            name=c92f656155d0d8e87d21471d7ea43e3ad0d42723.0.0-3.REDACTED.mp4.jpg pieceCount=4 torrentID=c92f656155d0d8e87d21471d7ea43e3ad0d42723
DEBU[0015] no more input. Seems that those were all the pieces 
DEBU[0016] image extracted successfully                  name=c92f656155d0d8e87d21471d7ea43e3ad0d42723.0.0-3.REDACTED.mp4.jpg torrentID=c92f656155d0d8e87d21471d7ea43e3ad0d42723
DEBU[0016] image persisted successfully                  name=c92f656155d0d8e87d21471d7ea43e3ad0d42723.0.0-3.REDACTED.mp4.jpg torrentID=c92f656155d0d8e87d21471d7ea43e3ad0d42723
```

> Warning! You're actually downloading a torrent that _might_ contain Copyrighted material. See LICENSE.

> Note: You're actually using bitTorrent. Your current network might be filtering ports, dropping traffic, might be slow, 
> or you won't find seeders at this moment. The videos might not download.

You can see the very same result at [TorrentPreview](http://torrentpreview.com/?id=c92f656155d0d8e87d21471d7ea43e3ad0d42723)

### Get torrent information from API

See file in `docs/http/api.postman_collection.json` and you'll find the requests and descriptions of what they do. 
See it [online](https://documenter.getpostman.com/view/10093390/TWDanw8n) as well.


## Testing

### Generate mocks

```bash
docker-compose run cli make generate
````

### Run tests

Commands:

```bash
docker-compose run cli make test # Generates mocks
docker-compose run cli make test-cover # Generates mocks and coverage
docker-compose run cli make test-fast # Without mock generation
```

You can run the tests from your IDE or CLI once the mocks are generated.

## Local install

You want to start the project locally? You'll need to:

- Install some system dependencies
    - make, [go >= 1.15](https://golang.org/dl/), [ffmpeg](https://ffmpeg.org/download.html), [sqlite3](https://www.sqlite.org/download.html), [mockery](https://github.com/vektra/mockery#installation)
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
Then create the schema by issuing:

```bash
touch prevtorrent.sqlite
sqlite3 prevtorrent.sqlite < infrastructure/database/sqlite.schema.sql

```


