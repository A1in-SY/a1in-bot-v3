package infrastructure

/*
Infrastructure 定义为直接与外部依赖交互的部分，仅提供可靠的基础能力
可靠指与外部依赖的连接好坏只在本层关心，本层承诺交付一个可用的服务
仅基础能力指只提供需要的外部依赖所能提供的能力，不进行任何包装
本层间不应有依赖关系
*/
type Infrastructure interface {
	InitInfra(cbs []byte) error
}
