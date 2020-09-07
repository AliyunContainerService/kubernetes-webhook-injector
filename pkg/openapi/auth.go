package openapi

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	log "k8s.io/klog"
	"os"
	"time"
)

const tokenConfigPath = "/var/addon/token-config"
const timeLayout = "2006-01-02T15:04:05Z"

var cachedAkInfo *AKInfo

func cacheExpired() bool {
	//when not initialized
	if cachedAkInfo == nil {
		return true
	}
	//AK never expires
	if cachedAkInfo.SecurityToken == "" {
		return false
	}
	t, err := time.Parse(timeLayout, cachedAkInfo.Expiration)
	if err != nil {
		log.Errorf(err.Error())
		return true
	}
	if t.Before(time.Now()) {
		return true
	} else {
		return false
	}
}

func GetAuthInfo() (*AKInfo, error) {

	if !cacheExpired() {
		return cachedAkInfo, nil
	}

	f, err := os.Open(tokenConfigPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tokenBytes, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	akInfo := LoadFromSts(string(tokenBytes))
	cachedAkInfo = &akInfo

	return &akInfo, nil
}

// following functions are copied from metrics-server
func LoadFromSts(stsToken string) AKInfo {
	akInfo := AKInfo{}
	if err := json.Unmarshal([]byte(stsToken), &akInfo); err != nil {
		log.Fatal(fmt.Sprintf("failed to parse encoded stsToken by error %v\n", err))
	}
	ak, err := Decrypt(akInfo.AccessKeyId, []byte(akInfo.Keyring))
	if err != nil {
		log.Fatal(fmt.Sprintf("failed to decode access key by error %v\n", err))
	}

	sk, err := Decrypt(akInfo.AccessKeySecret, []byte(akInfo.Keyring))
	if err != nil {
		log.Fatal(fmt.Sprintf("failed to decode access secret by error %v\n", err))
	}

	token, err := Decrypt(akInfo.SecurityToken, []byte(akInfo.Keyring))
	if err != nil {
		log.Fatal(fmt.Sprintf("failed to decode security stsToken by error %v\n", err))
	}

	layout := "2006-01-02T15:04:05Z"
	t, err := time.Parse(layout, akInfo.Expiration)
	if err != nil {
		log.Errorf(err.Error())
	}
	if t.Before(time.Now()) {
		log.Fatal(fmt.Sprintf("stsToken had expired"))
	}
	akInfo.AccessKeyId = string(ak)
	akInfo.AccessKeySecret = string(sk)
	akInfo.SecurityToken = string(token)
	return akInfo
}

func Decrypt(s string, keyring []byte) ([]byte, error) {
	cdata, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		log.Errorf("failed to decode base64 string, err: %v", err)
		return nil, err
	}
	block, err := aes.NewCipher(keyring)
	if err != nil {
		log.Errorf("failed to new cipher, err: %v", err)
		return nil, err
	}
	blockSize := block.BlockSize()

	iv := cdata[:blockSize]
	blockMode := cipher.NewCBCDecrypter(block, iv)
	origData := make([]byte, len(cdata)-blockSize)

	blockMode.CryptBlocks(origData, cdata[blockSize:])

	origData = PKCS5UnPadding(origData)
	return origData, nil
}

func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}
