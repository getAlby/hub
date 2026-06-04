//go:build windows

package bark

// The bark FFI static library (libbark_ffi_go.a) is built for the GNU/mingw
// target and embeds Rust's getrandom/ring code, which references Windows CNG
// symbols such as BCryptGenRandom. The upstream bark bindings only link
// -lbark_ffi_go, so the mingw linker fails with an undefined reference to
// BCryptGenRandom (and related Windows system symbols). cgo merges LDFLAGS
// across packages, so we supply the missing Windows system libraries here
// without modifying the vendored bark module.
//
// Remove once bark's bindings link these upstream (the list comes from their
// Rust crate's native-static-libs output).

/*
#cgo windows,amd64 LDFLAGS: -lbcrypt -lntdll -luserenv -lws2_32 -lcrypt32 -lncrypt -lsecur32 -ladvapi32
*/
import "C"
