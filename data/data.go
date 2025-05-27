package data

import (
	"encoding/json"
	"io"
	"net/http"
	"os/exec"
	"strconv"
)

const ROOT = "https://hacker-news.firebaseio.com/v0/"

type Item struct {
	Id          uint32
	Deleted     bool
	Type        string
	By          string
	Time        uint64
	Dead        bool
	Parent      uint32
	Poll        uint32
	Kids        []uint32
	Url         string
	Score       uint16
	Title       string
	Parts       []uint32
	Descendants uint32
	Text        string
}

func FetchResource[T any](route string, result T) error {
	resp, err := http.Get(ROOT + route)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, io_err := io.ReadAll(resp.Body)
	if io_err != nil {
		return io_err
	}

	err = json.Unmarshal(body, result)
	if err != nil {
		return err
	}

	return nil
}

func FetchItems(item_ids []uint32) ([]Item, error) {
	items := make([]Item, len(item_ids))
	for i, item_id := range item_ids {
		// TODO: Refactor to be async fetch
		var item Item
		err := FetchResource("/item/"+strconv.FormatUint(uint64(item_id), 10)+".json", &item)
		if err != nil {
			return nil, err
		}

		items[i] = item
	}

	return items, nil
}

func FetchUrlContent(url string) (string, error) {
	var content string

	cmd := exec.Command("lynx", "-dump", url)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}

	if err := cmd.Start(); err != nil {
		return "", err
	}

	bytes, err := io.ReadAll(stdout)
	if err != nil {
		return "", err
	}

	content = string(bytes)

	if err := cmd.Wait(); err != nil {
		return "", err
	}

	return content, nil

}
