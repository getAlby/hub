package ldk_node

// #include <ldk_node.h>
import "C"

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"runtime"
	"sync/atomic"
	"unsafe"
)

type RustBuffer = C.RustBuffer

type RustBufferI interface {
	AsReader() *bytes.Reader
	Free()
	ToGoBytes() []byte
	Data() unsafe.Pointer
	Len() int
	Capacity() int
}

func RustBufferFromExternal(b RustBufferI) RustBuffer {
	return RustBuffer{
		capacity: C.int(b.Capacity()),
		len:      C.int(b.Len()),
		data:     (*C.uchar)(b.Data()),
	}
}

func (cb RustBuffer) Capacity() int {
	return int(cb.capacity)
}

func (cb RustBuffer) Len() int {
	return int(cb.len)
}

func (cb RustBuffer) Data() unsafe.Pointer {
	return unsafe.Pointer(cb.data)
}

func (cb RustBuffer) AsReader() *bytes.Reader {
	b := unsafe.Slice((*byte)(cb.data), C.int(cb.len))
	return bytes.NewReader(b)
}

func (cb RustBuffer) Free() {
	rustCall(func(status *C.RustCallStatus) bool {
		C.ffi_ldk_node_rustbuffer_free(cb, status)
		return false
	})
}

func (cb RustBuffer) ToGoBytes() []byte {
	return C.GoBytes(unsafe.Pointer(cb.data), C.int(cb.len))
}

func stringToRustBuffer(str string) RustBuffer {
	return bytesToRustBuffer([]byte(str))
}

func bytesToRustBuffer(b []byte) RustBuffer {
	if len(b) == 0 {
		return RustBuffer{}
	}
	// We can pass the pointer along here, as it is pinned
	// for the duration of this call
	foreign := C.ForeignBytes{
		len:  C.int(len(b)),
		data: (*C.uchar)(unsafe.Pointer(&b[0])),
	}

	return rustCall(func(status *C.RustCallStatus) RustBuffer {
		return C.ffi_ldk_node_rustbuffer_from_bytes(foreign, status)
	})
}

type BufLifter[GoType any] interface {
	Lift(value RustBufferI) GoType
}

type BufLowerer[GoType any] interface {
	Lower(value GoType) RustBuffer
}

type FfiConverter[GoType any, FfiType any] interface {
	Lift(value FfiType) GoType
	Lower(value GoType) FfiType
}

type BufReader[GoType any] interface {
	Read(reader io.Reader) GoType
}

type BufWriter[GoType any] interface {
	Write(writer io.Writer, value GoType)
}

type FfiRustBufConverter[GoType any, FfiType any] interface {
	FfiConverter[GoType, FfiType]
	BufReader[GoType]
}

func LowerIntoRustBuffer[GoType any](bufWriter BufWriter[GoType], value GoType) RustBuffer {
	// This might be not the most efficient way but it does not require knowing allocation size
	// beforehand
	var buffer bytes.Buffer
	bufWriter.Write(&buffer, value)

	bytes, err := io.ReadAll(&buffer)
	if err != nil {
		panic(fmt.Errorf("reading written data: %w", err))
	}
	return bytesToRustBuffer(bytes)
}

func LiftFromRustBuffer[GoType any](bufReader BufReader[GoType], rbuf RustBufferI) GoType {
	defer rbuf.Free()
	reader := rbuf.AsReader()
	item := bufReader.Read(reader)
	if reader.Len() > 0 {
		// TODO: Remove this
		leftover, _ := io.ReadAll(reader)
		panic(fmt.Errorf("Junk remaining in buffer after lifting: %s", string(leftover)))
	}
	return item
}

func rustCallWithError[U any](converter BufLifter[error], callback func(*C.RustCallStatus) U) (U, error) {
	var status C.RustCallStatus
	returnValue := callback(&status)
	err := checkCallStatus(converter, status)

	return returnValue, err
}

func checkCallStatus(converter BufLifter[error], status C.RustCallStatus) error {
	switch status.code {
	case 0:
		return nil
	case 1:
		return converter.Lift(status.errorBuf)
	case 2:
		// when the rust code sees a panic, it tries to construct a rustbuffer
		// with the message.  but if that code panics, then it just sends back
		// an empty buffer.
		if status.errorBuf.len > 0 {
			panic(fmt.Errorf("%s", FfiConverterStringINSTANCE.Lift(status.errorBuf)))
		} else {
			panic(fmt.Errorf("Rust panicked while handling Rust panic"))
		}
	default:
		return fmt.Errorf("unknown status code: %d", status.code)
	}
}

func checkCallStatusUnknown(status C.RustCallStatus) error {
	switch status.code {
	case 0:
		return nil
	case 1:
		panic(fmt.Errorf("function not returning an error returned an error"))
	case 2:
		// when the rust code sees a panic, it tries to construct a rustbuffer
		// with the message.  but if that code panics, then it just sends back
		// an empty buffer.
		if status.errorBuf.len > 0 {
			panic(fmt.Errorf("%s", FfiConverterStringINSTANCE.Lift(status.errorBuf)))
		} else {
			panic(fmt.Errorf("Rust panicked while handling Rust panic"))
		}
	default:
		return fmt.Errorf("unknown status code: %d", status.code)
	}
}

func rustCall[U any](callback func(*C.RustCallStatus) U) U {
	returnValue, err := rustCallWithError(nil, callback)
	if err != nil {
		panic(err)
	}
	return returnValue
}

func writeInt8(writer io.Writer, value int8) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeUint8(writer io.Writer, value uint8) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeInt16(writer io.Writer, value int16) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeUint16(writer io.Writer, value uint16) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeInt32(writer io.Writer, value int32) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeUint32(writer io.Writer, value uint32) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeInt64(writer io.Writer, value int64) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeUint64(writer io.Writer, value uint64) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeFloat32(writer io.Writer, value float32) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeFloat64(writer io.Writer, value float64) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func readInt8(reader io.Reader) int8 {
	var result int8
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readUint8(reader io.Reader) uint8 {
	var result uint8
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readInt16(reader io.Reader) int16 {
	var result int16
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readUint16(reader io.Reader) uint16 {
	var result uint16
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readInt32(reader io.Reader) int32 {
	var result int32
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readUint32(reader io.Reader) uint32 {
	var result uint32
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readInt64(reader io.Reader) int64 {
	var result int64
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readUint64(reader io.Reader) uint64 {
	var result uint64
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readFloat32(reader io.Reader) float32 {
	var result float32
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readFloat64(reader io.Reader) float64 {
	var result float64
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func init() {

	uniffiCheckChecksums()
}

func uniffiCheckChecksums() {
	// Get the bindings contract version from our ComponentInterface
	bindingsContractVersion := 24
	// Get the scaffolding contract version by calling the into the dylib
	scaffoldingContractVersion := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint32_t {
		return C.ffi_ldk_node_uniffi_contract_version(uniffiStatus)
	})
	if bindingsContractVersion != int(scaffoldingContractVersion) {
		// If this happens try cleaning and rebuilding your project
		panic("ldk_node: UniFFI contract version mismatch")
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_func_default_config(uniffiStatus)
		})
		if checksum != 62308 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_func_default_config: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_func_generate_entropy_mnemonic(uniffiStatus)
		})
		if checksum != 7251 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_func_generate_entropy_mnemonic: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_bolt11payment_receive(uniffiStatus)
		})
		if checksum != 44074 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_bolt11payment_receive: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_bolt11payment_receive_variable_amount(uniffiStatus)
		})
		if checksum != 50172 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_bolt11payment_receive_variable_amount: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_bolt11payment_receive_variable_amount_via_jit_channel(uniffiStatus)
		})
		if checksum != 6695 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_bolt11payment_receive_variable_amount_via_jit_channel: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_bolt11payment_receive_via_jit_channel(uniffiStatus)
		})
		if checksum != 10006 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_bolt11payment_receive_via_jit_channel: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_bolt11payment_send(uniffiStatus)
		})
		if checksum != 54619 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_bolt11payment_send: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_bolt11payment_send_probes(uniffiStatus)
		})
		if checksum != 1481 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_bolt11payment_send_probes: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_bolt11payment_send_probes_using_amount(uniffiStatus)
		})
		if checksum != 40103 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_bolt11payment_send_probes_using_amount: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_bolt11payment_send_using_amount(uniffiStatus)
		})
		if checksum != 52842 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_bolt11payment_send_using_amount: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_builder_build(uniffiStatus)
		})
		if checksum != 46255 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_builder_build: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_builder_build_with_fs_store(uniffiStatus)
		})
		if checksum != 15423 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_builder_build_with_fs_store: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_builder_set_entropy_bip39_mnemonic(uniffiStatus)
		})
		if checksum != 35659 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_builder_set_entropy_bip39_mnemonic: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_builder_set_entropy_seed_bytes(uniffiStatus)
		})
		if checksum != 26795 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_builder_set_entropy_seed_bytes: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_builder_set_entropy_seed_path(uniffiStatus)
		})
		if checksum != 64056 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_builder_set_entropy_seed_path: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_builder_set_esplora_server(uniffiStatus)
		})
		if checksum != 7044 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_builder_set_esplora_server: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_builder_set_gossip_source_p2p(uniffiStatus)
		})
		if checksum != 9279 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_builder_set_gossip_source_p2p: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_builder_set_gossip_source_rgs(uniffiStatus)
		})
		if checksum != 64312 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_builder_set_gossip_source_rgs: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_builder_set_liquidity_source_lsps2(uniffiStatus)
		})
		if checksum != 26412 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_builder_set_liquidity_source_lsps2: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_builder_set_listening_addresses(uniffiStatus)
		})
		if checksum != 18689 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_builder_set_listening_addresses: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_builder_set_network(uniffiStatus)
		})
		if checksum != 38526 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_builder_set_network: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_builder_set_storage_dir_path(uniffiStatus)
		})
		if checksum != 59019 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_builder_set_storage_dir_path: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_channelconfig_accept_underpaying_htlcs(uniffiStatus)
		})
		if checksum != 45655 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_channelconfig_accept_underpaying_htlcs: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_channelconfig_cltv_expiry_delta(uniffiStatus)
		})
		if checksum != 19044 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_channelconfig_cltv_expiry_delta: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_channelconfig_force_close_avoidance_max_fee_satoshis(uniffiStatus)
		})
		if checksum != 69 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_channelconfig_force_close_avoidance_max_fee_satoshis: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_channelconfig_forwarding_fee_base_msat(uniffiStatus)
		})
		if checksum != 3400 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_channelconfig_forwarding_fee_base_msat: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_channelconfig_forwarding_fee_proportional_millionths(uniffiStatus)
		})
		if checksum != 31794 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_channelconfig_forwarding_fee_proportional_millionths: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_channelconfig_set_accept_underpaying_htlcs(uniffiStatus)
		})
		if checksum != 27275 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_channelconfig_set_accept_underpaying_htlcs: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_channelconfig_set_cltv_expiry_delta(uniffiStatus)
		})
		if checksum != 40735 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_channelconfig_set_cltv_expiry_delta: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_channelconfig_set_force_close_avoidance_max_fee_satoshis(uniffiStatus)
		})
		if checksum != 48479 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_channelconfig_set_force_close_avoidance_max_fee_satoshis: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_channelconfig_set_forwarding_fee_base_msat(uniffiStatus)
		})
		if checksum != 29831 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_channelconfig_set_forwarding_fee_base_msat: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_channelconfig_set_forwarding_fee_proportional_millionths(uniffiStatus)
		})
		if checksum != 65060 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_channelconfig_set_forwarding_fee_proportional_millionths: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_channelconfig_set_max_dust_htlc_exposure_from_fee_rate_multiplier(uniffiStatus)
		})
		if checksum != 4707 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_channelconfig_set_max_dust_htlc_exposure_from_fee_rate_multiplier: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_channelconfig_set_max_dust_htlc_exposure_from_fixed_limit(uniffiStatus)
		})
		if checksum != 16864 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_channelconfig_set_max_dust_htlc_exposure_from_fixed_limit: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_bolt11_payment(uniffiStatus)
		})
		if checksum != 41402 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_bolt11_payment: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_close_channel(uniffiStatus)
		})
		if checksum != 47156 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_close_channel: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_config(uniffiStatus)
		})
		if checksum != 15339 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_config: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_connect(uniffiStatus)
		})
		if checksum != 15352 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_connect: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_connect_open_channel(uniffiStatus)
		})
		if checksum != 8778 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_connect_open_channel: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_disconnect(uniffiStatus)
		})
		if checksum != 47760 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_disconnect: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_event_handled(uniffiStatus)
		})
		if checksum != 47939 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_event_handled: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_list_balances(uniffiStatus)
		})
		if checksum != 24919 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_list_balances: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_list_channels(uniffiStatus)
		})
		if checksum != 62491 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_list_channels: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_list_payments(uniffiStatus)
		})
		if checksum != 47765 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_list_payments: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_list_peers(uniffiStatus)
		})
		if checksum != 12947 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_list_peers: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_listening_addresses(uniffiStatus)
		})
		if checksum != 55483 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_listening_addresses: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_next_event(uniffiStatus)
		})
		if checksum != 46767 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_next_event: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_node_id(uniffiStatus)
		})
		if checksum != 34585 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_node_id: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_onchain_payment(uniffiStatus)
		})
		if checksum != 6092 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_onchain_payment: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_payment(uniffiStatus)
		})
		if checksum != 59271 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_payment: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_remove_payment(uniffiStatus)
		})
		if checksum != 8539 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_remove_payment: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_reset_router(uniffiStatus)
		})
		if checksum != 49565 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_reset_router: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_sign_message(uniffiStatus)
		})
		if checksum != 22595 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_sign_message: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_spontaneous_payment(uniffiStatus)
		})
		if checksum != 37403 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_spontaneous_payment: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_start(uniffiStatus)
		})
		if checksum != 21524 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_start: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_status(uniffiStatus)
		})
		if checksum != 46945 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_status: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_stop(uniffiStatus)
		})
		if checksum != 12389 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_stop: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_sync_wallets(uniffiStatus)
		})
		if checksum != 29385 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_sync_wallets: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_update_channel_config(uniffiStatus)
		})
		if checksum != 21821 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_update_channel_config: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_verify_signature(uniffiStatus)
		})
		if checksum != 56945 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_verify_signature: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_wait_next_event(uniffiStatus)
		})
		if checksum != 30900 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_wait_next_event: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_onchainpayment_new_address(uniffiStatus)
		})
		if checksum != 23077 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_onchainpayment_new_address: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_onchainpayment_send_all_to_address(uniffiStatus)
		})
		if checksum != 35766 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_onchainpayment_send_all_to_address: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_onchainpayment_send_to_address(uniffiStatus)
		})
		if checksum != 36709 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_onchainpayment_send_to_address: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_spontaneouspayment_send(uniffiStatus)
		})
		if checksum != 11473 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_spontaneouspayment_send: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_spontaneouspayment_send_probes(uniffiStatus)
		})
		if checksum != 32884 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_spontaneouspayment_send_probes: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_constructor_builder_from_config(uniffiStatus)
		})
		if checksum != 56443 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_constructor_builder_from_config: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_constructor_builder_new(uniffiStatus)
		})
		if checksum != 48442 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_constructor_builder_new: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_constructor_channelconfig_new(uniffiStatus)
		})
		if checksum != 24987 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_constructor_channelconfig_new: UniFFI API checksum mismatch")
		}
	}
}

type FfiConverterUint8 struct{}

var FfiConverterUint8INSTANCE = FfiConverterUint8{}

func (FfiConverterUint8) Lower(value uint8) C.uint8_t {
	return C.uint8_t(value)
}

func (FfiConverterUint8) Write(writer io.Writer, value uint8) {
	writeUint8(writer, value)
}

func (FfiConverterUint8) Lift(value C.uint8_t) uint8 {
	return uint8(value)
}

func (FfiConverterUint8) Read(reader io.Reader) uint8 {
	return readUint8(reader)
}

type FfiDestroyerUint8 struct{}

func (FfiDestroyerUint8) Destroy(_ uint8) {}

type FfiConverterUint16 struct{}

var FfiConverterUint16INSTANCE = FfiConverterUint16{}

func (FfiConverterUint16) Lower(value uint16) C.uint16_t {
	return C.uint16_t(value)
}

func (FfiConverterUint16) Write(writer io.Writer, value uint16) {
	writeUint16(writer, value)
}

func (FfiConverterUint16) Lift(value C.uint16_t) uint16 {
	return uint16(value)
}

func (FfiConverterUint16) Read(reader io.Reader) uint16 {
	return readUint16(reader)
}

type FfiDestroyerUint16 struct{}

func (FfiDestroyerUint16) Destroy(_ uint16) {}

type FfiConverterUint32 struct{}

var FfiConverterUint32INSTANCE = FfiConverterUint32{}

func (FfiConverterUint32) Lower(value uint32) C.uint32_t {
	return C.uint32_t(value)
}

func (FfiConverterUint32) Write(writer io.Writer, value uint32) {
	writeUint32(writer, value)
}

func (FfiConverterUint32) Lift(value C.uint32_t) uint32 {
	return uint32(value)
}

func (FfiConverterUint32) Read(reader io.Reader) uint32 {
	return readUint32(reader)
}

type FfiDestroyerUint32 struct{}

func (FfiDestroyerUint32) Destroy(_ uint32) {}

type FfiConverterUint64 struct{}

var FfiConverterUint64INSTANCE = FfiConverterUint64{}

func (FfiConverterUint64) Lower(value uint64) C.uint64_t {
	return C.uint64_t(value)
}

func (FfiConverterUint64) Write(writer io.Writer, value uint64) {
	writeUint64(writer, value)
}

func (FfiConverterUint64) Lift(value C.uint64_t) uint64 {
	return uint64(value)
}

func (FfiConverterUint64) Read(reader io.Reader) uint64 {
	return readUint64(reader)
}

type FfiDestroyerUint64 struct{}

func (FfiDestroyerUint64) Destroy(_ uint64) {}

type FfiConverterBool struct{}

var FfiConverterBoolINSTANCE = FfiConverterBool{}

func (FfiConverterBool) Lower(value bool) C.int8_t {
	if value {
		return C.int8_t(1)
	}
	return C.int8_t(0)
}

func (FfiConverterBool) Write(writer io.Writer, value bool) {
	if value {
		writeInt8(writer, 1)
	} else {
		writeInt8(writer, 0)
	}
}

func (FfiConverterBool) Lift(value C.int8_t) bool {
	return value != 0
}

func (FfiConverterBool) Read(reader io.Reader) bool {
	return readInt8(reader) != 0
}

type FfiDestroyerBool struct{}

func (FfiDestroyerBool) Destroy(_ bool) {}

type FfiConverterString struct{}

var FfiConverterStringINSTANCE = FfiConverterString{}

func (FfiConverterString) Lift(rb RustBufferI) string {
	defer rb.Free()
	reader := rb.AsReader()
	b, err := io.ReadAll(reader)
	if err != nil {
		panic(fmt.Errorf("reading reader: %w", err))
	}
	return string(b)
}

func (FfiConverterString) Read(reader io.Reader) string {
	length := readInt32(reader)
	buffer := make([]byte, length)
	read_length, err := reader.Read(buffer)
	if err != nil {
		panic(err)
	}
	if read_length != int(length) {
		panic(fmt.Errorf("bad read length when reading string, expected %d, read %d", length, read_length))
	}
	return string(buffer)
}

func (FfiConverterString) Lower(value string) RustBuffer {
	return stringToRustBuffer(value)
}

func (FfiConverterString) Write(writer io.Writer, value string) {
	if len(value) > math.MaxInt32 {
		panic("String is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	write_length, err := io.WriteString(writer, value)
	if err != nil {
		panic(err)
	}
	if write_length != len(value) {
		panic(fmt.Errorf("bad write length when writing string, expected %d, written %d", len(value), write_length))
	}
}

type FfiDestroyerString struct{}

func (FfiDestroyerString) Destroy(_ string) {}

// Below is an implementation of synchronization requirements outlined in the link.
// https://github.com/mozilla/uniffi-rs/blob/0dc031132d9493ca812c3af6e7dd60ad2ea95bf0/uniffi_bindgen/src/bindings/kotlin/templates/ObjectRuntime.kt#L31

type FfiObject struct {
	pointer      unsafe.Pointer
	callCounter  atomic.Int64
	freeFunction func(unsafe.Pointer, *C.RustCallStatus)
	destroyed    atomic.Bool
}

func newFfiObject(pointer unsafe.Pointer, freeFunction func(unsafe.Pointer, *C.RustCallStatus)) FfiObject {
	return FfiObject{
		pointer:      pointer,
		freeFunction: freeFunction,
	}
}

func (ffiObject *FfiObject) incrementPointer(debugName string) unsafe.Pointer {
	for {
		counter := ffiObject.callCounter.Load()
		if counter <= -1 {
			panic(fmt.Errorf("%v object has already been destroyed", debugName))
		}
		if counter == math.MaxInt64 {
			panic(fmt.Errorf("%v object call counter would overflow", debugName))
		}
		if ffiObject.callCounter.CompareAndSwap(counter, counter+1) {
			break
		}
	}

	return ffiObject.pointer
}

func (ffiObject *FfiObject) decrementPointer() {
	if ffiObject.callCounter.Add(-1) == -1 {
		ffiObject.freeRustArcPtr()
	}
}

func (ffiObject *FfiObject) destroy() {
	if ffiObject.destroyed.CompareAndSwap(false, true) {
		if ffiObject.callCounter.Add(-1) == -1 {
			ffiObject.freeRustArcPtr()
		}
	}
}

func (ffiObject *FfiObject) freeRustArcPtr() {
	rustCall(func(status *C.RustCallStatus) int32 {
		ffiObject.freeFunction(ffiObject.pointer, status)
		return 0
	})
}

type Bolt11Payment struct {
	ffiObject FfiObject
}

func (_self *Bolt11Payment) Receive(amountMsat uint64, description string, expirySecs uint32) (Bolt11Invoice, error) {
	_pointer := _self.ffiObject.incrementPointer("*Bolt11Payment")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_bolt11payment_receive(
			_pointer, FfiConverterUint64INSTANCE.Lower(amountMsat), FfiConverterStringINSTANCE.Lower(description), FfiConverterUint32INSTANCE.Lower(expirySecs), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue Bolt11Invoice
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterTypeBolt11InvoiceINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func (_self *Bolt11Payment) ReceiveVariableAmount(description string, expirySecs uint32) (Bolt11Invoice, error) {
	_pointer := _self.ffiObject.incrementPointer("*Bolt11Payment")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_bolt11payment_receive_variable_amount(
			_pointer, FfiConverterStringINSTANCE.Lower(description), FfiConverterUint32INSTANCE.Lower(expirySecs), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue Bolt11Invoice
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterTypeBolt11InvoiceINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func (_self *Bolt11Payment) ReceiveVariableAmountViaJitChannel(description string, expirySecs uint32, maxProportionalLspFeeLimitPpmMsat *uint64) (Bolt11Invoice, error) {
	_pointer := _self.ffiObject.incrementPointer("*Bolt11Payment")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_bolt11payment_receive_variable_amount_via_jit_channel(
			_pointer, FfiConverterStringINSTANCE.Lower(description), FfiConverterUint32INSTANCE.Lower(expirySecs), FfiConverterOptionalUint64INSTANCE.Lower(maxProportionalLspFeeLimitPpmMsat), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue Bolt11Invoice
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterTypeBolt11InvoiceINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func (_self *Bolt11Payment) ReceiveViaJitChannel(amountMsat uint64, description string, expirySecs uint32, maxLspFeeLimitMsat *uint64) (Bolt11Invoice, error) {
	_pointer := _self.ffiObject.incrementPointer("*Bolt11Payment")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_bolt11payment_receive_via_jit_channel(
			_pointer, FfiConverterUint64INSTANCE.Lower(amountMsat), FfiConverterStringINSTANCE.Lower(description), FfiConverterUint32INSTANCE.Lower(expirySecs), FfiConverterOptionalUint64INSTANCE.Lower(maxLspFeeLimitMsat), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue Bolt11Invoice
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterTypeBolt11InvoiceINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func (_self *Bolt11Payment) Send(invoice Bolt11Invoice) (PaymentId, error) {
	_pointer := _self.ffiObject.incrementPointer("*Bolt11Payment")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_bolt11payment_send(
			_pointer, FfiConverterTypeBolt11InvoiceINSTANCE.Lower(invoice), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue PaymentId
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterTypePaymentIdINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func (_self *Bolt11Payment) SendProbes(invoice Bolt11Invoice) error {
	_pointer := _self.ffiObject.incrementPointer("*Bolt11Payment")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_bolt11payment_send_probes(
			_pointer, FfiConverterTypeBolt11InvoiceINSTANCE.Lower(invoice), _uniffiStatus)
		return false
	})
	return _uniffiErr
}

func (_self *Bolt11Payment) SendProbesUsingAmount(invoice Bolt11Invoice, amountMsat uint64) error {
	_pointer := _self.ffiObject.incrementPointer("*Bolt11Payment")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_bolt11payment_send_probes_using_amount(
			_pointer, FfiConverterTypeBolt11InvoiceINSTANCE.Lower(invoice), FfiConverterUint64INSTANCE.Lower(amountMsat), _uniffiStatus)
		return false
	})
	return _uniffiErr
}

func (_self *Bolt11Payment) SendUsingAmount(invoice Bolt11Invoice, amountMsat uint64) (PaymentId, error) {
	_pointer := _self.ffiObject.incrementPointer("*Bolt11Payment")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_bolt11payment_send_using_amount(
			_pointer, FfiConverterTypeBolt11InvoiceINSTANCE.Lower(invoice), FfiConverterUint64INSTANCE.Lower(amountMsat), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue PaymentId
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterTypePaymentIdINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func (object *Bolt11Payment) Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterBolt11Payment struct{}

var FfiConverterBolt11PaymentINSTANCE = FfiConverterBolt11Payment{}

func (c FfiConverterBolt11Payment) Lift(pointer unsafe.Pointer) *Bolt11Payment {
	result := &Bolt11Payment{
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_ldk_node_fn_free_bolt11payment(pointer, status)
			}),
	}
	runtime.SetFinalizer(result, (*Bolt11Payment).Destroy)
	return result
}

func (c FfiConverterBolt11Payment) Read(reader io.Reader) *Bolt11Payment {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterBolt11Payment) Lower(value *Bolt11Payment) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := value.ffiObject.incrementPointer("*Bolt11Payment")
	defer value.ffiObject.decrementPointer()
	return pointer
}

func (c FfiConverterBolt11Payment) Write(writer io.Writer, value *Bolt11Payment) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerBolt11Payment struct{}

func (_ FfiDestroyerBolt11Payment) Destroy(value *Bolt11Payment) {
	value.Destroy()
}

type Builder struct {
	ffiObject FfiObject
}

func NewBuilder() *Builder {
	return FfiConverterBuilderINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_ldk_node_fn_constructor_builder_new(_uniffiStatus)
	}))
}

func BuilderFromConfig(config Config) *Builder {
	return FfiConverterBuilderINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_ldk_node_fn_constructor_builder_from_config(FfiConverterTypeConfigINSTANCE.Lower(config), _uniffiStatus)
	}))
}

func (_self *Builder) Build() (*Node, error) {
	_pointer := _self.ffiObject.incrementPointer("*Builder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeBuildError{}, func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_ldk_node_fn_method_builder_build(
			_pointer, _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue *Node
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterNodeINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func (_self *Builder) BuildWithFsStore() (*Node, error) {
	_pointer := _self.ffiObject.incrementPointer("*Builder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeBuildError{}, func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_ldk_node_fn_method_builder_build_with_fs_store(
			_pointer, _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue *Node
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterNodeINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func (_self *Builder) SetEntropyBip39Mnemonic(mnemonic Mnemonic, passphrase *string) {
	_pointer := _self.ffiObject.incrementPointer("*Builder")
	defer _self.ffiObject.decrementPointer()
	rustCall(func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_builder_set_entropy_bip39_mnemonic(
			_pointer, FfiConverterTypeMnemonicINSTANCE.Lower(mnemonic), FfiConverterOptionalStringINSTANCE.Lower(passphrase), _uniffiStatus)
		return false
	})
}

func (_self *Builder) SetEntropySeedBytes(seedBytes []uint8) error {
	_pointer := _self.ffiObject.incrementPointer("*Builder")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeBuildError{}, func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_builder_set_entropy_seed_bytes(
			_pointer, FfiConverterSequenceUint8INSTANCE.Lower(seedBytes), _uniffiStatus)
		return false
	})
	return _uniffiErr
}

func (_self *Builder) SetEntropySeedPath(seedPath string) {
	_pointer := _self.ffiObject.incrementPointer("*Builder")
	defer _self.ffiObject.decrementPointer()
	rustCall(func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_builder_set_entropy_seed_path(
			_pointer, FfiConverterStringINSTANCE.Lower(seedPath), _uniffiStatus)
		return false
	})
}

func (_self *Builder) SetEsploraServer(esploraServerUrl string) {
	_pointer := _self.ffiObject.incrementPointer("*Builder")
	defer _self.ffiObject.decrementPointer()
	rustCall(func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_builder_set_esplora_server(
			_pointer, FfiConverterStringINSTANCE.Lower(esploraServerUrl), _uniffiStatus)
		return false
	})
}

