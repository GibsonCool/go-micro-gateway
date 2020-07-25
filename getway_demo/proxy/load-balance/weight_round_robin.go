package load_balance

import (
	"errors"
	"strconv"
)

type WeightRoundRobinBalance struct {
	curIndex int
	rss      []*WeightNode
	rsw      []int
}

/*
	参考 Nginx  加权负载轮询实现机制
		1、每次取值遍历所有节点,过程中统计出所有权重之和  totalWeight=(nod1.weight+node2.weigh+...)

		2、计算各个单节点当前临时权重为节点当前临时权重+节点有效权重：currentWeight += effectiveWeight

		3、遍历中选出 当前临时权重最大的节点为命中节点

		4、遍历结束。修改命中节点的当前临时权重为当前临时权重-所有权重之和  currentWeight-=totalWeight

		5、返回命中节点的节点信息
*/
func (w *WeightRoundRobinBalance) Add(params ...string) error {
	if len(params) != 2 {
		return errors.New("参数错误，限制2个值例如：addr,weight")
	}
	parInt, err := strconv.Atoi(params[1])
	if err != nil {
		return err
	}

	node := &WeightNode{addr: params[0], weight: parInt}
	node.effectiveWeight = node.weight
	w.rss = append(w.rss, node)
	return nil
}

func (w *WeightRoundRobinBalance) Get(key string) (string, error) {
	return w.Next(), nil
}

func (w *WeightRoundRobinBalance) Next() string {
	totalWeight := 0
	var best *WeightNode

	// 通过遍历选出当前选中节点
	for i := 0; i < len(w.rss); i++ {
		w := w.rss[i]
		// 统计所有有效权重之和
		totalWeight += w.effectiveWeight

		// 变更节点当前临时权重为  节点当前临时权重+节点有效权重
		w.currentWeight += w.effectiveWeight

		// 有效权重默认与初始权重相同，如果通讯异常的时候权重下降-1，通讯成功则回复+1，知道回味到初始权重大小
		if w.effectiveWeight < w.weight {
			w.effectiveWeight++
		}

		// 挑选最大当前临时权重节点
		if best == nil || w.currentWeight > best.currentWeight {
			best = w
		}
	}
	if best == nil {
		return ""
	}

	// 变更选中节点的当前临时权重为   当前临时权重-有效权重之和
	best.currentWeight -= totalWeight

	// 返回节点信息地址
	return best.addr
}

type WeightNode struct {
	addr            string
	weight          int // 权重初始值
	currentWeight   int // 节点当前临时权重
	effectiveWeight int // 有效权重
}
