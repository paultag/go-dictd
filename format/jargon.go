package format

import (
	"bufio"
	"log"
	"os"
	"strings"

	"github.com/paultag/go-dictd/dictd"
)

func ParseJargonFormat(path string) []*dictd.Definition {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	var def = ""
	var word = ""
	var defs = make([]*dictd.Definition, 0)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, ":") {
			tokens := strings.SplitN(line, ":", 3)
			if len(tokens) == 3 {
				if word != "" {
					defs = append(defs, &dictd.Definition{
						Word:       word,
						Definition: def,
					})
				}
				word = strings.Trim(tokens[1], " \t\n\r")
				def = strings.Trim(tokens[2], " \t\n\r")
				continue
			}
		}
		def = def + "\r\n" + line
	}

	defs = append(defs, &dictd.Definition{
		Word:       word,
		Definition: def,
	})

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return defs
}
