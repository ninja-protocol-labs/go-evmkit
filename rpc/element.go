package rpc

// Standard JSON-RPC method names for the ETHXxx/NetVersion/Web3ClientVersion
// helpers below. Any other method (engine_*, erigon_*, debug_*, ...) has no
// constant here — pass its name directly to NewElement.
const (
	MethodETHChainID                             = "eth_chainId"
	MethodETHGasPrice                            = "eth_gasPrice"
	MethodETHBlockNumber                         = "eth_blockNumber"
	MethodETHGetBalance                          = "eth_getBalance"
	MethodETHGetCode                             = "eth_getCode"
	MethodETHCall                                = "eth_call"
	MethodETHEstimateGas                         = "eth_estimateGas"
	MethodETHGetTransactionCount                 = "eth_getTransactionCount"
	MethodETHMaxPriorityFeePerGas                = "eth_maxPriorityFeePerGas"
	MethodETHSendRawTransaction                  = "eth_sendRawTransaction"
	MethodETHGetBlockByNumber                    = "eth_getBlockByNumber"
	MethodETHGetBlockByHash                      = "eth_getBlockByHash"
	MethodETHGetTransactionByHash                = "eth_getTransactionByHash"
	MethodETHGetTransactionByBlockHashAndIndex   = "eth_getTransactionByBlockHashAndIndex"
	MethodETHGetTransactionByBlockNumberAndIndex = "eth_getTransactionByBlockNumberAndIndex"
	MethodETHGetTransactionReceipt               = "eth_getTransactionReceipt"
	MethodETHGetBlockReceipts                    = "eth_getBlockReceipts"
	MethodETHGetBlockTransactionCountByHash      = "eth_getBlockTransactionCountByHash"
	MethodETHGetBlockTransactionCountByNumber    = "eth_getBlockTransactionCountByNumber"
	MethodETHGetStorageAt                        = "eth_getStorageAt"
	MethodETHGetLogs                             = "eth_getLogs"
	MethodETHFeeHistory                          = "eth_feeHistory"
	MethodETHCreateAccessList                    = "eth_createAccessList"
	MethodETHSyncing                             = "eth_syncing"

	MethodNetVersion = "net_version"

	MethodWeb3ClientVersion = "web3_clientVersion"

	MethodTxpoolStatus      = "txpool_status"
	MethodTxpoolContent     = "txpool_content"
	MethodTxpoolContentFrom = "txpool_contentFrom"
	MethodTxpoolInspect     = "txpool_inspect"

	MethodDebugTraceTransaction   = "debug_traceTransaction"
	MethodDebugTraceCall          = "debug_traceCall"
	MethodDebugTraceBlockByNumber = "debug_traceBlockByNumber"
	MethodDebugTraceBlockByHash   = "debug_traceBlockByHash"
	MethodDebugGetRawTransaction  = "debug_getRawTransaction"
	MethodDebugGetRawReceipts     = "debug_getRawReceipts"
	MethodDebugGetRawHeader       = "debug_getRawHeader"
	MethodDebugGetRawBlock        = "debug_getRawBlock"
	MethodDebugGetBadBlocks       = "debug_getBadBlocks"
	MethodDebugStorageRangeAt     = "debug_storageRangeAt"

	MethodTraceTransaction             = "trace_transaction"
	MethodTraceBlock                   = "trace_block"
	MethodTraceFilter                  = "trace_filter"
	MethodTraceCall                    = "trace_call"
	MethodTraceReplayTransaction       = "trace_replayTransaction"
	MethodTraceReplayBlockTransactions = "trace_replayBlockTransactions"
)

// BlockTag is a symbolic block parameter, usable wherever a method takes a
// block: string in place of a hex block number.
const (
	BlockTagEarliest  = "earliest"
	BlockTagLatest    = "latest"
	BlockTagPending   = "pending"
	BlockTagSafe      = "safe"
	BlockTagFinalized = "finalized"
)

