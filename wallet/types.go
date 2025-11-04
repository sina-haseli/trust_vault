package wallet

// #cgo CFLAGS: -I${SRCDIR}/../../third_party/wallet-core/include -I/usr/local/include
// #cgo LDFLAGS: -L/usr/local/lib -lTrustWalletCore -lwallet_core_rs -lTrezorCrypto -lprotobuf -lstdc++ -lm -lpthread
// #include <TrustWalletCore/TWCoinType.h>
// #include <TrustWalletCore/TWPrivateKey.h>
// #include <TrustWalletCore/TWPublicKey.h>
import "C"

// Type aliases for C types to help with CGo
type (
	TWCoinType   = C.enum_TWCoinType
	TWPrivateKey = C.struct_TWPrivateKey
	TWPublicKey  = C.struct_TWPublicKey
)
