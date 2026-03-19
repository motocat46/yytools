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
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestIsFileExist(t *testing.T) {
	// 创建临时目录用于测试
	tempDir := t.TempDir()

	// 创建测试文件
	testFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}

	tests := []struct {
		name     string
		filePath string
		exists   bool
	}{
		{"存在的文件", testFile, true},
		{"不存在的文件", filepath.Join(tempDir, "nonexistent.txt"), false},
		{"空路径", "", false},
		{"目录路径", tempDir, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists, err := IsFileExist(tt.filePath)

			if tt.exists {
				if err != nil {
					t.Errorf("IsFileExist(%q) 返回了错误: %v", tt.filePath, err)
				}
				if !exists {
					t.Errorf("IsFileExist(%q) = %t, 期望 true", tt.filePath, exists)
				}
			} else {
				if exists {
					t.Errorf("IsFileExist(%q) = %t, 期望 false", tt.filePath, exists)
				}
			}
		})
	}
}

func TestIsFileNormalStat(t *testing.T) {
	// 创建临时目录用于测试
	tempDir := t.TempDir()

	// 创建测试文件
	testFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}

	tests := []struct {
		name     string
		filePath string
		exists   bool
	}{
		{"存在的文件", testFile, true},
		{"不存在的文件", filepath.Join(tempDir, "nonexistent.txt"), false},
		{"空路径", "", false},
		{"目录路径", tempDir, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists, err := IsFileNormalStat(tt.filePath)

			if tt.exists {
				if err != nil {
					t.Errorf("IsFileNormalStat(%q) 返回了错误: %v", tt.filePath, err)
				}
				if !exists {
					t.Errorf("IsFileNormalStat(%q) = %t, 期望 true", tt.filePath, exists)
				}
			} else {
				if exists {
					t.Errorf("IsFileNormalStat(%q) = %t, 期望 false", tt.filePath, exists)
				}
			}
		})
	}
}

func TestBackupFile(t *testing.T) {
	// 创建临时目录用于测试
	tempDir := t.TempDir()

	// 创建测试文件
	testFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}

	t.Run("备份存在的文件", func(t *testing.T) {
		success, err := BackupFile(testFile)

		if err != nil {
			t.Errorf("BackupFile(%q) 返回了错误: %v", testFile, err)
		}
		if !success {
			t.Errorf("BackupFile(%q) = %t, 期望 true", testFile, success)
		}

		// 检查原文件是否被重命名
		exists, _ := IsFileExist(testFile)
		if exists {
			t.Errorf("原文件 %q 仍然存在，应该被重命名", testFile)
		}

		// 检查备份文件是否存在
		backupPattern := filepath.Join(tempDir, "test.txt_*")
		matches, _ := filepath.Glob(backupPattern)
		if len(matches) == 0 {
			t.Errorf("未找到备份文件")
		}
	})

	t.Run("备份不存在的文件", func(t *testing.T) {
		nonexistentFile := filepath.Join(tempDir, "nonexistent.txt")
		success, _ := BackupFile(nonexistentFile)

		if success {
			t.Errorf("BackupFile(%q) = %t, 期望 false", nonexistentFile, success)
		}
	})

	t.Run("备份空路径", func(t *testing.T) {
		success, _ := BackupFile("")

		if success {
			t.Errorf("BackupFile(\"\") = %t, 期望 false", success)
		}
	})
}

func TestBackupFileMultipleTimes(t *testing.T) {
	// 测试多次备份同一文件
	tempDir := t.TempDir()

	for i := 0; i < 3; i++ {
		// 创建测试文件
		testFile := filepath.Join(tempDir, "test.txt")
		err := os.WriteFile(testFile, []byte("test content"), 0644)
		if err != nil {
			t.Fatalf("创建测试文件失败: %v", err)
		}

		// 备份文件
		success, err := BackupFile(testFile)
		if err != nil {
			t.Errorf("第 %d 次备份失败: %v", i+1, err)
		}
		if !success {
			t.Errorf("第 %d 次备份返回 false", i+1)
		}

		// 短暂延迟，确保时间戳不同
		time.Sleep(10 * time.Millisecond)
	}

	// 检查所有备份文件
	backupPattern := filepath.Join(tempDir, "test.txt_*")
	matches, _ := filepath.Glob(backupPattern)
	if len(matches) != 3 {
		t.Errorf("期望 3 个备份文件，实际找到 %d 个", len(matches))
	}
}

func TestBackupFileWithSpecialCharacters(t *testing.T) {
	// 测试包含特殊字符的文件名
	tempDir := t.TempDir()

	specialFile := filepath.Join(tempDir, "test file with spaces.txt")
	err := os.WriteFile(specialFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}

	success, err := BackupFile(specialFile)
	if err != nil {
		t.Errorf("备份特殊字符文件名失败: %v", err)
	}
	if !success {
		t.Errorf("备份特殊字符文件名返回 false")
	}

	// 检查备份文件是否存在
	backupPattern := filepath.Join(tempDir, "test file with spaces.txt_*")
	matches, _ := filepath.Glob(backupPattern)
	if len(matches) == 0 {
		t.Errorf("未找到备份文件")
	}
}

func TestBackupFilePerformance(t *testing.T) {
	// 性能测试
	tempDir := t.TempDir()

	for i := 0; i < 100; i++ {
		testFile := filepath.Join(tempDir, "test.txt")
		err := os.WriteFile(testFile, []byte("test content"), 0644)
		if err != nil {
			t.Fatalf("创建测试文件失败: %v", err)
		}

		success, err := BackupFile(testFile)
		if err != nil {
			t.Errorf("性能测试中备份失败: %v", err)
		}
		if !success {
			t.Errorf("性能测试中备份返回 false")
		}
	}
}
