package ios

import (
	"archive/zip"
	"errors"
	"io"
	"os"
	"os/exec"
	"path"
	"runtime"

	"internal/core"
)

type IPA struct {
	zipReader      *zip.Reader
	uuid           string
	binaryFileName string
	checkSumItem   map[string][]byte
}

func (ipa *IPA) CheckSums() map[string][]byte {
	return ipa.checkSumItem
}

func NewIPA(reader io.ReaderAt, size int64, binaryFileName string) (core.Release, error) {

	ipa := &IPA{
		checkSumItem: map[string][]byte{},
	}

	err := ipa.loadZipReader(reader, size)
	if err != nil {
		return nil, err
	}
	err = ipa.loadUUID(binaryFileName)
	if err != nil {
		return nil, err
	}
	return ipa, nil
}

func (ipa *IPA) loadZipReader(reader io.ReaderAt, size int64) error {
	var err error
	ipa.zipReader, err = zip.NewReader(reader, size)
	if err != nil {
		return err
	}
	return nil
}

var ErrBinaryFileNotFound = errors.New(`binary_file not found`)
var ErrParseUUIDFail = errors.New(`fail to parse ipa.binary_file.uuid`)

const dwarfdumpUuidOffSet = 6
const uuidLength = 36

// 从 ipa 中解压出指定的可执行文件，并下载到系统临时目录中
// 通过类似 dwarfdump 的工具，从文件中提取 uuid 字段
func (ipa *IPA) loadUUID(binaryFileName string) error {
	if ipa.zipReader == nil {
		return errors.New(`ipa.reader is nil`)
	}

	var binaryFile *zip.File
	var err error
	for _, file := range ipa.zipReader.File {
		if file.Name == binaryFileName {
			binaryFile = file
			break
		}
	}
	if binaryFile == nil {
		return ErrBinaryFileNotFound
	}

	destinationFile, err := os.CreateTemp(``, path.Base(binaryFile.Name))
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	zippedFile, err := binaryFile.Open()
	if err != nil {
		return err
	}
	defer func() {
		zippedFile.Close()
	}()

	_, err = io.Copy(destinationFile, zippedFile)
	if err != nil {
		return err
	}
	defer os.Remove(destinationFile.Name())

	// TODO @funco.lin 需要寻找替代方案
	cmdName := `dwarfdump`
	if runtime.GOOS == `linux` {
		cmdName = `llvm-dwarfdump`
	}
	cmd := exec.Command(cmdName, `--uuid`, destinationFile.Name())
	output, err := cmd.Output()
	if err != nil {
		return err
	}
	if len(output) < dwarfdumpUuidOffSet+uuidLength {
		return ErrParseUUIDFail
	}
	ipa.checkSumItem[binaryFileName] = output[dwarfdumpUuidOffSet : dwarfdumpUuidOffSet+uuidLength]

	return nil
}
