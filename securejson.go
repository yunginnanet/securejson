package securejson

import (
	"errors"
	"encoding/json"
)

type Storage interface {
	Put(key string, value []byte) (error)
	Get(key string) ([]byte, error)
}

type SecureJson struct {
	storageStrategy Storage
}

type Json struct {
	UserName string
	Signature string
	EncryptedData string
	Timestamp string
	PublicKey string
	//TODO: NewPublicKey string
}

func (obj *SecureJson) GenerateJson(user string, passwd string, data string) (outputJson []byte, err error) {
	privKey,_ := obj.hash([]byte(passwd))	

	userData := []byte(user)
	encryptedData,_ := obj.encrypt(data, privKey)
	timeData,_ := obj.getTimestamp()
	pubkeyData,_ := obj.getPubKey(privKey)

	fullHash := obj.genHash(userData, encryptedData, timeData, pubkeyData)
	sigData,_ := obj.sign(fullHash, privKey) 
	
	var jsonMap Json
	jsonMap.UserName = user
	jsonMap.Signature = obj.bytesToString(sigData) 
	jsonMap.EncryptedData = obj.bytesToString(encryptedData) 
	jsonMap.Timestamp = obj.bytesToString(timeData)
	jsonMap.PublicKey = obj.bytesToString(pubkeyData)

	outputJson, err = json.Marshal(jsonMap)
	return
}

func (obj *SecureJson) VerifyJson(jsonBytes []byte) (ok bool, err error) {
	var jsonMap Json
	err = json.Unmarshal(jsonBytes, &jsonMap)
	if err != nil {
		return false, err
	}	

	if !obj.checkTimestampBeforeNow(jsonMap.Timestamp) {
		return false, err
	}
	
	userData := []byte(jsonMap.UserName)
	encryptedData := obj.stringToBytes(jsonMap.EncryptedData)
	timeData := obj.stringToBytes(jsonMap.Timestamp)
	pubkeyData := obj.stringToBytes(jsonMap.PublicKey)
	sigData := obj.stringToBytes(jsonMap.Signature)
	fullHash := obj.genHash(userData, encryptedData, timeData, pubkeyData)

	ok = obj.verify(fullHash, pubkeyData, sigData)
	if ok {
		return
	} else {
		err = errors.New("Signature verify fail")
		return false, err
	}
}

func (obj *SecureJson) PutJson(inputJson []byte) (err error) {
	var ok bool
	if ok, err = obj.VerifyJson(inputJson); !ok || err!=nil {
		return
	}
	outputJson, err := obj.getJsonFromStorage(inputJson)
	if err != nil {
		err = obj.putJsonToStorage(inputJson)
		return
	}
	if ok, err = obj.checkInputOutputJson(inputJson, outputJson); err!=nil || !ok {
		return
	}

	err = obj.putJsonToStorage(inputJson)
	return
}

func (obj *SecureJson) GetJson(inputJson []byte) (outputJson []byte, err error) {
	var ok bool
	if ok, err = obj.VerifyJson(inputJson); !ok || err!=nil {
		return
	}
	outputJson, err = obj.getJsonFromStorage(inputJson)
	if err != nil {
		return
	}
	if ok, err = obj.checkInputOutputJson(inputJson, outputJson); err!=nil || !ok {
		outputJson = []byte{}
		return
	}
	return
}

func New(storageObj Storage) *SecureJson {
	obj := new(SecureJson)
	obj.storageStrategy = storageObj
	return obj
}