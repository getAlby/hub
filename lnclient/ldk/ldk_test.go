package ldk

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/getAlby/hub/tests"
)

func TestGetVssNodeIdentifier(t *testing.T) {
	mnemonic := "thought turkey ask pottery head say catalog desk pledge elbow naive mimic"
	expectedVssNodeIdentifier := "751636"

	svc, err := tests.CreateTestServiceWithMnemonic(mnemonic, "123")
	require.NoError(t, err)
	defer svc.Remove()

	vssNodeIdentifier, err := GetVssNodeIdentifier(svc.Keys)
	require.NoError(t, err)

	assert.Equal(t, expectedVssNodeIdentifier, vssNodeIdentifier)
}
func TestGetVssNodeIdentifier2(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	expectedVssNodeIdentifier := "770256"

	svc, err := tests.CreateTestServiceWithMnemonic(mnemonic, "123")
	require.NoError(t, err)
	defer svc.Remove()

	vssNodeIdentifier, err := GetVssNodeIdentifier(svc.Keys)
	require.NoError(t, err)

	assert.Equal(t, expectedVssNodeIdentifier, vssNodeIdentifier)
}
