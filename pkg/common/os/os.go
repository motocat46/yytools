// Package os.

// 版权所有(Copyright)[yangyuan]
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

// http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// 作者:  yangyuan
// 创建日期:2023/6/7
package os

import (
	"fmt"
	"os"
	"time"
)

// 检测路径是否存在
func IsFileExist(file string) (error, bool) {
	err, ok := IsFileNormalStat(file)
	if ok {
		return err, ok
	}
	return err, os.IsExist(err)
}

func IsFileNormalStat(file string) (error, bool) {
	_, err := os.Stat(file)
	if err == nil {
		return nil, true
	}
	return err, false
}

// 备份指定路径文件
// 重命名原文件(添加日期时间后缀)
// 例如 ~/work/test.go -> ~/work/test.go_202306071537010001
func BackupFile(file string) (error, bool) {
	err, ok := IsFileNormalStat(file)
	if ok {
		now := time.Now()
		for i := 1; i < 1e4; i++ {
			newName := fmt.Sprintf("%v_%d%02d%02d%02d%02d%02d%04d",
				file, now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), i)
			if _, isExist := IsFileExist(newName); isExist {
				continue
			}
			err1 := os.Rename(file, newName)
			return err1, ok
		}
		return os.ErrExist, false
	}
	return err, ok
}