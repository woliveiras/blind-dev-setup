package target

import "testing"

func TestDecodeCandidatesNormalizesAndSortsPowerShellResult(t *testing.T) {
	input := []byte(`[
		{"DiskNumber":3,"Model":" USB Flash Disk ","DriveLetter":"f:","Label":"ESTUDOS","FileSystem":"ntfs","SizeBytes":34359738368,"FreeBytes":21474836480},
		{"DiskNumber":2,"Model":"Kingston DataTraveler","DriveLetter":"E","Label":"PROGRAMAÇÃO","FileSystem":"NTFS","SizeBytes":17179869184,"FreeBytes":10737418240}
	]`)

	candidates, err := decodeCandidates(input)
	if err != nil {
		t.Fatalf("decodeCandidates() error = %v", err)
	}
	if len(candidates) != 2 {
		t.Fatalf("len(candidates) = %d", len(candidates))
	}
	if candidates[0].DiskNumber != 2 || candidates[0].RootPath() != "E:\\" {
		t.Errorf("candidates[0] = %#v", candidates[0])
	}
	if candidates[0].Label != "PROGRAMAÇÃO" {
		t.Errorf("candidates[0].Label = %q", candidates[0].Label)
	}
	if candidates[1].DiskNumber != 3 || candidates[1].RootPath() != "F:\\" {
		t.Errorf("candidates[1] = %#v", candidates[1])
	}
	if candidates[1].Model != "USB Flash Disk" || candidates[1].FileSystem != "NTFS" {
		t.Errorf("candidates[1] = %#v", candidates[1])
	}
}

func TestDecodeCandidatesRejectsUnexpectedPowerShellOutput(t *testing.T) {
	if _, err := decodeCandidates([]byte(`not json`)); err == nil {
		t.Fatal("decodeCandidates() error = nil")
	}
}
