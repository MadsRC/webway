package uuid

import (
	"encoding/base64"
	"fmt"
	"github.com/hashicorp/go-uuid"
	"github.com/stretchr/testify/require"
	"regexp"
	"testing"
)

func TestClusterID(t *testing.T) {
	KIP74Regex := regexp.MustCompile(`^[a-zA-Z0-9_\-]+$`)
	require.Regexp(t, KIP74Regex, "vPeOCWypqUOSepEvx0cbog")

	strUUID, err := uuid.GenerateUUID()
	require.NoError(t, err)

	b, err := uuid.ParseUUID(strUUID)
	require.NoError(t, err)

	clusterID := base64.RawURLEncoding.EncodeToString(b)

	require.Regexp(t, KIP74Regex, clusterID)

	decodedClusterID, err := base64.RawURLEncoding.DecodeString("vPeOCWypqUOSepEvx0cbog")
	require.NoError(t, err)

	uuidClusterID, err := uuid.FormatUUID(decodedClusterID)
	require.NoError(t, err)

	fmt.Println(uuidClusterID)
	t.Fail()
}
