package sm4

// #cgo LDFLAGS: -L../../librarys -lsmcryptokit -lcrypto
// #include "../../include/sm/smcryptokit.h"
import "C"
import "unsafe"

import (
	// "crypto/cipher"
	"fmt"
	// "strconv"

	"bytes"
	// "crypto/aes"
	"crypto/rand"
	"errors"
	"io"
)

// The SM4 block size in bytes.
const (
	BlockSize   = 16
	SM4_ENCRYPT = 1
	SM4_DECRYPT = 0

	// SM4KeyLength is the default SM4 key length
	SM4KeyLength = 16
)

// A cipher is an instance of AES encryption using a particular key.
type sm4Cipher struct {
	enc [32]uint32
	dec [32]uint32
}

// GetRandomBytes returns len random looking bytes
func GetRandomBytes(len int) ([]byte, error) {
	key := make([]byte, len)

	// TODO: rand could fill less bytes then len
	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}

	return key, nil
}

// NewCipher creates and returns a new cipher.Block.
// The key argument should be the SM4 key,
// 16 bytes to select SM4
func NewCipher(key []byte) (*sm4Cipher, error) {
	if len(key) != 16 {
		return nil, fmt.Errorf("NewCipher invalid key length %v", len(key))
	}

	block := &sm4Cipher{}
	C.sm4_setkey_enc(unsafe.Pointer(&block.enc[0]), unsafe.Pointer(&key[0]), 16)
	C.sm4_setkey_dec(unsafe.Pointer(&block.dec[0]), unsafe.Pointer(&key[0]), 16)
	return block, nil
}

func (c *sm4Cipher) BlockSize() int { return BlockSize }

func (c *sm4Cipher) Encrypt(dst, src []byte) {}

// 	if len(src) < BlockSize {
// 		panic("sm4Cipher Encrypt: input not full block")
// 	}
// 	if len(dst) < BlockSize {
// 		panic("sm4Cipher Encrypt: output not full block")
// 	}
// 	// C.sm4_crypt_cbc(unsafe.Pointer(&block.enc[0]), SM4_ENCRYPT, C.int(len(src)),
// 	// 				unsigned char iv[16],
// 	//                    unsigned char *input, unsigned char *output);
// 	C.sm4_crypt_cbc(unsafe.Pointer(&block.enc[0]), SM4_ENCRYPT, C.int(len(src)),
// 		unsafe.Pointer(&iv[0]), unsafe.Pointer(&s[0]), unsafe.Pointer(&ciphertext[BlockSize]))
// }

func (c *sm4Cipher) Decrypt(dst, src []byte) {}

// 	if len(src) < BlockSize {
// 		panic("sm4Cipher Decrypt: input not full block")
// 	}
// 	if len(dst) < BlockSize {
// 		panic("sm4Cipher Decrypt: output not full block")
// 	}
// 	decryptBlockGo(c.dec, dst, src)
// }

// GenSM4Key returns a random SM4 key of length SM4KeyLength
func GenSM4Key() ([]byte, error) {
	return GetRandomBytes(SM4KeyLength)
}

// PKCS7Padding pads as prescribed by the PKCS7 standard
func PKCS7Padding(src []byte) []byte {
	padding := BlockSize - len(src)%BlockSize
	// fmt.Printf("padding:%v\n", padding)
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padtext...)
}

// PKCS7UnPadding unpads as prescribed by the PKCS7 standard
func PKCS7UnPadding(src []byte) ([]byte, error) {
	length := len(src)
	unpadding := int(src[length-1])

	//fmt.Printf("unpadding:%v ---\n", unpadding)

	if unpadding > BlockSize || unpadding == 0 {
		return nil, fmt.Errorf("invalid padding")
	}

	//fmt.Println("unpading....",unpadding)
	pad := src[len(src)-unpadding:]
	for i := 0; i < unpadding; i++ {
		if pad[i] != byte(unpadding) {
			return nil, fmt.Errorf("invalid padding")
		}
	}

	return src[:(length - unpadding)], nil
}

