package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestHealthcheck(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	want := ResponseMsg{
		Message: "I'm fine, Thank!",
	}

	if assert.NoError(t, Healthcheck(c)) {
		var got ResponseMsg
		err := json.Unmarshal([]byte(rec.Body.String()), &got)

		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		equal := reflect.DeepEqual(want, got)

		if !equal {
			assert.Fail(t, fmt.Sprintf("expected %v, but got %v", want, got))
		}
	}
}
