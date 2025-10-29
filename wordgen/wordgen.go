package wordgen

import (
	"fmt"
	"math/rand"
	"time"
)

var adjectives = []string{
	"happy", "bright", "swift", "calm", "bold",
	"clever", "gentle", "mighty", "noble", "quiet",
	"rapid", "steady", "vivid", "warm", "wild",
	"amber", "azure", "crystal", "golden", "silver",
	"cosmic", "electric", "mystic", "sonic", "lunar",
}

var nouns = []string{
	"mountain", "river", "ocean", "forest", "meadow",
	"thunder", "lightning", "sunset", "sunrise", "moon",
	"star", "comet", "nebula", "galaxy", "planet",
	"falcon", "eagle", "wolf", "tiger", "dragon",
	"phoenix", "wind", "storm", "wave", "flame",
}

// Generate creates a random word combination in format YYYYMMDD_word1_word2_word3
func Generate() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	
	datePrefix := time.Now().Format("20060102")
	
	word1 := adjectives[r.Intn(len(adjectives))]
	word2 := nouns[r.Intn(len(nouns))]
	word3 := nouns[r.Intn(len(nouns))]
	
	// Make sure word2 and word3 are different
	for word3 == word2 {
		word3 = nouns[r.Intn(len(nouns))]
	}
	
	return fmt.Sprintf("%s_%s_%s_%s", datePrefix, word1, word2, word3)
}

