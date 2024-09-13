package cmd

import (
	"encoding/json"
	"fmt"
	ignore "github.com/sabhiram/go-gitignore"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

type Template struct {
	Name  string     `json:"name"`
	Files []FileInfo `json:"files"`
}

type FileInfo struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

func NewTemplateCommand() *cobra.Command {
	var templateName string

	cmd := &cobra.Command{
		Use:   "template -n <模板名字> <项目目录>",
		Short: "创建项目模板",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectDir := args[0]
			return createTemplate(templateName, projectDir)
		},
	}

	cmd.Flags().StringVarP(&templateName, "name", "n", "", "模板名字")
	cmd.MarkFlagRequired("name")

	return cmd
}

func createTemplate(templateName, projectDir string) error {
	var template Template
	template.Name = templateName

	// 解析 .gitignore 文件
	gitignorePatterns, err := getGitignorePatterns(projectDir)
	if err != nil {
		return err
	}

	// 遍历项目目录，生成模板文件信息
	err = filepath.Walk(projectDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 忽略 .git 文件夹
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}

		// 检查是否需要忽略文件
		if shouldIgnoreFile(path, projectDir, gitignorePatterns) {
			return nil
		}

		// 读取文件内容并保存到模板中
		if !info.IsDir() {
			relPath, err := filepath.Rel(projectDir, path)
			if err != nil {
				return err
			}

			content, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}

			fileInfo := FileInfo{
				Path:    filepath.ToSlash(relPath), // 确保使用 / 分隔符保存路径
				Content: string(content),
			}
			template.Files = append(template.Files, fileInfo)
		}

		return nil
	})

	if err != nil {
		return err
	}

	// 将模板写入 JSON 文件
	jsonData, err := json.MarshalIndent(template, "", "  ")
	if err != nil {
		return err
	}

	templateFileName := fmt.Sprintf("%s.json", templateName)
	if err := ioutil.WriteFile(templateFileName, jsonData, 0644); err != nil {
		return err
	}

	fmt.Printf("模板 %s 已成功创建并保存到 %s\n", templateName, templateFileName)
	return nil
}

// getGitignorePatterns 读取 .gitignore 文件并解析规则
func getGitignorePatterns(projectDir string) ([]string, error) {
	gitignorePath := filepath.Join(projectDir, ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		return nil, nil
	}

	content, err := ioutil.ReadFile(gitignorePath)
	if err != nil {
		return nil, err
	}

	var patterns []string
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			patterns = append(patterns, line)
		}
	}
	return patterns, nil
}

// shouldIgnoreFile 根据 .gitignore 规则判断是否忽略文件
func shouldIgnoreFile(path string, projectDir string, patterns []string) bool {
	if len(patterns) == 0 {
		return false
	}

	ignoreMatcher := ignore.CompileIgnoreLines(patterns...)
	relPath, err := filepath.Rel(projectDir, path)
	if err != nil {
		return false
	}

	return ignoreMatcher.MatchesPath(filepath.ToSlash(relPath))
}