func (_self *Builder) SetGossipSourceP2p() {
	_pointer := _self.ffiObject.incrementPointer("*Builder")
	defer _self.ffiObject.decrementPointer()
	rustCall(func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_builder_set_gossip_source_p2p(
			_pointer, _uniffiStatus)
		return false
	})
}

func (_self *Builder) SetGossipSourceRgs(rgsServerUrl string) {
	_pointer := _self.ffiObject.incrementPointer("*Builder")
	defer _self.ffiObject.decrementPointer()
	rustCall(func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_builder_set_gossip_source_rgs(
			_pointer, FfiConverterStringINSTANCE.Lower(rgsServerUrl), _uniffiStatus)
		return false
	})
}

func (_self *Builder) SetLiquiditySourceLsps2(address SocketAddress, nodeId PublicKey, token *string) {
	_pointer := _self.ffiObject.incrementPointer("*Builder")
	defer _self.ffiObject.decrementPointer()
	rustCall(func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_builder_set_liquidity_source_lsps2(
			_pointer, FfiConverterTypeSocketAddressINSTANCE.Lower(address), FfiConverterTypePublicKeyINSTANCE.Lower(nodeId), FfiConverterOptionalStringINSTANCE.Lower(token), _uniffiStatus)
		return false
	})
}

func (_self *Builder) SetListeningAddresses(listeningAddresses []SocketAddress) error {
	_pointer := _self.ffiObject.incrementPointer("*Builder")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeBuildError{}, func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_builder_set_listening_addresses(
			_pointer, FfiConverterSequenceTypeSocketAddressINSTANCE.Lower(listeningAddresses), _uniffiStatus)
		return false
	})
	return _uniffiErr
}

func (_self *Builder) SetNetwork(network Network) {
	_pointer := _self.ffiObject.incrementPointer("*Builder")
	defer _self.ffiObject.decrementPointer()
	rustCall(func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_builder_set_network(
			_pointer, FfiConverterTypeNetworkINSTANCE.Lower(network), _uniffiStatus)
		return false
	})
}

func (_self *Builder) SetStorageDirPath(storageDirPath string) {
	_pointer := _self.ffiObject.incrementPointer("*Builder")
	defer _self.ffiObject.decrementPointer()
	rustCall(func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_builder_set_storage_dir_path(
			_pointer, FfiConverterStringINSTANCE.Lower(storageDirPath), _uniffiStatus)
		return false
	})
}

func (object *Builder) Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterBuilder struct{}

var FfiConverterBuilderINSTANCE = FfiConverterBuilder{}

func (c FfiConverterBuilder) Lift(pointer unsafe.Pointer) *Builder {
	result := &Builder{
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_ldk_node_fn_free_builder(pointer, status)
			}),
	}
	runtime.SetFinalizer(result, (*Builder).Destroy)
	return result
}

func (c FfiConverterBuilder) Read(reader io.Reader) *Builder {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterBuilder) Lower(value *Builder) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := value.ffiObject.incrementPointer("*Builder")
	defer value.ffiObject.decrementPointer()
	return pointer
}

func (c FfiConverterBuilder) Write(writer io.Writer, value *Builder) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerBuilder struct{}

func (_ FfiDestroyerBuilder) Destroy(value *Builder) {
	value.Destroy()
}

type ChannelConfig struct {
	ffiObject FfiObject
}

func NewChannelConfig() *ChannelConfig {
	return FfiConverterChannelConfigINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_ldk_node_fn_constructor_channelconfig_new(_uniffiStatus)
	}))
}

func (_self *ChannelConfig) AcceptUnderpayingHtlcs() bool {
	_pointer := _self.ffiObject.incrementPointer("*ChannelConfig")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterBoolINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) C.int8_t {
		return C.uniffi_ldk_node_fn_method_channelconfig_accept_underpaying_htlcs(
			_pointer, _uniffiStatus)
	}))
}

func (_self *ChannelConfig) CltvExpiryDelta() uint16 {
	_pointer := _self.ffiObject.incrementPointer("*ChannelConfig")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterUint16INSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
		return C.uniffi_ldk_node_fn_method_channelconfig_cltv_expiry_delta(
			_pointer, _uniffiStatus)
	}))
}

func (_self *ChannelConfig) ForceCloseAvoidanceMaxFeeSatoshis() uint64 {
	_pointer := _self.ffiObject.incrementPointer("*ChannelConfig")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterUint64INSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint64_t {
		return C.uniffi_ldk_node_fn_method_channelconfig_force_close_avoidance_max_fee_satoshis(
			_pointer, _uniffiStatus)
	}))
}

func (_self *ChannelConfig) ForwardingFeeBaseMsat() uint32 {
	_pointer := _self.ffiObject.incrementPointer("*ChannelConfig")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterUint32INSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint32_t {
		return C.uniffi_ldk_node_fn_method_channelconfig_forwarding_fee_base_msat(
			_pointer, _uniffiStatus)
	}))
}

func (_self *ChannelConfig) ForwardingFeeProportionalMillionths() uint32 {
	_pointer := _self.ffiObject.incrementPointer("*ChannelConfig")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterUint32INSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint32_t {
		return C.uniffi_ldk_node_fn_method_channelconfig_forwarding_fee_proportional_millionths(
			_pointer, _uniffiStatus)
	}))
}

func (_self *ChannelConfig) SetAcceptUnderpayingHtlcs(value bool) {
	_pointer := _self.ffiObject.incrementPointer("*ChannelConfig")
	defer _self.ffiObject.decrementPointer()
	rustCall(func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_channelconfig_set_accept_underpaying_htlcs(
			_pointer, FfiConverterBoolINSTANCE.Lower(value), _uniffiStatus)
		return false
	})
}

func (_self *ChannelConfig) SetCltvExpiryDelta(value uint16) {
	_pointer := _self.ffiObject.incrementPointer("*ChannelConfig")
	defer _self.ffiObject.decrementPointer()
	rustCall(func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_channelconfig_set_cltv_expiry_delta(
			_pointer, FfiConverterUint16INSTANCE.Lower(value), _uniffiStatus)
		return false
	})
}

func (_self *ChannelConfig) SetForceCloseAvoidanceMaxFeeSatoshis(valueSat uint64) {
	_pointer := _self.ffiObject.incrementPointer("*ChannelConfig")
	defer _self.ffiObject.decrementPointer()
	rustCall(func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_channelconfig_set_force_close_avoidance_max_fee_satoshis(
			_pointer, FfiConverterUint64INSTANCE.Lower(valueSat), _uniffiStatus)
		return false
	})
}

func (_self *ChannelConfig) SetForwardingFeeBaseMsat(feeMsat uint32) {
	_pointer := _self.ffiObject.incrementPointer("*ChannelConfig")
	defer _self.ffiObject.decrementPointer()
	rustCall(func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_channelconfig_set_forwarding_fee_base_msat(
			_pointer, FfiConverterUint32INSTANCE.Lower(feeMsat), _uniffiStatus)
		return false
	})
}

func (_self *ChannelConfig) SetForwardingFeeProportionalMillionths(value uint32) {
	_pointer := _self.ffiObject.incrementPointer("*ChannelConfig")
	defer _self.ffiObject.decrementPointer()
	rustCall(func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_channelconfig_set_forwarding_fee_proportional_millionths(
			_pointer, FfiConverterUint32INSTANCE.Lower(value), _uniffiStatus)
		return false
	})
}

func (_self *ChannelConfig) SetMaxDustHtlcExposureFromFeeRateMultiplier(multiplier uint64) {
	_pointer := _self.ffiObject.incrementPointer("*ChannelConfig")
	defer _self.ffiObject.decrementPointer()
	rustCall(func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_channelconfig_set_max_dust_htlc_exposure_from_fee_rate_multiplier(
			_pointer, FfiConverterUint64INSTANCE.Lower(multiplier), _uniffiStatus)
		return false
	})
}

func (_self *ChannelConfig) SetMaxDustHtlcExposureFromFixedLimit(limitMsat uint64) {
	_pointer := _self.ffiObject.incrementPointer("*ChannelConfig")
	defer _self.ffiObject.decrementPointer()
	rustCall(func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_channelconfig_set_max_dust_htlc_exposure_from_fixed_limit(
			_pointer, FfiConverterUint64INSTANCE.Lower(limitMsat), _uniffiStatus)
		return false
	})
}

func (object *ChannelConfig) Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterChannelConfig struct{}

var FfiConverterChannelConfigINSTANCE = FfiConverterChannelConfig{}

func (c FfiConverterChannelConfig) Lift(pointer unsafe.Pointer) *ChannelConfig {
	result := &ChannelConfig{
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_ldk_node_fn_free_channelconfig(pointer, status)
			}),
	}
	runtime.SetFinalizer(result, (*ChannelConfig).Destroy)
	return result
}

func (c FfiConverterChannelConfig) Read(reader io.Reader) *ChannelConfig {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterChannelConfig) Lower(value *ChannelConfig) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := value.ffiObject.incrementPointer("*ChannelConfig")
	defer value.ffiObject.decrementPointer()
	return pointer
}

func (c FfiConverterChannelConfig) Write(writer io.Writer, value *ChannelConfig) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerChannelConfig struct{}

func (_ FfiDestroyerChannelConfig) Destroy(value *ChannelConfig) {
	value.Destroy()
}

type Node struct {
	ffiObject FfiObject
}

func (_self *Node) Bolt11Payment() *Bolt11Payment {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterBolt11PaymentINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_ldk_node_fn_method_node_bolt11_payment(
			_pointer, _uniffiStatus)
	}))
}

func (_self *Node) CloseChannel(userChannelId UserChannelId, counterpartyNodeId PublicKey, force bool) error {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_node_close_channel(
			_pointer, FfiConverterTypeUserChannelIdINSTANCE.Lower(userChannelId), FfiConverterTypePublicKeyINSTANCE.Lower(counterpartyNodeId), FfiConverterBoolINSTANCE.Lower(force), _uniffiStatus)
		return false
	})
	return _uniffiErr
}

func (_self *Node) Config() Config {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterTypeConfigINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_node_config(
			_pointer, _uniffiStatus)
	}))
}

func (_self *Node) Connect(nodeId PublicKey, address SocketAddress, persist bool) error {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_node_connect(
			_pointer, FfiConverterTypePublicKeyINSTANCE.Lower(nodeId), FfiConverterTypeSocketAddressINSTANCE.Lower(address), FfiConverterBoolINSTANCE.Lower(persist), _uniffiStatus)
		return false
	})
	return _uniffiErr
}

func (_self *Node) ConnectOpenChannel(nodeId PublicKey, address SocketAddress, channelAmountSats uint64, pushToCounterpartyMsat *uint64, channelConfig **ChannelConfig, announceChannel bool) (UserChannelId, error) {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_node_connect_open_channel(
			_pointer, FfiConverterTypePublicKeyINSTANCE.Lower(nodeId), FfiConverterTypeSocketAddressINSTANCE.Lower(address), FfiConverterUint64INSTANCE.Lower(channelAmountSats), FfiConverterOptionalUint64INSTANCE.Lower(pushToCounterpartyMsat), FfiConverterOptionalChannelConfigINSTANCE.Lower(channelConfig), FfiConverterBoolINSTANCE.Lower(announceChannel), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue UserChannelId
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterTypeUserChannelIdINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func (_self *Node) Disconnect(nodeId PublicKey) error {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_node_disconnect(
			_pointer, FfiConverterTypePublicKeyINSTANCE.Lower(nodeId), _uniffiStatus)
		return false
	})
	return _uniffiErr
}

func (_self *Node) EventHandled() {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	rustCall(func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_node_event_handled(
			_pointer, _uniffiStatus)
		return false
	})
}

func (_self *Node) ListBalances() BalanceDetails {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterTypeBalanceDetailsINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_node_list_balances(
			_pointer, _uniffiStatus)
	}))
}

func (_self *Node) ListChannels() []ChannelDetails {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterSequenceTypeChannelDetailsINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_node_list_channels(
			_pointer, _uniffiStatus)
	}))
}

func (_self *Node) ListPayments() []PaymentDetails {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterSequenceTypePaymentDetailsINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_node_list_payments(
			_pointer, _uniffiStatus)
	}))
}

func (_self *Node) ListPeers() []PeerDetails {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterSequenceTypePeerDetailsINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_node_list_peers(
			_pointer, _uniffiStatus)
	}))
}

func (_self *Node) ListeningAddresses() *[]SocketAddress {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterOptionalSequenceTypeSocketAddressINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_node_listening_addresses(
			_pointer, _uniffiStatus)
	}))
}

func (_self *Node) NextEvent() *Event {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterOptionalTypeEventINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_node_next_event(
			_pointer, _uniffiStatus)
	}))
}

func (_self *Node) NodeId() PublicKey {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterTypePublicKeyINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_node_node_id(
			_pointer, _uniffiStatus)
	}))
}

func (_self *Node) OnchainPayment() *OnchainPayment {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterOnchainPaymentINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_ldk_node_fn_method_node_onchain_payment(
			_pointer, _uniffiStatus)
	}))
}

func (_self *Node) Payment(paymentId PaymentId) *PaymentDetails {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterOptionalTypePaymentDetailsINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_node_payment(
			_pointer, FfiConverterTypePaymentIdINSTANCE.Lower(paymentId), _uniffiStatus)
	}))
}

func (_self *Node) RemovePayment(paymentId PaymentId) error {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_node_remove_payment(
			_pointer, FfiConverterTypePaymentIdINSTANCE.Lower(paymentId), _uniffiStatus)
		return false
	})
	return _uniffiErr
}

func (_self *Node) ResetRouter() error {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_node_reset_router(
			_pointer, _uniffiStatus)
		return false
	})
	return _uniffiErr
}

func (_self *Node) SignMessage(msg []uint8) (string, error) {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_node_sign_message(
			_pointer, FfiConverterSequenceUint8INSTANCE.Lower(msg), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue string
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterStringINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func (_self *Node) SpontaneousPayment() *SpontaneousPayment {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterSpontaneousPaymentINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_ldk_node_fn_method_node_spontaneous_payment(
			_pointer, _uniffiStatus)
	}))
}

func (_self *Node) Start() error {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_node_start(
			_pointer, _uniffiStatus)
		return false
	})
	return _uniffiErr
}

func (_self *Node) Status() NodeStatus {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterTypeNodeStatusINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_node_status(
			_pointer, _uniffiStatus)
	}))
}

func (_self *Node) Stop() error {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_node_stop(
			_pointer, _uniffiStatus)
		return false
	})
	return _uniffiErr
}

func (_self *Node) SyncWallets() error {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_node_sync_wallets(
			_pointer, _uniffiStatus)
		return false
	})
	return _uniffiErr
}

func (_self *Node) UpdateChannelConfig(userChannelId UserChannelId, counterpartyNodeId PublicKey, channelConfig *ChannelConfig) error {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_node_update_channel_config(
			_pointer, FfiConverterTypeUserChannelIdINSTANCE.Lower(userChannelId), FfiConverterTypePublicKeyINSTANCE.Lower(counterpartyNodeId), FfiConverterChannelConfigINSTANCE.Lower(channelConfig), _uniffiStatus)
		return false
	})
	return _uniffiErr
}

func (_self *Node) VerifySignature(msg []uint8, sig string, pkey PublicKey) bool {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterBoolINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) C.int8_t {
		return C.uniffi_ldk_node_fn_method_node_verify_signature(
			_pointer, FfiConverterSequenceUint8INSTANCE.Lower(msg), FfiConverterStringINSTANCE.Lower(sig), FfiConverterTypePublicKeyINSTANCE.Lower(pkey), _uniffiStatus)
	}))
}

func (_self *Node) WaitNextEvent() Event {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterTypeEventINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_node_wait_next_event(
			_pointer, _uniffiStatus)
	}))
}

func (object *Node) Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterNode struct{}

var FfiConverterNodeINSTANCE = FfiConverterNode{}

func (c FfiConverterNode) Lift(pointer unsafe.Pointer) *Node {
	result := &Node{
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_ldk_node_fn_free_node(pointer, status)
			}),
	}
	runtime.SetFinalizer(result, (*Node).Destroy)
	return result
}

func (c FfiConverterNode) Read(reader io.Reader) *Node {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterNode) Lower(value *Node) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := value.ffiObject.incrementPointer("*Node")
	defer value.ffiObject.decrementPointer()
	return pointer
}

func (c FfiConverterNode) Write(writer io.Writer, value *Node) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerNode struct{}

func (_ FfiDestroyerNode) Destroy(value *Node) {
	value.Destroy()
}

type OnchainPayment struct {
	ffiObject FfiObject
}

func (_self *OnchainPayment) NewAddress() (Address, error) {
	_pointer := _self.ffiObject.incrementPointer("*OnchainPayment")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_onchainpayment_new_address(
			_pointer, _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue Address
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterTypeAddressINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func (_self *OnchainPayment) SendAllToAddress(address Address) (Txid, error) {
	_pointer := _self.ffiObject.incrementPointer("*OnchainPayment")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_onchainpayment_send_all_to_address(
			_pointer, FfiConverterTypeAddressINSTANCE.Lower(address), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue Txid
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterTypeTxidINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func (_self *OnchainPayment) SendToAddress(address Address, amountMsat uint64) (Txid, error) {
	_pointer := _self.ffiObject.incrementPointer("*OnchainPayment")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_onchainpayment_send_to_address(
			_pointer, FfiConverterTypeAddressINSTANCE.Lower(address), FfiConverterUint64INSTANCE.Lower(amountMsat), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue Txid
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterTypeTxidINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func (object *OnchainPayment) Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterOnchainPayment struct{}

var FfiConverterOnchainPaymentINSTANCE = FfiConverterOnchainPayment{}

func (c FfiConverterOnchainPayment) Lift(pointer unsafe.Pointer) *OnchainPayment {
	result := &OnchainPayment{
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_ldk_node_fn_free_onchainpayment(pointer, status)
			}),
	}
	runtime.SetFinalizer(result, (*OnchainPayment).Destroy)
	return result
}

func (c FfiConverterOnchainPayment) Read(reader io.Reader) *OnchainPayment {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterOnchainPayment) Lower(value *OnchainPayment) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := value.ffiObject.incrementPointer("*OnchainPayment")
	defer value.ffiObject.decrementPointer()
	return pointer
}

func (c FfiConverterOnchainPayment) Write(writer io.Writer, value *OnchainPayment) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerOnchainPayment struct{}

func (_ FfiDestroyerOnchainPayment) Destroy(value *OnchainPayment) {
	value.Destroy()
}

type SpontaneousPayment struct {
	ffiObject FfiObject
}

func (_self *SpontaneousPayment) Send(amountMsat uint64, nodeId PublicKey, customTlvs []TlvEntry) (PaymentId, error) {
	_pointer := _self.ffiObject.incrementPointer("*SpontaneousPayment")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_spontaneouspayment_send(
			_pointer, FfiConverterUint64INSTANCE.Lower(amountMsat), FfiConverterTypePublicKeyINSTANCE.Lower(nodeId), FfiConverterSequenceTypeTlvEntryINSTANCE.Lower(customTlvs), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue PaymentId
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterTypePaymentIdINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func (_self *SpontaneousPayment) SendProbes(amountMsat uint64, nodeId PublicKey) error {
	_pointer := _self.ffiObject.incrementPointer("*SpontaneousPayment")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_spontaneouspayment_send_probes(
			_pointer, FfiConverterUint64INSTANCE.Lower(amountMsat), FfiConverterTypePublicKeyINSTANCE.Lower(nodeId), _uniffiStatus)
		return false
	})
	return _uniffiErr
}

func (object *SpontaneousPayment) Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterSpontaneousPayment struct{}

var FfiConverterSpontaneousPaymentINSTANCE = FfiConverterSpontaneousPayment{}

func (c FfiConverterSpontaneousPayment) Lift(pointer unsafe.Pointer) *SpontaneousPayment {
	result := &SpontaneousPayment{
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_ldk_node_fn_free_spontaneouspayment(pointer, status)
			}),
	}
	runtime.SetFinalizer(result, (*SpontaneousPayment).Destroy)
	return result
}

func (c FfiConverterSpontaneousPayment) Read(reader io.Reader) *SpontaneousPayment {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterSpontaneousPayment) Lower(value *SpontaneousPayment) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := value.ffiObject.incrementPointer("*SpontaneousPayment")
	defer value.ffiObject.decrementPointer()
	return pointer
}

