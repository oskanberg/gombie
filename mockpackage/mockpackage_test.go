package mockpackage_test

import (
	"testing"

	"github.com/oskanberg/gombie/mockpackage"
)

func TestReturnTrueReturnsTrue(t *testing.T) {
	result := mockpackage.ReturnTrue()
	if !result {
		t.Error("Returned False!")
	}
}
