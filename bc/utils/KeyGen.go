/*
 * @author DooQY
 * @createDate 2019/11/8 - 下午1:25
 */

package utils

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"github.com/Nik-U/pbc"
	"io/ioutil"
	"math/big"
)

const str = `type a
	q 8780710799663312522437781984754049815806883199414208211028653399266475630880222957078625179422662221423155858769582317459277713367317481324925129998224791
	h 12016012264891146079388821366740534204802954401251311822919615131047207289359704531102844802183906537786776
	r 730750818665451621361119245571504901405976559617
	exp2 159
	exp1 107
	sign1 1
	sign0 1`

var pairing, _ = pbc.NewPairingFromString(str)
var StaticG = []byte{48, 157, 94, 118, 198, 28, 66, 139, 61, 41, 240, 150, 225, 195, 202, 153, 238, 172, 139, 131, 157, 199, 62, 150, 165, 113, 115, 229, 20, 235, 118, 73, 87, 16, 152, 177, 55, 1, 5, 103, 235, 108, 93, 28, 42, 106, 89, 57, 22, 100, 148, 143, 175, 94, 73, 83, 37, 80, 78, 254, 133, 47, 151, 178}
var StaticG1 = pairing.NewG1().SetXBytes(StaticG)

func GenerateZru() (UserID, Z, pow *pbc.Element) {
	UserID = pairing.NewZr().Rand()
	Z = pairing.NewZr().Rand()
	r := pairing.NewZr().Rand()
	r2 := pairing.NewZr().Add(r, UserID)
	pow = pairing.NewZr().Div(Z, r2)
	return
}

func GenerateABC(Z, pow, g, g1 *pbc.Element) (Ak, Bk, Ck *pbc.Element) {

	Ak = pairing.NewG1().PowZn(g, Z)
	Bk = pairing.NewG1().PowZn(g1, pow)
	Ck = pairing.NewG1().PowZn(g, pow)
	return
}

func HashAttr(M map[string]string, pow *pbc.Element) []*pbc.Element {

	ret := make([]*pbc.Element, len(M))
	index := 0
	for _, attribute := range M {
		ret[index] = H(attribute)
		ret[index].PowZn(ret[index], pow)
		index++
	}
	return ret
}

func GetUserInfo(ch chan map[string]string) {

	userinfo := make(map[string]string)
	file, err := ioutil.ReadFile("/home/forsim/go/LSH/bc/UserInfo.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(file, &userinfo)
	if err != nil {
		panic(err)
	}
	ch <- userinfo
}

func H(M string) *pbc.Element {

	Ms := sha256.Sum256([]byte(M))
	sl := Ms[0:4]
	si := binary.BigEndian.Uint32(sl)
	sr := pairing.NewZr().SetBig(big.NewInt(int64(si)))
	h := pairing.NewG1().MulZn(StaticG1, sr)

	return h
}

func H1(ID string) *pbc.Element {

	s := sha256.Sum256([]byte(ID))
	sl := s[0:4]
	si := binary.BigEndian.Uint32(sl)
	sr := pairing.NewZr().SetBig(big.NewInt(int64(si)))

	return sr
}

func H0(element *pbc.Element) [32]byte {
	return sha256.Sum256(element.Bytes())
}
