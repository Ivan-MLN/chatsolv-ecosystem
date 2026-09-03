package knowledge

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

func TestListKnowledgeUsesPublicJSONContract(t *testing.T) {
	createdAt := time.Date(2026, time.August, 25, 12, 0, 0, 0, time.UTC)
	repository := &fakeRepository{
		listed: []SourceRecord{{
			ID: "source-id", WorkspaceID: "workspace-id", SecondBrainID: "brain-id",
			Type: "document", Title: "Katalog Produk", Status: "queued",
			ObjectKey: "private/object-key", Checksum: "private-checksum",
			CreatedAt: createdAt,
		}},
	}
	app := fiber.New()
	app.Get("/", NewHandler(NewService(repository, nil, 0, 0)).List)

	result, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/?workspace_id=workspace-id", nil))

	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, result.StatusCode)
	var payload struct {
		Data []map[string]any `json:"data"`
	}
	require.NoError(t, json.NewDecoder(result.Body).Decode(&payload))
	require.Len(t, payload.Data, 1)
	require.Equal(t, "source-id", payload.Data[0]["id"])
	require.Equal(t, "Katalog Produk", payload.Data[0]["title"])
	require.Equal(t, "document", payload.Data[0]["type"])
	require.Equal(t, "queued", payload.Data[0]["status"])
	require.Equal(t, createdAt.Format(time.RFC3339), payload.Data[0]["created_at"])
	require.NotContains(t, payload.Data[0], "WorkspaceID")
	require.NotContains(t, payload.Data[0], "ObjectKey")
	require.NotContains(t, payload.Data[0], "Checksum")
}

func TestCreateTextReadsSnakeCaseWorkspaceID(t *testing.T) {
	repository := &fakeRepository{access: WorkspaceAccess{Role: "owner", SecondBrainID: "brain-id"}}
	app := fiber.New()
	app.Post("/", NewHandler(NewService(repository, &fakeStore{}, 1024, 10)).CreateText)
	request := httptest.NewRequest(fiber.MethodPost, "/", strings.NewReader(`{
		"workspace_id":"workspace-id",
		"title":"SOP Retur",
		"content":"Barang dapat diretur dalam tiga hari."
	}`))
	request.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	result, err := app.Test(request)

	require.NoError(t, err)
	require.Equal(t, fiber.StatusAccepted, result.StatusCode)
	require.Equal(t, "workspace-id", repository.created.WorkspaceID)
}

func TestCreateFAQAcceptsDocumentedFAQArray(t *testing.T) {
	repository := &fakeRepository{access: WorkspaceAccess{Role: "owner", SecondBrainID: "brain-id"}}
	app := fiber.New()
	app.Post("/", NewHandler(NewService(repository, &fakeStore{}, 1024, 10)).CreateFAQ)
	request := httptest.NewRequest(fiber.MethodPost, "/", strings.NewReader(`{
		"workspace_id":"workspace-id",
		"title":"FAQ Pengiriman",
		"faqs":[
			{"question":"Berapa lama pengiriman?","answer":"Dua sampai tiga hari."},
			{"question":"Ada nomor resi?","answer":"Ya, resi dikirim otomatis."}
		]
	}`))
	request.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	result, err := app.Test(request)

	require.NoError(t, err)
	require.Equal(t, fiber.StatusAccepted, result.StatusCode)
	require.Equal(t, "workspace-id", repository.created.WorkspaceID)
	require.Contains(t, repository.created.ObjectKey, "Berapa lama pengiriman?")
	require.Contains(t, repository.created.ObjectKey, "Ada nomor resi?")
}

func TestUpdateKnowledgeRejectsMalformedJSON(t *testing.T) {
	app := fiber.New()
	app.Patch("/:id", NewHandler(nil).Update)
	request := httptest.NewRequest(fiber.MethodPatch, "/source-id", strings.NewReader(`{"title":`))
	request.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	result, err := app.Test(request)

	require.NoError(t, err)
	require.Equal(t, fiber.StatusBadRequest, result.StatusCode)
}

func TestUpdateKnowledgeRequiresJSONContentType(t *testing.T) {
	app := fiber.New()
	app.Patch("/:id", NewHandler(nil).Update)
	request := httptest.NewRequest(fiber.MethodPatch, "/source-id", strings.NewReader(`{"title":"Updated"}`))

	result, err := app.Test(request)

	require.NoError(t, err)
	require.Equal(t, fiber.StatusBadRequest, result.StatusCode)
}
