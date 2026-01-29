package android

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"io"
	"path"
	"sort"

	"golang.org/x/exp/maps"

	"internal/core"
)

type APK struct {
	zipReader     *zip.Reader
	coreFiles     map[string]*zip.File
	checkSumItems map[string][]byte
	versionCode   int64
	isAAB         bool
}

// file-extension used to APK.generate
var checkSumExt = map[string]bool{
	`.dex`:  true,
	`.so`:   true,
	`.arsc`: true,
}

type Option func(apk *APK) error

func WithVersionCode(versionCode int64) Option {
	return func(apk *APK) error {
		apk.versionCode = versionCode
		return nil
	}
}

func WithAAB(isAAB bool) Option {
	return func(apk *APK) error {
		apk.isAAB = isAAB
		return nil
	}
}

func NewRelease(apkBytes []byte, options ...Option) (core.Release, error) {
	apkBytesReader := bytes.NewReader(apkBytes)

	apk := &APK{
		coreFiles:     make(map[string]*zip.File),
		checkSumItems: make(map[string][]byte),
	}

	err := apk.SetReader(apkBytesReader, int64(len(apkBytes)))
	if err != nil {
		return nil, err
	}
	for _, option := range options {
		err = option(apk)
		if err != nil {
			return nil, err
		}
	}
	apk.calculateCheckSum()

	return apk, nil
}

func (apk *APK) SetReader(reader io.ReaderAt, size int64) error {
	var err error
	apk.zipReader, err = zip.NewReader(reader, size)
	if err != nil {
		return err
	}
	for _, file := range apk.zipReader.File {
		ext := path.Ext(file.Name)
		if isEnabled, isExists := checkSumExt[ext]; !isExists || !isEnabled {
			continue
		}
		apk.coreFiles[file.Name] = file
	}
	return nil
}

func (apk *APK) CheckSums() map[string][]byte {
	return apk.checkSumItems
}

func (apk *APK) VersionCode() int64 {
	return apk.versionCode
}

// zip 自带 crc 校验和，且 crc32 计算比 md5 快
// 因此，这里直接使用 core files 的 crc32 checksum 合成 apk 的 crc32
func (apk *APK) calculateCheckSum() {
	if len(apk.coreFiles) == 0 {
		return
	}
	coreFilenames := maps.Keys(apk.coreFiles)
	sort.Strings(coreFilenames)

	// combine crc32
	for _, filename := range coreFilenames {
		crcBytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(crcBytes, apk.coreFiles[filename].CRC32)
		if apk.isAAB {
			if filename[0:9] == `base/dex/` {
				filename = filename[9:]
			} else {
				filename = filename[5:]
			}
		}
		apk.checkSumItems[filename] = crcBytes
	}
}