// CBCEncrypt encrypts using CBC mode
func CBCEncrypt(key, s []byte) ([]byte, error) {
	// CBC mode works on blocks so plaintexts may need to be padded to the
	// next whole block. For an example of such padding, see
	// https://tools.ietf.org/html/rfc5246#section-6.2.3.2. Here we'll
	// assume that the plaintext is already of the correct length.
	if len(s)%BlockSize != 0 {
		return nil, errors.New("plaintext is not a multiple of the block size")
	}

	block, err := NewCipher(key)
	if err != nil {
		return nil, err
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	ciphertext := make([]byte, BlockSize+len(s))
	iv := ciphertext[:BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	// mode := cipher.NewCBCEncrypter(block, iv)
	// mode.CryptBlocks(ciphertext[aes.BlockSize:], s)

	C.sm4_crypt_cbc(unsafe.Pointer(&block.enc[0]), SM4_ENCRYPT, C.int(len(s)),
		unsafe.Pointer(&iv[0]), unsafe.Pointer(&s[0]), unsafe.Pointer(&ciphertext[BlockSize]))

	// It's important to remember that ciphertexts must be authenticated
	// (i.e. by using crypto/hmac) as well as being encrypted in order to
	// be secure.
	return ciphertext, nil
}

// CBCDecrypt decrypts using CBC mode
func CBCDecrypt(key, src []byte) ([]byte, error) {
	block, err := NewCipher(key)
	if err != nil {
		return nil, err
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	if len(src) < BlockSize {
		return nil, errors.New("ciphertext too short")
	}
	iv := src[:BlockSize]
	src = src[BlockSize:]

	// CBC mode always works in whole blocks.
	if len(src)%BlockSize != 0 {
		return nil, errors.New("ciphertext is not a multiple of the block size")
	}

	dest := make([]byte, len(src), len(src))

	C.sm4_crypt_cbc(unsafe.Pointer(&block.dec[0]), SM4_DECRYPT, C.int(len(src)),
		unsafe.Pointer(&iv[0]), unsafe.Pointer(&src[0]), unsafe.Pointer(&dest[0]))

	return dest, nil
}

// EBCEncrypt encrypts using EBC mode
func EBCEncrypt(key, s []byte) ([]byte, error) {
	// CBC mode works on blocks so plaintexts may need to be padded to the
	// next whole block. For an example of such padding, see
	// https://tools.ietf.org/html/rfc5246#section-6.2.3.2. Here we'll
	// assume that the plaintext is already of the correct length.
	if len(s)%BlockSize != 0 {
		return nil, errors.New("plaintext is not a multiple of the block size")
	}

	block, err := NewCipher(key)
	if err != nil {
		return nil, err
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	ciphertext := make([]byte, len(s))

	// mode := cipher.NewCBCEncrypter(block, iv)
	// mode.CryptBlocks(ciphertext[aes.BlockSize:], s)

	C.sm4_crypt_ecb(unsafe.Pointer(&block.enc[0]), SM4_ENCRYPT, C.int(len(s)),
		unsafe.Pointer(&s[0]), unsafe.Pointer(&ciphertext[0]))

	// It's important to remember that ciphertexts must be authenticated
	// (i.e. by using crypto/hmac) as well as being encrypted in order to
	// be secure.
	return ciphertext, nil
}

// EBCDecrypt decrypts using EBC mode
func EBCDecrypt(key, src []byte) ([]byte, error) {
	block, err := NewCipher(key)
	if err != nil {
		return nil, err
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	if len(src) < BlockSize {
		return nil, errors.New("ciphertext too short")
	}

	// CBC mode always works in whole blocks.
	if len(src)%BlockSize != 0 {
		return nil, errors.New("ciphertext is not a multiple of the block size")
	}

	dest := make([]byte, len(src))

	C.sm4_crypt_ecb(unsafe.Pointer(&block.dec[0]), SM4_DECRYPT, C.int(len(src)),
		unsafe.Pointer(&src[0]), unsafe.Pointer(&dest[0]))

	return dest, nil
}

// CBCPKCS7Encrypt combines CBC encryption and PKCS7 padding
func CBCPKCS7Encrypt(key, src []byte) ([]byte, error) {
	tmp := PKCS7Padding(src)
	// fmt.Printf("tmp:%v\n", tmp)
	return CBCEncrypt(key, tmp)
}

// CBCPKCS7Decrypt combines CBC decryption and PKCS7 unpadding
func CBCPKCS7Decrypt(key, src []byte) ([]byte, error) {
	pt, err := CBCDecrypt(key, src)
	if err != nil {
		return nil, err
	}

	// fmt.Printf("pt:%v\n", pt)

	original, err := PKCS7UnPadding(pt)
	if err != nil {
		return nil, err
	}

	return original, nil
}

// EBCPKCS7Encrypt combines CBC encryption and PKCS7 padding
func EBCPKCS7Encrypt(key, src []byte) ([]byte, error) {
	tmp := PKCS7Padding(src)
	// fmt.Printf("tmp:%v\n", tmp)
	return EBCEncrypt(key, tmp)
}

// EBCPKCS7Decrypt combines CBC decryption and PKCS7 unpadding
func EBCPKCS7Decrypt(key, src []byte) ([]byte, error) {
	pt, err := EBCDecrypt(key, src)
	if err != nil {
		return nil, err
	}

	// fmt.Printf("pt:%v\n", pt)

	original, err := PKCS7UnPadding(pt)
	if err != nil {
		return nil, err
	}

	return original, nil
}

func Encrypt(key, src []byte) ([]byte, error) {
	return EBCPKCS7Encrypt(key, src)
}

func Decrypt(key, src []byte) ([]byte, error) {
	return EBCPKCS7Decrypt(key, src)
}
