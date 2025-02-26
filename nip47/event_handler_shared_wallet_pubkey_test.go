package nip47

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getAlby/hub/tests"
)

func TestHandleResponse_SharedWalletPubkey_Nip04_WithPermission(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	doTestHandleResponse_WithPermission(t, svc, tests.CreateAppWithSharedWalletPubkey, "nip04")
}

func TestHandleResponse_SharedWalletPubkey_Nip44_WithPermission(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	doTestHandleResponse_WithPermission(t, svc, tests.CreateAppWithSharedWalletPubkey, "nip44_v2")
}

func TestHandleResponse_SharedWalletPubkey_Nip04_DuplicateRequest(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	doTestHandleResponse_DuplicateRequest(t, svc, tests.CreateAppWithSharedWalletPubkey, "nip04")
}

func TestHandleResponse_SharedWalletPubkey_Nip44_DuplicateRequest(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	doTestHandleResponse_DuplicateRequest(t, svc, tests.CreateAppWithSharedWalletPubkey, "nip44_v2")
}

func TestHandleResponse_SharedWalletPubkey_Nip04_NoPermission(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	doTestHandleResponse_NoPermission(t, svc, tests.CreateAppWithSharedWalletPubkey, "nip04")
}

func TestHandleResponse_SharedWalletPubkey_Nip44_NoPermission(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	doTestHandleResponse_NoPermission(t, svc, tests.CreateAppWithSharedWalletPubkey, "nip44_v2")
}

func TestHandleResponse_SharedWalletPubkey_Nip04_IncorrectPubkey(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	doTestHandleResponse_IncorrectPubkey(t, svc, tests.CreateAppWithSharedWalletPubkey, "nip04")
}

func TestHandleResponse_SharedWalletPubkey_Nip44_IncorrectPubkey(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	doTestHandleResponse_IncorrectPubkey(t, svc, tests.CreateAppWithSharedWalletPubkey, "nip44_v2")
}

func TestHandleResponse_SharedWalletPubkey_Nip04_OldRequestForPayment(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	doTestHandleResponse_OldRequestForPayment(t, svc, tests.CreateAppWithSharedWalletPubkey, "nip04")
}

func TestHandleResponse_SharedWalletPubkey_Nip44_OldRequestForPayment(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	doTestHandleResponse_OldRequestForPayment(t, svc, tests.CreateAppWithSharedWalletPubkey, "nip44_v2")
}