// Tracer is a built-in geth tracer name, for the "tracer" field of the
// traceConfig object passed to DebugTraceTransaction, DebugTraceCall,
// DebugTraceBlockByNumber and DebugTraceBlockByHash.
const (
	TracerCall     = "callTracer"
	TracerPrestate = "prestateTracer"
	Tracer4Byte    = "4byteTracer"
	TracerNoop     = "noopTracer"
	TracerMux      = "muxTracer"
	TracerFlatCall = "flatCallTracer"
)

// TraceType selects what TraceCall, TraceReplayTransaction and
// TraceReplayBlockTransactions compute, as an element of their traceTypes
// slice.
const (
	TraceTypeTrace     = "trace"
	TraceTypeVMTrace   = "vmTrace"
	TraceTypeStateDiff = "stateDiff"
)

// Element is a single JSON-RPC request/response pairing: the caller
// supplies ID so Client stays stateless, and Result is a pointer the
// response gets decoded into.
type Element struct {
	ID     string
	Method string
	Params any
	Result any
}

// NewElement builds an Element for any JSON-RPC method, including ones this
// package has no named helper for (engine_*, erigon_*, debug_*, or any
// other node-specific extension).
func NewElement(id, method string, params, result any) Element {
	return Element{
		ID:     id,
		Method: method,
		Params: params,
		Result: result,
	}
}

