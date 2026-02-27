//go:build wasip1

package guest

//go:wasmimport ghost_ops kv_get
func kv_get(keyPtr, keyLen, valPtr, valLen uint32) uint32

//go:wasmimport ghost_ops rpc
func rpc(svcPtr, svcLen, methPtr, methLen, payPtr, payLen, outPtr, outLen uint32) uint32

//go:wasmimport ghost_ops next_command
func next_command(methPtr, methCap uint32) uint64 // returns (payLen << 32) | methLen

//go:wasmimport ghost_ops read_payload
func read_payload(ptr, cap uint32)

//go:wasmimport ghost_ops submit_result
func submit_result(ptr, len uint32)
