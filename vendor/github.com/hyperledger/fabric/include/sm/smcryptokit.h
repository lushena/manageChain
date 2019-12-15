#include <stdio.h>
#include <string.h>
#include <syslog.h>
#include <openssl/des.h>
#include <openssl/ec.h>
#include <openssl/ecdsa.h>
#include <openssl/evp.h>
#include <openssl/bn.h>
#include <openssl/bio.h>
#include <openssl/obj_mac.h>
#include <openssl/pem.h>
#include <openssl/err.h>
#include <openssl/rand.h>

#include "sm2/sm2.h"
#include "sm3/sm3.h"
#include "sm4/sms4.h"
#ifndef HEADER_SMTEST_H
#define HEADER_SMTEST_H

#ifdef  __cplusplus
extern "C" {
#endif

#if __GNUC__
#define EXPORTDLL __attribute__ ((visibility ("default")))
#else
#define EXPORTDLL _declspec(dllexport)
#endif

/**
 * Creates a new EC_KEY object using NID_prime256v1SM2TEST curve.
 * [RET]EC_KEY*: return object or NULL if an error occurred. 
 */
EXPORTDLL EC_KEY* SM2NewEcKey();

/**
 * Frees a EC_KEY object..
 * [IN] ecKey:   EC_KEY object to be freed.
 * [RET]void.
 */
EXPORTDLL void SM2FreeEcKey(EC_KEY * ecKey);

/**
 * Load private key file from specified path.
 * [IN] path:    the private key file path.
 * [IN] len:     lenght of path
 * [RET]EC_KEY*: return object or NULL if an error occurred. 
 */
EXPORTDLL EC_KEY* LoadSM2PrivKeyFromFile(void* path, int len);

/**
 * Load private key from specified memory.
 * [IN] keybytes:pointer to the private key to load.
 * [IN] len:     lenght of private key
 * [RET]EC_KEY*: return object or NULL if an error occurred. 
 */
EXPORTDLL EC_KEY* LoadSM2PrivKeyFromBytes(void* keybytes,int len);

/**
 * Frees a X509 object..
 * [IN] x:   X509 object to be freed.
 * [RET]void.
 */
EXPORTDLL void SM2FreeX509(X509* x);

/**
 * Load X509 certificate file from specified path.
 * [IN]  path:   the certificate file path.
 * [IN] len:     lenght of path
 * [RET]X509*:   return object or NULL if an error occurred.
 */
EXPORTDLL X509* LoadSM2CertFromFile(void* path, int len);

/**
 * Load public key from specified memory.
 * [IN] keybytes:pointer to the public key to load.
 * [IN] len:     lenght of public key
 * [RET]EC_KEY*:   return object or NULL if an error occurred. 
 */
EXPORTDLL EC_KEY* LoadSM2PubKeyFromBytes(void* keybytes,int len);

/**
 * Computes ECDSA signature of a given hash value using the supplied private key
 * [IN] type:    this parameter is ignored
 * [IN] dgst:    pointer to the hash value to sign
 * [IN] dLen:    length of the hash value
 * [OUT] sig:    buffer to hold the DER encoded signature. sig must point to ECDSA_size(eckey) bytes of memory.
 * [OUT]sigLen:  pointer to the length of the returned signature
 * [IN]eckey:    EC_KEY object containing a private EC key
 * [RET] int:    return 1 on success and 0 otherwise
 */
EXPORTDLL int SM2Sign(int type, void* dgst, int dLen, void* sig,
                      unsigned int* sigLen, void * eckey);

/**
 * Computes ECDSA signature of a given hash value using the supplied private key
 * [IN] type:    this parameter is ignored
 * [IN] dgst:    pointer to the hash value to sign
 * [IN] dLen:    length of the hash value
 * [OUT] r:      buffer to hold the DER encoded signature->r. sig must point to 32 bytes of memory
 * [OUT] rLen:   pointer to the length of the returned signature->r
 * [OUT] s:      buffer to hold the DER encoded signature->s. sig must point to 32 bytes of memory
 * [OUT] sLen:   pointer to the length of the returned signature->s
 * [IN] eckey:   EC_KEY object containing a private EC key
 * [RET] int:    return 1 on success and 0 otherwise
 */
EXPORTDLL int SM2SignDirect(int type, void* dgst, int dLen, void* r, 
                            int* rLen, void* s, int * sLen, void* eckey);

/**
 * Verifies that the given signature is valid ECDSA signature of the supplied hash value using the specified public key.
 * [IN] type:    this parameter is ignored
 * [IN] dgst:    pointer to the hash value
 * [IN] dLen:    length of the hash value
 * [IN] sig:     pointer to the DER encoded signature
 * [IN] sigLen:  length of the DER encoded signature
 * [IN] eckey:   EC_KEY object containing a public EC key
 * [RET] int:    return 1 if the signature is valid and 0 otherwise
 */
EXPORTDLL int SM2Verify(int type, void* dgst, int dLen, void* sig, int sigLen, void* eckey);

/**
 * Verifies that the given signature is valid ECDSA signature of the supplied hash value using the specified public key.
 * [IN] type:    this parameter is ignored
 * [IN] dgst:    pointer to the hash value
 * [IN] dLen:    length of the hash value
 * [IN] r:       pointer to the DER encoded signature->r
 * [IN] rLen:    length of the returned signature->r
 * [IN] s:       pointer to the DER encoded signature->s
 * [IN] sLen:    length of the returned signature->s
 * [IN] eckey:   EC_KEY object containing a public EC key
 * [RET] int:    return 1 if the signature is valid and 0 otherwise
 */
EXPORTDLL int SM2VerifyDirect(int type, void* dgst, int dLen, void* r, int rLen, void* s, int sLen, void* eckey);

/**
 * Initialize SM3
 * [OUT] total:  total is an array of 2 integers to be initialized
 * [IN] totalLen:2
 * [OUT] state:  state is an array of 8 integers to be initialized
 * [IN] stateLen:8
 * [RET] void
 */
EXPORTDLL void Sm3Starts(void * total, int totalLen, void * state, int stateLen);

/**
 * Update sm3 hash
 * [IN] total:   total is an array of 2 integers has been initialized
 * [IN] totalLen:2
 * [IN] state:   state is an array of 8 integers has been initialized
 * [IN] stateLen:8
 * [OUT] buffer: state is an array of 64 unsigned char
 * [IN]bufferLen:64
 * [IN] input:   pointer to the data to be added to the hash object
 * [IN] iLen:    length of the input buffer
 * [RET] void
 */
EXPORTDLL void Sm3Update(void *total, int totalLen, void *state, int stateLen,
                         void *buffer, int bufferLen, void *input, int iLen);

/**
 * Finish sm3 hash
 * [IN] total:   total is an array of 2 integers has been initialized
 * [IN] totalLen:2
 * [IN] state:   state is an array of 8 integers has been initialized
 * [IN] stateLen:8
 * [OUT] buffer: state is an array of 64 unsigned char
 * [IN]bufferLen:64
 * [IN] output:  pointer to hash result.output must point to 32 bytes of memory.
 * [IN] oLen:    32
 */
EXPORTDLL void Sm3Finish(void *total, int totalLen, void *state, int stateLen,
                         void *buffer, int bufferLen, void *output, int oLen);


/**
 * Generate encrypt round key using the specified encrypt key
 * [OUT] encKey: buffer to hold the round key. encKey must point to 128 bytes of memory
 * [IN] key:     pointer to encrypt key. key size is 16 bytes
 * [IN] keyLen:  16
 * [RET] void
 */
EXPORTDLL void sm4_setkey_enc(void *encKey, void *key, int keyLen);

/**
 * Generate decrypt round key using the specified decrypt key
 * [OUT] decKey: buffer to hold the round key. decKey must point to 128 bytes of memory
 * [IN] key:     pointer to encrypt key. key size is 16 bytes
 * [IN] keyLen:  16
 * [RET] void
 */
EXPORTDLL void sm4_setkey_dec(void *decKey, void *key, int keyLen);

/**
 * SM4 ecb mode encrypt/decrypt.
 * [IN] seckey:  pointer to hold the round key
 * [IN] mode:    1 encrypt or 0 decrypt
 * [IN] length:  length of the input buffer
 * [IN] in:      pointer to the data to be encrypted/decrypted
 * [OUT] out:    buffer to hold the out data
 * [RET] void
 */
EXPORTDLL void sm4_crypt_ecb(void *seckey, int mode, int length, void *in, void *out);

/**
 * SM4 cbc mode encrypt/decrypt.
 * [IN] seckey:  pointer to hold the round key
 * [IN] mode:    1 encrypt or 0 decrypt
 * [IN] length:  length of the input buffer
 * [IN] initVec: pointer to the initialize vector
 * [IN] in:      pointer to the data to be encrypted/decrypted
 * [OUT] out:    buffer to hold the out data
 * [RET] int:    return 1 on success and 0 otherwise
 */
EXPORTDLL void sm4_crypt_cbc(void *seckey, int mode, int length, void *initVec, void *in, void *out);

#ifdef  __cplusplus
}
#endif

#endif
