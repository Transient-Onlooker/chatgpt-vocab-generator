package main

import (
	"regexp"
	"strings"
)

type VocabPair struct {
	Word     string
	Meanings []string
}

// parseVocabBlock parses a block of text in "word = meaning1, meaning2; meaning3" format.
func parseVocabBlock(vocabBlock string) []VocabPair {
	var pairs []VocabPair
	re := regexp.MustCompile(`[;,]`)

	for _, raw := range strings.Split(vocabBlock, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) < 2 {
			continue
		}

		word := strings.TrimSpace(parts[0])
		meaningsRaw := strings.TrimSpace(parts[1])

		senses := re.Split(meaningsRaw, -1)
		var cleanSenses []string
		for _, s := range senses {
			s = strings.TrimSpace(s)
			if s != "" {
				cleanSenses = append(cleanSenses, s)
			}
		}

		if word != "" && len(cleanSenses) > 0 {
			pairs = append(pairs, VocabPair{Word: word, Meanings: cleanSenses})
		}
	}
	return pairs
}
