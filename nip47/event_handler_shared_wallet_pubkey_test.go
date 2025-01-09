package nip47

import (
	"testing"

	"github.com/getAlby/hub/tests"
	"github.com/stretchr/testify/require"
)

func TestHandleResponse_SharedWalletPubkey_Nip04_WithPermission(t *testing.T) {
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	doTestHandleResponse_WithPermission(t, svc, tests.CreateAppWithSharedWalletPubkey, "0.0")
}

func TestHandleResponse_SharedWalletPubkey_Nip44_WithPermission(t *testing.T) {
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	doTestHandleResponse_WithPermission(t, svc, tests.CreateAppWithSharedWalletPubkey, "1.0")
}

func TestHandleResponse_SharedWalletPubkey_Nip04_DuplicateRequest(t *testing.T) {
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	doTestHandleResponse_DuplicateRequest(t, svc, tests.CreateAppWithSharedWalletPubkey, "0.0")
}

func TestHandleResponse_SharedWalletPubkey_Nip44_DuplicateRequest(t *testing.T) {
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	doTestHandleResponse_DuplicateRequest(t, svc, tests.CreateAppWithSharedWalletPubkey, "1.0")
}

func TestHandleResponse_SharedWalletPubkey_Nip04_NoPermission(t *testing.T) {
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	doTestHandleResponse_NoPermission(t, svc, tests.CreateAppWithSharedWalletPubkey, "0.0")
}

func TestHandleResponse_SharedWalletPubkey_Nip44_NoPermission(t *testing.T) {
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	doTestHandleResponse_NoPermission(t, svc, tests.CreateAppWithSharedWalletPubkey, "1.0")
}

func TestHandleResponse_SharedWalletPubkey_Nip04_IncorrectPubkey(t *testing.T) {
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	doTestHandleResponse_IncorrectPubkey(t, svc, tests.CreateAppWithSharedWalletPubkey, "0.0")
}

func TestHandleResponse_SharedWalletPubkey_Nip44_IncorrectPubkey(t *testing.T) {
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	doTestHandleResponse_IncorrectPubkey(t, svc, tests.CreateAppWithSharedWalletPubkey, "1.0")
}

func TestHandleResponse_SharedWalletPubkey_Nip04_OldRequestForPayment(t *testing.T) {
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	doTestHandleResponse_OldRequestForPayment(t, svc, tests.CreateAppWithSharedWalletPubkey, "0.0")
}

func TestHandleResponse_SharedWalletPubkey_Nip44_OldRequestForPayment(t *testing.T) {
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	doTestHandleResponse_OldRequestForPayment(t, svc, tests.CreateAppWithSharedWalletPubkey, "1.0")
}
