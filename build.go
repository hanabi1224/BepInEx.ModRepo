package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"sort"
	"strings"
)

const (
	prebuiltDir = "prebuilt"
	publishDir  = "publish"
)

func main() {
	if entries, err := ioutil.ReadDir(prebuiltDir); err == nil {
		if _, err := os.Stat(publishDir); os.IsNotExist(err) {
			os.MkdirAll(publishDir, os.ModePerm)
		}
		for _, entry := range entries {
			if entry.IsDir() {
				println(entry.Name())
				fromDir := path.Join(prebuiltDir, entry.Name())
				if err := exec.Command("cp", "-r", fromDir, publishDir).Run(); err != nil {
					println(err.Error())
				}
			}
		}
	}

	if entries, err := ioutil.ReadDir(publishDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				gameId := entry.Name()
				indexFile := path.Join(publishDir, gameId+".json")
				indexBytes := generateIndexContent(gameId, path.Join(publishDir, gameId))
				os.WriteFile(indexFile, indexBytes, fs.ModePerm)
			}
		}
	}
}

func generateIndexContent(gameId, root string) []byte {
	links := map[string]bool{}
	buckets := []Bucket{}
	if entries, err := ioutil.ReadDir(root); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				entryName := entry.Name()
				if strings.HasSuffix(entryName, ".zip") {
					url := fmt.Sprintf("/%s/%s", gameId, entryName)
					if !links[url] {
						buckets = append(buckets, Bucket{Name: entryName, Url: url})
					}
				} else if strings.HasSuffix(entryName, ".json") {
					var extraBuckets []Bucket
					if bytes, err := ioutil.ReadFile(path.Join(root, entryName)); err == nil {
						if json.Unmarshal(bytes, &extraBuckets) == nil {
							for _, eb := range extraBuckets {
								if eb.isValid() && !links[eb.Url] {
									buckets = append(buckets, eb)
									links[eb.Url] = true
								}
							}
						}
					}
				}
			}
		}
	}
	sort.Slice(buckets, func(i, j int) bool { return buckets[i].Name < buckets[j].Name })
	if jsonBytes, err := json.MarshalIndent(buckets, "", "  "); err == nil {
		return jsonBytes
	}
	return []byte{}
}

type Bucket struct {
	Name string `json:"name"`
	Url  string `json:url`
}

func (b *Bucket) isValid() bool {
	return len(b.Name) > 0 && len(b.Url) > 0
}
