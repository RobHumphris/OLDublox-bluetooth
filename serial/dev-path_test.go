package serial

import (
	"fmt"
	"testing"
)

func TestGetFTDIDevPath(t *testing.T) {
	path, err := GetFTDIDevPaths()
	if err != nil {
		t.Errorf("GetFTDIDevPath failed: %v\n", err)
	}
	fmt.Println("Path returned: ", path)
}