// request is the JSON-RPC 2.0 wire representation of an Element.
type request struct {
	JSONRPC string `json:"jsonrpc"`
	ID      string `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

// toRequest converts e to its JSON-RPC 2.0 wire representation.
func (e Element) toRequest() request {
	return request{
		JSONRPC: "2.0",
		ID:      e.ID,
		Method:  e.Method,
		Params:  e.Params,
	}
}

// Elements is a batch of Element, sent together as one JSON-RPC batch request.
type Elements []Element

// With appends elem and returns the receiver, for chaining.
func (e *Elements) With(elem Element) *Elements {
	*e = append(*e, elem)
	return e
}

// Len returns the number of elements.
func (e *Elements) Len() int {
	return len(*e)
}

// GetID returns the ID of the i-th element.
func (e *Elements) GetID(i int) string {
	return (*e)[i].ID
}

// GetMethod returns the method of the i-th element.
func (e *Elements) GetMethod(i int) string {
	return (*e)[i].Method
}

// GetParams returns the params of the i-th element.
func (e *Elements) GetParams(i int) any {
	return (*e)[i].Params
}

// GetResult returns the result destination of the i-th element.
func (e *Elements) GetResult(i int) any {
	return (*e)[i].Result
}

// ETHChainID builds the eth_chainId element.
func ETHChainID(id string, result *string) Element {
	return Element{
		ID:     id,
		Method: MethodETHChainID,
		Result: result,
	}
}

// ETHGasPrice builds the eth_gasPrice element.
func ETHGasPrice(id string, result *string) Element {
	return Element{
		ID:     id,
		Method: MethodETHGasPrice,
		Result: result,
	}
}

// ETHBlockNumber builds the eth_blockNumber element.
func ETHBlockNumber(id string, result *string) Element {
	return Element{
		ID:     id,
		Method: MethodETHBlockNumber,
		Result: result,
	}
}

// ETHGetBalance builds the eth_getBalance element.
func ETHGetBalance(id string, address, block string, result *string) Element {
	return Element{
		ID:     id,
		Method: MethodETHGetBalance,
		Params: []any{address, block},
		Result: result,
	}
}

// ETHGetCode builds the eth_getCode element.
func ETHGetCode(id string, address, block string, result *string) Element {
	return Element{
		ID:     id,
		Method: MethodETHGetCode,
		Params: []any{address, block},
		Result: result,
	}
}

// ETHCall builds the eth_call element.
func ETHCall(id string, params any, block string, result *string) Element {
	return Element{
		ID:     id,
		Method: MethodETHCall,
		Params: []any{params, block},
		Result: result,
	}
}

// ETHEstimateGas builds the eth_estimateGas element.
func ETHEstimateGas(id string, params any, block string, result *string) Element {
	return Element{
		ID:     id,
		Method: MethodETHEstimateGas,
		Params: []any{params, block},
		Result: result,
	}
}

// ETHGetTransactionCount builds the eth_getTransactionCount element.
func ETHGetTransactionCount(id string, address, block string, result *string) Element {
	return Element{
		ID:     id,
		Method: MethodETHGetTransactionCount,
		Params: []any{address, block},
		Result: result,
	}
}

// ETHMaxPriorityFeePerGas builds the eth_maxPriorityFeePerGas element.
func ETHMaxPriorityFeePerGas(id string, result *string) Element {
	return Element{
		ID:     id,
		Method: MethodETHMaxPriorityFeePerGas,
		Result: result,
	}
}

// ETHSendRawTransaction builds the eth_sendRawTransaction element.
func ETHSendRawTransaction(id string, rawTx string, result *string) Element {
	return Element{
		ID:     id,
		Method: MethodETHSendRawTransaction,
		Params: []any{rawTx},
		Result: result,
	}
}

// ETHGetBlockByNumber builds the eth_getBlockByNumber element.
func ETHGetBlockByNumber(id string, block string, fullTx bool, result *map[string]any) Element {
	return Element{
		ID:     id,
		Method: MethodETHGetBlockByNumber,
		Params: []any{block, fullTx},
		Result: result,
	}
}

// ETHGetBlockByHash builds the eth_getBlockByHash element.
func ETHGetBlockByHash(id string, blockHash string, fullTx bool, result *map[string]any) Element {
	return Element{
		ID:     id,
		Method: MethodETHGetBlockByHash,
		Params: []any{blockHash, fullTx},
		Result: result,
	}
}

// ETHGetTransactionByHash builds the eth_getTransactionByHash element.
func ETHGetTransactionByHash(id string, txHash string, result *map[string]any) Element {
	return Element{
		ID:     id,
		Method: MethodETHGetTransactionByHash,
		Params: []any{txHash},
		Result: result,
	}
}

// ETHGetTransactionByBlockHashAndIndex builds the eth_getTransactionByBlockHashAndIndex element.
func ETHGetTransactionByBlockHashAndIndex(id string, blockHash, index string, result *map[string]any) Element {
	return Element{
		ID:     id,
		Method: MethodETHGetTransactionByBlockHashAndIndex,
		Params: []any{blockHash, index},
		Result: result,
	}
}

// ETHGetTransactionByBlockNumberAndIndex builds the eth_getTransactionByBlockNumberAndIndex element.
func ETHGetTransactionByBlockNumberAndIndex(id string, block, index string, result *map[string]any) Element {
	return Element{
		ID:     id,
		Method: MethodETHGetTransactionByBlockNumberAndIndex,
		Params: []any{block, index},
		Result: result,
	}
}

// ETHGetTransactionReceipt builds the eth_getTransactionReceipt element.
func ETHGetTransactionReceipt(id string, txHash string, result *map[string]any) Element {
	return Element{
		ID:     id,
		Method: MethodETHGetTransactionReceipt,
		Params: []any{txHash},
		Result: result,
	}
}

// ETHGetBlockReceipts builds the eth_getBlockReceipts element.
func ETHGetBlockReceipts(id string, block string, result *[]map[string]any) Element {
	return Element{
		ID:     id,
		Method: MethodETHGetBlockReceipts,
		Params: []any{block},
		Result: result,
	}
}

// ETHGetBlockTransactionCountByHash builds the eth_getBlockTransactionCountByHash element.
func ETHGetBlockTransactionCountByHash(id string, blockHash string, result *string) Element {
	return Element{
		ID:     id,
		Method: MethodETHGetBlockTransactionCountByHash,
		Params: []any{blockHash},
		Result: result,
	}
}

// ETHGetBlockTransactionCountByNumber builds the eth_getBlockTransactionCountByNumber element.
func ETHGetBlockTransactionCountByNumber(id string, block string, result *string) Element {
	return Element{
		ID:     id,
		Method: MethodETHGetBlockTransactionCountByNumber,
		Params: []any{block},
		Result: result,
	}
}

// ETHGetStorageAt builds the eth_getStorageAt element.
func ETHGetStorageAt(id string, address, slot, block string, result *string) Element {
	return Element{
		ID:     id,
		Method: MethodETHGetStorageAt,
		Params: []any{address, slot, block},
		Result: result,
	}
}

// ETHGetLogs builds the eth_getLogs element. filter is the JSON-RPC filter object.
func ETHGetLogs(id string, filter any, result *[]map[string]any) Element {
	return Element{
		ID:     id,
		Method: MethodETHGetLogs,
		Params: []any{filter},
		Result: result,
	}
}

// ETHFeeHistory builds the eth_feeHistory element.
func ETHFeeHistory(id string, blockCount string, newestBlock string, rewardPercentiles []float64, result *map[string]any) Element {
	return Element{
		ID:     id,
		Method: MethodETHFeeHistory,
		Params: []any{blockCount, newestBlock, rewardPercentiles},
		Result: result,
	}
}

// ETHCreateAccessList builds the eth_createAccessList element.
func ETHCreateAccessList(id string, params any, block string, result *map[string]any) Element {
	return Element{
		ID:     id,
		Method: MethodETHCreateAccessList,
		Params: []any{params, block},
		Result: result,
	}
}

// ETHSyncing builds the eth_syncing element. result decodes to false or a sync-status object.
func ETHSyncing(id string, result *any) Element {
	return Element{
		ID:     id,
		Method: MethodETHSyncing,
		Result: result,
	}
}

// NetVersion builds the net_version element.
func NetVersion(id string, result *string) Element {
	return Element{
		ID:     id,
		Method: MethodNetVersion,
		Result: result,
	}
}

// Web3ClientVersion builds the web3_clientVersion element.
func Web3ClientVersion(id string, result *string) Element {
	return Element{
		ID:     id,
		Method: MethodWeb3ClientVersion,
		Result: result,
	}
}

// TxpoolStatus builds the txpool_status element.
func TxpoolStatus(id string, result *map[string]any) Element {
	return Element{
		ID:     id,
		Method: MethodTxpoolStatus,
		Result: result,
	}
}

// TxpoolContent builds the txpool_content element.
func TxpoolContent(id string, result *map[string]any) Element {
	return Element{
		ID:     id,
		Method: MethodTxpoolContent,
		Result: result,
	}
}

// TxpoolContentFrom builds the txpool_contentFrom element.
func TxpoolContentFrom(id string, address string, result *map[string]any) Element {
	return Element{
		ID:     id,
		Method: MethodTxpoolContentFrom,
		Params: []any{address},
		Result: result,
	}
}

// TxpoolInspect builds the txpool_inspect element.
func TxpoolInspect(id string, result *map[string]any) Element {
	return Element{
		ID:     id,
		Method: MethodTxpoolInspect,
		Result: result,
	}
}

// DebugTraceTransaction builds the debug_traceTransaction element. traceConfig may be nil.
func DebugTraceTransaction(id string, txHash string, traceConfig any, result *any) Element {
	return Element{
		ID:     id,
		Method: MethodDebugTraceTransaction,
		Params: []any{txHash, traceConfig},
		Result: result,
	}
}

// DebugTraceCall builds the debug_traceCall element. traceConfig may be nil.
func DebugTraceCall(id string, params any, block string, traceConfig any, result *any) Element {
	return Element{
		ID:     id,
		Method: MethodDebugTraceCall,
		Params: []any{params, block, traceConfig},
		Result: result,
	}
}

// DebugTraceBlockByNumber builds the debug_traceBlockByNumber element. traceConfig may be nil.
func DebugTraceBlockByNumber(id string, block string, traceConfig any, result *[]map[string]any) Element {
	return Element{
		ID:     id,
		Method: MethodDebugTraceBlockByNumber,
		Params: []any{block, traceConfig},
		Result: result,
	}
}

// DebugTraceBlockByHash builds the debug_traceBlockByHash element. traceConfig may be nil.
func DebugTraceBlockByHash(id string, blockHash string, traceConfig any, result *[]map[string]any) Element {
	return Element{
		ID:     id,
		Method: MethodDebugTraceBlockByHash,
		Params: []any{blockHash, traceConfig},
		Result: result,
	}
}

// DebugGetRawTransaction builds the debug_getRawTransaction element.
func DebugGetRawTransaction(id string, txHash string, result *string) Element {
	return Element{
		ID:     id,
		Method: MethodDebugGetRawTransaction,
		Params: []any{txHash},
		Result: result,
	}
}

// DebugGetRawReceipts builds the debug_getRawReceipts element.
func DebugGetRawReceipts(id string, block string, result *[]string) Element {
	return Element{
		ID:     id,
		Method: MethodDebugGetRawReceipts,
		Params: []any{block},
		Result: result,
	}
}

// DebugGetRawHeader builds the debug_getRawHeader element.
func DebugGetRawHeader(id string, block string, result *string) Element {
	return Element{
		ID:     id,
		Method: MethodDebugGetRawHeader,
		Params: []any{block},
		Result: result,
	}
}

// DebugGetRawBlock builds the debug_getRawBlock element.
func DebugGetRawBlock(id string, block string, result *string) Element {
	return Element{
		ID:     id,
		Method: MethodDebugGetRawBlock,
		Params: []any{block},
		Result: result,
	}
}

// DebugGetBadBlocks builds the debug_getBadBlocks element.
func DebugGetBadBlocks(id string, result *[]map[string]any) Element {
	return Element{
		ID:     id,
		Method: MethodDebugGetBadBlocks,
		Result: result,
	}
}

// DebugStorageRangeAt builds the debug_storageRangeAt element.
func DebugStorageRangeAt(id string, blockHash string, txIndex int, contractAddress string, keyStart string, maxResult int, result *map[string]any) Element {
	return Element{
		ID:     id,
		Method: MethodDebugStorageRangeAt,
		Params: []any{blockHash, txIndex, contractAddress, keyStart, maxResult},
		Result: result,
	}
}

// TraceTransaction builds the trace_transaction element.
func TraceTransaction(id string, txHash string, result *[]map[string]any) Element {
	return Element{
		ID:     id,
		Method: MethodTraceTransaction,
		Params: []any{txHash},
		Result: result,
	}
}

// TraceBlock builds the trace_block element.
func TraceBlock(id string, block string, result *[]map[string]any) Element {
	return Element{
		ID:     id,
		Method: MethodTraceBlock,
		Params: []any{block},
		Result: result,
	}
}

// TraceFilter builds the trace_filter element. filter is the trace filter object.
func TraceFilter(id string, filter any, result *[]map[string]any) Element {
	return Element{
		ID:     id,
		Method: MethodTraceFilter,
		Params: []any{filter},
		Result: result,
	}
}

// TraceCall builds the trace_call element.
func TraceCall(id string, params any, traceTypes []string, block string, result *map[string]any) Element {
	return Element{
		ID:     id,
		Method: MethodTraceCall,
		Params: []any{params, traceTypes, block},
		Result: result,
	}
}

// TraceReplayTransaction builds the trace_replayTransaction element.
func TraceReplayTransaction(id string, txHash string, traceTypes []string, result *map[string]any) Element {
	return Element{
		ID:     id,
		Method: MethodTraceReplayTransaction,
		Params: []any{txHash, traceTypes},
		Result: result,
	}
}

// TraceReplayBlockTransactions builds the trace_replayBlockTransactions element.
func TraceReplayBlockTransactions(id string, block string, traceTypes []string, result *[]map[string]any) Element {
	return Element{
		ID:     id,
		Method: MethodTraceReplayBlockTransactions,
		Params: []any{block, traceTypes},
		Result: result,
	}
}
