package main

import (
	"time"
)

const (
	CONG_SUCCESS = 1 << 0
	CONG_LOSS    = 1 << 1
	CONG_TIMEOUT = 1 << 2
	CONG_EARLY   = 1 << 3
	CONG_NOACK   = 1 << 4
)

type CongestionControlAlgorithm interface {
	Add(status uint8, sendAt time.Time, rtt time.Duration)
	WindowSize() int
}

type SimpleControl struct {
	size int
}

func NewSimpleControl(initial int) CongestionControlAlgorithm {
	return &SimpleControl{size: initial}
}

func (c *SimpleControl) Add(status uint8, sendAt time.Time, rtt time.Duration) {
	if status&CONG_SUCCESS != 0 {
		c.size++
	} else if c.size > 1 {
		c.size /= 2
	}
}

func (c *SimpleControl) WindowSize() int {
	return c.size
}

type VegasControl struct {
	size int
	rtt  *RTTCollecter
	a, b float64
}

func NewVegasControl(inital int, rtt *RTTCollecter, a, b float64) CongestionControlAlgorithm {
	return &VegasControl{
		size: inital,
		rtt:  rtt,
		a:    a,
		b:    b,
	}
}

func (c *VegasControl) Add(status uint8, sendAt time.Time, rtt time.Duration) {
	if status&CONG_SUCCESS != 0 {
		cwnd := float64(c.size)
		diff := (cwnd/c.rtt.min - cwnd/float64(rtt)) * c.rtt.min
		//logger.Error("diff :", diff)
		if c.size > 1 && diff > c.b {
			c.size--
		} else if diff < c.a {
			c.size++
		}
	}
}

func (c *VegasControl) WindowSize() int {
	return c.size
}

type Vegas2Control struct {
	size      int
	rtt       *RTTCollecter
	a, b      float64
	lossCount int
	loss      int
}

func NewVegas2Control(inital int, rtt *RTTCollecter, a, b float64, loss int) CongestionControlAlgorithm {
	return &Vegas2Control{
		size: inital,
		rtt:  rtt,
		a:    a,
		b:    b,
		loss: loss,
	}
}

func (c *Vegas2Control) Add(status uint8, sendAt time.Time, rtt time.Duration) {
	if status&CONG_SUCCESS != 0 {
		cwnd := float64(c.size)
		diff := (cwnd/c.rtt.min - cwnd/float64(rtt)) * c.rtt.min
		//logger.Error("diff :", diff)
		if c.size > 1 && diff > c.b {
			c.size--
		} else if diff < c.a {
			c.size++
		}
	} else if status&CONG_LOSS != 0 {
		if c.size > 1 && c.lossCount <= 0 {
			c.size--
			c.lossCount = c.loss
		}
		c.lossCount--
	}
}

func (c *Vegas2Control) WindowSize() int {
	return c.size
}