func (c FfiConverterSpontaneousPayment) Write(writer io.Writer, value *SpontaneousPayment) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerSpontaneousPayment struct{}

func (_ FfiDestroyerSpontaneousPayment) Destroy(value *SpontaneousPayment) {
	value.Destroy()
}

type AnchorChannelsConfig struct {
	TrustedPeersNoReserve []PublicKey
	PerChannelReserveSats uint64
}

func (r *AnchorChannelsConfig) Destroy() {
	FfiDestroyerSequenceTypePublicKey{}.Destroy(r.TrustedPeersNoReserve)
	FfiDestroyerUint64{}.Destroy(r.PerChannelReserveSats)
}

type FfiConverterTypeAnchorChannelsConfig struct{}

var FfiConverterTypeAnchorChannelsConfigINSTANCE = FfiConverterTypeAnchorChannelsConfig{}

func (c FfiConverterTypeAnchorChannelsConfig) Lift(rb RustBufferI) AnchorChannelsConfig {
	return LiftFromRustBuffer[AnchorChannelsConfig](c, rb)
}

func (c FfiConverterTypeAnchorChannelsConfig) Read(reader io.Reader) AnchorChannelsConfig {
	return AnchorChannelsConfig{
		FfiConverterSequenceTypePublicKeyINSTANCE.Read(reader),
		FfiConverterUint64INSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeAnchorChannelsConfig) Lower(value AnchorChannelsConfig) RustBuffer {
	return LowerIntoRustBuffer[AnchorChannelsConfig](c, value)
}

func (c FfiConverterTypeAnchorChannelsConfig) Write(writer io.Writer, value AnchorChannelsConfig) {
	FfiConverterSequenceTypePublicKeyINSTANCE.Write(writer, value.TrustedPeersNoReserve)
	FfiConverterUint64INSTANCE.Write(writer, value.PerChannelReserveSats)
}

type FfiDestroyerTypeAnchorChannelsConfig struct{}

func (_ FfiDestroyerTypeAnchorChannelsConfig) Destroy(value AnchorChannelsConfig) {
	value.Destroy()
}

type BalanceDetails struct {
	TotalOnchainBalanceSats            uint64
	SpendableOnchainBalanceSats        uint64
	TotalAnchorChannelsReserveSats     uint64
	TotalLightningBalanceSats          uint64
	LightningBalances                  []LightningBalance
	PendingBalancesFromChannelClosures []PendingSweepBalance
}

func (r *BalanceDetails) Destroy() {
	FfiDestroyerUint64{}.Destroy(r.TotalOnchainBalanceSats)
	FfiDestroyerUint64{}.Destroy(r.SpendableOnchainBalanceSats)
	FfiDestroyerUint64{}.Destroy(r.TotalAnchorChannelsReserveSats)
	FfiDestroyerUint64{}.Destroy(r.TotalLightningBalanceSats)
	FfiDestroyerSequenceTypeLightningBalance{}.Destroy(r.LightningBalances)
	FfiDestroyerSequenceTypePendingSweepBalance{}.Destroy(r.PendingBalancesFromChannelClosures)
}

type FfiConverterTypeBalanceDetails struct{}

var FfiConverterTypeBalanceDetailsINSTANCE = FfiConverterTypeBalanceDetails{}

func (c FfiConverterTypeBalanceDetails) Lift(rb RustBufferI) BalanceDetails {
	return LiftFromRustBuffer[BalanceDetails](c, rb)
}

func (c FfiConverterTypeBalanceDetails) Read(reader io.Reader) BalanceDetails {
	return BalanceDetails{
		FfiConverterUint64INSTANCE.Read(reader),
		FfiConverterUint64INSTANCE.Read(reader),
		FfiConverterUint64INSTANCE.Read(reader),
		FfiConverterUint64INSTANCE.Read(reader),
		FfiConverterSequenceTypeLightningBalanceINSTANCE.Read(reader),
		FfiConverterSequenceTypePendingSweepBalanceINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeBalanceDetails) Lower(value BalanceDetails) RustBuffer {
	return LowerIntoRustBuffer[BalanceDetails](c, value)
}

func (c FfiConverterTypeBalanceDetails) Write(writer io.Writer, value BalanceDetails) {
	FfiConverterUint64INSTANCE.Write(writer, value.TotalOnchainBalanceSats)
	FfiConverterUint64INSTANCE.Write(writer, value.SpendableOnchainBalanceSats)
	FfiConverterUint64INSTANCE.Write(writer, value.TotalAnchorChannelsReserveSats)
	FfiConverterUint64INSTANCE.Write(writer, value.TotalLightningBalanceSats)
	FfiConverterSequenceTypeLightningBalanceINSTANCE.Write(writer, value.LightningBalances)
	FfiConverterSequenceTypePendingSweepBalanceINSTANCE.Write(writer, value.PendingBalancesFromChannelClosures)
}

type FfiDestroyerTypeBalanceDetails struct{}

func (_ FfiDestroyerTypeBalanceDetails) Destroy(value BalanceDetails) {
	value.Destroy()
}

type BestBlock struct {
	BlockHash BlockHash
	Height    uint32
}

func (r *BestBlock) Destroy() {
	FfiDestroyerTypeBlockHash{}.Destroy(r.BlockHash)
	FfiDestroyerUint32{}.Destroy(r.Height)
}

type FfiConverterTypeBestBlock struct{}

var FfiConverterTypeBestBlockINSTANCE = FfiConverterTypeBestBlock{}

func (c FfiConverterTypeBestBlock) Lift(rb RustBufferI) BestBlock {
	return LiftFromRustBuffer[BestBlock](c, rb)
}

func (c FfiConverterTypeBestBlock) Read(reader io.Reader) BestBlock {
	return BestBlock{
		FfiConverterTypeBlockHashINSTANCE.Read(reader),
		FfiConverterUint32INSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeBestBlock) Lower(value BestBlock) RustBuffer {
	return LowerIntoRustBuffer[BestBlock](c, value)
}

func (c FfiConverterTypeBestBlock) Write(writer io.Writer, value BestBlock) {
	FfiConverterTypeBlockHashINSTANCE.Write(writer, value.BlockHash)
	FfiConverterUint32INSTANCE.Write(writer, value.Height)
}

type FfiDestroyerTypeBestBlock struct{}

func (_ FfiDestroyerTypeBestBlock) Destroy(value BestBlock) {
	value.Destroy()
}

type ChannelDetails struct {
	ChannelId                                           ChannelId
	CounterpartyNodeId                                  PublicKey
	FundingTxo                                          *OutPoint
	ChannelType                                         *ChannelType
	ChannelValueSats                                    uint64
	UnspendablePunishmentReserve                        *uint64
	UserChannelId                                       UserChannelId
	FeerateSatPer1000Weight                             uint32
	OutboundCapacityMsat                                uint64
	InboundCapacityMsat                                 uint64
	ConfirmationsRequired                               *uint32
	Confirmations                                       *uint32
	IsOutbound                                          bool
	IsChannelReady                                      bool
	IsUsable                                            bool
	IsPublic                                            bool
	CltvExpiryDelta                                     *uint16
	CounterpartyUnspendablePunishmentReserve            uint64
	CounterpartyOutboundHtlcMinimumMsat                 *uint64
	CounterpartyOutboundHtlcMaximumMsat                 *uint64
	CounterpartyForwardingInfoFeeBaseMsat               *uint32
	CounterpartyForwardingInfoFeeProportionalMillionths *uint32
	CounterpartyForwardingInfoCltvExpiryDelta           *uint16
	NextOutboundHtlcLimitMsat                           uint64
	NextOutboundHtlcMinimumMsat                         uint64
	ForceCloseSpendDelay                                *uint16
	InboundHtlcMinimumMsat                              uint64
	InboundHtlcMaximumMsat                              *uint64
	Config                                              *ChannelConfig
}

func (r *ChannelDetails) Destroy() {
	FfiDestroyerTypeChannelId{}.Destroy(r.ChannelId)
	FfiDestroyerTypePublicKey{}.Destroy(r.CounterpartyNodeId)
	FfiDestroyerOptionalTypeOutPoint{}.Destroy(r.FundingTxo)
	FfiDestroyerOptionalTypeChannelType{}.Destroy(r.ChannelType)
	FfiDestroyerUint64{}.Destroy(r.ChannelValueSats)
	FfiDestroyerOptionalUint64{}.Destroy(r.UnspendablePunishmentReserve)
	FfiDestroyerTypeUserChannelId{}.Destroy(r.UserChannelId)
	FfiDestroyerUint32{}.Destroy(r.FeerateSatPer1000Weight)
	FfiDestroyerUint64{}.Destroy(r.OutboundCapacityMsat)
	FfiDestroyerUint64{}.Destroy(r.InboundCapacityMsat)
	FfiDestroyerOptionalUint32{}.Destroy(r.ConfirmationsRequired)
	FfiDestroyerOptionalUint32{}.Destroy(r.Confirmations)
	FfiDestroyerBool{}.Destroy(r.IsOutbound)
	FfiDestroyerBool{}.Destroy(r.IsChannelReady)
	FfiDestroyerBool{}.Destroy(r.IsUsable)
	FfiDestroyerBool{}.Destroy(r.IsPublic)
	FfiDestroyerOptionalUint16{}.Destroy(r.CltvExpiryDelta)
	FfiDestroyerUint64{}.Destroy(r.CounterpartyUnspendablePunishmentReserve)
	FfiDestroyerOptionalUint64{}.Destroy(r.CounterpartyOutboundHtlcMinimumMsat)
	FfiDestroyerOptionalUint64{}.Destroy(r.CounterpartyOutboundHtlcMaximumMsat)
	FfiDestroyerOptionalUint32{}.Destroy(r.CounterpartyForwardingInfoFeeBaseMsat)
	FfiDestroyerOptionalUint32{}.Destroy(r.CounterpartyForwardingInfoFeeProportionalMillionths)
	FfiDestroyerOptionalUint16{}.Destroy(r.CounterpartyForwardingInfoCltvExpiryDelta)
	FfiDestroyerUint64{}.Destroy(r.NextOutboundHtlcLimitMsat)
	FfiDestroyerUint64{}.Destroy(r.NextOutboundHtlcMinimumMsat)
	FfiDestroyerOptionalUint16{}.Destroy(r.ForceCloseSpendDelay)
	FfiDestroyerUint64{}.Destroy(r.InboundHtlcMinimumMsat)
	FfiDestroyerOptionalUint64{}.Destroy(r.InboundHtlcMaximumMsat)
	FfiDestroyerChannelConfig{}.Destroy(r.Config)
}

type FfiConverterTypeChannelDetails struct{}

var FfiConverterTypeChannelDetailsINSTANCE = FfiConverterTypeChannelDetails{}

func (c FfiConverterTypeChannelDetails) Lift(rb RustBufferI) ChannelDetails {
	return LiftFromRustBuffer[ChannelDetails](c, rb)
}

func (c FfiConverterTypeChannelDetails) Read(reader io.Reader) ChannelDetails {
	return ChannelDetails{
		FfiConverterTypeChannelIdINSTANCE.Read(reader),
		FfiConverterTypePublicKeyINSTANCE.Read(reader),
		FfiConverterOptionalTypeOutPointINSTANCE.Read(reader),
		FfiConverterOptionalTypeChannelTypeINSTANCE.Read(reader),
		FfiConverterUint64INSTANCE.Read(reader),
		FfiConverterOptionalUint64INSTANCE.Read(reader),
		FfiConverterTypeUserChannelIdINSTANCE.Read(reader),
		FfiConverterUint32INSTANCE.Read(reader),
		FfiConverterUint64INSTANCE.Read(reader),
		FfiConverterUint64INSTANCE.Read(reader),
		FfiConverterOptionalUint32INSTANCE.Read(reader),
		FfiConverterOptionalUint32INSTANCE.Read(reader),
		FfiConverterBoolINSTANCE.Read(reader),
		FfiConverterBoolINSTANCE.Read(reader),
		FfiConverterBoolINSTANCE.Read(reader),
		FfiConverterBoolINSTANCE.Read(reader),
		FfiConverterOptionalUint16INSTANCE.Read(reader),
		FfiConverterUint64INSTANCE.Read(reader),
		FfiConverterOptionalUint64INSTANCE.Read(reader),
		FfiConverterOptionalUint64INSTANCE.Read(reader),
		FfiConverterOptionalUint32INSTANCE.Read(reader),
		FfiConverterOptionalUint32INSTANCE.Read(reader),
		FfiConverterOptionalUint16INSTANCE.Read(reader),
		FfiConverterUint64INSTANCE.Read(reader),
		FfiConverterUint64INSTANCE.Read(reader),
		FfiConverterOptionalUint16INSTANCE.Read(reader),
		FfiConverterUint64INSTANCE.Read(reader),
		FfiConverterOptionalUint64INSTANCE.Read(reader),
		FfiConverterChannelConfigINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeChannelDetails) Lower(value ChannelDetails) RustBuffer {
	return LowerIntoRustBuffer[ChannelDetails](c, value)
}

func (c FfiConverterTypeChannelDetails) Write(writer io.Writer, value ChannelDetails) {
	FfiConverterTypeChannelIdINSTANCE.Write(writer, value.ChannelId)
	FfiConverterTypePublicKeyINSTANCE.Write(writer, value.CounterpartyNodeId)
	FfiConverterOptionalTypeOutPointINSTANCE.Write(writer, value.FundingTxo)
	FfiConverterOptionalTypeChannelTypeINSTANCE.Write(writer, value.ChannelType)
	FfiConverterUint64INSTANCE.Write(writer, value.ChannelValueSats)
	FfiConverterOptionalUint64INSTANCE.Write(writer, value.UnspendablePunishmentReserve)
	FfiConverterTypeUserChannelIdINSTANCE.Write(writer, value.UserChannelId)
	FfiConverterUint32INSTANCE.Write(writer, value.FeerateSatPer1000Weight)
	FfiConverterUint64INSTANCE.Write(writer, value.OutboundCapacityMsat)
	FfiConverterUint64INSTANCE.Write(writer, value.InboundCapacityMsat)
	FfiConverterOptionalUint32INSTANCE.Write(writer, value.ConfirmationsRequired)
	FfiConverterOptionalUint32INSTANCE.Write(writer, value.Confirmations)
	FfiConverterBoolINSTANCE.Write(writer, value.IsOutbound)
	FfiConverterBoolINSTANCE.Write(writer, value.IsChannelReady)
	FfiConverterBoolINSTANCE.Write(writer, value.IsUsable)
	FfiConverterBoolINSTANCE.Write(writer, value.IsPublic)
	FfiConverterOptionalUint16INSTANCE.Write(writer, value.CltvExpiryDelta)
	FfiConverterUint64INSTANCE.Write(writer, value.CounterpartyUnspendablePunishmentReserve)
	FfiConverterOptionalUint64INSTANCE.Write(writer, value.CounterpartyOutboundHtlcMinimumMsat)
	FfiConverterOptionalUint64INSTANCE.Write(writer, value.CounterpartyOutboundHtlcMaximumMsat)
	FfiConverterOptionalUint32INSTANCE.Write(writer, value.CounterpartyForwardingInfoFeeBaseMsat)
	FfiConverterOptionalUint32INSTANCE.Write(writer, value.CounterpartyForwardingInfoFeeProportionalMillionths)
	FfiConverterOptionalUint16INSTANCE.Write(writer, value.CounterpartyForwardingInfoCltvExpiryDelta)
	FfiConverterUint64INSTANCE.Write(writer, value.NextOutboundHtlcLimitMsat)
	FfiConverterUint64INSTANCE.Write(writer, value.NextOutboundHtlcMinimumMsat)
	FfiConverterOptionalUint16INSTANCE.Write(writer, value.ForceCloseSpendDelay)
	FfiConverterUint64INSTANCE.Write(writer, value.InboundHtlcMinimumMsat)
	FfiConverterOptionalUint64INSTANCE.Write(writer, value.InboundHtlcMaximumMsat)
	FfiConverterChannelConfigINSTANCE.Write(writer, value.Config)
}

type FfiDestroyerTypeChannelDetails struct{}

func (_ FfiDestroyerTypeChannelDetails) Destroy(value ChannelDetails) {
	value.Destroy()
}

type Config struct {
	StorageDirPath                  string
	LogDirPath                      *string
	Network                         Network
	ListeningAddresses              *[]SocketAddress
	DefaultCltvExpiryDelta          uint32
	OnchainWalletSyncIntervalSecs   uint64
	WalletSyncIntervalSecs          uint64
	FeeRateCacheUpdateIntervalSecs  uint64
	TrustedPeers0conf               []PublicKey
	ProbingLiquidityLimitMultiplier uint64
	LogLevel                        LogLevel
	AnchorChannelsConfig            *AnchorChannelsConfig
}

func (r *Config) Destroy() {
	FfiDestroyerString{}.Destroy(r.StorageDirPath)
	FfiDestroyerOptionalString{}.Destroy(r.LogDirPath)
	FfiDestroyerTypeNetwork{}.Destroy(r.Network)
	FfiDestroyerOptionalSequenceTypeSocketAddress{}.Destroy(r.ListeningAddresses)
	FfiDestroyerUint32{}.Destroy(r.DefaultCltvExpiryDelta)
	FfiDestroyerUint64{}.Destroy(r.OnchainWalletSyncIntervalSecs)
	FfiDestroyerUint64{}.Destroy(r.WalletSyncIntervalSecs)
	FfiDestroyerUint64{}.Destroy(r.FeeRateCacheUpdateIntervalSecs)
	FfiDestroyerSequenceTypePublicKey{}.Destroy(r.TrustedPeers0conf)
	FfiDestroyerUint64{}.Destroy(r.ProbingLiquidityLimitMultiplier)
	FfiDestroyerTypeLogLevel{}.Destroy(r.LogLevel)
	FfiDestroyerOptionalTypeAnchorChannelsConfig{}.Destroy(r.AnchorChannelsConfig)
}

type FfiConverterTypeConfig struct{}

var FfiConverterTypeConfigINSTANCE = FfiConverterTypeConfig{}

func (c FfiConverterTypeConfig) Lift(rb RustBufferI) Config {
	return LiftFromRustBuffer[Config](c, rb)
}

func (c FfiConverterTypeConfig) Read(reader io.Reader) Config {
	return Config{
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterOptionalStringINSTANCE.Read(reader),
		FfiConverterTypeNetworkINSTANCE.Read(reader),
		FfiConverterOptionalSequenceTypeSocketAddressINSTANCE.Read(reader),
		FfiConverterUint32INSTANCE.Read(reader),
		FfiConverterUint64INSTANCE.Read(reader),
		FfiConverterUint64INSTANCE.Read(reader),
		FfiConverterUint64INSTANCE.Read(reader),
		FfiConverterSequenceTypePublicKeyINSTANCE.Read(reader),
		FfiConverterUint64INSTANCE.Read(reader),
		FfiConverterTypeLogLevelINSTANCE.Read(reader),
		FfiConverterOptionalTypeAnchorChannelsConfigINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeConfig) Lower(value Config) RustBuffer {
	return LowerIntoRustBuffer[Config](c, value)
}

func (c FfiConverterTypeConfig) Write(writer io.Writer, value Config) {
	FfiConverterStringINSTANCE.Write(writer, value.StorageDirPath)
	FfiConverterOptionalStringINSTANCE.Write(writer, value.LogDirPath)
	FfiConverterTypeNetworkINSTANCE.Write(writer, value.Network)
	FfiConverterOptionalSequenceTypeSocketAddressINSTANCE.Write(writer, value.ListeningAddresses)
	FfiConverterUint32INSTANCE.Write(writer, value.DefaultCltvExpiryDelta)
	FfiConverterUint64INSTANCE.Write(writer, value.OnchainWalletSyncIntervalSecs)
	FfiConverterUint64INSTANCE.Write(writer, value.WalletSyncIntervalSecs)
	FfiConverterUint64INSTANCE.Write(writer, value.FeeRateCacheUpdateIntervalSecs)
	FfiConverterSequenceTypePublicKeyINSTANCE.Write(writer, value.TrustedPeers0conf)
	FfiConverterUint64INSTANCE.Write(writer, value.ProbingLiquidityLimitMultiplier)
	FfiConverterTypeLogLevelINSTANCE.Write(writer, value.LogLevel)
	FfiConverterOptionalTypeAnchorChannelsConfigINSTANCE.Write(writer, value.AnchorChannelsConfig)
}

type FfiDestroyerTypeConfig struct{}

func (_ FfiDestroyerTypeConfig) Destroy(value Config) {
	value.Destroy()
}

type LspFeeLimits struct {
	MaxTotalOpeningFeeMsat           *uint64
	MaxProportionalOpeningFeePpmMsat *uint64
}

func (r *LspFeeLimits) Destroy() {
	FfiDestroyerOptionalUint64{}.Destroy(r.MaxTotalOpeningFeeMsat)
	FfiDestroyerOptionalUint64{}.Destroy(r.MaxProportionalOpeningFeePpmMsat)
}

type FfiConverterTypeLSPFeeLimits struct{}

var FfiConverterTypeLSPFeeLimitsINSTANCE = FfiConverterTypeLSPFeeLimits{}

func (c FfiConverterTypeLSPFeeLimits) Lift(rb RustBufferI) LspFeeLimits {
	return LiftFromRustBuffer[LspFeeLimits](c, rb)
}

func (c FfiConverterTypeLSPFeeLimits) Read(reader io.Reader) LspFeeLimits {
	return LspFeeLimits{
		FfiConverterOptionalUint64INSTANCE.Read(reader),
		FfiConverterOptionalUint64INSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeLSPFeeLimits) Lower(value LspFeeLimits) RustBuffer {
	return LowerIntoRustBuffer[LspFeeLimits](c, value)
}

func (c FfiConverterTypeLSPFeeLimits) Write(writer io.Writer, value LspFeeLimits) {
	FfiConverterOptionalUint64INSTANCE.Write(writer, value.MaxTotalOpeningFeeMsat)
	FfiConverterOptionalUint64INSTANCE.Write(writer, value.MaxProportionalOpeningFeePpmMsat)
}

type FfiDestroyerTypeLspFeeLimits struct{}

func (_ FfiDestroyerTypeLspFeeLimits) Destroy(value LspFeeLimits) {
	value.Destroy()
}

type NodeStatus struct {
	IsRunning                                bool
	IsListening                              bool
	CurrentBestBlock                         BestBlock
	LatestWalletSyncTimestamp                *uint64
	LatestOnchainWalletSyncTimestamp         *uint64
	LatestFeeRateCacheUpdateTimestamp        *uint64
	LatestRgsSnapshotTimestamp               *uint64
	LatestNodeAnnouncementBroadcastTimestamp *uint64
}

func (r *NodeStatus) Destroy() {
	FfiDestroyerBool{}.Destroy(r.IsRunning)
	FfiDestroyerBool{}.Destroy(r.IsListening)
	FfiDestroyerTypeBestBlock{}.Destroy(r.CurrentBestBlock)
	FfiDestroyerOptionalUint64{}.Destroy(r.LatestWalletSyncTimestamp)
	FfiDestroyerOptionalUint64{}.Destroy(r.LatestOnchainWalletSyncTimestamp)
	FfiDestroyerOptionalUint64{}.Destroy(r.LatestFeeRateCacheUpdateTimestamp)
	FfiDestroyerOptionalUint64{}.Destroy(r.LatestRgsSnapshotTimestamp)
	FfiDestroyerOptionalUint64{}.Destroy(r.LatestNodeAnnouncementBroadcastTimestamp)
}

type FfiConverterTypeNodeStatus struct{}

var FfiConverterTypeNodeStatusINSTANCE = FfiConverterTypeNodeStatus{}

func (c FfiConverterTypeNodeStatus) Lift(rb RustBufferI) NodeStatus {
	return LiftFromRustBuffer[NodeStatus](c, rb)
}

func (c FfiConverterTypeNodeStatus) Read(reader io.Reader) NodeStatus {
	return NodeStatus{
		FfiConverterBoolINSTANCE.Read(reader),
		FfiConverterBoolINSTANCE.Read(reader),
		FfiConverterTypeBestBlockINSTANCE.Read(reader),
		FfiConverterOptionalUint64INSTANCE.Read(reader),
		FfiConverterOptionalUint64INSTANCE.Read(reader),
		FfiConverterOptionalUint64INSTANCE.Read(reader),
		FfiConverterOptionalUint64INSTANCE.Read(reader),
		FfiConverterOptionalUint64INSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeNodeStatus) Lower(value NodeStatus) RustBuffer {
	return LowerIntoRustBuffer[NodeStatus](c, value)
}

func (c FfiConverterTypeNodeStatus) Write(writer io.Writer, value NodeStatus) {
	FfiConverterBoolINSTANCE.Write(writer, value.IsRunning)
	FfiConverterBoolINSTANCE.Write(writer, value.IsListening)
	FfiConverterTypeBestBlockINSTANCE.Write(writer, value.CurrentBestBlock)
	FfiConverterOptionalUint64INSTANCE.Write(writer, value.LatestWalletSyncTimestamp)
	FfiConverterOptionalUint64INSTANCE.Write(writer, value.LatestOnchainWalletSyncTimestamp)
	FfiConverterOptionalUint64INSTANCE.Write(writer, value.LatestFeeRateCacheUpdateTimestamp)
	FfiConverterOptionalUint64INSTANCE.Write(writer, value.LatestRgsSnapshotTimestamp)
	FfiConverterOptionalUint64INSTANCE.Write(writer, value.LatestNodeAnnouncementBroadcastTimestamp)
}

type FfiDestroyerTypeNodeStatus struct{}

func (_ FfiDestroyerTypeNodeStatus) Destroy(value NodeStatus) {
	value.Destroy()
}

type OutPoint struct {
	Txid Txid
	Vout uint32
}

func (r *OutPoint) Destroy() {
	FfiDestroyerTypeTxid{}.Destroy(r.Txid)
	FfiDestroyerUint32{}.Destroy(r.Vout)
}

type FfiConverterTypeOutPoint struct{}

var FfiConverterTypeOutPointINSTANCE = FfiConverterTypeOutPoint{}

func (c FfiConverterTypeOutPoint) Lift(rb RustBufferI) OutPoint {
	return LiftFromRustBuffer[OutPoint](c, rb)
}

func (c FfiConverterTypeOutPoint) Read(reader io.Reader) OutPoint {
	return OutPoint{
		FfiConverterTypeTxidINSTANCE.Read(reader),
		FfiConverterUint32INSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeOutPoint) Lower(value OutPoint) RustBuffer {
	return LowerIntoRustBuffer[OutPoint](c, value)
}

func (c FfiConverterTypeOutPoint) Write(writer io.Writer, value OutPoint) {
	FfiConverterTypeTxidINSTANCE.Write(writer, value.Txid)
	FfiConverterUint32INSTANCE.Write(writer, value.Vout)
}

type FfiDestroyerTypeOutPoint struct{}

func (_ FfiDestroyerTypeOutPoint) Destroy(value OutPoint) {
	value.Destroy()
}

type PaymentDetails struct {
	Id         PaymentId
	Kind       PaymentKind
	AmountMsat *uint64
	Direction  PaymentDirection
	Status     PaymentStatus
}

func (r *PaymentDetails) Destroy() {
	FfiDestroyerTypePaymentId{}.Destroy(r.Id)
	FfiDestroyerTypePaymentKind{}.Destroy(r.Kind)
	FfiDestroyerOptionalUint64{}.Destroy(r.AmountMsat)
	FfiDestroyerTypePaymentDirection{}.Destroy(r.Direction)
	FfiDestroyerTypePaymentStatus{}.Destroy(r.Status)
}

type FfiConverterTypePaymentDetails struct{}

var FfiConverterTypePaymentDetailsINSTANCE = FfiConverterTypePaymentDetails{}

func (c FfiConverterTypePaymentDetails) Lift(rb RustBufferI) PaymentDetails {
	return LiftFromRustBuffer[PaymentDetails](c, rb)
}

func (c FfiConverterTypePaymentDetails) Read(reader io.Reader) PaymentDetails {
	return PaymentDetails{
		FfiConverterTypePaymentIdINSTANCE.Read(reader),
		FfiConverterTypePaymentKindINSTANCE.Read(reader),
		FfiConverterOptionalUint64INSTANCE.Read(reader),
		FfiConverterTypePaymentDirectionINSTANCE.Read(reader),
		FfiConverterTypePaymentStatusINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypePaymentDetails) Lower(value PaymentDetails) RustBuffer {
	return LowerIntoRustBuffer[PaymentDetails](c, value)
}

func (c FfiConverterTypePaymentDetails) Write(writer io.Writer, value PaymentDetails) {
	FfiConverterTypePaymentIdINSTANCE.Write(writer, value.Id)
	FfiConverterTypePaymentKindINSTANCE.Write(writer, value.Kind)
	FfiConverterOptionalUint64INSTANCE.Write(writer, value.AmountMsat)
	FfiConverterTypePaymentDirectionINSTANCE.Write(writer, value.Direction)
	FfiConverterTypePaymentStatusINSTANCE.Write(writer, value.Status)
}

type FfiDestroyerTypePaymentDetails struct{}

func (_ FfiDestroyerTypePaymentDetails) Destroy(value PaymentDetails) {
	value.Destroy()
}

type PeerDetails struct {
	NodeId      PublicKey
	Address     SocketAddress
	IsPersisted bool
	IsConnected bool
}

func (r *PeerDetails) Destroy() {
	FfiDestroyerTypePublicKey{}.Destroy(r.NodeId)
	FfiDestroyerTypeSocketAddress{}.Destroy(r.Address)
	FfiDestroyerBool{}.Destroy(r.IsPersisted)
	FfiDestroyerBool{}.Destroy(r.IsConnected)
}

type FfiConverterTypePeerDetails struct{}

var FfiConverterTypePeerDetailsINSTANCE = FfiConverterTypePeerDetails{}

func (c FfiConverterTypePeerDetails) Lift(rb RustBufferI) PeerDetails {
	return LiftFromRustBuffer[PeerDetails](c, rb)
}

func (c FfiConverterTypePeerDetails) Read(reader io.Reader) PeerDetails {
	return PeerDetails{
		FfiConverterTypePublicKeyINSTANCE.Read(reader),
		FfiConverterTypeSocketAddressINSTANCE.Read(reader),
		FfiConverterBoolINSTANCE.Read(reader),
		FfiConverterBoolINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypePeerDetails) Lower(value PeerDetails) RustBuffer {
	return LowerIntoRustBuffer[PeerDetails](c, value)
}

func (c FfiConverterTypePeerDetails) Write(writer io.Writer, value PeerDetails) {
	FfiConverterTypePublicKeyINSTANCE.Write(writer, value.NodeId)
	FfiConverterTypeSocketAddressINSTANCE.Write(writer, value.Address)
	FfiConverterBoolINSTANCE.Write(writer, value.IsPersisted)
	FfiConverterBoolINSTANCE.Write(writer, value.IsConnected)
}

type FfiDestroyerTypePeerDetails struct{}

func (_ FfiDestroyerTypePeerDetails) Destroy(value PeerDetails) {
	value.Destroy()
}

type TlvEntry struct {
	Type  uint64
	Value []uint8
}

func (r *TlvEntry) Destroy() {
	FfiDestroyerUint64{}.Destroy(r.Type)
	FfiDestroyerSequenceUint8{}.Destroy(r.Value)
}

type FfiConverterTypeTlvEntry struct{}

var FfiConverterTypeTlvEntryINSTANCE = FfiConverterTypeTlvEntry{}

func (c FfiConverterTypeTlvEntry) Lift(rb RustBufferI) TlvEntry {
	return LiftFromRustBuffer[TlvEntry](c, rb)
}

func (c FfiConverterTypeTlvEntry) Read(reader io.Reader) TlvEntry {
	return TlvEntry{
		FfiConverterUint64INSTANCE.Read(reader),
		FfiConverterSequenceUint8INSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeTlvEntry) Lower(value TlvEntry) RustBuffer {
	return LowerIntoRustBuffer[TlvEntry](c, value)
}

func (c FfiConverterTypeTlvEntry) Write(writer io.Writer, value TlvEntry) {
	FfiConverterUint64INSTANCE.Write(writer, value.Type)
	FfiConverterSequenceUint8INSTANCE.Write(writer, value.Value)
}

type FfiDestroyerTypeTlvEntry struct{}

func (_ FfiDestroyerTypeTlvEntry) Destroy(value TlvEntry) {
	value.Destroy()
}

type BuildError struct {
	err error
}

func (err BuildError) Error() string {
	return fmt.Sprintf("BuildError: %s", err.err.Error())
}

func (err BuildError) Unwrap() error {
	return err.err
}

// Err* are used for checking error type with `errors.Is`
var ErrBuildErrorInvalidSeedBytes = fmt.Errorf("BuildErrorInvalidSeedBytes")
var ErrBuildErrorInvalidSeedFile = fmt.Errorf("BuildErrorInvalidSeedFile")
var ErrBuildErrorInvalidSystemTime = fmt.Errorf("BuildErrorInvalidSystemTime")
var ErrBuildErrorInvalidChannelMonitor = fmt.Errorf("BuildErrorInvalidChannelMonitor")
var ErrBuildErrorInvalidListeningAddresses = fmt.Errorf("BuildErrorInvalidListeningAddresses")
var ErrBuildErrorReadFailed = fmt.Errorf("BuildErrorReadFailed")
var ErrBuildErrorWriteFailed = fmt.Errorf("BuildErrorWriteFailed")
var ErrBuildErrorStoragePathAccessFailed = fmt.Errorf("BuildErrorStoragePathAccessFailed")
var ErrBuildErrorKvStoreSetupFailed = fmt.Errorf("BuildErrorKvStoreSetupFailed")
var ErrBuildErrorWalletSetupFailed = fmt.Errorf("BuildErrorWalletSetupFailed")
var ErrBuildErrorLoggerSetupFailed = fmt.Errorf("BuildErrorLoggerSetupFailed")

// Variant structs
type BuildErrorInvalidSeedBytes struct {
	message string
}

func NewBuildErrorInvalidSeedBytes() *BuildError {
	return &BuildError{
		err: &BuildErrorInvalidSeedBytes{},
	}
}

func (err BuildErrorInvalidSeedBytes) Error() string {
	return fmt.Sprintf("InvalidSeedBytes: %s", err.message)
}

func (self BuildErrorInvalidSeedBytes) Is(target error) bool {
	return target == ErrBuildErrorInvalidSeedBytes
}

type BuildErrorInvalidSeedFile struct {
	message string
}

func NewBuildErrorInvalidSeedFile() *BuildError {
	return &BuildError{
		err: &BuildErrorInvalidSeedFile{},
	}
}

func (err BuildErrorInvalidSeedFile) Error() string {
	return fmt.Sprintf("InvalidSeedFile: %s", err.message)
}

func (self BuildErrorInvalidSeedFile) Is(target error) bool {
	return target == ErrBuildErrorInvalidSeedFile
}

type BuildErrorInvalidSystemTime struct {
	message string
}

func NewBuildErrorInvalidSystemTime() *BuildError {
	return &BuildError{
		err: &BuildErrorInvalidSystemTime{},
	}
}

func (err BuildErrorInvalidSystemTime) Error() string {
	return fmt.Sprintf("InvalidSystemTime: %s", err.message)
}

func (self BuildErrorInvalidSystemTime) Is(target error) bool {
	return target == ErrBuildErrorInvalidSystemTime
}

type BuildErrorInvalidChannelMonitor struct {
	message string
}

func NewBuildErrorInvalidChannelMonitor() *BuildError {
	return &BuildError{
		err: &BuildErrorInvalidChannelMonitor{},
	}
}

func (err BuildErrorInvalidChannelMonitor) Error() string {
	return fmt.Sprintf("InvalidChannelMonitor: %s", err.message)
}

func (self BuildErrorInvalidChannelMonitor) Is(target error) bool {
	return target == ErrBuildErrorInvalidChannelMonitor
}

type BuildErrorInvalidListeningAddresses struct {
	message string
}

func NewBuildErrorInvalidListeningAddresses() *BuildError {
	return &BuildError{
		err: &BuildErrorInvalidListeningAddresses{},
	}
}

func (err BuildErrorInvalidListeningAddresses) Error() string {
	return fmt.Sprintf("InvalidListeningAddresses: %s", err.message)
}

func (self BuildErrorInvalidListeningAddresses) Is(target error) bool {
	return target == ErrBuildErrorInvalidListeningAddresses
}

type BuildErrorReadFailed struct {
	message string
}

func NewBuildErrorReadFailed() *BuildError {
	return &BuildError{
		err: &BuildErrorReadFailed{},
	}
}

func (err BuildErrorReadFailed) Error() string {
	return fmt.Sprintf("ReadFailed: %s", err.message)
}

func (self BuildErrorReadFailed) Is(target error) bool {
	return target == ErrBuildErrorReadFailed
}

type BuildErrorWriteFailed struct {
	message string
}

func NewBuildErrorWriteFailed() *BuildError {
	return &BuildError{
		err: &BuildErrorWriteFailed{},
	}
}

func (err BuildErrorWriteFailed) Error() string {
	return fmt.Sprintf("WriteFailed: %s", err.message)
}

func (self BuildErrorWriteFailed) Is(target error) bool {
	return target == ErrBuildErrorWriteFailed
}

type BuildErrorStoragePathAccessFailed struct {
	message string
}

func NewBuildErrorStoragePathAccessFailed() *BuildError {
	return &BuildError{
		err: &BuildErrorStoragePathAccessFailed{},
	}
}

func (err BuildErrorStoragePathAccessFailed) Error() string {
	return fmt.Sprintf("StoragePathAccessFailed: %s", err.message)
}

func (self BuildErrorStoragePathAccessFailed) Is(target error) bool {
	return target == ErrBuildErrorStoragePathAccessFailed
}

type BuildErrorKvStoreSetupFailed struct {
	message string
}

func NewBuildErrorKvStoreSetupFailed() *BuildError {
	return &BuildError{
		err: &BuildErrorKvStoreSetupFailed{},
	}
}

func (err BuildErrorKvStoreSetupFailed) Error() string {
	return fmt.Sprintf("KvStoreSetupFailed: %s", err.message)
}

func (self BuildErrorKvStoreSetupFailed) Is(target error) bool {
	return target == ErrBuildErrorKvStoreSetupFailed
}

type BuildErrorWalletSetupFailed struct {
	message string
}

func NewBuildErrorWalletSetupFailed() *BuildError {
	return &BuildError{
		err: &BuildErrorWalletSetupFailed{},
	}
}

func (err BuildErrorWalletSetupFailed) Error() string {
	return fmt.Sprintf("WalletSetupFailed: %s", err.message)
}

func (self BuildErrorWalletSetupFailed) Is(target error) bool {
	return target == ErrBuildErrorWalletSetupFailed
}

type BuildErrorLoggerSetupFailed struct {
	message string
}

func NewBuildErrorLoggerSetupFailed() *BuildError {
	return &BuildError{
		err: &BuildErrorLoggerSetupFailed{},
	}
}

func (err BuildErrorLoggerSetupFailed) Error() string {
	return fmt.Sprintf("LoggerSetupFailed: %s", err.message)
}

func (self BuildErrorLoggerSetupFailed) Is(target error) bool {
	return target == ErrBuildErrorLoggerSetupFailed
}

type FfiConverterTypeBuildError struct{}

var FfiConverterTypeBuildErrorINSTANCE = FfiConverterTypeBuildError{}

func (c FfiConverterTypeBuildError) Lift(eb RustBufferI) error {
	return LiftFromRustBuffer[error](c, eb)
}

func (c FfiConverterTypeBuildError) Lower(value *BuildError) RustBuffer {
	return LowerIntoRustBuffer[*BuildError](c, value)
}

func (c FfiConverterTypeBuildError) Read(reader io.Reader) error {
	errorID := readUint32(reader)

	message := FfiConverterStringINSTANCE.Read(reader)
	switch errorID {
	case 1:
		return &BuildError{&BuildErrorInvalidSeedBytes{message}}
	case 2:
		return &BuildError{&BuildErrorInvalidSeedFile{message}}
	case 3:
		return &BuildError{&BuildErrorInvalidSystemTime{message}}
	case 4:
		return &BuildError{&BuildErrorInvalidChannelMonitor{message}}
	case 5:
		return &BuildError{&BuildErrorInvalidListeningAddresses{message}}
	case 6:
		return &BuildError{&BuildErrorReadFailed{message}}
	case 7:
		return &BuildError{&BuildErrorWriteFailed{message}}
	case 8:
		return &BuildError{&BuildErrorStoragePathAccessFailed{message}}
	case 9:
		return &BuildError{&BuildErrorKvStoreSetupFailed{message}}
	case 10:
		return &BuildError{&BuildErrorWalletSetupFailed{message}}
	case 11:
		return &BuildError{&BuildErrorLoggerSetupFailed{message}}
	default:
		panic(fmt.Sprintf("Unknown error code %d in FfiConverterTypeBuildError.Read()", errorID))
	}

}

func (c FfiConverterTypeBuildError) Write(writer io.Writer, value *BuildError) {
	switch variantValue := value.err.(type) {
	case *BuildErrorInvalidSeedBytes:
		writeInt32(writer, 1)
	case *BuildErrorInvalidSeedFile:
		writeInt32(writer, 2)
	case *BuildErrorInvalidSystemTime:
		writeInt32(writer, 3)
	case *BuildErrorInvalidChannelMonitor:
		writeInt32(writer, 4)
	case *BuildErrorInvalidListeningAddresses:
		writeInt32(writer, 5)
	case *BuildErrorReadFailed:
		writeInt32(writer, 6)
	case *BuildErrorWriteFailed:
		writeInt32(writer, 7)
	case *BuildErrorStoragePathAccessFailed:
		writeInt32(writer, 8)
	case *BuildErrorKvStoreSetupFailed:
		writeInt32(writer, 9)
	case *BuildErrorWalletSetupFailed:
		writeInt32(writer, 10)
	case *BuildErrorLoggerSetupFailed:
		writeInt32(writer, 11)
	default:
		_ = variantValue
		panic(fmt.Sprintf("invalid error value `%v` in FfiConverterTypeBuildError.Write", value))
	}
}

type ChannelType uint

const (
	ChannelTypeStaticRemoteKey ChannelType = 1
	ChannelTypeAnchors         ChannelType = 2
)

type FfiConverterTypeChannelType struct{}

var FfiConverterTypeChannelTypeINSTANCE = FfiConverterTypeChannelType{}

func (c FfiConverterTypeChannelType) Lift(rb RustBufferI) ChannelType {
	return LiftFromRustBuffer[ChannelType](c, rb)
}

func (c FfiConverterTypeChannelType) Lower(value ChannelType) RustBuffer {
	return LowerIntoRustBuffer[ChannelType](c, value)
}
func (FfiConverterTypeChannelType) Read(reader io.Reader) ChannelType {
	id := readInt32(reader)
	return ChannelType(id)
}

func (FfiConverterTypeChannelType) Write(writer io.Writer, value ChannelType) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerTypeChannelType struct{}

func (_ FfiDestroyerTypeChannelType) Destroy(value ChannelType) {
}

type ClosureReason interface {
	Destroy()
}
type ClosureReasonCounterpartyForceClosed struct {
	PeerMsg UntrustedString
}

func (e ClosureReasonCounterpartyForceClosed) Destroy() {
	FfiDestroyerTypeUntrustedString{}.Destroy(e.PeerMsg)
}

type ClosureReasonHolderForceClosed struct {
}

func (e ClosureReasonHolderForceClosed) Destroy() {
}

type ClosureReasonLegacyCooperativeClosure struct {
}

func (e ClosureReasonLegacyCooperativeClosure) Destroy() {
}

type ClosureReasonCounterpartyInitiatedCooperativeClosure struct {
}

func (e ClosureReasonCounterpartyInitiatedCooperativeClosure) Destroy() {
}

type ClosureReasonLocallyInitiatedCooperativeClosure struct {
}

func (e ClosureReasonLocallyInitiatedCooperativeClosure) Destroy() {
}

type ClosureReasonCommitmentTxConfirmed struct {
}

func (e ClosureReasonCommitmentTxConfirmed) Destroy() {
}

type ClosureReasonFundingTimedOut struct {
}

func (e ClosureReasonFundingTimedOut) Destroy() {
}

type ClosureReasonProcessingError struct {
	Err string
}

func (e ClosureReasonProcessingError) Destroy() {
	FfiDestroyerString{}.Destroy(e.Err)
}

type ClosureReasonDisconnectedPeer struct {
}

func (e ClosureReasonDisconnectedPeer) Destroy() {
}

type ClosureReasonOutdatedChannelManager struct {
}

func (e ClosureReasonOutdatedChannelManager) Destroy() {
}

type ClosureReasonCounterpartyCoopClosedUnfundedChannel struct {
}

func (e ClosureReasonCounterpartyCoopClosedUnfundedChannel) Destroy() {
}

type ClosureReasonFundingBatchClosure struct {
}

func (e ClosureReasonFundingBatchClosure) Destroy() {
}

type FfiConverterTypeClosureReason struct{}

var FfiConverterTypeClosureReasonINSTANCE = FfiConverterTypeClosureReason{}

func (c FfiConverterTypeClosureReason) Lift(rb RustBufferI) ClosureReason {
	return LiftFromRustBuffer[ClosureReason](c, rb)
}

func (c FfiConverterTypeClosureReason) Lower(value ClosureReason) RustBuffer {
	return LowerIntoRustBuffer[ClosureReason](c, value)
}
func (FfiConverterTypeClosureReason) Read(reader io.Reader) ClosureReason {
	id := readInt32(reader)
	switch id {
	case 1:
		return ClosureReasonCounterpartyForceClosed{
			FfiConverterTypeUntrustedStringINSTANCE.Read(reader),
		}
	case 2:
		return ClosureReasonHolderForceClosed{}
	case 3:
		return ClosureReasonLegacyCooperativeClosure{}
	case 4:
		return ClosureReasonCounterpartyInitiatedCooperativeClosure{}
	case 5:
		return ClosureReasonLocallyInitiatedCooperativeClosure{}
	case 6:
		return ClosureReasonCommitmentTxConfirmed{}
	case 7:
		return ClosureReasonFundingTimedOut{}
	case 8:
		return ClosureReasonProcessingError{
			FfiConverterStringINSTANCE.Read(reader),
		}
	case 9:
		return ClosureReasonDisconnectedPeer{}
	case 10:
		return ClosureReasonOutdatedChannelManager{}
	case 11:
		return ClosureReasonCounterpartyCoopClosedUnfundedChannel{}
	case 12:
		return ClosureReasonFundingBatchClosure{}
	default:
		panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeClosureReason.Read()", id))
	}
}

func (FfiConverterTypeClosureReason) Write(writer io.Writer, value ClosureReason) {
	switch variant_value := value.(type) {
	case ClosureReasonCounterpartyForceClosed:
		writeInt32(writer, 1)
		FfiConverterTypeUntrustedStringINSTANCE.Write(writer, variant_value.PeerMsg)
	case ClosureReasonHolderForceClosed:
		writeInt32(writer, 2)
	case ClosureReasonLegacyCooperativeClosure:
		writeInt32(writer, 3)
	case ClosureReasonCounterpartyInitiatedCooperativeClosure:
		writeInt32(writer, 4)
	case ClosureReasonLocallyInitiatedCooperativeClosure:
		writeInt32(writer, 5)
	case ClosureReasonCommitmentTxConfirmed:
		writeInt32(writer, 6)
	case ClosureReasonFundingTimedOut:
		writeInt32(writer, 7)
	case ClosureReasonProcessingError:
		writeInt32(writer, 8)
		FfiConverterStringINSTANCE.Write(writer, variant_value.Err)
	case ClosureReasonDisconnectedPeer:
		writeInt32(writer, 9)
	case ClosureReasonOutdatedChannelManager:
		writeInt32(writer, 10)
	case ClosureReasonCounterpartyCoopClosedUnfundedChannel:
		writeInt32(writer, 11)
	case ClosureReasonFundingBatchClosure:
		writeInt32(writer, 12)
	default:
		_ = variant_value
		panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeClosureReason.Write", value))
	}
}

type FfiDestroyerTypeClosureReason struct{}

func (_ FfiDestroyerTypeClosureReason) Destroy(value ClosureReason) {
	value.Destroy()
}

type Event interface {
	Destroy()
}
type EventPaymentSuccessful struct {
	PaymentId   *PaymentId
	PaymentHash PaymentHash
	FeePaidMsat *uint64
}

func (e EventPaymentSuccessful) Destroy() {
	FfiDestroyerOptionalTypePaymentId{}.Destroy(e.PaymentId)
	FfiDestroyerTypePaymentHash{}.Destroy(e.PaymentHash)
	FfiDestroyerOptionalUint64{}.Destroy(e.FeePaidMsat)
}

type EventPaymentFailed struct {
	PaymentId   *PaymentId
	PaymentHash PaymentHash
	Reason      *PaymentFailureReason
}

func (e EventPaymentFailed) Destroy() {
	FfiDestroyerOptionalTypePaymentId{}.Destroy(e.PaymentId)
	FfiDestroyerTypePaymentHash{}.Destroy(e.PaymentHash)
	FfiDestroyerOptionalTypePaymentFailureReason{}.Destroy(e.Reason)
}

type EventPaymentReceived struct {
	PaymentId   *PaymentId
	PaymentHash PaymentHash
	AmountMsat  uint64
}

func (e EventPaymentReceived) Destroy() {
	FfiDestroyerOptionalTypePaymentId{}.Destroy(e.PaymentId)
	FfiDestroyerTypePaymentHash{}.Destroy(e.PaymentHash)
	FfiDestroyerUint64{}.Destroy(e.AmountMsat)
}

type EventChannelPending struct {
	ChannelId                ChannelId
	UserChannelId            UserChannelId
	FormerTemporaryChannelId ChannelId
	CounterpartyNodeId       PublicKey
	FundingTxo               OutPoint
}

func (e EventChannelPending) Destroy() {
	FfiDestroyerTypeChannelId{}.Destroy(e.ChannelId)
	FfiDestroyerTypeUserChannelId{}.Destroy(e.UserChannelId)
	FfiDestroyerTypeChannelId{}.Destroy(e.FormerTemporaryChannelId)
	FfiDestroyerTypePublicKey{}.Destroy(e.CounterpartyNodeId)
	FfiDestroyerTypeOutPoint{}.Destroy(e.FundingTxo)
}

type EventChannelReady struct {
	ChannelId          ChannelId
	UserChannelId      UserChannelId
	CounterpartyNodeId *PublicKey
}

func (e EventChannelReady) Destroy() {
	FfiDestroyerTypeChannelId{}.Destroy(e.ChannelId)
	FfiDestroyerTypeUserChannelId{}.Destroy(e.UserChannelId)
	FfiDestroyerOptionalTypePublicKey{}.Destroy(e.CounterpartyNodeId)
}

type EventChannelClosed struct {
	ChannelId          ChannelId
	UserChannelId      UserChannelId
	CounterpartyNodeId *PublicKey
	Reason             *ClosureReason
}

func (e EventChannelClosed) Destroy() {
	FfiDestroyerTypeChannelId{}.Destroy(e.ChannelId)
	FfiDestroyerTypeUserChannelId{}.Destroy(e.UserChannelId)
	FfiDestroyerOptionalTypePublicKey{}.Destroy(e.CounterpartyNodeId)
	FfiDestroyerOptionalTypeClosureReason{}.Destroy(e.Reason)
}

type FfiConverterTypeEvent struct{}

var FfiConverterTypeEventINSTANCE = FfiConverterTypeEvent{}

func (c FfiConverterTypeEvent) Lift(rb RustBufferI) Event {
	return LiftFromRustBuffer[Event](c, rb)
}

func (c FfiConverterTypeEvent) Lower(value Event) RustBuffer {
	return LowerIntoRustBuffer[Event](c, value)
}
func (FfiConverterTypeEvent) Read(reader io.Reader) Event {
	id := readInt32(reader)
	switch id {
	case 1:
		return EventPaymentSuccessful{
			FfiConverterOptionalTypePaymentIdINSTANCE.Read(reader),
			FfiConverterTypePaymentHashINSTANCE.Read(reader),
			FfiConverterOptionalUint64INSTANCE.Read(reader),
		}
	case 2:
		return EventPaymentFailed{
			FfiConverterOptionalTypePaymentIdINSTANCE.Read(reader),
			FfiConverterTypePaymentHashINSTANCE.Read(reader),
			FfiConverterOptionalTypePaymentFailureReasonINSTANCE.Read(reader),
		}
	case 3:
		return EventPaymentReceived{
			FfiConverterOptionalTypePaymentIdINSTANCE.Read(reader),
			FfiConverterTypePaymentHashINSTANCE.Read(reader),
			FfiConverterUint64INSTANCE.Read(reader),
		}
	case 4:
		return EventChannelPending{
			FfiConverterTypeChannelIdINSTANCE.Read(reader),
			FfiConverterTypeUserChannelIdINSTANCE.Read(reader),
			FfiConverterTypeChannelIdINSTANCE.Read(reader),
			FfiConverterTypePublicKeyINSTANCE.Read(reader),
			FfiConverterTypeOutPointINSTANCE.Read(reader),
		}
	case 5:
		return EventChannelReady{
			FfiConverterTypeChannelIdINSTANCE.Read(reader),
			FfiConverterTypeUserChannelIdINSTANCE.Read(reader),
			FfiConverterOptionalTypePublicKeyINSTANCE.Read(reader),
		}
	case 6:
		return EventChannelClosed{
			FfiConverterTypeChannelIdINSTANCE.Read(reader),
			FfiConverterTypeUserChannelIdINSTANCE.Read(reader),
			FfiConverterOptionalTypePublicKeyINSTANCE.Read(reader),
			FfiConverterOptionalTypeClosureReasonINSTANCE.Read(reader),
		}
	default:
		panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeEvent.Read()", id))
	}
}

func (FfiConverterTypeEvent) Write(writer io.Writer, value Event) {
	switch variant_value := value.(type) {
	case EventPaymentSuccessful:
		writeInt32(writer, 1)
		FfiConverterOptionalTypePaymentIdINSTANCE.Write(writer, variant_value.PaymentId)
		FfiConverterTypePaymentHashINSTANCE.Write(writer, variant_value.PaymentHash)
		FfiConverterOptionalUint64INSTANCE.Write(writer, variant_value.FeePaidMsat)
	case EventPaymentFailed:
		writeInt32(writer, 2)
		FfiConverterOptionalTypePaymentIdINSTANCE.Write(writer, variant_value.PaymentId)
		FfiConverterTypePaymentHashINSTANCE.Write(writer, variant_value.PaymentHash)
		FfiConverterOptionalTypePaymentFailureReasonINSTANCE.Write(writer, variant_value.Reason)
	case EventPaymentReceived:
		writeInt32(writer, 3)
		FfiConverterOptionalTypePaymentIdINSTANCE.Write(writer, variant_value.PaymentId)
		FfiConverterTypePaymentHashINSTANCE.Write(writer, variant_value.PaymentHash)
		FfiConverterUint64INSTANCE.Write(writer, variant_value.AmountMsat)
	case EventChannelPending:
		writeInt32(writer, 4)
		FfiConverterTypeChannelIdINSTANCE.Write(writer, variant_value.ChannelId)
		FfiConverterTypeUserChannelIdINSTANCE.Write(writer, variant_value.UserChannelId)
		FfiConverterTypeChannelIdINSTANCE.Write(writer, variant_value.FormerTemporaryChannelId)
		FfiConverterTypePublicKeyINSTANCE.Write(writer, variant_value.CounterpartyNodeId)
		FfiConverterTypeOutPointINSTANCE.Write(writer, variant_value.FundingTxo)
	case EventChannelReady:
		writeInt32(writer, 5)
		FfiConverterTypeChannelIdINSTANCE.Write(writer, variant_value.ChannelId)
		FfiConverterTypeUserChannelIdINSTANCE.Write(writer, variant_value.UserChannelId)
		FfiConverterOptionalTypePublicKeyINSTANCE.Write(writer, variant_value.CounterpartyNodeId)
	case EventChannelClosed:
		writeInt32(writer, 6)
		FfiConverterTypeChannelIdINSTANCE.Write(writer, variant_value.ChannelId)
		FfiConverterTypeUserChannelIdINSTANCE.Write(writer, variant_value.UserChannelId)
		FfiConverterOptionalTypePublicKeyINSTANCE.Write(writer, variant_value.CounterpartyNodeId)
		FfiConverterOptionalTypeClosureReasonINSTANCE.Write(writer, variant_value.Reason)
	default:
		_ = variant_value
		panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeEvent.Write", value))
	}
}

type FfiDestroyerTypeEvent struct{}

func (_ FfiDestroyerTypeEvent) Destroy(value Event) {
	value.Destroy()
}

type LightningBalance interface {
	Destroy()
}
type LightningBalanceClaimableOnChannelClose struct {
	ChannelId          ChannelId
	CounterpartyNodeId PublicKey
	AmountSatoshis     uint64
}

func (e LightningBalanceClaimableOnChannelClose) Destroy() {
	FfiDestroyerTypeChannelId{}.Destroy(e.ChannelId)
	FfiDestroyerTypePublicKey{}.Destroy(e.CounterpartyNodeId)
	FfiDestroyerUint64{}.Destroy(e.AmountSatoshis)
}

type LightningBalanceClaimableAwaitingConfirmations struct {
	ChannelId          ChannelId
	CounterpartyNodeId PublicKey
	AmountSatoshis     uint64
	ConfirmationHeight uint32
}

func (e LightningBalanceClaimableAwaitingConfirmations) Destroy() {
	FfiDestroyerTypeChannelId{}.Destroy(e.ChannelId)
	FfiDestroyerTypePublicKey{}.Destroy(e.CounterpartyNodeId)
	FfiDestroyerUint64{}.Destroy(e.AmountSatoshis)
	FfiDestroyerUint32{}.Destroy(e.ConfirmationHeight)
}

type LightningBalanceContentiousClaimable struct {
	ChannelId          ChannelId
	CounterpartyNodeId PublicKey
	AmountSatoshis     uint64
	TimeoutHeight      uint32
	PaymentHash        PaymentHash
	PaymentPreimage    PaymentPreimage
}

func (e LightningBalanceContentiousClaimable) Destroy() {
	FfiDestroyerTypeChannelId{}.Destroy(e.ChannelId)
	FfiDestroyerTypePublicKey{}.Destroy(e.CounterpartyNodeId)
	FfiDestroyerUint64{}.Destroy(e.AmountSatoshis)
	FfiDestroyerUint32{}.Destroy(e.TimeoutHeight)
	FfiDestroyerTypePaymentHash{}.Destroy(e.PaymentHash)
	FfiDestroyerTypePaymentPreimage{}.Destroy(e.PaymentPreimage)
}

type LightningBalanceMaybeTimeoutClaimableHtlc struct {
	ChannelId          ChannelId
	CounterpartyNodeId PublicKey
	AmountSatoshis     uint64
	ClaimableHeight    uint32
	PaymentHash        PaymentHash
}

func (e LightningBalanceMaybeTimeoutClaimableHtlc) Destroy() {
	FfiDestroyerTypeChannelId{}.Destroy(e.ChannelId)
	FfiDestroyerTypePublicKey{}.Destroy(e.CounterpartyNodeId)
	FfiDestroyerUint64{}.Destroy(e.AmountSatoshis)
	FfiDestroyerUint32{}.Destroy(e.ClaimableHeight)
	FfiDestroyerTypePaymentHash{}.Destroy(e.PaymentHash)
}

type LightningBalanceMaybePreimageClaimableHtlc struct {
	ChannelId          ChannelId
	CounterpartyNodeId PublicKey
	AmountSatoshis     uint64
	ExpiryHeight       uint32
	PaymentHash        PaymentHash
}

func (e LightningBalanceMaybePreimageClaimableHtlc) Destroy() {
	FfiDestroyerTypeChannelId{}.Destroy(e.ChannelId)
	FfiDestroyerTypePublicKey{}.Destroy(e.CounterpartyNodeId)
	FfiDestroyerUint64{}.Destroy(e.AmountSatoshis)
	FfiDestroyerUint32{}.Destroy(e.ExpiryHeight)
	FfiDestroyerTypePaymentHash{}.Destroy(e.PaymentHash)
}

type LightningBalanceCounterpartyRevokedOutputClaimable struct {
	ChannelId          ChannelId
	CounterpartyNodeId PublicKey
	AmountSatoshis     uint64
}

func (e LightningBalanceCounterpartyRevokedOutputClaimable) Destroy() {
	FfiDestroyerTypeChannelId{}.Destroy(e.ChannelId)
	FfiDestroyerTypePublicKey{}.Destroy(e.CounterpartyNodeId)
	FfiDestroyerUint64{}.Destroy(e.AmountSatoshis)
}

type FfiConverterTypeLightningBalance struct{}

var FfiConverterTypeLightningBalanceINSTANCE = FfiConverterTypeLightningBalance{}

func (c FfiConverterTypeLightningBalance) Lift(rb RustBufferI) LightningBalance {
	return LiftFromRustBuffer[LightningBalance](c, rb)
}

func (c FfiConverterTypeLightningBalance) Lower(value LightningBalance) RustBuffer {
	return LowerIntoRustBuffer[LightningBalance](c, value)
}
func (FfiConverterTypeLightningBalance) Read(reader io.Reader) LightningBalance {
	id := readInt32(reader)
	switch id {
	case 1:
		return LightningBalanceClaimableOnChannelClose{
			FfiConverterTypeChannelIdINSTANCE.Read(reader),
			FfiConverterTypePublicKeyINSTANCE.Read(reader),
			FfiConverterUint64INSTANCE.Read(reader),
		}
	case 2:
		return LightningBalanceClaimableAwaitingConfirmations{
			FfiConverterTypeChannelIdINSTANCE.Read(reader),
			FfiConverterTypePublicKeyINSTANCE.Read(reader),
			FfiConverterUint64INSTANCE.Read(reader),
			FfiConverterUint32INSTANCE.Read(reader),
		}
	case 3:
		return LightningBalanceContentiousClaimable{
			FfiConverterTypeChannelIdINSTANCE.Read(reader),
			FfiConverterTypePublicKeyINSTANCE.Read(reader),
			FfiConverterUint64INSTANCE.Read(reader),
			FfiConverterUint32INSTANCE.Read(reader),
			FfiConverterTypePaymentHashINSTANCE.Read(reader),
			FfiConverterTypePaymentPreimageINSTANCE.Read(reader),
		}
	case 4:
		return LightningBalanceMaybeTimeoutClaimableHtlc{
			FfiConverterTypeChannelIdINSTANCE.Read(reader),
			FfiConverterTypePublicKeyINSTANCE.Read(reader),
			FfiConverterUint64INSTANCE.Read(reader),
			FfiConverterUint32INSTANCE.Read(reader),
			FfiConverterTypePaymentHashINSTANCE.Read(reader),
		}
	case 5:
		return LightningBalanceMaybePreimageClaimableHtlc{
			FfiConverterTypeChannelIdINSTANCE.Read(reader),
			FfiConverterTypePublicKeyINSTANCE.Read(reader),
			FfiConverterUint64INSTANCE.Read(reader),
			FfiConverterUint32INSTANCE.Read(reader),
			FfiConverterTypePaymentHashINSTANCE.Read(reader),
		}
	case 6:
		return LightningBalanceCounterpartyRevokedOutputClaimable{
			FfiConverterTypeChannelIdINSTANCE.Read(reader),
			FfiConverterTypePublicKeyINSTANCE.Read(reader),
			FfiConverterUint64INSTANCE.Read(reader),
		}
	default:
		panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeLightningBalance.Read()", id))
	}
}

func (FfiConverterTypeLightningBalance) Write(writer io.Writer, value LightningBalance) {
	switch variant_value := value.(type) {
	case LightningBalanceClaimableOnChannelClose:
		writeInt32(writer, 1)
		FfiConverterTypeChannelIdINSTANCE.Write(writer, variant_value.ChannelId)
		FfiConverterTypePublicKeyINSTANCE.Write(writer, variant_value.CounterpartyNodeId)
		FfiConverterUint64INSTANCE.Write(writer, variant_value.AmountSatoshis)
	case LightningBalanceClaimableAwaitingConfirmations:
		writeInt32(writer, 2)
		FfiConverterTypeChannelIdINSTANCE.Write(writer, variant_value.ChannelId)
		FfiConverterTypePublicKeyINSTANCE.Write(writer, variant_value.CounterpartyNodeId)
		FfiConverterUint64INSTANCE.Write(writer, variant_value.AmountSatoshis)
		FfiConverterUint32INSTANCE.Write(writer, variant_value.ConfirmationHeight)
	case LightningBalanceContentiousClaimable:
		writeInt32(writer, 3)
		FfiConverterTypeChannelIdINSTANCE.Write(writer, variant_value.ChannelId)
		FfiConverterTypePublicKeyINSTANCE.Write(writer, variant_value.CounterpartyNodeId)
		FfiConverterUint64INSTANCE.Write(writer, variant_value.AmountSatoshis)
		FfiConverterUint32INSTANCE.Write(writer, variant_value.TimeoutHeight)
		FfiConverterTypePaymentHashINSTANCE.Write(writer, variant_value.PaymentHash)
		FfiConverterTypePaymentPreimageINSTANCE.Write(writer, variant_value.PaymentPreimage)
	case LightningBalanceMaybeTimeoutClaimableHtlc:
		writeInt32(writer, 4)
		FfiConverterTypeChannelIdINSTANCE.Write(writer, variant_value.ChannelId)
		FfiConverterTypePublicKeyINSTANCE.Write(writer, variant_value.CounterpartyNodeId)
		FfiConverterUint64INSTANCE.Write(writer, variant_value.AmountSatoshis)
		FfiConverterUint32INSTANCE.Write(writer, variant_value.ClaimableHeight)
		FfiConverterTypePaymentHashINSTANCE.Write(writer, variant_value.PaymentHash)
	case LightningBalanceMaybePreimageClaimableHtlc:
		writeInt32(writer, 5)
		FfiConverterTypeChannelIdINSTANCE.Write(writer, variant_value.ChannelId)
		FfiConverterTypePublicKeyINSTANCE.Write(writer, variant_value.CounterpartyNodeId)
		FfiConverterUint64INSTANCE.Write(writer, variant_value.AmountSatoshis)
		FfiConverterUint32INSTANCE.Write(writer, variant_value.ExpiryHeight)
		FfiConverterTypePaymentHashINSTANCE.Write(writer, variant_value.PaymentHash)
	case LightningBalanceCounterpartyRevokedOutputClaimable:
		writeInt32(writer, 6)
		FfiConverterTypeChannelIdINSTANCE.Write(writer, variant_value.ChannelId)
		FfiConverterTypePublicKeyINSTANCE.Write(writer, variant_value.CounterpartyNodeId)
		FfiConverterUint64INSTANCE.Write(writer, variant_value.AmountSatoshis)
	default:
		_ = variant_value
		panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeLightningBalance.Write", value))
	}
}

type FfiDestroyerTypeLightningBalance struct{}

func (_ FfiDestroyerTypeLightningBalance) Destroy(value LightningBalance) {
	value.Destroy()
}

type LogLevel uint

const (
	LogLevelGossip LogLevel = 1
	LogLevelTrace  LogLevel = 2
	LogLevelDebug  LogLevel = 3
	LogLevelInfo   LogLevel = 4
	LogLevelWarn   LogLevel = 5
	LogLevelError  LogLevel = 6
)

type FfiConverterTypeLogLevel struct{}

var FfiConverterTypeLogLevelINSTANCE = FfiConverterTypeLogLevel{}

func (c FfiConverterTypeLogLevel) Lift(rb RustBufferI) LogLevel {
	return LiftFromRustBuffer[LogLevel](c, rb)
}

func (c FfiConverterTypeLogLevel) Lower(value LogLevel) RustBuffer {
	return LowerIntoRustBuffer[LogLevel](c, value)
}
func (FfiConverterTypeLogLevel) Read(reader io.Reader) LogLevel {
	id := readInt32(reader)
	return LogLevel(id)
}

func (FfiConverterTypeLogLevel) Write(writer io.Writer, value LogLevel) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerTypeLogLevel struct{}

func (_ FfiDestroyerTypeLogLevel) Destroy(value LogLevel) {
}

type NodeError struct {
	err error
}

func (err NodeError) Error() string {
	return fmt.Sprintf("NodeError: %s", err.err.Error())
}

func (err NodeError) Unwrap() error {
	return err.err
}

// Err* are used for checking error type with `errors.Is`
var ErrNodeErrorAlreadyRunning = fmt.Errorf("NodeErrorAlreadyRunning")
var ErrNodeErrorNotRunning = fmt.Errorf("NodeErrorNotRunning")
var ErrNodeErrorOnchainTxCreationFailed = fmt.Errorf("NodeErrorOnchainTxCreationFailed")
var ErrNodeErrorConnectionFailed = fmt.Errorf("NodeErrorConnectionFailed")
var ErrNodeErrorInvoiceCreationFailed = fmt.Errorf("NodeErrorInvoiceCreationFailed")
var ErrNodeErrorPaymentSendingFailed = fmt.Errorf("NodeErrorPaymentSendingFailed")
var ErrNodeErrorProbeSendingFailed = fmt.Errorf("NodeErrorProbeSendingFailed")
var ErrNodeErrorChannelCreationFailed = fmt.Errorf("NodeErrorChannelCreationFailed")
var ErrNodeErrorChannelClosingFailed = fmt.Errorf("NodeErrorChannelClosingFailed")
var ErrNodeErrorChannelConfigUpdateFailed = fmt.Errorf("NodeErrorChannelConfigUpdateFailed")
var ErrNodeErrorPersistenceFailed = fmt.Errorf("NodeErrorPersistenceFailed")
var ErrNodeErrorFeerateEstimationUpdateFailed = fmt.Errorf("NodeErrorFeerateEstimationUpdateFailed")
var ErrNodeErrorWalletOperationFailed = fmt.Errorf("NodeErrorWalletOperationFailed")
var ErrNodeErrorOnchainTxSigningFailed = fmt.Errorf("NodeErrorOnchainTxSigningFailed")
var ErrNodeErrorMessageSigningFailed = fmt.Errorf("NodeErrorMessageSigningFailed")
var ErrNodeErrorTxSyncFailed = fmt.Errorf("NodeErrorTxSyncFailed")
var ErrNodeErrorGossipUpdateFailed = fmt.Errorf("NodeErrorGossipUpdateFailed")
var ErrNodeErrorLiquidityRequestFailed = fmt.Errorf("NodeErrorLiquidityRequestFailed")
var ErrNodeErrorInvalidAddress = fmt.Errorf("NodeErrorInvalidAddress")
var ErrNodeErrorInvalidSocketAddress = fmt.Errorf("NodeErrorInvalidSocketAddress")
var ErrNodeErrorInvalidPublicKey = fmt.Errorf("NodeErrorInvalidPublicKey")
var ErrNodeErrorInvalidSecretKey = fmt.Errorf("NodeErrorInvalidSecretKey")
var ErrNodeErrorInvalidPaymentId = fmt.Errorf("NodeErrorInvalidPaymentId")
var ErrNodeErrorInvalidPaymentHash = fmt.Errorf("NodeErrorInvalidPaymentHash")
var ErrNodeErrorInvalidPaymentPreimage = fmt.Errorf("NodeErrorInvalidPaymentPreimage")
var ErrNodeErrorInvalidPaymentSecret = fmt.Errorf("NodeErrorInvalidPaymentSecret")
var ErrNodeErrorInvalidAmount = fmt.Errorf("NodeErrorInvalidAmount")
var ErrNodeErrorInvalidInvoice = fmt.Errorf("NodeErrorInvalidInvoice")
var ErrNodeErrorInvalidChannelId = fmt.Errorf("NodeErrorInvalidChannelId")
var ErrNodeErrorInvalidNetwork = fmt.Errorf("NodeErrorInvalidNetwork")
var ErrNodeErrorInvalidCustomTlv = fmt.Errorf("NodeErrorInvalidCustomTlv")
var ErrNodeErrorDuplicatePayment = fmt.Errorf("NodeErrorDuplicatePayment")
var ErrNodeErrorInsufficientFunds = fmt.Errorf("NodeErrorInsufficientFunds")
var ErrNodeErrorLiquiditySourceUnavailable = fmt.Errorf("NodeErrorLiquiditySourceUnavailable")
var ErrNodeErrorLiquidityFeeTooHigh = fmt.Errorf("NodeErrorLiquidityFeeTooHigh")

// Variant structs
type NodeErrorAlreadyRunning struct {
	message string
}

func NewNodeErrorAlreadyRunning() *NodeError {
	return &NodeError{
		err: &NodeErrorAlreadyRunning{},
	}
}

func (err NodeErrorAlreadyRunning) Error() string {
	return fmt.Sprintf("AlreadyRunning: %s", err.message)
}

func (self NodeErrorAlreadyRunning) Is(target error) bool {
	return target == ErrNodeErrorAlreadyRunning
}

type NodeErrorNotRunning struct {
	message string
}

func NewNodeErrorNotRunning() *NodeError {
	return &NodeError{
		err: &NodeErrorNotRunning{},
	}
}

func (err NodeErrorNotRunning) Error() string {
	return fmt.Sprintf("NotRunning: %s", err.message)
}

func (self NodeErrorNotRunning) Is(target error) bool {
	return target == ErrNodeErrorNotRunning
}

type NodeErrorOnchainTxCreationFailed struct {
	message string
}

func NewNodeErrorOnchainTxCreationFailed() *NodeError {
	return &NodeError{
		err: &NodeErrorOnchainTxCreationFailed{},
	}
}

func (err NodeErrorOnchainTxCreationFailed) Error() string {
	return fmt.Sprintf("OnchainTxCreationFailed: %s", err.message)
}

func (self NodeErrorOnchainTxCreationFailed) Is(target error) bool {
	return target == ErrNodeErrorOnchainTxCreationFailed
}

type NodeErrorConnectionFailed struct {
	message string
}

func NewNodeErrorConnectionFailed() *NodeError {
	return &NodeError{
		err: &NodeErrorConnectionFailed{},
	}
}

func (err NodeErrorConnectionFailed) Error() string {
	return fmt.Sprintf("ConnectionFailed: %s", err.message)
}

func (self NodeErrorConnectionFailed) Is(target error) bool {
	return target == ErrNodeErrorConnectionFailed
}

type NodeErrorInvoiceCreationFailed struct {
	message string
}

func NewNodeErrorInvoiceCreationFailed() *NodeError {
	return &NodeError{
		err: &NodeErrorInvoiceCreationFailed{},
	}
}

func (err NodeErrorInvoiceCreationFailed) Error() string {
	return fmt.Sprintf("InvoiceCreationFailed: %s", err.message)
}

func (self NodeErrorInvoiceCreationFailed) Is(target error) bool {
	return target == ErrNodeErrorInvoiceCreationFailed
}

type NodeErrorPaymentSendingFailed struct {
	message string
}

func NewNodeErrorPaymentSendingFailed() *NodeError {
	return &NodeError{
		err: &NodeErrorPaymentSendingFailed{},
	}
}

func (err NodeErrorPaymentSendingFailed) Error() string {
	return fmt.Sprintf("PaymentSendingFailed: %s", err.message)
}

func (self NodeErrorPaymentSendingFailed) Is(target error) bool {
	return target == ErrNodeErrorPaymentSendingFailed
}

type NodeErrorProbeSendingFailed struct {
	message string
}

func NewNodeErrorProbeSendingFailed() *NodeError {
	return &NodeError{
		err: &NodeErrorProbeSendingFailed{},
	}
}

func (err NodeErrorProbeSendingFailed) Error() string {
	return fmt.Sprintf("ProbeSendingFailed: %s", err.message)
}

func (self NodeErrorProbeSendingFailed) Is(target error) bool {
	return target == ErrNodeErrorProbeSendingFailed
}

type NodeErrorChannelCreationFailed struct {
	message string
}

func NewNodeErrorChannelCreationFailed() *NodeError {
	return &NodeError{
		err: &NodeErrorChannelCreationFailed{},
	}
}

func (err NodeErrorChannelCreationFailed) Error() string {
	return fmt.Sprintf("ChannelCreationFailed: %s", err.message)
}

func (self NodeErrorChannelCreationFailed) Is(target error) bool {
	return target == ErrNodeErrorChannelCreationFailed
}

type NodeErrorChannelClosingFailed struct {
	message string
}

func NewNodeErrorChannelClosingFailed() *NodeError {
	return &NodeError{
		err: &NodeErrorChannelClosingFailed{},
	}
}

func (err NodeErrorChannelClosingFailed) Error() string {
	return fmt.Sprintf("ChannelClosingFailed: %s", err.message)
}

func (self NodeErrorChannelClosingFailed) Is(target error) bool {
	return target == ErrNodeErrorChannelClosingFailed
}

type NodeErrorChannelConfigUpdateFailed struct {
	message string
}

func NewNodeErrorChannelConfigUpdateFailed() *NodeError {
	return &NodeError{
		err: &NodeErrorChannelConfigUpdateFailed{},
	}
}

func (err NodeErrorChannelConfigUpdateFailed) Error() string {
	return fmt.Sprintf("ChannelConfigUpdateFailed: %s", err.message)
}

func (self NodeErrorChannelConfigUpdateFailed) Is(target error) bool {
	return target == ErrNodeErrorChannelConfigUpdateFailed
}

type NodeErrorPersistenceFailed struct {
	message string
}

func NewNodeErrorPersistenceFailed() *NodeError {
	return &NodeError{
		err: &NodeErrorPersistenceFailed{},
	}
}

func (err NodeErrorPersistenceFailed) Error() string {
	return fmt.Sprintf("PersistenceFailed: %s", err.message)
}

func (self NodeErrorPersistenceFailed) Is(target error) bool {
	return target == ErrNodeErrorPersistenceFailed
}

type NodeErrorFeerateEstimationUpdateFailed struct {
	message string
}

func NewNodeErrorFeerateEstimationUpdateFailed() *NodeError {
	return &NodeError{
		err: &NodeErrorFeerateEstimationUpdateFailed{},
	}
}

func (err NodeErrorFeerateEstimationUpdateFailed) Error() string {
	return fmt.Sprintf("FeerateEstimationUpdateFailed: %s", err.message)
}

func (self NodeErrorFeerateEstimationUpdateFailed) Is(target error) bool {
	return target == ErrNodeErrorFeerateEstimationUpdateFailed
}

type NodeErrorWalletOperationFailed struct {
	message string
}

func NewNodeErrorWalletOperationFailed() *NodeError {
	return &NodeError{
		err: &NodeErrorWalletOperationFailed{},
	}
}

func (err NodeErrorWalletOperationFailed) Error() string {
	return fmt.Sprintf("WalletOperationFailed: %s", err.message)
}

func (self NodeErrorWalletOperationFailed) Is(target error) bool {
	return target == ErrNodeErrorWalletOperationFailed
}

type NodeErrorOnchainTxSigningFailed struct {
	message string
}

func NewNodeErrorOnchainTxSigningFailed() *NodeError {
	return &NodeError{
		err: &NodeErrorOnchainTxSigningFailed{},
	}
}

func (err NodeErrorOnchainTxSigningFailed) Error() string {
	return fmt.Sprintf("OnchainTxSigningFailed: %s", err.message)
}

func (self NodeErrorOnchainTxSigningFailed) Is(target error) bool {
	return target == ErrNodeErrorOnchainTxSigningFailed
}

type NodeErrorMessageSigningFailed struct {
	message string
}

func NewNodeErrorMessageSigningFailed() *NodeError {
	return &NodeError{
		err: &NodeErrorMessageSigningFailed{},
	}
}

func (err NodeErrorMessageSigningFailed) Error() string {
	return fmt.Sprintf("MessageSigningFailed: %s", err.message)
}

func (self NodeErrorMessageSigningFailed) Is(target error) bool {
	return target == ErrNodeErrorMessageSigningFailed
}

type NodeErrorTxSyncFailed struct {
	message string
}

func NewNodeErrorTxSyncFailed() *NodeError {
	return &NodeError{
		err: &NodeErrorTxSyncFailed{},
	}
}

func (err NodeErrorTxSyncFailed) Error() string {
	return fmt.Sprintf("TxSyncFailed: %s", err.message)
}

func (self NodeErrorTxSyncFailed) Is(target error) bool {
	return target == ErrNodeErrorTxSyncFailed
}

type NodeErrorGossipUpdateFailed struct {
	message string
}

func NewNodeErrorGossipUpdateFailed() *NodeError {
	return &NodeError{
		err: &NodeErrorGossipUpdateFailed{},
	}
}

func (err NodeErrorGossipUpdateFailed) Error() string {
	return fmt.Sprintf("GossipUpdateFailed: %s", err.message)
}

func (self NodeErrorGossipUpdateFailed) Is(target error) bool {
	return target == ErrNodeErrorGossipUpdateFailed
}

type NodeErrorLiquidityRequestFailed struct {
	message string
}

func NewNodeErrorLiquidityRequestFailed() *NodeError {
	return &NodeError{
		err: &NodeErrorLiquidityRequestFailed{},
	}
}

func (err NodeErrorLiquidityRequestFailed) Error() string {
	return fmt.Sprintf("LiquidityRequestFailed: %s", err.message)
}

func (self NodeErrorLiquidityRequestFailed) Is(target error) bool {
	return target == ErrNodeErrorLiquidityRequestFailed
}

type NodeErrorInvalidAddress struct {
	message string
}

func NewNodeErrorInvalidAddress() *NodeError {
	return &NodeError{
		err: &NodeErrorInvalidAddress{},
	}
}

func (err NodeErrorInvalidAddress) Error() string {
	return fmt.Sprintf("InvalidAddress: %s", err.message)
}

func (self NodeErrorInvalidAddress) Is(target error) bool {
	return target == ErrNodeErrorInvalidAddress
}

type NodeErrorInvalidSocketAddress struct {
	message string
}

func NewNodeErrorInvalidSocketAddress() *NodeError {
	return &NodeError{
		err: &NodeErrorInvalidSocketAddress{},
	}
}

func (err NodeErrorInvalidSocketAddress) Error() string {
	return fmt.Sprintf("InvalidSocketAddress: %s", err.message)
}

func (self NodeErrorInvalidSocketAddress) Is(target error) bool {
	return target == ErrNodeErrorInvalidSocketAddress
}

type NodeErrorInvalidPublicKey struct {
	message string
}

func NewNodeErrorInvalidPublicKey() *NodeError {
	return &NodeError{
		err: &NodeErrorInvalidPublicKey{},
	}
}

func (err NodeErrorInvalidPublicKey) Error() string {
	return fmt.Sprintf("InvalidPublicKey: %s", err.message)
}

func (self NodeErrorInvalidPublicKey) Is(target error) bool {
	return target == ErrNodeErrorInvalidPublicKey
}

type NodeErrorInvalidSecretKey struct {
	message string
}

func NewNodeErrorInvalidSecretKey() *NodeError {
	return &NodeError{
		err: &NodeErrorInvalidSecretKey{},
	}
}

func (err NodeErrorInvalidSecretKey) Error() string {
	return fmt.Sprintf("InvalidSecretKey: %s", err.message)
}

func (self NodeErrorInvalidSecretKey) Is(target error) bool {
	return target == ErrNodeErrorInvalidSecretKey
}

type NodeErrorInvalidPaymentId struct {
	message string
}

func NewNodeErrorInvalidPaymentId() *NodeError {
	return &NodeError{
		err: &NodeErrorInvalidPaymentId{},
	}
}

func (err NodeErrorInvalidPaymentId) Error() string {
	return fmt.Sprintf("InvalidPaymentId: %s", err.message)
}

func (self NodeErrorInvalidPaymentId) Is(target error) bool {
	return target == ErrNodeErrorInvalidPaymentId
}

type NodeErrorInvalidPaymentHash struct {
	message string
}

func NewNodeErrorInvalidPaymentHash() *NodeError {
	return &NodeError{
		err: &NodeErrorInvalidPaymentHash{},
	}
}

func (err NodeErrorInvalidPaymentHash) Error() string {
	return fmt.Sprintf("InvalidPaymentHash: %s", err.message)
}

func (self NodeErrorInvalidPaymentHash) Is(target error) bool {
	return target == ErrNodeErrorInvalidPaymentHash
}

type NodeErrorInvalidPaymentPreimage struct {
	message string
}

func NewNodeErrorInvalidPaymentPreimage() *NodeError {
	return &NodeError{
		err: &NodeErrorInvalidPaymentPreimage{},
	}
}

func (err NodeErrorInvalidPaymentPreimage) Error() string {
	return fmt.Sprintf("InvalidPaymentPreimage: %s", err.message)
}

func (self NodeErrorInvalidPaymentPreimage) Is(target error) bool {
	return target == ErrNodeErrorInvalidPaymentPreimage
}

type NodeErrorInvalidPaymentSecret struct {
	message string
}

func NewNodeErrorInvalidPaymentSecret() *NodeError {
	return &NodeError{
		err: &NodeErrorInvalidPaymentSecret{},
	}
}

func (err NodeErrorInvalidPaymentSecret) Error() string {
	return fmt.Sprintf("InvalidPaymentSecret: %s", err.message)
}

func (self NodeErrorInvalidPaymentSecret) Is(target error) bool {
	return target == ErrNodeErrorInvalidPaymentSecret
}

type NodeErrorInvalidAmount struct {
	message string
}

func NewNodeErrorInvalidAmount() *NodeError {
	return &NodeError{
		err: &NodeErrorInvalidAmount{},
	}
}

func (err NodeErrorInvalidAmount) Error() string {
	return fmt.Sprintf("InvalidAmount: %s", err.message)
}

func (self NodeErrorInvalidAmount) Is(target error) bool {
	return target == ErrNodeErrorInvalidAmount
}

type NodeErrorInvalidInvoice struct {
	message string
}

func NewNodeErrorInvalidInvoice() *NodeError {
	return &NodeError{
		err: &NodeErrorInvalidInvoice{},
	}
}

func (err NodeErrorInvalidInvoice) Error() string {
	return fmt.Sprintf("InvalidInvoice: %s", err.message)
}

func (self NodeErrorInvalidInvoice) Is(target error) bool {
	return target == ErrNodeErrorInvalidInvoice
}

type NodeErrorInvalidChannelId struct {
	message string
}

func NewNodeErrorInvalidChannelId() *NodeError {
	return &NodeError{
		err: &NodeErrorInvalidChannelId{},
	}
}

func (err NodeErrorInvalidChannelId) Error() string {
	return fmt.Sprintf("InvalidChannelId: %s", err.message)
}

func (self NodeErrorInvalidChannelId) Is(target error) bool {
	return target == ErrNodeErrorInvalidChannelId
}

type NodeErrorInvalidNetwork struct {
	message string
}

func NewNodeErrorInvalidNetwork() *NodeError {
	return &NodeError{
		err: &NodeErrorInvalidNetwork{},
	}
}

func (err NodeErrorInvalidNetwork) Error() string {
	return fmt.Sprintf("InvalidNetwork: %s", err.message)
}

func (self NodeErrorInvalidNetwork) Is(target error) bool {
	return target == ErrNodeErrorInvalidNetwork
}

type NodeErrorInvalidCustomTlv struct {
	message string
}

func NewNodeErrorInvalidCustomTlv() *NodeError {
	return &NodeError{
		err: &NodeErrorInvalidCustomTlv{},
	}
}

func (err NodeErrorInvalidCustomTlv) Error() string {
	return fmt.Sprintf("InvalidCustomTlv: %s", err.message)
}

func (self NodeErrorInvalidCustomTlv) Is(target error) bool {
	return target == ErrNodeErrorInvalidCustomTlv
}

type NodeErrorDuplicatePayment struct {
	message string
}

func NewNodeErrorDuplicatePayment() *NodeError {
	return &NodeError{
		err: &NodeErrorDuplicatePayment{},
	}
}

func (err NodeErrorDuplicatePayment) Error() string {
	return fmt.Sprintf("DuplicatePayment: %s", err.message)
}

func (self NodeErrorDuplicatePayment) Is(target error) bool {
	return target == ErrNodeErrorDuplicatePayment
}

type NodeErrorInsufficientFunds struct {
	message string
}

func NewNodeErrorInsufficientFunds() *NodeError {
	return &NodeError{
		err: &NodeErrorInsufficientFunds{},
	}
}

func (err NodeErrorInsufficientFunds) Error() string {
	return fmt.Sprintf("InsufficientFunds: %s", err.message)
}

func (self NodeErrorInsufficientFunds) Is(target error) bool {
	return target == ErrNodeErrorInsufficientFunds
}

type NodeErrorLiquiditySourceUnavailable struct {
	message string
}

func NewNodeErrorLiquiditySourceUnavailable() *NodeError {
	return &NodeError{
		err: &NodeErrorLiquiditySourceUnavailable{},
	}
}

func (err NodeErrorLiquiditySourceUnavailable) Error() string {
	return fmt.Sprintf("LiquiditySourceUnavailable: %s", err.message)
}

func (self NodeErrorLiquiditySourceUnavailable) Is(target error) bool {
	return target == ErrNodeErrorLiquiditySourceUnavailable
}

type NodeErrorLiquidityFeeTooHigh struct {
	message string
}

func NewNodeErrorLiquidityFeeTooHigh() *NodeError {
	return &NodeError{
		err: &NodeErrorLiquidityFeeTooHigh{},
	}
}

func (err NodeErrorLiquidityFeeTooHigh) Error() string {
	return fmt.Sprintf("LiquidityFeeTooHigh: %s", err.message)
}

func (self NodeErrorLiquidityFeeTooHigh) Is(target error) bool {
	return target == ErrNodeErrorLiquidityFeeTooHigh
}

type FfiConverterTypeNodeError struct{}

var FfiConverterTypeNodeErrorINSTANCE = FfiConverterTypeNodeError{}

func (c FfiConverterTypeNodeError) Lift(eb RustBufferI) error {
	return LiftFromRustBuffer[error](c, eb)
}

func (c FfiConverterTypeNodeError) Lower(value *NodeError) RustBuffer {
	return LowerIntoRustBuffer[*NodeError](c, value)
}

func (c FfiConverterTypeNodeError) Read(reader io.Reader) error {
	errorID := readUint32(reader)

	message := FfiConverterStringINSTANCE.Read(reader)
	switch errorID {
	case 1:
		return &NodeError{&NodeErrorAlreadyRunning{message}}
	case 2:
		return &NodeError{&NodeErrorNotRunning{message}}
	case 3:
		return &NodeError{&NodeErrorOnchainTxCreationFailed{message}}
	case 4:
		return &NodeError{&NodeErrorConnectionFailed{message}}
	case 5:
		return &NodeError{&NodeErrorInvoiceCreationFailed{message}}
	case 6:
		return &NodeError{&NodeErrorPaymentSendingFailed{message}}
	case 7:
		return &NodeError{&NodeErrorProbeSendingFailed{message}}
	case 8:
		return &NodeError{&NodeErrorChannelCreationFailed{message}}
	case 9:
		return &NodeError{&NodeErrorChannelClosingFailed{message}}
	case 10:
		return &NodeError{&NodeErrorChannelConfigUpdateFailed{message}}
	case 11:
		return &NodeError{&NodeErrorPersistenceFailed{message}}
	case 12:
		return &NodeError{&NodeErrorFeerateEstimationUpdateFailed{message}}
	case 13:
		return &NodeError{&NodeErrorWalletOperationFailed{message}}
	case 14:
		return &NodeError{&NodeErrorOnchainTxSigningFailed{message}}
	case 15:
		return &NodeError{&NodeErrorMessageSigningFailed{message}}
	case 16:
		return &NodeError{&NodeErrorTxSyncFailed{message}}
	case 17:
		return &NodeError{&NodeErrorGossipUpdateFailed{message}}
	case 18:
		return &NodeError{&NodeErrorLiquidityRequestFailed{message}}
	case 19:
		return &NodeError{&NodeErrorInvalidAddress{message}}
	case 20:
		return &NodeError{&NodeErrorInvalidSocketAddress{message}}
	case 21:
		return &NodeError{&NodeErrorInvalidPublicKey{message}}
	case 22:
		return &NodeError{&NodeErrorInvalidSecretKey{message}}
	case 23:
		return &NodeError{&NodeErrorInvalidPaymentId{message}}
	case 24:
		return &NodeError{&NodeErrorInvalidPaymentHash{message}}
	case 25:
		return &NodeError{&NodeErrorInvalidPaymentPreimage{message}}
	case 26:
		return &NodeError{&NodeErrorInvalidPaymentSecret{message}}
	case 27:
		return &NodeError{&NodeErrorInvalidAmount{message}}
	case 28:
		return &NodeError{&NodeErrorInvalidInvoice{message}}
	case 29:
		return &NodeError{&NodeErrorInvalidChannelId{message}}
	case 30:
		return &NodeError{&NodeErrorInvalidNetwork{message}}
	case 31:
		return &NodeError{&NodeErrorInvalidCustomTlv{message}}
	case 32:
		return &NodeError{&NodeErrorDuplicatePayment{message}}
	case 33:
		return &NodeError{&NodeErrorInsufficientFunds{message}}
	case 34:
		return &NodeError{&NodeErrorLiquiditySourceUnavailable{message}}
	case 35:
		return &NodeError{&NodeErrorLiquidityFeeTooHigh{message}}
	default:
		panic(fmt.Sprintf("Unknown error code %d in FfiConverterTypeNodeError.Read()", errorID))
	}

}

func (c FfiConverterTypeNodeError) Write(writer io.Writer, value *NodeError) {
	switch variantValue := value.err.(type) {
	case *NodeErrorAlreadyRunning:
		writeInt32(writer, 1)
	case *NodeErrorNotRunning:
		writeInt32(writer, 2)
	case *NodeErrorOnchainTxCreationFailed:
		writeInt32(writer, 3)
	case *NodeErrorConnectionFailed:
		writeInt32(writer, 4)
	case *NodeErrorInvoiceCreationFailed:
		writeInt32(writer, 5)
	case *NodeErrorPaymentSendingFailed:
		writeInt32(writer, 6)
	case *NodeErrorProbeSendingFailed:
		writeInt32(writer, 7)
	case *NodeErrorChannelCreationFailed:
		writeInt32(writer, 8)
	case *NodeErrorChannelClosingFailed:
		writeInt32(writer, 9)
	case *NodeErrorChannelConfigUpdateFailed:
		writeInt32(writer, 10)
	case *NodeErrorPersistenceFailed:
		writeInt32(writer, 11)
	case *NodeErrorFeerateEstimationUpdateFailed:
		writeInt32(writer, 12)
	case *NodeErrorWalletOperationFailed:
		writeInt32(writer, 13)
	case *NodeErrorOnchainTxSigningFailed:
		writeInt32(writer, 14)
	case *NodeErrorMessageSigningFailed:
		writeInt32(writer, 15)
	case *NodeErrorTxSyncFailed:
		writeInt32(writer, 16)
	case *NodeErrorGossipUpdateFailed:
		writeInt32(writer, 17)
	case *NodeErrorLiquidityRequestFailed:
		writeInt32(writer, 18)
	case *NodeErrorInvalidAddress:
		writeInt32(writer, 19)
	case *NodeErrorInvalidSocketAddress:
		writeInt32(writer, 20)
	case *NodeErrorInvalidPublicKey:
		writeInt32(writer, 21)
	case *NodeErrorInvalidSecretKey:
		writeInt32(writer, 22)
	case *NodeErrorInvalidPaymentId:
		writeInt32(writer, 23)
	case *NodeErrorInvalidPaymentHash:
		writeInt32(writer, 24)
	case *NodeErrorInvalidPaymentPreimage:
		writeInt32(writer, 25)
	case *NodeErrorInvalidPaymentSecret:
		writeInt32(writer, 26)
	case *NodeErrorInvalidAmount:
		writeInt32(writer, 27)
	case *NodeErrorInvalidInvoice:
		writeInt32(writer, 28)
	case *NodeErrorInvalidChannelId:
		writeInt32(writer, 29)
	case *NodeErrorInvalidNetwork:
		writeInt32(writer, 30)
	case *NodeErrorInvalidCustomTlv:
		writeInt32(writer, 31)
	case *NodeErrorDuplicatePayment:
		writeInt32(writer, 32)
	case *NodeErrorInsufficientFunds:
		writeInt32(writer, 33)
	case *NodeErrorLiquiditySourceUnavailable:
		writeInt32(writer, 34)
	case *NodeErrorLiquidityFeeTooHigh:
		writeInt32(writer, 35)
	default:
		_ = variantValue
		panic(fmt.Sprintf("invalid error value `%v` in FfiConverterTypeNodeError.Write", value))
	}
}

type PaymentDirection uint

const (
	PaymentDirectionInbound  PaymentDirection = 1
	PaymentDirectionOutbound PaymentDirection = 2
)

type FfiConverterTypePaymentDirection struct{}

var FfiConverterTypePaymentDirectionINSTANCE = FfiConverterTypePaymentDirection{}

func (c FfiConverterTypePaymentDirection) Lift(rb RustBufferI) PaymentDirection {
	return LiftFromRustBuffer[PaymentDirection](c, rb)
}

func (c FfiConverterTypePaymentDirection) Lower(value PaymentDirection) RustBuffer {
	return LowerIntoRustBuffer[PaymentDirection](c, value)
}
func (FfiConverterTypePaymentDirection) Read(reader io.Reader) PaymentDirection {
	id := readInt32(reader)
	return PaymentDirection(id)
}

func (FfiConverterTypePaymentDirection) Write(writer io.Writer, value PaymentDirection) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerTypePaymentDirection struct{}

func (_ FfiDestroyerTypePaymentDirection) Destroy(value PaymentDirection) {
}

type PaymentFailureReason uint

const (
	PaymentFailureReasonRecipientRejected PaymentFailureReason = 1
	PaymentFailureReasonUserAbandoned     PaymentFailureReason = 2
	PaymentFailureReasonRetriesExhausted  PaymentFailureReason = 3
	PaymentFailureReasonPaymentExpired    PaymentFailureReason = 4
	PaymentFailureReasonRouteNotFound     PaymentFailureReason = 5
	PaymentFailureReasonUnexpectedError   PaymentFailureReason = 6
)

type FfiConverterTypePaymentFailureReason struct{}

var FfiConverterTypePaymentFailureReasonINSTANCE = FfiConverterTypePaymentFailureReason{}

func (c FfiConverterTypePaymentFailureReason) Lift(rb RustBufferI) PaymentFailureReason {
	return LiftFromRustBuffer[PaymentFailureReason](c, rb)
}

func (c FfiConverterTypePaymentFailureReason) Lower(value PaymentFailureReason) RustBuffer {
	return LowerIntoRustBuffer[PaymentFailureReason](c, value)
}
func (FfiConverterTypePaymentFailureReason) Read(reader io.Reader) PaymentFailureReason {
	id := readInt32(reader)
	return PaymentFailureReason(id)
}

func (FfiConverterTypePaymentFailureReason) Write(writer io.Writer, value PaymentFailureReason) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerTypePaymentFailureReason struct{}

func (_ FfiDestroyerTypePaymentFailureReason) Destroy(value PaymentFailureReason) {
}

type PaymentKind interface {
	Destroy()
}
type PaymentKindOnchain struct {
}

func (e PaymentKindOnchain) Destroy() {
}

type PaymentKindBolt11 struct {
	Hash          PaymentHash
	Preimage      *PaymentPreimage
	Secret        *PaymentSecret
	Bolt11Invoice *string
}

func (e PaymentKindBolt11) Destroy() {
	FfiDestroyerTypePaymentHash{}.Destroy(e.Hash)
	FfiDestroyerOptionalTypePaymentPreimage{}.Destroy(e.Preimage)
	FfiDestroyerOptionalTypePaymentSecret{}.Destroy(e.Secret)
	FfiDestroyerOptionalString{}.Destroy(e.Bolt11Invoice)
}

type PaymentKindBolt11Jit struct {
	Hash         PaymentHash
	Preimage     *PaymentPreimage
	Secret       *PaymentSecret
	LspFeeLimits LspFeeLimits
}

func (e PaymentKindBolt11Jit) Destroy() {
	FfiDestroyerTypePaymentHash{}.Destroy(e.Hash)
	FfiDestroyerOptionalTypePaymentPreimage{}.Destroy(e.Preimage)
	FfiDestroyerOptionalTypePaymentSecret{}.Destroy(e.Secret)
	FfiDestroyerTypeLspFeeLimits{}.Destroy(e.LspFeeLimits)
}

type PaymentKindSpontaneous struct {
	Hash     PaymentHash
	Preimage *PaymentPreimage
}

func (e PaymentKindSpontaneous) Destroy() {
	FfiDestroyerTypePaymentHash{}.Destroy(e.Hash)
	FfiDestroyerOptionalTypePaymentPreimage{}.Destroy(e.Preimage)
}

type FfiConverterTypePaymentKind struct{}

var FfiConverterTypePaymentKindINSTANCE = FfiConverterTypePaymentKind{}

func (c FfiConverterTypePaymentKind) Lift(rb RustBufferI) PaymentKind {
	return LiftFromRustBuffer[PaymentKind](c, rb)
}

func (c FfiConverterTypePaymentKind) Lower(value PaymentKind) RustBuffer {
	return LowerIntoRustBuffer[PaymentKind](c, value)
}
func (FfiConverterTypePaymentKind) Read(reader io.Reader) PaymentKind {
	id := readInt32(reader)
	switch id {
	case 1:
		return PaymentKindOnchain{}
	case 2:
		return PaymentKindBolt11{
			FfiConverterTypePaymentHashINSTANCE.Read(reader),
			FfiConverterOptionalTypePaymentPreimageINSTANCE.Read(reader),
			FfiConverterOptionalTypePaymentSecretINSTANCE.Read(reader),
			FfiConverterOptionalStringINSTANCE.Read(reader),
		}
	case 3:
		return PaymentKindBolt11Jit{
			FfiConverterTypePaymentHashINSTANCE.Read(reader),
			FfiConverterOptionalTypePaymentPreimageINSTANCE.Read(reader),
			FfiConverterOptionalTypePaymentSecretINSTANCE.Read(reader),
			FfiConverterTypeLSPFeeLimitsINSTANCE.Read(reader),
		}
	case 4:
		return PaymentKindSpontaneous{
			FfiConverterTypePaymentHashINSTANCE.Read(reader),
			FfiConverterOptionalTypePaymentPreimageINSTANCE.Read(reader),
		}
	default:
		panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypePaymentKind.Read()", id))
	}
}

func (FfiConverterTypePaymentKind) Write(writer io.Writer, value PaymentKind) {
	switch variant_value := value.(type) {
	case PaymentKindOnchain:
		writeInt32(writer, 1)
	case PaymentKindBolt11:
		writeInt32(writer, 2)
		FfiConverterTypePaymentHashINSTANCE.Write(writer, variant_value.Hash)
		FfiConverterOptionalTypePaymentPreimageINSTANCE.Write(writer, variant_value.Preimage)
		FfiConverterOptionalTypePaymentSecretINSTANCE.Write(writer, variant_value.Secret)
		FfiConverterOptionalStringINSTANCE.Write(writer, variant_value.Bolt11Invoice)
	case PaymentKindBolt11Jit:
		writeInt32(writer, 3)
		FfiConverterTypePaymentHashINSTANCE.Write(writer, variant_value.Hash)
		FfiConverterOptionalTypePaymentPreimageINSTANCE.Write(writer, variant_value.Preimage)
		FfiConverterOptionalTypePaymentSecretINSTANCE.Write(writer, variant_value.Secret)
		FfiConverterTypeLSPFeeLimitsINSTANCE.Write(writer, variant_value.LspFeeLimits)
	case PaymentKindSpontaneous:
		writeInt32(writer, 4)
		FfiConverterTypePaymentHashINSTANCE.Write(writer, variant_value.Hash)
		FfiConverterOptionalTypePaymentPreimageINSTANCE.Write(writer, variant_value.Preimage)
	default:
		_ = variant_value
		panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypePaymentKind.Write", value))
	}
}

type FfiDestroyerTypePaymentKind struct{}

func (_ FfiDestroyerTypePaymentKind) Destroy(value PaymentKind) {
	value.Destroy()
}

type PaymentStatus uint

const (
	PaymentStatusPending   PaymentStatus = 1
	PaymentStatusSucceeded PaymentStatus = 2
	PaymentStatusFailed    PaymentStatus = 3
)

type FfiConverterTypePaymentStatus struct{}

var FfiConverterTypePaymentStatusINSTANCE = FfiConverterTypePaymentStatus{}

func (c FfiConverterTypePaymentStatus) Lift(rb RustBufferI) PaymentStatus {
	return LiftFromRustBuffer[PaymentStatus](c, rb)
}

func (c FfiConverterTypePaymentStatus) Lower(value PaymentStatus) RustBuffer {
	return LowerIntoRustBuffer[PaymentStatus](c, value)
}
func (FfiConverterTypePaymentStatus) Read(reader io.Reader) PaymentStatus {
	id := readInt32(reader)
	return PaymentStatus(id)
}

func (FfiConverterTypePaymentStatus) Write(writer io.Writer, value PaymentStatus) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerTypePaymentStatus struct{}

func (_ FfiDestroyerTypePaymentStatus) Destroy(value PaymentStatus) {
}

type PendingSweepBalance interface {
	Destroy()
}
type PendingSweepBalancePendingBroadcast struct {
	ChannelId      *ChannelId
	AmountSatoshis uint64
}

func (e PendingSweepBalancePendingBroadcast) Destroy() {
	FfiDestroyerOptionalTypeChannelId{}.Destroy(e.ChannelId)
	FfiDestroyerUint64{}.Destroy(e.AmountSatoshis)
}

type PendingSweepBalanceBroadcastAwaitingConfirmation struct {
	ChannelId             *ChannelId
	LatestBroadcastHeight uint32
	LatestSpendingTxid    Txid
	AmountSatoshis        uint64
}

func (e PendingSweepBalanceBroadcastAwaitingConfirmation) Destroy() {
	FfiDestroyerOptionalTypeChannelId{}.Destroy(e.ChannelId)
	FfiDestroyerUint32{}.Destroy(e.LatestBroadcastHeight)
	FfiDestroyerTypeTxid{}.Destroy(e.LatestSpendingTxid)
	FfiDestroyerUint64{}.Destroy(e.AmountSatoshis)
}

type PendingSweepBalanceAwaitingThresholdConfirmations struct {
	ChannelId          *ChannelId
	LatestSpendingTxid Txid
	ConfirmationHash   BlockHash
	ConfirmationHeight uint32
	AmountSatoshis     uint64
}

func (e PendingSweepBalanceAwaitingThresholdConfirmations) Destroy() {
	FfiDestroyerOptionalTypeChannelId{}.Destroy(e.ChannelId)
	FfiDestroyerTypeTxid{}.Destroy(e.LatestSpendingTxid)
	FfiDestroyerTypeBlockHash{}.Destroy(e.ConfirmationHash)
	FfiDestroyerUint32{}.Destroy(e.ConfirmationHeight)
	FfiDestroyerUint64{}.Destroy(e.AmountSatoshis)
}

type FfiConverterTypePendingSweepBalance struct{}

var FfiConverterTypePendingSweepBalanceINSTANCE = FfiConverterTypePendingSweepBalance{}

func (c FfiConverterTypePendingSweepBalance) Lift(rb RustBufferI) PendingSweepBalance {
	return LiftFromRustBuffer[PendingSweepBalance](c, rb)
}

func (c FfiConverterTypePendingSweepBalance) Lower(value PendingSweepBalance) RustBuffer {
	return LowerIntoRustBuffer[PendingSweepBalance](c, value)
}
func (FfiConverterTypePendingSweepBalance) Read(reader io.Reader) PendingSweepBalance {
	id := readInt32(reader)
	switch id {
	case 1:
		return PendingSweepBalancePendingBroadcast{
			FfiConverterOptionalTypeChannelIdINSTANCE.Read(reader),
			FfiConverterUint64INSTANCE.Read(reader),
		}
	case 2:
		return PendingSweepBalanceBroadcastAwaitingConfirmation{
			FfiConverterOptionalTypeChannelIdINSTANCE.Read(reader),
			FfiConverterUint32INSTANCE.Read(reader),
			FfiConverterTypeTxidINSTANCE.Read(reader),
			FfiConverterUint64INSTANCE.Read(reader),
		}
	case 3:
		return PendingSweepBalanceAwaitingThresholdConfirmations{
			FfiConverterOptionalTypeChannelIdINSTANCE.Read(reader),
			FfiConverterTypeTxidINSTANCE.Read(reader),
			FfiConverterTypeBlockHashINSTANCE.Read(reader),
			FfiConverterUint32INSTANCE.Read(reader),
			FfiConverterUint64INSTANCE.Read(reader),
		}
	default:
		panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypePendingSweepBalance.Read()", id))
	}
}

func (FfiConverterTypePendingSweepBalance) Write(writer io.Writer, value PendingSweepBalance) {
	switch variant_value := value.(type) {
	case PendingSweepBalancePendingBroadcast:
		writeInt32(writer, 1)
		FfiConverterOptionalTypeChannelIdINSTANCE.Write(writer, variant_value.ChannelId)
		FfiConverterUint64INSTANCE.Write(writer, variant_value.AmountSatoshis)
	case PendingSweepBalanceBroadcastAwaitingConfirmation:
		writeInt32(writer, 2)
		FfiConverterOptionalTypeChannelIdINSTANCE.Write(writer, variant_value.ChannelId)
		FfiConverterUint32INSTANCE.Write(writer, variant_value.LatestBroadcastHeight)
		FfiConverterTypeTxidINSTANCE.Write(writer, variant_value.LatestSpendingTxid)
		FfiConverterUint64INSTANCE.Write(writer, variant_value.AmountSatoshis)
	case PendingSweepBalanceAwaitingThresholdConfirmations:
		writeInt32(writer, 3)
		FfiConverterOptionalTypeChannelIdINSTANCE.Write(writer, variant_value.ChannelId)
		FfiConverterTypeTxidINSTANCE.Write(writer, variant_value.LatestSpendingTxid)
		FfiConverterTypeBlockHashINSTANCE.Write(writer, variant_value.ConfirmationHash)
		FfiConverterUint32INSTANCE.Write(writer, variant_value.ConfirmationHeight)
		FfiConverterUint64INSTANCE.Write(writer, variant_value.AmountSatoshis)
	default:
		_ = variant_value
		panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypePendingSweepBalance.Write", value))
	}
}

type FfiDestroyerTypePendingSweepBalance struct{}

func (_ FfiDestroyerTypePendingSweepBalance) Destroy(value PendingSweepBalance) {
	value.Destroy()
}

type FfiConverterOptionalUint16 struct{}

var FfiConverterOptionalUint16INSTANCE = FfiConverterOptionalUint16{}

func (c FfiConverterOptionalUint16) Lift(rb RustBufferI) *uint16 {
	return LiftFromRustBuffer[*uint16](c, rb)
}

func (_ FfiConverterOptionalUint16) Read(reader io.Reader) *uint16 {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterUint16INSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalUint16) Lower(value *uint16) RustBuffer {
	return LowerIntoRustBuffer[*uint16](c, value)
}

func (_ FfiConverterOptionalUint16) Write(writer io.Writer, value *uint16) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterUint16INSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalUint16 struct{}

func (_ FfiDestroyerOptionalUint16) Destroy(value *uint16) {
	if value != nil {
		FfiDestroyerUint16{}.Destroy(*value)
	}
}

type FfiConverterOptionalUint32 struct{}

var FfiConverterOptionalUint32INSTANCE = FfiConverterOptionalUint32{}

func (c FfiConverterOptionalUint32) Lift(rb RustBufferI) *uint32 {
	return LiftFromRustBuffer[*uint32](c, rb)
}

func (_ FfiConverterOptionalUint32) Read(reader io.Reader) *uint32 {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterUint32INSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalUint32) Lower(value *uint32) RustBuffer {
	return LowerIntoRustBuffer[*uint32](c, value)
}

func (_ FfiConverterOptionalUint32) Write(writer io.Writer, value *uint32) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterUint32INSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalUint32 struct{}

func (_ FfiDestroyerOptionalUint32) Destroy(value *uint32) {
	if value != nil {
		FfiDestroyerUint32{}.Destroy(*value)
	}
}

type FfiConverterOptionalUint64 struct{}

var FfiConverterOptionalUint64INSTANCE = FfiConverterOptionalUint64{}

func (c FfiConverterOptionalUint64) Lift(rb RustBufferI) *uint64 {
	return LiftFromRustBuffer[*uint64](c, rb)
}

func (_ FfiConverterOptionalUint64) Read(reader io.Reader) *uint64 {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterUint64INSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalUint64) Lower(value *uint64) RustBuffer {
	return LowerIntoRustBuffer[*uint64](c, value)
}

func (_ FfiConverterOptionalUint64) Write(writer io.Writer, value *uint64) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterUint64INSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalUint64 struct{}

func (_ FfiDestroyerOptionalUint64) Destroy(value *uint64) {
	if value != nil {
		FfiDestroyerUint64{}.Destroy(*value)
	}
}

type FfiConverterOptionalString struct{}

var FfiConverterOptionalStringINSTANCE = FfiConverterOptionalString{}

func (c FfiConverterOptionalString) Lift(rb RustBufferI) *string {
	return LiftFromRustBuffer[*string](c, rb)
}

func (_ FfiConverterOptionalString) Read(reader io.Reader) *string {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterStringINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalString) Lower(value *string) RustBuffer {
	return LowerIntoRustBuffer[*string](c, value)
}

func (_ FfiConverterOptionalString) Write(writer io.Writer, value *string) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterStringINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalString struct{}

func (_ FfiDestroyerOptionalString) Destroy(value *string) {
	if value != nil {
		FfiDestroyerString{}.Destroy(*value)
	}
}

type FfiConverterOptionalChannelConfig struct{}

var FfiConverterOptionalChannelConfigINSTANCE = FfiConverterOptionalChannelConfig{}

func (c FfiConverterOptionalChannelConfig) Lift(rb RustBufferI) **ChannelConfig {
	return LiftFromRustBuffer[**ChannelConfig](c, rb)
}

func (_ FfiConverterOptionalChannelConfig) Read(reader io.Reader) **ChannelConfig {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterChannelConfigINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalChannelConfig) Lower(value **ChannelConfig) RustBuffer {
	return LowerIntoRustBuffer[**ChannelConfig](c, value)
}

func (_ FfiConverterOptionalChannelConfig) Write(writer io.Writer, value **ChannelConfig) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterChannelConfigINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalChannelConfig struct{}

func (_ FfiDestroyerOptionalChannelConfig) Destroy(value **ChannelConfig) {
	if value != nil {
		FfiDestroyerChannelConfig{}.Destroy(*value)
	}
}

type FfiConverterOptionalTypeAnchorChannelsConfig struct{}

var FfiConverterOptionalTypeAnchorChannelsConfigINSTANCE = FfiConverterOptionalTypeAnchorChannelsConfig{}

func (c FfiConverterOptionalTypeAnchorChannelsConfig) Lift(rb RustBufferI) *AnchorChannelsConfig {
	return LiftFromRustBuffer[*AnchorChannelsConfig](c, rb)
}

func (_ FfiConverterOptionalTypeAnchorChannelsConfig) Read(reader io.Reader) *AnchorChannelsConfig {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypeAnchorChannelsConfigINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypeAnchorChannelsConfig) Lower(value *AnchorChannelsConfig) RustBuffer {
	return LowerIntoRustBuffer[*AnchorChannelsConfig](c, value)
}

func (_ FfiConverterOptionalTypeAnchorChannelsConfig) Write(writer io.Writer, value *AnchorChannelsConfig) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypeAnchorChannelsConfigINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypeAnchorChannelsConfig struct{}

func (_ FfiDestroyerOptionalTypeAnchorChannelsConfig) Destroy(value *AnchorChannelsConfig) {
	if value != nil {
		FfiDestroyerTypeAnchorChannelsConfig{}.Destroy(*value)
	}
}

type FfiConverterOptionalTypeOutPoint struct{}

var FfiConverterOptionalTypeOutPointINSTANCE = FfiConverterOptionalTypeOutPoint{}

func (c FfiConverterOptionalTypeOutPoint) Lift(rb RustBufferI) *OutPoint {
	return LiftFromRustBuffer[*OutPoint](c, rb)
}

func (_ FfiConverterOptionalTypeOutPoint) Read(reader io.Reader) *OutPoint {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypeOutPointINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypeOutPoint) Lower(value *OutPoint) RustBuffer {
	return LowerIntoRustBuffer[*OutPoint](c, value)
}

func (_ FfiConverterOptionalTypeOutPoint) Write(writer io.Writer, value *OutPoint) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypeOutPointINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypeOutPoint struct{}

func (_ FfiDestroyerOptionalTypeOutPoint) Destroy(value *OutPoint) {
	if value != nil {
		FfiDestroyerTypeOutPoint{}.Destroy(*value)
	}
}

type FfiConverterOptionalTypePaymentDetails struct{}

var FfiConverterOptionalTypePaymentDetailsINSTANCE = FfiConverterOptionalTypePaymentDetails{}

func (c FfiConverterOptionalTypePaymentDetails) Lift(rb RustBufferI) *PaymentDetails {
	return LiftFromRustBuffer[*PaymentDetails](c, rb)
}

func (_ FfiConverterOptionalTypePaymentDetails) Read(reader io.Reader) *PaymentDetails {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypePaymentDetailsINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypePaymentDetails) Lower(value *PaymentDetails) RustBuffer {
	return LowerIntoRustBuffer[*PaymentDetails](c, value)
}

func (_ FfiConverterOptionalTypePaymentDetails) Write(writer io.Writer, value *PaymentDetails) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypePaymentDetailsINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypePaymentDetails struct{}

func (_ FfiDestroyerOptionalTypePaymentDetails) Destroy(value *PaymentDetails) {
	if value != nil {
		FfiDestroyerTypePaymentDetails{}.Destroy(*value)
	}
}

type FfiConverterOptionalTypeChannelType struct{}

var FfiConverterOptionalTypeChannelTypeINSTANCE = FfiConverterOptionalTypeChannelType{}

func (c FfiConverterOptionalTypeChannelType) Lift(rb RustBufferI) *ChannelType {
	return LiftFromRustBuffer[*ChannelType](c, rb)
}

func (_ FfiConverterOptionalTypeChannelType) Read(reader io.Reader) *ChannelType {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypeChannelTypeINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypeChannelType) Lower(value *ChannelType) RustBuffer {
	return LowerIntoRustBuffer[*ChannelType](c, value)
}

func (_ FfiConverterOptionalTypeChannelType) Write(writer io.Writer, value *ChannelType) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypeChannelTypeINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypeChannelType struct{}

func (_ FfiDestroyerOptionalTypeChannelType) Destroy(value *ChannelType) {
	if value != nil {
		FfiDestroyerTypeChannelType{}.Destroy(*value)
	}
}

type FfiConverterOptionalTypeClosureReason struct{}

var FfiConverterOptionalTypeClosureReasonINSTANCE = FfiConverterOptionalTypeClosureReason{}

func (c FfiConverterOptionalTypeClosureReason) Lift(rb RustBufferI) *ClosureReason {
	return LiftFromRustBuffer[*ClosureReason](c, rb)
}

func (_ FfiConverterOptionalTypeClosureReason) Read(reader io.Reader) *ClosureReason {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypeClosureReasonINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypeClosureReason) Lower(value *ClosureReason) RustBuffer {
	return LowerIntoRustBuffer[*ClosureReason](c, value)
}

func (_ FfiConverterOptionalTypeClosureReason) Write(writer io.Writer, value *ClosureReason) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypeClosureReasonINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypeClosureReason struct{}

func (_ FfiDestroyerOptionalTypeClosureReason) Destroy(value *ClosureReason) {
	if value != nil {
		FfiDestroyerTypeClosureReason{}.Destroy(*value)
	}
}

type FfiConverterOptionalTypeEvent struct{}

var FfiConverterOptionalTypeEventINSTANCE = FfiConverterOptionalTypeEvent{}

func (c FfiConverterOptionalTypeEvent) Lift(rb RustBufferI) *Event {
	return LiftFromRustBuffer[*Event](c, rb)
}

func (_ FfiConverterOptionalTypeEvent) Read(reader io.Reader) *Event {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypeEventINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypeEvent) Lower(value *Event) RustBuffer {
	return LowerIntoRustBuffer[*Event](c, value)
}

func (_ FfiConverterOptionalTypeEvent) Write(writer io.Writer, value *Event) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypeEventINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypeEvent struct{}

func (_ FfiDestroyerOptionalTypeEvent) Destroy(value *Event) {
	if value != nil {
		FfiDestroyerTypeEvent{}.Destroy(*value)
	}
}

type FfiConverterOptionalTypePaymentFailureReason struct{}

var FfiConverterOptionalTypePaymentFailureReasonINSTANCE = FfiConverterOptionalTypePaymentFailureReason{}

func (c FfiConverterOptionalTypePaymentFailureReason) Lift(rb RustBufferI) *PaymentFailureReason {
	return LiftFromRustBuffer[*PaymentFailureReason](c, rb)
}

func (_ FfiConverterOptionalTypePaymentFailureReason) Read(reader io.Reader) *PaymentFailureReason {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypePaymentFailureReasonINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypePaymentFailureReason) Lower(value *PaymentFailureReason) RustBuffer {
	return LowerIntoRustBuffer[*PaymentFailureReason](c, value)
}

func (_ FfiConverterOptionalTypePaymentFailureReason) Write(writer io.Writer, value *PaymentFailureReason) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypePaymentFailureReasonINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypePaymentFailureReason struct{}

func (_ FfiDestroyerOptionalTypePaymentFailureReason) Destroy(value *PaymentFailureReason) {
	if value != nil {
		FfiDestroyerTypePaymentFailureReason{}.Destroy(*value)
	}
}

type FfiConverterOptionalSequenceTypeSocketAddress struct{}

var FfiConverterOptionalSequenceTypeSocketAddressINSTANCE = FfiConverterOptionalSequenceTypeSocketAddress{}

func (c FfiConverterOptionalSequenceTypeSocketAddress) Lift(rb RustBufferI) *[]SocketAddress {
	return LiftFromRustBuffer[*[]SocketAddress](c, rb)
}

func (_ FfiConverterOptionalSequenceTypeSocketAddress) Read(reader io.Reader) *[]SocketAddress {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterSequenceTypeSocketAddressINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalSequenceTypeSocketAddress) Lower(value *[]SocketAddress) RustBuffer {
	return LowerIntoRustBuffer[*[]SocketAddress](c, value)
}

func (_ FfiConverterOptionalSequenceTypeSocketAddress) Write(writer io.Writer, value *[]SocketAddress) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterSequenceTypeSocketAddressINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalSequenceTypeSocketAddress struct{}

func (_ FfiDestroyerOptionalSequenceTypeSocketAddress) Destroy(value *[]SocketAddress) {
	if value != nil {
		FfiDestroyerSequenceTypeSocketAddress{}.Destroy(*value)
	}
}

type FfiConverterOptionalTypeChannelId struct{}

var FfiConverterOptionalTypeChannelIdINSTANCE = FfiConverterOptionalTypeChannelId{}

func (c FfiConverterOptionalTypeChannelId) Lift(rb RustBufferI) *ChannelId {
	return LiftFromRustBuffer[*ChannelId](c, rb)
}

func (_ FfiConverterOptionalTypeChannelId) Read(reader io.Reader) *ChannelId {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypeChannelIdINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypeChannelId) Lower(value *ChannelId) RustBuffer {
	return LowerIntoRustBuffer[*ChannelId](c, value)
}

func (_ FfiConverterOptionalTypeChannelId) Write(writer io.Writer, value *ChannelId) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypeChannelIdINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypeChannelId struct{}

func (_ FfiDestroyerOptionalTypeChannelId) Destroy(value *ChannelId) {
	if value != nil {
		FfiDestroyerTypeChannelId{}.Destroy(*value)
	}
}

type FfiConverterOptionalTypePaymentId struct{}

var FfiConverterOptionalTypePaymentIdINSTANCE = FfiConverterOptionalTypePaymentId{}

func (c FfiConverterOptionalTypePaymentId) Lift(rb RustBufferI) *PaymentId {
	return LiftFromRustBuffer[*PaymentId](c, rb)
}

func (_ FfiConverterOptionalTypePaymentId) Read(reader io.Reader) *PaymentId {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypePaymentIdINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypePaymentId) Lower(value *PaymentId) RustBuffer {
	return LowerIntoRustBuffer[*PaymentId](c, value)
}

func (_ FfiConverterOptionalTypePaymentId) Write(writer io.Writer, value *PaymentId) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypePaymentIdINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypePaymentId struct{}

func (_ FfiDestroyerOptionalTypePaymentId) Destroy(value *PaymentId) {
	if value != nil {
		FfiDestroyerTypePaymentId{}.Destroy(*value)
	}
}

type FfiConverterOptionalTypePaymentPreimage struct{}

var FfiConverterOptionalTypePaymentPreimageINSTANCE = FfiConverterOptionalTypePaymentPreimage{}

func (c FfiConverterOptionalTypePaymentPreimage) Lift(rb RustBufferI) *PaymentPreimage {
	return LiftFromRustBuffer[*PaymentPreimage](c, rb)
}

func (_ FfiConverterOptionalTypePaymentPreimage) Read(reader io.Reader) *PaymentPreimage {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypePaymentPreimageINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypePaymentPreimage) Lower(value *PaymentPreimage) RustBuffer {
	return LowerIntoRustBuffer[*PaymentPreimage](c, value)
}

func (_ FfiConverterOptionalTypePaymentPreimage) Write(writer io.Writer, value *PaymentPreimage) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypePaymentPreimageINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypePaymentPreimage struct{}

func (_ FfiDestroyerOptionalTypePaymentPreimage) Destroy(value *PaymentPreimage) {
	if value != nil {
		FfiDestroyerTypePaymentPreimage{}.Destroy(*value)
	}
}

type FfiConverterOptionalTypePaymentSecret struct{}

var FfiConverterOptionalTypePaymentSecretINSTANCE = FfiConverterOptionalTypePaymentSecret{}

func (c FfiConverterOptionalTypePaymentSecret) Lift(rb RustBufferI) *PaymentSecret {
	return LiftFromRustBuffer[*PaymentSecret](c, rb)
}

func (_ FfiConverterOptionalTypePaymentSecret) Read(reader io.Reader) *PaymentSecret {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypePaymentSecretINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypePaymentSecret) Lower(value *PaymentSecret) RustBuffer {
	return LowerIntoRustBuffer[*PaymentSecret](c, value)
}

func (_ FfiConverterOptionalTypePaymentSecret) Write(writer io.Writer, value *PaymentSecret) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypePaymentSecretINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypePaymentSecret struct{}

func (_ FfiDestroyerOptionalTypePaymentSecret) Destroy(value *PaymentSecret) {
	if value != nil {
		FfiDestroyerTypePaymentSecret{}.Destroy(*value)
	}
}

type FfiConverterOptionalTypePublicKey struct{}

var FfiConverterOptionalTypePublicKeyINSTANCE = FfiConverterOptionalTypePublicKey{}

func (c FfiConverterOptionalTypePublicKey) Lift(rb RustBufferI) *PublicKey {
	return LiftFromRustBuffer[*PublicKey](c, rb)
}

func (_ FfiConverterOptionalTypePublicKey) Read(reader io.Reader) *PublicKey {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypePublicKeyINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypePublicKey) Lower(value *PublicKey) RustBuffer {
	return LowerIntoRustBuffer[*PublicKey](c, value)
}

func (_ FfiConverterOptionalTypePublicKey) Write(writer io.Writer, value *PublicKey) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypePublicKeyINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypePublicKey struct{}

func (_ FfiDestroyerOptionalTypePublicKey) Destroy(value *PublicKey) {
	if value != nil {
		FfiDestroyerTypePublicKey{}.Destroy(*value)
	}
}

type FfiConverterSequenceUint8 struct{}

var FfiConverterSequenceUint8INSTANCE = FfiConverterSequenceUint8{}

func (c FfiConverterSequenceUint8) Lift(rb RustBufferI) []uint8 {
	return LiftFromRustBuffer[[]uint8](c, rb)
}

func (c FfiConverterSequenceUint8) Read(reader io.Reader) []uint8 {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]uint8, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterUint8INSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceUint8) Lower(value []uint8) RustBuffer {
	return LowerIntoRustBuffer[[]uint8](c, value)
}

func (c FfiConverterSequenceUint8) Write(writer io.Writer, value []uint8) {
	if len(value) > math.MaxInt32 {
		panic("[]uint8 is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterUint8INSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceUint8 struct{}

func (FfiDestroyerSequenceUint8) Destroy(sequence []uint8) {
	for _, value := range sequence {
		FfiDestroyerUint8{}.Destroy(value)
	}
}

type FfiConverterSequenceTypeChannelDetails struct{}

var FfiConverterSequenceTypeChannelDetailsINSTANCE = FfiConverterSequenceTypeChannelDetails{}

func (c FfiConverterSequenceTypeChannelDetails) Lift(rb RustBufferI) []ChannelDetails {
	return LiftFromRustBuffer[[]ChannelDetails](c, rb)
}

func (c FfiConverterSequenceTypeChannelDetails) Read(reader io.Reader) []ChannelDetails {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]ChannelDetails, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypeChannelDetailsINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypeChannelDetails) Lower(value []ChannelDetails) RustBuffer {
	return LowerIntoRustBuffer[[]ChannelDetails](c, value)
}

func (c FfiConverterSequenceTypeChannelDetails) Write(writer io.Writer, value []ChannelDetails) {
	if len(value) > math.MaxInt32 {
		panic("[]ChannelDetails is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypeChannelDetailsINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypeChannelDetails struct{}

func (FfiDestroyerSequenceTypeChannelDetails) Destroy(sequence []ChannelDetails) {
	for _, value := range sequence {
		FfiDestroyerTypeChannelDetails{}.Destroy(value)
	}
}

type FfiConverterSequenceTypePaymentDetails struct{}

var FfiConverterSequenceTypePaymentDetailsINSTANCE = FfiConverterSequenceTypePaymentDetails{}

func (c FfiConverterSequenceTypePaymentDetails) Lift(rb RustBufferI) []PaymentDetails {
	return LiftFromRustBuffer[[]PaymentDetails](c, rb)
}

func (c FfiConverterSequenceTypePaymentDetails) Read(reader io.Reader) []PaymentDetails {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]PaymentDetails, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypePaymentDetailsINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypePaymentDetails) Lower(value []PaymentDetails) RustBuffer {
	return LowerIntoRustBuffer[[]PaymentDetails](c, value)
}

func (c FfiConverterSequenceTypePaymentDetails) Write(writer io.Writer, value []PaymentDetails) {
	if len(value) > math.MaxInt32 {
		panic("[]PaymentDetails is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypePaymentDetailsINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypePaymentDetails struct{}

func (FfiDestroyerSequenceTypePaymentDetails) Destroy(sequence []PaymentDetails) {
	for _, value := range sequence {
		FfiDestroyerTypePaymentDetails{}.Destroy(value)
	}
}

type FfiConverterSequenceTypePeerDetails struct{}

var FfiConverterSequenceTypePeerDetailsINSTANCE = FfiConverterSequenceTypePeerDetails{}

func (c FfiConverterSequenceTypePeerDetails) Lift(rb RustBufferI) []PeerDetails {
	return LiftFromRustBuffer[[]PeerDetails](c, rb)
}

func (c FfiConverterSequenceTypePeerDetails) Read(reader io.Reader) []PeerDetails {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]PeerDetails, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypePeerDetailsINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypePeerDetails) Lower(value []PeerDetails) RustBuffer {
	return LowerIntoRustBuffer[[]PeerDetails](c, value)
}

func (c FfiConverterSequenceTypePeerDetails) Write(writer io.Writer, value []PeerDetails) {
	if len(value) > math.MaxInt32 {
		panic("[]PeerDetails is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypePeerDetailsINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypePeerDetails struct{}

func (FfiDestroyerSequenceTypePeerDetails) Destroy(sequence []PeerDetails) {
	for _, value := range sequence {
		FfiDestroyerTypePeerDetails{}.Destroy(value)
	}
}

type FfiConverterSequenceTypeTlvEntry struct{}

var FfiConverterSequenceTypeTlvEntryINSTANCE = FfiConverterSequenceTypeTlvEntry{}

func (c FfiConverterSequenceTypeTlvEntry) Lift(rb RustBufferI) []TlvEntry {
	return LiftFromRustBuffer[[]TlvEntry](c, rb)
}

func (c FfiConverterSequenceTypeTlvEntry) Read(reader io.Reader) []TlvEntry {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]TlvEntry, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypeTlvEntryINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypeTlvEntry) Lower(value []TlvEntry) RustBuffer {
	return LowerIntoRustBuffer[[]TlvEntry](c, value)
}

func (c FfiConverterSequenceTypeTlvEntry) Write(writer io.Writer, value []TlvEntry) {
	if len(value) > math.MaxInt32 {
		panic("[]TlvEntry is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypeTlvEntryINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypeTlvEntry struct{}

func (FfiDestroyerSequenceTypeTlvEntry) Destroy(sequence []TlvEntry) {
	for _, value := range sequence {
		FfiDestroyerTypeTlvEntry{}.Destroy(value)
	}
}

type FfiConverterSequenceTypeLightningBalance struct{}

var FfiConverterSequenceTypeLightningBalanceINSTANCE = FfiConverterSequenceTypeLightningBalance{}

func (c FfiConverterSequenceTypeLightningBalance) Lift(rb RustBufferI) []LightningBalance {
	return LiftFromRustBuffer[[]LightningBalance](c, rb)
}

func (c FfiConverterSequenceTypeLightningBalance) Read(reader io.Reader) []LightningBalance {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]LightningBalance, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypeLightningBalanceINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypeLightningBalance) Lower(value []LightningBalance) RustBuffer {
	return LowerIntoRustBuffer[[]LightningBalance](c, value)
}

func (c FfiConverterSequenceTypeLightningBalance) Write(writer io.Writer, value []LightningBalance) {
	if len(value) > math.MaxInt32 {
		panic("[]LightningBalance is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypeLightningBalanceINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypeLightningBalance struct{}

func (FfiDestroyerSequenceTypeLightningBalance) Destroy(sequence []LightningBalance) {
	for _, value := range sequence {
		FfiDestroyerTypeLightningBalance{}.Destroy(value)
	}
}

type FfiConverterSequenceTypePendingSweepBalance struct{}

var FfiConverterSequenceTypePendingSweepBalanceINSTANCE = FfiConverterSequenceTypePendingSweepBalance{}

func (c FfiConverterSequenceTypePendingSweepBalance) Lift(rb RustBufferI) []PendingSweepBalance {
	return LiftFromRustBuffer[[]PendingSweepBalance](c, rb)
}

func (c FfiConverterSequenceTypePendingSweepBalance) Read(reader io.Reader) []PendingSweepBalance {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]PendingSweepBalance, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypePendingSweepBalanceINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypePendingSweepBalance) Lower(value []PendingSweepBalance) RustBuffer {
	return LowerIntoRustBuffer[[]PendingSweepBalance](c, value)
}

func (c FfiConverterSequenceTypePendingSweepBalance) Write(writer io.Writer, value []PendingSweepBalance) {
	if len(value) > math.MaxInt32 {
		panic("[]PendingSweepBalance is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypePendingSweepBalanceINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypePendingSweepBalance struct{}

func (FfiDestroyerSequenceTypePendingSweepBalance) Destroy(sequence []PendingSweepBalance) {
	for _, value := range sequence {
		FfiDestroyerTypePendingSweepBalance{}.Destroy(value)
	}
}

type FfiConverterSequenceTypePublicKey struct{}

var FfiConverterSequenceTypePublicKeyINSTANCE = FfiConverterSequenceTypePublicKey{}

func (c FfiConverterSequenceTypePublicKey) Lift(rb RustBufferI) []PublicKey {
	return LiftFromRustBuffer[[]PublicKey](c, rb)
}

func (c FfiConverterSequenceTypePublicKey) Read(reader io.Reader) []PublicKey {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]PublicKey, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypePublicKeyINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypePublicKey) Lower(value []PublicKey) RustBuffer {
	return LowerIntoRustBuffer[[]PublicKey](c, value)
}

func (c FfiConverterSequenceTypePublicKey) Write(writer io.Writer, value []PublicKey) {
	if len(value) > math.MaxInt32 {
		panic("[]PublicKey is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypePublicKeyINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypePublicKey struct{}

func (FfiDestroyerSequenceTypePublicKey) Destroy(sequence []PublicKey) {
	for _, value := range sequence {
		FfiDestroyerTypePublicKey{}.Destroy(value)
	}
}

type FfiConverterSequenceTypeSocketAddress struct{}

var FfiConverterSequenceTypeSocketAddressINSTANCE = FfiConverterSequenceTypeSocketAddress{}

func (c FfiConverterSequenceTypeSocketAddress) Lift(rb RustBufferI) []SocketAddress {
	return LiftFromRustBuffer[[]SocketAddress](c, rb)
}

func (c FfiConverterSequenceTypeSocketAddress) Read(reader io.Reader) []SocketAddress {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]SocketAddress, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypeSocketAddressINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypeSocketAddress) Lower(value []SocketAddress) RustBuffer {
	return LowerIntoRustBuffer[[]SocketAddress](c, value)
}

func (c FfiConverterSequenceTypeSocketAddress) Write(writer io.Writer, value []SocketAddress) {
	if len(value) > math.MaxInt32 {
		panic("[]SocketAddress is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypeSocketAddressINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypeSocketAddress struct{}

func (FfiDestroyerSequenceTypeSocketAddress) Destroy(sequence []SocketAddress) {
	for _, value := range sequence {
		FfiDestroyerTypeSocketAddress{}.Destroy(value)
	}
}

/**
 * Typealias from the type name used in the UDL file to the builtin type.  This
 * is needed because the UDL type name is used in function/method signatures.
 * It's also what we have an external type that references a custom type.
 */
type Address = string
type FfiConverterTypeAddress = FfiConverterString
type FfiDestroyerTypeAddress = FfiDestroyerString

var FfiConverterTypeAddressINSTANCE = FfiConverterString{}

/**
 * Typealias from the type name used in the UDL file to the builtin type.  This
 * is needed because the UDL type name is used in function/method signatures.
 * It's also what we have an external type that references a custom type.
 */
type BlockHash = string
type FfiConverterTypeBlockHash = FfiConverterString
type FfiDestroyerTypeBlockHash = FfiDestroyerString

var FfiConverterTypeBlockHashINSTANCE = FfiConverterString{}

/**
 * Typealias from the type name used in the UDL file to the builtin type.  This
 * is needed because the UDL type name is used in function/method signatures.
 * It's also what we have an external type that references a custom type.
 */
type Bolt11Invoice = string
type FfiConverterTypeBolt11Invoice = FfiConverterString
type FfiDestroyerTypeBolt11Invoice = FfiDestroyerString

var FfiConverterTypeBolt11InvoiceINSTANCE = FfiConverterString{}

/**
 * Typealias from the type name used in the UDL file to the builtin type.  This
 * is needed because the UDL type name is used in function/method signatures.
 * It's also what we have an external type that references a custom type.
 */
type ChannelId = string
type FfiConverterTypeChannelId = FfiConverterString
type FfiDestroyerTypeChannelId = FfiDestroyerString

var FfiConverterTypeChannelIdINSTANCE = FfiConverterString{}

/**
 * Typealias from the type name used in the UDL file to the builtin type.  This
 * is needed because the UDL type name is used in function/method signatures.
 * It's also what we have an external type that references a custom type.
 */
type Mnemonic = string
type FfiConverterTypeMnemonic = FfiConverterString
type FfiDestroyerTypeMnemonic = FfiDestroyerString

var FfiConverterTypeMnemonicINSTANCE = FfiConverterString{}

/**
 * Typealias from the type name used in the UDL file to the builtin type.  This
 * is needed because the UDL type name is used in function/method signatures.
 * It's also what we have an external type that references a custom type.
 */
type Network = string
type FfiConverterTypeNetwork = FfiConverterString
type FfiDestroyerTypeNetwork = FfiDestroyerString

var FfiConverterTypeNetworkINSTANCE = FfiConverterString{}

/**
 * Typealias from the type name used in the UDL file to the builtin type.  This
 * is needed because the UDL type name is used in function/method signatures.
 * It's also what we have an external type that references a custom type.
 */
type PaymentHash = string
type FfiConverterTypePaymentHash = FfiConverterString
type FfiDestroyerTypePaymentHash = FfiDestroyerString

var FfiConverterTypePaymentHashINSTANCE = FfiConverterString{}

/**
 * Typealias from the type name used in the UDL file to the builtin type.  This
 * is needed because the UDL type name is used in function/method signatures.
 * It's also what we have an external type that references a custom type.
 */
type PaymentId = string
type FfiConverterTypePaymentId = FfiConverterString
type FfiDestroyerTypePaymentId = FfiDestroyerString

var FfiConverterTypePaymentIdINSTANCE = FfiConverterString{}

/**
 * Typealias from the type name used in the UDL file to the builtin type.  This
 * is needed because the UDL type name is used in function/method signatures.
 * It's also what we have an external type that references a custom type.
 */
type PaymentPreimage = string
type FfiConverterTypePaymentPreimage = FfiConverterString
type FfiDestroyerTypePaymentPreimage = FfiDestroyerString

var FfiConverterTypePaymentPreimageINSTANCE = FfiConverterString{}

/**
 * Typealias from the type name used in the UDL file to the builtin type.  This
 * is needed because the UDL type name is used in function/method signatures.
 * It's also what we have an external type that references a custom type.
 */
type PaymentSecret = string
type FfiConverterTypePaymentSecret = FfiConverterString
type FfiDestroyerTypePaymentSecret = FfiDestroyerString

var FfiConverterTypePaymentSecretINSTANCE = FfiConverterString{}

/**
 * Typealias from the type name used in the UDL file to the builtin type.  This
 * is needed because the UDL type name is used in function/method signatures.
 * It's also what we have an external type that references a custom type.
 */
type PublicKey = string
type FfiConverterTypePublicKey = FfiConverterString
type FfiDestroyerTypePublicKey = FfiDestroyerString

var FfiConverterTypePublicKeyINSTANCE = FfiConverterString{}

/**
 * Typealias from the type name used in the UDL file to the builtin type.  This
 * is needed because the UDL type name is used in function/method signatures.
 * It's also what we have an external type that references a custom type.
 */
type SocketAddress = string
type FfiConverterTypeSocketAddress = FfiConverterString
type FfiDestroyerTypeSocketAddress = FfiDestroyerString

var FfiConverterTypeSocketAddressINSTANCE = FfiConverterString{}

/**
 * Typealias from the type name used in the UDL file to the builtin type.  This
 * is needed because the UDL type name is used in function/method signatures.
 * It's also what we have an external type that references a custom type.
 */
type Txid = string
type FfiConverterTypeTxid = FfiConverterString
type FfiDestroyerTypeTxid = FfiDestroyerString

var FfiConverterTypeTxidINSTANCE = FfiConverterString{}

/**
 * Typealias from the type name used in the UDL file to the builtin type.  This
 * is needed because the UDL type name is used in function/method signatures.
 * It's also what we have an external type that references a custom type.
 */
type UntrustedString = string
type FfiConverterTypeUntrustedString = FfiConverterString
type FfiDestroyerTypeUntrustedString = FfiDestroyerString

var FfiConverterTypeUntrustedStringINSTANCE = FfiConverterString{}

/**
 * Typealias from the type name used in the UDL file to the builtin type.  This
 * is needed because the UDL type name is used in function/method signatures.
 * It's also what we have an external type that references a custom type.
 */
type UserChannelId = string
type FfiConverterTypeUserChannelId = FfiConverterString
type FfiDestroyerTypeUserChannelId = FfiDestroyerString

var FfiConverterTypeUserChannelIdINSTANCE = FfiConverterString{}

func DefaultConfig() Config {
	return FfiConverterTypeConfigINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_func_default_config(_uniffiStatus)
	}))
}

func GenerateEntropyMnemonic() Mnemonic {
	return FfiConverterTypeMnemonicINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_func_generate_entropy_mnemonic(_uniffiStatus)
	}))
}
