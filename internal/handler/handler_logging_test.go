package handler_test

import (
	"bytes"
	"log"
	"strings"
	"testing"

	"familyshare/internal/config"
	"familyshare/internal/handler"
	"familyshare/internal/storage"
	"familyshare/internal/testutil"
	"familyshare/web"
)

func TestHandlerTemplateLogging_DebugFlag(t *testing.T) {
	db, _, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	store := storage.New(t.TempDir())

	origOutput := log.Writer()
	origFlags := log.Flags()
	defer func() {
		log.SetOutput(origOutput)
		log.SetFlags(origFlags)
	}()

	buf := &bytes.Buffer{}
	log.SetOutput(buf)
	log.SetFlags(0)

	handler.New(db, store, web.EmbedFS, &config.Config{Debug: false}, nil)
	if strings.Contains(buf.String(), "template files to parse") || strings.Contains(buf.String(), "loaded templates") {
		t.Fatalf("expected no verbose template logs when debug disabled, got: %s", buf.String())
	}

	buf.Reset()
	handler.New(db, store, web.EmbedFS, &config.Config{Debug: true}, nil)
	if !strings.Contains(buf.String(), "loaded templates") {
		t.Fatalf("expected verbose template logs when debug enabled, got: %s", buf.String())
	}
}
