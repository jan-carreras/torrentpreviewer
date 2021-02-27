// +build integration

package http

import (
	"database/sql"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"prevtorrent/internal/platform/container"
	"prevtorrent/internal/platform/services"
	"testing"
)

// Database is automatically created but we should remove it afterwards
// It should be created at testdata, at least

func Test_GetTorrent_Found(t *testing.T) {
	c, err := container.NewTestingContainer()
	require.NoError(t, err)

	defer removeDB(c.Config().SqlitePath)

	createDB(t, c.GetSQLDatabase())
	populateDB(t, c.GetSQLDatabase())

	s, err := services.NewServices(c)
	require.NoError(t, err)

	server := NewServer(s)

	ts := httptest.NewServer(setupServer(server))
	// Shut down the server and block until all requests have gone through
	defer ts.Close()

	torrentID := "cb84ccc10f296df72d6c40ba7a07c178a4323a14"
	// Make a request to our server with the {base url}/ping
	resp, err := http.Get(fmt.Sprintf("%s/torrent/%v", ts.URL, torrentID))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	val, ok := resp.Header["Content-Type"]

	assert.True(t, ok)
	assert.Equal(t, "application/json; charset=utf-8", val[0])

	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	expectedRsp, err := ioutil.ReadFile("./testdata/torrent.cb84.response.json")
	require.NoError(t, err)

	assert.JSONEq(t, string(expectedRsp), string(body))
}

func Test_GetTorrent_NotFound(t *testing.T) {
	c, err := container.NewTestingContainer()
	require.NoError(t, err)

	defer removeDB(c.Config().SqlitePath)

	createDB(t, c.GetSQLDatabase())
	populateDB(t, c.GetSQLDatabase())

	s, err := services.NewServices(c)
	require.NoError(t, err)

	server := NewServer(s)

	ts := httptest.NewServer(setupServer(server))
	// Shut down the server and block until all requests have gone through
	defer ts.Close()

	torrentID := "cb84ccc10f296df72d6c40ba7a07c178a4323000"
	// Make a request to our server with the {base url}/ping
	resp, err := http.Get(fmt.Sprintf("%s/torrent/%v", ts.URL, torrentID))
	require.NoError(t, err)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	val, ok := resp.Header["Content-Type"]

	assert.True(t, ok)
	assert.Equal(t, "application/json; charset=utf-8", val[0])
}

func createDB(t *testing.T, db *sql.DB) {
	createSchemaDDL, err := ioutil.ReadFile("./testdata/schema.sql")
	require.NoError(t, err)

	_, err = db.Exec(string(createSchemaDDL))
	require.NoError(t, err)
}

func populateDB(t *testing.T, db *sql.DB) {
	insertDataStatements, err := ioutil.ReadFile("./testdata/testdata.sql")
	require.NoError(t, err)

	_, err = db.Exec(string(insertDataStatements))
	require.NoError(t, err)
}

func removeDB(dbPath string) {
	os.Remove(dbPath)
}
