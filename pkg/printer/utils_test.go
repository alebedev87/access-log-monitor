package printer

import (
	"testing"
)

func TestTable2dMessageNominal(t *testing.T) {
	longLine := "veeeeeeeery looooooooong liiiiiiiineeee"
	notThatLongLine := "not that long line but still"
	sepLen := 4

	tbl := NewTable2dMessage("TEST", "Column 1", "Other Column")
	tbl.AddRow("here is my text", notThatLongLine)
	tbl.AddRow(longLine, "here is other text")

	expectedLen := len(longLine) + len(notThatLongLine) + sepLen
	gotLen := tbl.Length()
	if expectedLen != gotLen {
		t.Fatalf("Expected table length %d, got %d", expectedLen, gotLen)
	}

	expectedOutput := `
---------------------------------TEST----------------------------------
               Column 1                            Other Column        
---------------------------------------    ----------------------------
here is my text                            not that long line but still
veeeeeeeery looooooooong liiiiiiiineeee              here is other text
`
	gotOutput := tbl.Format()
	if expectedOutput != gotOutput {
		t.Fatalf("Expected table output %s, got %s", expectedOutput, gotOutput)
	}

	// enlarge case
	tbl.Enlarge(tbl.Length() + 10)

	expectedLen += 10
	gotLen = tbl.Length()
	if expectedLen != gotLen {
		t.Fatalf("Expected table length %d, got %d", expectedLen, gotLen)
	}

	expectedEnlarge := `
--------------------------------------TEST---------------------------------------
               Column 1                                 Other Column             
---------------------------------------    --------------------------------------
here is my text                                      not that long line but still
veeeeeeeery looooooooong liiiiiiiineeee                        here is other text
`
	gotEnlarge := tbl.Format()
	if expectedEnlarge != gotEnlarge {
		t.Fatalf("Expected table output %s, got %s", expectedEnlarge, gotEnlarge)
	}
}
