package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

// 正则表达式匹配模板变量格式：@变量名@
var variableRegex = regexp.MustCompile(`@([a-zA-Z0-9_-]+)@`)

func NewCreateCommand() *cobra.Command {
	var templateFilePath string

	cmd := &cobra.Command{
		Use:   "create -t <模板json文件路径> <项目文件夹>",
		Short: "应用项目模板并在当前目录下创建项目文件夹",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectDir := args[0]
			return applyTemplate(templateFilePath, projectDir)
		},
	}

	cmd.Flags().StringVarP(&templateFilePath, "template", "t", "", "模板 JSON 文件路径")
	cmd.MarkFlagRequired("template")

	return cmd
}

func applyTemplate(templateFilePath, projectDir string) error {
	// 读取模板 JSON 文件
	jsonData, err := ioutil.ReadFile(templateFilePath)
	if err != nil {
		return err
	}

	var template Template
	if err := json.Unmarshal(jsonData, &template); err != nil {
		return err
	}

	// 收集所有模板变量
	variables := make(map[string]string)

	for _, file := range template.Files {
		findVariables(file.Content, variables)
	}

	// 询问用户填写每个模板变量的值
	askForVariableValues(variables)

	// 创建目标文件夹（在当前目录下）
	targetDir := filepath.Join(".", projectDir)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return err
	}

	// 在目标文件夹中应用模板
	for _, file := range template.Files {
		targetPath := filepath.Join(targetDir, filepath.FromSlash(file.Path)) // 使用 FromSlash 兼容 Windows 路径

		// 创建文件夹
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}

		// 替换模板变量并写入文件
		finalContent := replaceVariables(file.Content, variables)
		if err := ioutil.WriteFile(targetPath, []byte(finalContent), 0644); err != nil {
			return err
		}
	}

	fmt.Printf("模板 %s 已成功应用到 %s\n", template.Name, targetDir)
	return nil
}

// findVariables 查找文件内容中的模板变量
func findVariables(content string, variables map[string]string) {
	matches := variableRegex.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		variableName := match[1]
		if _, exists := variables[variableName]; !exists {
			variables[variableName] = "" // 初始化为空字符串，等待用户输入
		}
	}
}

// askForVariableValues 询问用户为每个模板变量提供值
func askForVariableValues(variables map[string]string) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("请为以下模板变量提供值:")

	for variable := range variables {
		fmt.Printf("%s: ", variable)
		value, _ := reader.ReadString('\n')
		variables[variable] = strings.TrimSpace(value)
	}
}

var replaceAnnotations = []string{
	"//@replace",
	"// @replace",
}

// replaceVariables 替换文件内容中的模板变量
func replaceVariables(content string, variables map[string]string) string {
	variableReplacedContent := variableRegex.ReplaceAllStringFunc(content, func(match string) string {
		variableName := match[1 : len(match)-1] // 去掉@符号
		if value, exists := variables[variableName]; exists {
			return value
		}
		return match // 如果没有对应的值，保留原模板变量
	})

	var finalContent string
	lines := strings.Split(variableReplacedContent, "\n")
	for i := 0; i < len(lines)-1; i++ {
		current := lines[i]
		for _, annotation := range replaceAnnotations {
			replaceIndex := strings.Index(current, annotation)
			if replaceIndex != -1 {
				start := current[:replaceIndex]
				end := current[replaceIndex+len(annotation):]
				end = strings.TrimLeft(end, " ")
				current = start + end
				i++
				break
			}
		}
		finalContent += current + "\n"
	}

	return finalContent
}
