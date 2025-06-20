package utils

import (
	"fmt"
	"github.com/qctc/fabric2-api-server/define"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
)

var ConfigPath string

// UpdateUserInConfig 动态添加组织或用户，保留原始格式（缩进、注释、顺序）
// 返回值:
//   - bool: 如果没有修改配置文件则返回 true，否则返回 false
//   - error: 错误信息
func UpdateUserInConfig(req define.UserConfigRequest, chainName string) (bool, error) {
	fabricPath, ok := define.GlobalConfig.Fabric[chainName]
	if !ok {
		return false, fmt.Errorf("chain %s not found in global config", chainName)
	}
	ConfigPath = fabricPath.ConfigFilePath

	// 读取原始 YAML 数据（保留完整结构）
	data, err := ioutil.ReadFile(ConfigPath)
	if err != nil {
		return false, fmt.Errorf("failed to read config file: %v", err)
	}

	// 使用 yaml.Node 解析，保留结构
	var root yaml.Node
	err = yaml.Unmarshal(data, &root)
	if err != nil {
		return false, fmt.Errorf("failed to unmarshal config with Node: %v", err)
	}

	// ======================
	// 1️⃣ 查找 organizations 节点
	// ======================
	var orgsNode *yaml.Node
	for i := 0; i < len(root.Content[0].Content); i += 2 {
		keyNode := root.Content[0].Content[i]
		if keyNode.Value == "organizations" {
			orgsNode = root.Content[0].Content[i+1]
			break
		}
	}

	if orgsNode == nil {
		return false, fmt.Errorf("organizations section not found")
	}

	// ======================
	// 2️⃣ 遍历 organizations 节点查找 mspid
	// ======================
	foundMspid := false
	foundUserName := false

	for i := 0; i < len(orgsNode.Content); i += 2 {
		orgKey := orgsNode.Content[i]
		orgVal := orgsNode.Content[i+1]

		if orgKey.Value == "ordererorg" {
			continue // 跳过 ordererorg
		}

		var mspidNode *yaml.Node
		for j := 0; j < len(orgVal.Content); j += 2 {
			fieldKey := orgVal.Content[j]
			if fieldKey.Value == "mspid" {
				mspidNode = orgVal.Content[j+1]
				break
			}
		}

		if mspidNode != nil && mspidNode.Value == req.MspId {
			foundMspid = true

			// 检查 users 是否存在当前用户名
			var usersNode *yaml.Node
			for j := 0; j < len(orgVal.Content); j += 2 {
				fieldKey := orgVal.Content[j]
				if fieldKey.Value == "users" {
					usersNode = orgVal.Content[j+1]
					break
				}
			}

			if usersNode == nil {
				// 创建新的 users 节点
				usersNode = &yaml.Node{
					Kind: yaml.MappingNode,
					Tag:  "!!map",
				}
				newUsersKey := &yaml.Node{
					Kind:  yaml.ScalarNode,
					Value: "users",
					Tag:   "!!str",
				}
				orgVal.Content = append(orgVal.Content, newUsersKey, usersNode)
			}

			// 检查是否已有该用户
			foundUserName = false
			for j := 0; j < len(usersNode.Content); j += 2 {
				userKey := usersNode.Content[j]
				if userKey.Value == req.UserName {
					foundUserName = true
					break
				}
			}

			// 用户不存在 → 添加新用户
			if !foundUserName {
				userKeyNode := &yaml.Node{
					Kind:  yaml.ScalarNode,
					Value: req.UserName,
					Tag:   "!!str",
				}
				certPathNode := &yaml.Node{
					Kind:  yaml.ScalarNode,
					Value: CertPath(req.PathId),
					Tag:   "!!str",
				}
				keyPathNode := &yaml.Node{
					Kind:  yaml.ScalarNode,
					Value: KeyPath(req.PathId),
					Tag:   "!!str",
				}
				userValueNode := &yaml.Node{
					Kind: yaml.MappingNode,
					Content: []*yaml.Node{
						{Kind: yaml.ScalarNode, Value: "cert", Tag: "!!str"},
						&yaml.Node{
							Kind: yaml.MappingNode,
							Content: []*yaml.Node{
								{Kind: yaml.ScalarNode, Value: "path", Tag: "!!str"},
								certPathNode,
							},
						},
						{Kind: yaml.ScalarNode, Value: "key", Tag: "!!str"},
						&yaml.Node{
							Kind: yaml.MappingNode,
							Content: []*yaml.Node{
								{Kind: yaml.ScalarNode, Value: "path", Tag: "!!str"},
								keyPathNode,
							},
						},
					},
				}

				usersNode.Content = append(usersNode.Content, userKeyNode, userValueNode)
			}
			break
		}
	}

	// ======================
	// 3️⃣ 如果未找到 mspid → 复制 org1 并新建
	// ======================
	if !foundMspid {
		// 查找 org1 的结构模板
		var org1Node *yaml.Node
		var org1Key *yaml.Node
		for i := 0; i < len(orgsNode.Content); i += 2 {
			orgKey := orgsNode.Content[i]
			orgVal := orgsNode.Content[i+1]
			if orgKey.Value == "org1" {
				org1Key = orgKey
				org1Node = orgVal
				break
			}
		}

		if org1Node == nil {
			return false, fmt.Errorf("template org1 not found")
		}

		// 创建新组织节点，仅复制必要的结构，不复制 users
		orgCopy := &yaml.Node{
			Kind:    yaml.MappingNode,
			Tag:     "!!map",
			Content: []*yaml.Node{},
		}

		// 复制除 users 以外的所有字段
		for i := 0; i < len(org1Node.Content); i += 2 {
			fieldKey := org1Node.Content[i]
			fieldVal := org1Node.Content[i+1]

			if fieldKey.Value == "users" {
				continue // 跳过 users 字段
			}

			// 深拷贝字段
			newKey := &yaml.Node{
				Kind:  yaml.ScalarNode,
				Value: fieldKey.Value,
				Tag:   fieldKey.Tag,
			}
			newVal := deepCopyYAML(fieldVal)
			orgCopy.Content = append(orgCopy.Content, newKey, newVal)
		}

		// 设置新的 mspid
		for i := 0; i < len(orgCopy.Content); i += 2 {
			fieldKey := orgCopy.Content[i]
			if fieldKey.Value == "mspid" {
				orgCopy.Content[i+1].Value = req.MspId
				break
			}
		}

		// 添加新的 users 节点
		userKeyNode := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: req.UserName,
			Tag:   "!!str",
		}
		certPathNode := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: CertPath(req.PathId),
			Tag:   "!!str",
		}
		keyPathNode := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: KeyPath(req.PathId),
			Tag:   "!!str",
		}
		userValueNode := &yaml.Node{
			Kind: yaml.MappingNode,
			Content: []*yaml.Node{
				{Kind: yaml.ScalarNode, Value: "cert", Tag: "!!str"},
				&yaml.Node{
					Kind: yaml.MappingNode,
					Content: []*yaml.Node{
						{Kind: yaml.ScalarNode, Value: "path", Tag: "!!str"},
						certPathNode,
					},
				},
				{Kind: yaml.ScalarNode, Value: "key", Tag: "!!str"},
				&yaml.Node{
					Kind: yaml.MappingNode,
					Content: []*yaml.Node{
						{Kind: yaml.ScalarNode, Value: "path", Tag: "!!str"},
						keyPathNode,
					},
				},
			},
		}

		usersKey := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: "users",
			Tag:   "!!str",
		}
		usersValue := &yaml.Node{
			Kind: yaml.MappingNode,
			Content: []*yaml.Node{
				userKeyNode,
				userValueNode,
			},
		}
		orgCopy.Content = append(orgCopy.Content, usersKey, usersValue)

		// 构造新组织名称并插入到 organizations 下
		newOrgNum := findMaxOrgNumberUsingNode(orgsNode) + 1
		newOrgKey := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: fmt.Sprintf("org%d", newOrgNum),
			Tag:   "!!str",
		}
		orgsNode.Content = append(orgsNode.Content, newOrgKey, orgCopy)
		newOrgKey.Line = org1Key.Line // 保留原缩进

	}

	// ======================
	// 4️⃣ 更新 client.organization 字段
	// ======================
	//var clientNode *yaml.Node
	//for i := 0; i < len(root.Content[0].Content); i += 2 {
	//	keyNode := root.Content[0].Content[i]
	//	if keyNode.Value == "client" {
	//		clientNode = root.Content[0].Content[i+1]
	//		break
	//	}
	//}
	//
	//if clientNode != nil {
	//	for i := 0; i < len(clientNode.Content); i += 2 {
	//		keyNode := clientNode.Content[i]
	//		if keyNode.Value == "organization" {
	//			clientNode.Content[i+1].Value = newOrgName
	//			break
	//		}
	//	}
	//}

	// ======================
	// 5️⃣ 写回配置文件
	// ======================
	// 如果 mspid 和用户名都存在，并且没有新增任何内容，则直接返回 true
	if foundMspid && foundUserName {
		return true, nil
	}

	newData, _ := yaml.Marshal(&root)
	err = ioutil.WriteFile(ConfigPath, newData, os.ModePerm)
	if err != nil {
		return false, fmt.Errorf("failed to write config file: %v", err)
	}

	return false, nil
}

