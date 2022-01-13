package util

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/andreburgaud/crypt2go/ecb"
	"github.com/andreburgaud/crypt2go/padding"
	"io"
)

// GetMd5V1 获取字符串md5
func GetMd5V1(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

// GetMd5V2 获取字符串md5
func GetMd5V2(str string) string {
	data := []byte(str)
	has := md5.Sum(data)
	md5str := fmt.Sprintf("%x", has)
	return md5str
}

// GetMd5V3 获取字符串md5
func GetMd5V3(str string) string {
	w := md5.New()
	io.WriteString(w, str)
	md5str := fmt.Sprintf("%x", w.Sum(nil))
	return md5str
}

// AESECBEncrypt encrypt aes加密 ecb blockSize:16
func AESECBEncrypt(srcData, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	mode := ecb.NewECBEncrypter(block)
	padder := padding.NewPkcs7Padding(block.BlockSize())
	srcData, err = padder.Pad(srcData) // padd last block of plaintext if block size less than block cipher size
	if err != nil {
		return nil, err
	}
	ct := make([]byte, len(srcData))
	mode.CryptBlocks(ct, srcData)
	return ct, nil
}

// AESECBDecrypt decrypt aes解密  ecb blockSize:16
func AESECBDecrypt(encryptedData, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	mode := ecb.NewECBDecrypter(block)
	pt := make([]byte, len(encryptedData))
	if len(encryptedData)%mode.BlockSize() != 0 {
		return nil, err
	}
	mode.CryptBlocks(pt, encryptedData)
	padder := padding.NewPkcs7Padding(block.BlockSize())
	pt, err = padder.Unpad(pt) // unpad plaintext after decryption
	if err != nil {
		return nil, err
	}
	return pt, nil
}

// AESCBCDecrypt aes解密  cbc blockSize:16
func AESCBCDecrypt(encryptedData, key, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, iv[:blockSize])
	origData := make([]byte, len(encryptedData))
	blockMode.CryptBlocks(origData, encryptedData)
	padder := padding.NewPkcs7Padding(block.BlockSize())
	origData, err = padder.Unpad(origData) // unpad plaintext after decryption
	if err != nil {
		return nil, err
	}
	return origData, nil
}
