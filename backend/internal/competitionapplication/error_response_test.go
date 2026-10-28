package competitionapplication

import (
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestInternalErrorDoesNotLeakCause(t *testing.T) {
	w := httptest.NewRecorder()
	handleError(w, errors.New("SQL secret details"))
	if w.Code != 500 || !strings.Contains(w.Body.String(), `"code":"internal_error"`) || strings.Contains(w.Body.String(), "SQL secret") {
		t.Fatalf("unsafe response: %d %s", w.Code, w.Body.String())
	}
}
