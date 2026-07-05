package prompt

import "fmt"

type TaskBreakdownBuilder struct{}

func NewTaskBreakdownBuilder() *TaskBreakdownBuilder {
	return &TaskBreakdownBuilder{}
}

func (b *TaskBreakdownBuilder) Build(input string) string {
	return fmt.Sprintf(`你是一个资深 Go 后端架构师和研发任务拆解专家。

请根据下面的需求内容，拆解研发任务。

输出结构：
1. 需求理解
2. 总体目标
3. 功能模块拆分
4. 后端接口设计
5. 数据结构设计
6. 核心流程
7. 开发任务列表
8. 测试点
9. 风险点
10. 建议开发顺序

要求：
- 任务要能落地开发
- 每个任务要有明确边界
- 优先使用 Go 后端项目视角
- 不要过度设计
- 保持核心功能优先

需求内容：
%s`, input)
}
