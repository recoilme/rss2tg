package rss2tg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_RssList(t *testing.T) {
	lines, err := RssList("../rss.txt")

	assert.NoError(t, err)
	assert.Equal(t, true, len(lines) >= 10)
}

func Test_FeedItems(t *testing.T) {
	lines, err := RssList("../rss.txt")

	assert.NoError(t, err)
	words, err := WordsList("../words.txt")
	assert.NoError(t, err)

	assert.Equal(t, true, len(lines) >= 10)
	FeedItems(lines[0], words)
}

func Test_WordsList(t *testing.T) {
	words, err := WordsList("../words.txt")

	assert.NoError(t, err)
	//fmt.Printf("%+v",lines)
	assert.Equal(t, true, len(words) >= 10)
}
