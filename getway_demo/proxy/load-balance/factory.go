package load_balance

// 定义 负载均衡 枚举
type LbType int

const (
	LbRandom           LbType = iota // 随机
	LbRoundRobin                     // 轮询
	LbWeightRoundRobin               // 权重轮询
	LbConsistentHash                 // 一致性hash
)

func LoadBalanceFactory(lbType LbType) LoadBalance {
	switch lbType {
	case LbRandom:
		return &RandomBalance{}
	case LbRoundRobin:
		return &RoundRobinBalance{}
	case LbWeightRoundRobin:
		return &WeightRoundRobinBalance{}
	case LbConsistentHash:
		return NewConsistentHashBalance(nil, 10)
	default:
		return &RandomBalance{}
	}
}
