package wallet

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
