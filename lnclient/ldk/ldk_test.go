package ldk

import (
	"testing"

	"github.com/getAlby/hub/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetVssNodeIdentifier(t *testing.T) {
	mnemonic := "thought turkey ask pottery head say catalog desk pledge elbow naive mimic"
	expectedVssNodeIdentifier := "751636"

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestServiceWithMnemonic(mnemonic, "123")
	require.NoError(t, err)

	vssNodeIdentifier, err := GetVssNodeIdentifier(svc.Keys)
	require.NoError(t, err)

	assert.Equal(t, expectedVssNodeIdentifier, vssNodeIdentifier)
}
func TestGetVssNodeIdentifier2(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	expectedVssNodeIdentifier := "770256"

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestServiceWithMnemonic(mnemonic, "123")
	require.NoError(t, err)

	vssNodeIdentifier, err := GetVssNodeIdentifier(svc.Keys)
	require.NoError(t, err)

	assert.Equal(t, expectedVssNodeIdentifier, vssNodeIdentifier)
}
