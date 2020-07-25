package load_balance

import (
	"errors"
)

type RoundRobinBalance struct {
	curIndex int
	rss      []string
}

func (r *RoundRobinBalance) Add(params ...string) error {
	if len(params) == 0 {
		return errors.New("参数长度为空")
	}
	r.rss = append(r.rss, params...)
	return nil
}

func (r *RoundRobinBalance) Get(key string) (string, error) {
	return r.Next(), nil
}

func (r *RoundRobinBalance) Next() string {
	if len(r.rss) == 0 {
		return ""
	}
	lens := len(r.rss)
	if r.curIndex >= lens {
		r.curIndex = 0
	}
	curAddr := r.rss[r.curIndex]
	r.curIndex += 1
	return curAddr
}
