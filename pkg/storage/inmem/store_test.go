package inmem_test

import (
	"io"
	"io/fs"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cardil/qe-clusterlogging/pkg/clusterlogging"
	"github.com/cardil/qe-clusterlogging/pkg/kubernetes"
	"github.com/cardil/qe-clusterlogging/pkg/storage/inmem"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore(t *testing.T) {
	storage := inmem.NewStore()
	testmsg := map[string][]string{
		"foo": {
			"lorem ipsum",
			"dolor sit amet",
			"consectetur adipiscing elit",
		},
		"bar": {
			"Sed ut perspiciatis unde omnis",
			"iste natus error",
			"sit voluptatem accusantium doloremque laudantium",
			"totam rem aperiam",
		},
	}
	for name, texts := range testmsg {
		for _, text := range texts {
			err := storage.Store(&clusterlogging.Message{
				Timestamp: time.Now(),
				Message:   text,
				ContainerInfo: kubernetes.ContainerInfo{
					ContainerName:  "user",
					ContainerImage: "example.org/foo",
					NamespaceName:  "default",
					PodName:        name,
				},
			})
			require.NoError(t, err)
		}
	}

	stats := storage.Stats()
	assert.Len(t, stats, 2)
	wantStats := map[string]int{
		"foo": 3,
		"bar": 4,
	}
	gotStats := map[string]int{}
	for _, stat := range stats {
		gotStats[stat.PodName] = stat.MessageCount
	}
	assert.Equal(t, wantStats, gotStats)

	artifacts := storage.Download()
	rootdir := filepath.ToSlash(t.TempDir())
	for filename, readerFn := range artifacts {
		fp := path.Join(rootdir, filename)
		reader := readerFn()
		dir := path.Dir(fp)
		require.NoError(t, os.MkdirAll(dir, 0o755))
		f, err := os.Create(fp)
		require.NoError(t, err)
		writer := io.MultiWriter(f) // to avoid using ReaderFrom interface
		buf := make([]byte, rand.Intn(62)+2)
		_, err = io.CopyBuffer(writer, reader, buf)
		require.NoError(t, err)
	}
	wantFiles := map[string]string{
		"default/foo/user.json": `{
  "container_name": "user",
  "container_image": "example.org/foo",
  "namespace_name": "default",
  "pod_name": "foo"
}`,
		"default/foo/user.log": `lorem ipsum
dolor sit amet
consectetur adipiscing elit
`,
		"default/bar/user.json": `{
  "container_name": "user",
  "container_image": "example.org/foo",
  "namespace_name": "default",
  "pod_name": "bar"
}`,
		"default/bar/user.log": `Sed ut perspiciatis unde omnis
iste natus error
sit voluptatem accusantium doloremque laudantium
totam rem aperiam
`,
	}
	gotFiles := make(map[string]string, len(wantFiles))
	require.NoError(t, filepath.WalkDir(rootdir, func(pth string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		relpth := strings.TrimPrefix(filepath.ToSlash(pth), rootdir+"/")
		cont, rerr := os.ReadFile(pth)
		if rerr != nil {
			return rerr
		}
		gotFiles[relpth] = string(cont)
		return nil
	}))
	assert.Equal(t, wantFiles, gotFiles)
}
