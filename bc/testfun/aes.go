/*
 * @author DooQY
 * @createDate 2019/10/28 - 下午3:37
 */

package main

import (
	"crypto/aes"
	"fmt"
)

var key = []byte{69, 77, 3, 64, 110, 242, 213, 30, 169, 20, 182, 208, 141, 69, 190, 133, 89, 218, 240, 14, 117, 20, 233, 193, 12, 73, 224, 25, 140, 59, 192, 52}

func main() {
	enc := AESEnc([]byte("富强民主文明和谐自由平等公正法治爱国敬业诚信友善"))
	dec := AESDec(enc)
	fmt.Println(string(dec))
}

func AESEnc(input []byte) []byte {

	loop := len(input) / 16
	AesCipher, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	var result []byte
	part := make([]byte, 16)

	for i := 0; i <= loop; i++ {

		block := make([]byte, 16)
		copy(block, input[i*16:(i+1)*16])

		AesCipher.Encrypt(part, block)
		result = append(result, part...)
	}

	return result
}

func AESDec(input []byte) []byte {

	loop := len(input) / 16
	AesCipher, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	var result []byte
	part := make([]byte, 16)
	block := make([]byte, 16)

	for i := 0; i < loop; i++ {
		copy(block, input[i*16:(i+1)*16])

		AesCipher.Decrypt(part, block)
		result = append(result, part...)
	}

	return result
}
