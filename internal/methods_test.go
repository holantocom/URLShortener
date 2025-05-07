package internal

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

var (
	mockUrl   = "https://example.com"
	MockId    = 10000
	mockShort = "ClS"
)

func TestRedirect(t *testing.T) {
	e := echo.New()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error initializing mock db: %v", err)
	}
	defer db.Close()

	// Мокаем кэш
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT original FROM urls WHERE id = $1`)).
		WithArgs(MockId).
		WillReturnRows(sqlmock.NewRows([]string{"original"}).AddRow(mockUrl))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/:short")
	c.SetParamNames("short")
	c.SetParamValues(mockShort)
	c.Set("db", db)

	err = Redirect(c)
	if err != nil {
		t.Fatalf("Error during Redirect handler execution: %v", err)
	}

	assert.Equal(t, http.StatusFound, rec.Code)
	assert.Equal(t, mockUrl, rec.Header().Get("Location"))
}

func TestShorten(t *testing.T) {
	e := echo.New()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error initializing mock db: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO urls (original, clicks) VALUES ($1, 0) RETURNING id`)).
		WithArgs(mockUrl).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(MockId))

	e.GET("/shorten", Shorten)
	reqBody := `{"original": "https://example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.Set("db", db)

	err = Shorten(c)
	if err != nil {
		t.Fatalf("Error during Shorten handler execution: %v", err)
	}

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, rec.Body.String(), "{\"short\":\"ClS\",\"original\":\"https://example.com\",\"clicks\":0}\n")
}

func TestStats(t *testing.T) {
	e := echo.New()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error initializing mock db: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT original, clicks FROM urls WHERE id = $1`)).
		WithArgs(MockId).
		WillReturnRows(sqlmock.NewRows([]string{"original", "clicks"}).AddRow(mockUrl, 15))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/stats/:short")
	c.SetParamNames("short")
	c.SetParamValues(mockShort)
	c.Set("db", db)

	err = Stats(c)
	if err != nil {
		t.Fatalf("Error during Stats handler execution: %v", err)
	}

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), mockUrl)
	assert.Contains(t, rec.Body.String(), "15")
}

func Test_isValidURL(t *testing.T) {
	tests := []struct {
		url  string
		want bool
	}{
		{
			url:  "http://google.com",
			want: true,
		},
		{
			url:  "https//google.com",
			want: false,
		},
		{
			url:  "google.com",
			want: false,
		},
		{
			url:  "/some-path",
			want: false,
		},
		{
			url:  "http://google.com?param=value",
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			if got := isValidURL(tt.url); got != tt.want {
				t.Errorf("isValidURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
