package dictd

import (
	"testing"
)

func TestTokenizingSimple(t *testing.T) {
	query := `ONE FISH TWO FISH RED FISH BLUE FISH`
	tokens, err := tokenizeLine(query)

	if err != nil {
		t.Errorf("Error tokenizing query " + query)
	}

	if len(tokens) != 8 {
		t.Errorf("Bad token count out - didn't get 8")
	}
}

func TestTokenizingString(t *testing.T) {
	query := `ONE "FISH TWO FISH" RED FISH BLUE FISH`
	tokens, err := tokenizeLine(query)

	if err != nil {
		t.Errorf("Error tokenizing query " + query)
	}

	if len(tokens) != 6 {
		t.Errorf("Bad token count out - didn't get 6")
	}
}

func TestTokenizingEscape(t *testing.T) {
	query := `ONE "FISH \"TWO FISH" RED FISH BLUE FISH`
	tokens, err := tokenizeLine(query)

	if err != nil {
		t.Errorf("Error tokenizing query " + query)
	}

	if len(tokens) != 6 {
		t.Errorf("Bad token count out - didn't get 6")
	}

	if tokens[1] != "FISH \"TWO FISH" {
		t.Errorf("Bad escape handling")
	}
}

func TestTokenizingQuoteEscape(t *testing.T) {
	query := `ONE 'FISH \'TWO FISH' RED FISH BLUE FISH`
	tokens, err := tokenizeLine(query)

	if err != nil {
		t.Errorf("Error tokenizing query " + query)
	}

	if len(tokens) != 6 {
		t.Errorf("Bad token count out - didn't get 6")
	}

	if tokens[1] != "FISH 'TWO FISH" {
		t.Errorf("Bad escape handling")
	}
}
