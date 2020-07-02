package load_balance

import (
	"errors"
	"hash/crc32"
	"sort"
	"strconv"
	"sync"
)

/*
	一致性hash负载均衡实现
*/

type Hash func(data []byte) uint32

type UInt32Slice []uint32

func (s UInt32Slice) Len() int {
	return len(s)
}

func (s UInt32Slice) Less(i, j int) bool {
	return s[i] < s[j]
}

func (s UInt32Slice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type ConsistentHashBalance struct {
	mux      sync.RWMutex      // 同步锁
	hashFn   Hash              // 计算 hashFn 值的方法
	replicas int               // 复制因子,也就是每个节点对应需要几个虚拟节点
	keys     UInt32Slice       // 已排序的节点 hashFn 切片
	hashMap  map[uint32]string // 节点hash和key 的 map.
}

func NewConsistentHashBalance(fn Hash, replicas int) *ConsistentHashBalance {
	m := &ConsistentHashBalance{hashFn: fn,
		replicas: replicas,
		hashMap:  make(map[uint32]string),
	}
	if m.hashFn == nil {
		m.hashFn = crc32.ChecksumIEEE
	}
	return m
}

func (c *ConsistentHashBalance) Add(params ...string) error {
	if len(params) == 0 {
		return errors.New("参数长度不能为空")
	}

	c.mux.Lock()
	defer c.mux.Unlock()
	for _, addr := range params {
		for i := 0; i < c.replicas; i++ {
			// 计算出虚拟节点的 hash 值
			hashKey := c.hashFn([]byte(strconv.Itoa(i) + addr))
			c.keys = append(c.keys, hashKey)
			c.hashMap[hashKey] = addr
		}
	}

	// 对hash节点切边进行排序，便于后面获取的时候进行二分查找
	sort.Sort(c.keys)
	return nil
}

func (c *ConsistentHashBalance) IsEmpty() bool {
	return len(c.keys) == 0
}

func (c *ConsistentHashBalance) Get(key string) (string, error) {
	if c.IsEmpty() {
		return "", errors.New("节点切片为空")
	}

	hashKey := c.hashFn([]byte(key))
	idx := sort.Search(len(c.keys), func(i int) bool {
		return c.keys[i] >= hashKey
	})

	if idx == len(c.keys) {
		idx = 0
	}
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.hashMap[c.keys[idx]], nil
}
