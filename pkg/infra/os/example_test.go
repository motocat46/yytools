// 版权所有(Copyright)[yangyuan]
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package os_test

import (
	"fmt"
	"os"

	yos "github.com/motocat46/yytools/pkg/infra/os"
)

// ExampleIsFileExist 展示路径存在性检查——不存在返回 (false, nil)，而非 error。
func ExampleIsFileExist() {
	// 不存在的路径：false, nil（不存在不是错误）
	exists, err := yos.IsFileExist("/tmp/yytools-nonexistent-file-xyz123")
	fmt.Println(exists, err)
	// Output:
	// false <nil>
}

// ExampleBackupFile 展示文件备份——重命名原文件并追加时间戳后缀。
// 例：test.go → test.go_202306071537010001
func ExampleBackupFile() {
	// 创建临时文件
	f, err := os.CreateTemp("", "yytools-backup-example-*.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	name := f.Name()
	f.Close()

	// 备份：原文件被重命名，不再存在
	ok, err := yos.BackupFile(name)
	fmt.Println(ok, err)

	// 原路径已不存在
	exists, _ := yos.IsFileExist(name)
	fmt.Println(exists)
	// Output:
	// true <nil>
	// false
}
