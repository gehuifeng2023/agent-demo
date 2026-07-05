package prompt

import "fmt"

type LogAnalysisBuilder struct{}

func NewLogAnalysisBuilder() *LogAnalysisBuilder {
	return &LogAnalysisBuilder{}
}

func (b *LogAnalysisBuilder) Build(input string) string {
	return fmt.Sprintf(`你是一个资深后端问题排查专家，擅长分析 Go、Nginx、APISIX、Kubernetes、HTTP 网关和代理服务日志。

请分析下面的日志内容。

回答结构必须包含：
1. 结论
2. 关键错误点
3. 可能原因
4. 影响范围
5. 排查步骤
6. 修复建议
7. 如果信息不足，需要补充哪些日志或配置

要求：
- 不要只翻译日志
- 要结合上下文解释错误链路
- 对不确定的地方明确标注“可能”
- 优先给出可操作建议

日志内容：
%s`, input)
}
