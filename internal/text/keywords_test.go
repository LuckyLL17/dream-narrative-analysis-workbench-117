package text

import "testing"

// TestKeywordTitleBonusOutranksBodyFrequency guards the product rule: a
// word that appears in the title earns semantic weight that lets it rank
// ahead of a word that appears more often but only in the body. This is the
// "钟 should precede 雨" requirement even when the body term is more
// frequent (e.g. title twice vs body seven times).
func TestKeywordTitleBonusOutranksBodyFrequency(t *testing.T) {
	title := "时钟响了"
	content := "大雨 大雨 大雨 大雨 大雨 大雨 大雨"

	keywords := Keywords(title, content)
	// Compute positions of each word's first appearance in the ranked list.
	positions := map[string]int{}
	for i, k := range keywords {
		if _, seen := positions[k.Word]; !seen {
			positions[k.Word] = i
		}
	}

	clock, ok := positions["时钟"]
	if !ok {
		t.Fatalf("missing title term 时钟 in keywords: %v", keywords)
	}
	rain, ok := positions["大雨"]
	if !ok {
		t.Fatalf("missing body term 大雨 in keywords: %v", keywords)
	}
	if clock > rain {
		t.Fatalf("title term 时钟 (pos %d) should rank ahead of body term 大雨 (pos %d); keywords=%v",
			clock, rain, keywords)
	}
}

func TestKeywordBonusValue(t *testing.T) {
	if got := titleKeywordBonus("时钟", "时钟响了"); got != titleKeywordWeight {
		t.Fatalf("titleKeywordBonus = %v, want %v", got, titleKeywordWeight)
	}
	if got := titleKeywordBonus("大雨", "时钟响了"); got != 0 {
		t.Fatalf("titleKeywordBonus for non-title term = %v, want 0", got)
	}
	if got := titleKeywordBonus("时钟", ""); got != 0 {
		t.Fatalf("titleKeywordBonus with empty title = %v, want 0", got)
	}
}
