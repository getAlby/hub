package ldk

import (
	"testing"

	"github.com/getAlby/hub/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetVssNodeIdentifier(t *testing.T) {
	mnemonic := "thought turkey ask pottery head say catalog desk pledge elbow naive mimic"
	expectedPubkey := "0488b2"

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestServiceWithMnemonic(mnemonic, "123")
	require.NoError(t, err)

	vssNodeIdentifier, err := GetVssNodeIdentifier(svc.Keys)
	require.NoError(t, err)

	assert.Equal(t, expectedPubkey, vssNodeIdentifier)
}
