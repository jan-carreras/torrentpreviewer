package main

import (
	"bytes"
	"fmt"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/storage"
	"io/ioutil"
	"time"
)

func main() {
	f, err := ioutil.ReadFile("/Users/jan/Documents/projects/langs/go/src/prevtorrent/tmp/torrents/Star Wars Episode V The Empire Strikes Back (1980) [1080p].torrent")
	if err != nil {
		panic(err)
	}
	buff := bytes.NewBuffer(f)
	metaInfo, err := metainfo.Load(buff)
	if err != nil {
		panic(err)
	}

	conf := torrent.NewDefaultClientConfig()
	conf.DefaultStorage = storage.NewBoltDB("/Users/jan/Documents/projects/langs/go/src/prevtorrent/cmd")
	conf.DisableIPv6 = true

	torrentClient, err := torrent.NewClient(conf)
	if err != nil {
		panic(err)
	}

	t, err := torrentClient.AddTorrent(metaInfo)
	if err != nil {
		panic(err)
	}
	t.DownloadAll()

	go func() {
		for {
			fmt.Println(t.NumPieces(), "completed=", t.BytesCompleted(), "missing", t.BytesMissing())
			time.Sleep(time.Second)
		}
	}()

	torrentClient.WaitAll()

	fmt.Println("Downladed...")
}