// deepCopyYAML 深度复制 yaml.Node 树
func deepCopyYAML(in *yaml.Node) *yaml.Node {
	out := &yaml.Node{}
	*out = *in
	if in.Content != nil {
		out.Content = make([]*yaml.Node, len(in.Content))
		for i, ch := range in.Content {
			out.Content[i] = deepCopyYAML(ch)
		}
	}
	return out
}

// CertPath 构造 cert.path
func CertPath(path string) string {
	return fmt.Sprintf("%s/%s/account.crt", path)
}

// KeyPath 构造 key.path
func KeyPath(path string) string {
	return fmt.Sprintf("%s/%s/account.key", path)
}

// findMaxOrgNumberUsingNode 查找最大的 org 编号
func findMaxOrgNumberUsingNode(orgsNode *yaml.Node) int {
	maxN := 0
	re := regexp.MustCompile(`^org(\d+)$`)
	for i := 0; i < len(orgsNode.Content); i += 2 {
		name := orgsNode.Content[i].Value
		matches := re.FindStringSubmatch(name)
		if len(matches) > 1 {
			num, _ := strconv.Atoi(matches[1])
			if num > maxN {
				maxN = num
			}
		}
	}
	return maxN
}

func ConvertStringMatrixToByteMatrix(input [][]string) ([][]byte, error) {
	var result [][]byte

	for _, row := range input {
		var byteRow []byte
		for _, str := range row {
			byteRow = append(byteRow, []byte(str)...)
		}
		result = append(result, byteRow)
	}

	return result, nil
}
