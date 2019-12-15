/*
Copyright IBM Corp. 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package util

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/hyperledger/fabric/common/flogging"
)

var logger = flogging.MustGetLogger("kvledger.util")

// CreateDirIfMissing creates a dir for dirPath if not already exists. If the dir is empty it returns true
func CreateDirIfMissing(dirPath string) (bool, error) {
	// if dirPath does not end with a path separator, it leaves out the last segment while creating directories
	if !strings.HasSuffix(dirPath, "/") {
		dirPath = dirPath + "/"
	}
	logger.Debugf("CreateDirIfMissing [%s]", dirPath)
	logDirStatus("Before creating dir", dirPath)
	err := os.MkdirAll(path.Dir(dirPath), 0755)
	if err != nil {
		logger.Debugf("Error while creating dir [%s]", dirPath)
		return false, err
	}
	logDirStatus("After creating dir", dirPath)
	return DirEmpty(dirPath)
}

// DirEmpty returns true if the dir at dirPath is empty
func DirEmpty(dirPath string) (bool, error) {
	f, err := os.Open(dirPath)
	if err != nil {
		logger.Debugf("Error while opening dir [%s]: %s", dirPath, err)
		return false, err
	}
	defer f.Close()

	_, err = f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err
}

// FileExists checks whether the given file exists.
// If the file exists, this method also returns the size of the file.
func FileExists(filePath string) (bool, int64, error) {
	fileInfo, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false, 0, nil
	}
	return true, fileInfo.Size(), err
}

func FileModTime(filePath string) (time.Time, error) {
	if fileInfo, err := os.Stat(filePath); err != nil {
		return time.Now(), err
	} else {
		return fileInfo.ModTime(), nil
	}
}

//WriteFileToPackage writes a file to the tarball
func WriteFileToPackage(localpath string, packagepath string, tw *tar.Writer) error {
	fd, err := os.Open(localpath)
	if err != nil {
		return fmt.Errorf("%s: %s", localpath, err)
	}
	defer fd.Close()

	is := bufio.NewReader(fd)
	return WriteStreamToPackage(is, localpath, packagepath, tw)

}

//WriteStreamToPackage writes bytes (from a file reader) to the tarball
func WriteStreamToPackage(is io.Reader, localpath string, packagepath string, tw *tar.Writer) error {
	info, err := os.Stat(localpath)
	if err != nil {
		return fmt.Errorf("%s: %s", localpath, err)
	}
	header, err := tar.FileInfoHeader(info, localpath)
	if err != nil {
		return fmt.Errorf("Error getting FileInfoHeader: %s", err)
	}
	oldname := header.Name
	header.Name = packagepath
	if err = tw.WriteHeader(header); err != nil {
		return fmt.Errorf("Error write header for (path: %s, oldname:%s,newname:%s,sz:%d) : %s", localpath, oldname, packagepath, header.Size, err)
	}
	if _, err := io.Copy(tw, is); err != nil {
		return fmt.Errorf("Error copy (path: %s, oldname:%s,newname:%s,sz:%d) : %s", localpath, oldname, packagepath, header.Size, err)
	}

	return nil
}

func TarFiles(files []string, dstPath string) error {
	oldpwd, _ := os.Getwd()
	dstFile := filepath.Base(dstPath)
	dstDir := filepath.Dir(dstPath)
	if _, err := CreateDirIfMissing(dstDir); err != nil {
		return fmt.Errorf("Error create the dst dir %s for : %s", dstDir, err)
	}
	os.Chdir(dstDir)
	defer os.Chdir(oldpwd)
	fw, err := os.Create(dstFile)
	if err != nil {
		return fmt.Errorf("Cannot create file %s for %s", dstFile, err)
	}
	defer fw.Close()
	gw := gzip.NewWriter(fw)
	tw := tar.NewWriter(gw)

	for _, file := range files {
		name := filepath.Base(file)
		err = WriteFileToPackage(file, name, tw)
		if err != nil {
			return fmt.Errorf("Error writing %s to tar: %s", file, err)
		}
	}
	tw.Close()
	gw.Close()
	return nil
}

func RemoveFiles(files []string) error {
	var err error
	for _, file := range files {
		if err1 := os.Remove(file); err1 != nil {
			err = err1
		}
	}
	return err
}

func LoadFileByNumber(srcDir string, dstDir string, blockFileName string, number uint64, formatString string) error {
	logger.Debugf("Load Block file [%d] from %s", number, srcDir)
	if filename, err := findTarByNumber(srcDir, number, formatString); err != nil {
		return err
	} else {
		return Untar(filepath.Join(srcDir, filename), filepath.Join(dstDir, blockFileName), blockFileName)
	}
}

func Untar(srcTar string, dstFile string, filename string) error {
	fr, err := os.Open(srcTar)
	if err != nil {
		return fmt.Errorf("Cannot open the file %s for %s", srcTar, err)
	}
	defer fr.Close()

	gr, err := gzip.NewReader(fr)
	if err != nil {
		return fmt.Errorf("Cannot get the gzip reader %s", err)
	}
	defer gr.Close()

	tr := tar.NewReader(gr)

	for hdr, err := tr.Next(); err != io.EOF; hdr, err = tr.Next() {
		if err != nil {
			return fmt.Errorf("Cannot get the header of file %s", err)
		}
		if hdr.Name == filename {
			return UntarFile(dstFile, tr)
		}
	}
	return fmt.Errorf("Cannot find the file :[%s]", filename)
}
func UntarFile(dstFile string, tr *tar.Reader) error {
	fw, err := os.Create(dstFile)
	if err != nil {
		return fmt.Errorf("Cannot create the file %s for %s", dstFile, err)
	}
	defer fw.Close()

	_, err = io.Copy(fw, tr)
	if err != nil {
		return fmt.Errorf("Cannot untar the file %s for %s", dstFile, err)
	}
	return nil
}

func findTarByNumber(path string, number uint64, formatString string) (string, error) {
	file_list, err := ioutil.ReadDir(path)
	if err != nil {
		return "", err
	}
	for _, file := range file_list {
		fileName := file.Name()
		var startNum, endNum uint64
		fmt.Sscanf(fileName, formatString, &startNum, &endNum)
		if startNum <= number && number <= endNum {
			return fileName, nil
		}
	}
	return "", fmt.Errorf("block %d Not found", number)
}

// ListSubdirs returns the subdirectories
func ListSubdirs(dirPath string) ([]string, error) {
	subdirs := []string{}
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		if f.IsDir() {
			subdirs = append(subdirs, f.Name())
		}
	}
	return subdirs, nil
}

func logDirStatus(msg string, dirPath string) {
	exists, _, err := FileExists(dirPath)
	if err != nil {
		logger.Errorf("Error while checking for dir existence")
	}
	if exists {
		logger.Debugf("%s - [%s] exists", msg, dirPath)
	} else {
		logger.Debugf("%s - [%s] does not exist", msg, dirPath)
	}
}
