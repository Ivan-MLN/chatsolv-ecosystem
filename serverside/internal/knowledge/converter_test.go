package knowledge

import (
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestTextConversionIncludesRequiredFrontmatter(t *testing.T) {
	notes, err := Convert(Source{ID: "src_1", WorkspaceID: "wsp_1", Type: "text", Title: "Refund Policy"}, strings.NewReader("Refund within 7 days"), "text/plain")
	require.NoError(t, err)
	require.Len(t, notes, 1)
	require.Contains(t, notes[0].Content, "workspace_id: wsp_1")
	require.Contains(t, notes[0].Content, "source_id: src_1")
	require.Equal(t, "knowledge/refund-policy.md", notes[0].Path)
}
func TestCSVConversionCreatesOneProductPerRow(t *testing.T) {
	csv := "sku,name,price,description\nSKU1,Produk A,149000,Deskripsi A\nSKU2,Produk B,99000,Deskripsi B\n"
	notes, err := Convert(Source{ID: "src", WorkspaceID: "wsp", Type: "document", Title: "Catalog"}, strings.NewReader(csv), "text/csv")
	require.NoError(t, err)
	require.Len(t, notes, 2)
	require.Equal(t, "products/sku1.md", notes[0].Path)
	require.Contains(t, notes[0].Content, "sku: SKU1")
}
func TestUnsupportedDocumentFailsPermanently(t *testing.T) {
	_, err := Convert(Source{}, strings.NewReader("x"), "application/x-executable")
	require.ErrorIs(t, err, ErrUnsupportedType)
}
