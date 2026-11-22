package api

import "github.com/Mag1cFall/AIStudio2API/internal/aistudio"

// uncutModels 为支持函数调用的对话模型补充 -uncut 抗截断型号
func (s *server) uncutModels(models []aistudio.Model) []aistudio.Model {
	return aistudio.ModelsWithAntiCensorAliases(models)
}
