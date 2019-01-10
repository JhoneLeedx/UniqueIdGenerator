/**
 * @author	Gao Xinyuan<gaoxinyuan@yinhai.com>
 */
package UniqueIdGenerator

import (
	"fmt"
	"github.com/imroc/biu"
	"sync"
	"time"
)

// UniqueId 生成器结构体
type UniqueIdGenerator struct {
	// 机器编号
	workId uint64
	// 移位后的机器编号
	workIdAfterShift uint64
	// 上次生成的时间戳
	lastTimestamp uint64
	// 当前序列号
	currentSequence uint64
	// 初始时间戳
	startTimestamp uint64

	// 时间戳位数
	timestampBitSize uint8
	// 机器编号位数
	workIdBitSize uint8
	// 序列号位数
	sequenceBitSize uint8

	// 同步锁
	lock *sync.Mutex
	// 是否已初始化
	isInited bool

	// 机器编号最大值
	maxWorkId uint64
	// 序列号最大值
	maxSequence uint64
	// workId 左移步长
	workIdLeftShift uint8
	// 时间戳左移步长
	timestampLeftShift uint8
}

func CreateIdGenerator() *UniqueIdGenerator {
	return &UniqueIdGenerator{
		workId:             0,
		lastTimestamp:      0,
		currentSequence:    0,
		timestampBitSize:   41,
		workIdBitSize:      10,
		sequenceBitSize:    12,
		maxWorkId:          0,
		maxSequence:        0,
		workIdLeftShift:    0,
		timestampLeftShift: 0,
		lock:               new(sync.Mutex),
		isInited:           false,
		startTimestamp:     0,
	}
}

// set workId
func (generator *UniqueIdGenerator) SetWorkId(workId uint64) *UniqueIdGenerator {
	generator.lock.Lock()
	defer generator.lock.Unlock()
	generator.isInited = false
	generator.workId = workId
	return generator
}

// set timestampBitSize
func (generator *UniqueIdGenerator) SetTimestampBitSize(size uint8) *UniqueIdGenerator {
	generator.lock.Lock()
	defer generator.lock.Unlock()
	generator.isInited = false
	generator.timestampBitSize = size
	return generator
}

// set workIdBitSize
func (generator *UniqueIdGenerator) SetWorkIdBitSize(size uint8) *UniqueIdGenerator {
	generator.lock.Lock()
	defer generator.lock.Unlock()
	generator.isInited = false
	generator.workIdBitSize = size
	return generator
}

// set sequenceBitSize
func (generator *UniqueIdGenerator) SetSequenceBitSize(size uint8) *UniqueIdGenerator {
	generator.lock.Lock()
	defer generator.lock.Unlock()
	generator.isInited = false
	generator.sequenceBitSize = size
	return generator
}

func (generator *UniqueIdGenerator) checkSettings() bool {
	const min = 1
	const max = 60

	if generator.sequenceBitSize < min || generator.sequenceBitSize > max {
		return false
	}

	if generator.timestampBitSize < min || generator.timestampBitSize > max {
		return false
	}

	if generator.workIdBitSize < min || generator.workIdBitSize > max {
		return false
	}

	var totalSize = generator.sequenceBitSize + generator.timestampBitSize + generator.workIdBitSize
	if totalSize != 64 {
		return false
	}

	return true
}

// 初始化
func (generator *UniqueIdGenerator) Init() *UniqueIdGenerator {
	generator.lock.Lock()
	defer generator.lock.Unlock()

	if generator.isInited {
		return generator
	}
	if !generator.checkSettings() {
		fmt.Println("参数设置错误")
		return nil
	}

	// 计算时间戳和机器编号的位移数
	generator.workIdLeftShift = generator.sequenceBitSize
	generator.timestampLeftShift = generator.sequenceBitSize + generator.workIdLeftShift

	// 计算机器编号和序列号的最大值
	const uint64_MAX = ^uint64(0)
	generator.maxWorkId = ^(uint64_MAX >> generator.workIdBitSize << generator.workIdBitSize)
	generator.maxSequence = ^(uint64_MAX >> generator.sequenceBitSize << generator.sequenceBitSize)

	// workId 移位
	if generator.workId > generator.maxWorkId {
		fmt.Printf("workId 不能大于 %d", generator.maxWorkId)
		return nil
	}
	generator.workIdAfterShift = generator.workId << generator.workIdLeftShift

	generator.isInited = true
	generator.lastTimestamp = 0
	generator.currentSequence = 0
	generator.startTimestamp = generator.getTimestamp()
	return generator
}

func (generator *UniqueIdGenerator) getTimestamp() uint64 {
	// 微秒时间戳
	timestamp := uint64(time.Now().UnixNano())
	// 保留设置的位数
	rest := 64 - generator.timestampBitSize
	return (timestamp >> rest) << rest
}

func (generator *UniqueIdGenerator) createNextTimestamp(last uint64) uint64 {
	for {
		current := generator.getTimestamp()
		if current > last {
			return current
		}
	}
}

func (generator *UniqueIdGenerator) CreateNextId() (uint64, error) {
	generator.lock.Lock()
	defer generator.lock.Unlock()

	if !generator.isInited {
		return 0, fmt.Errorf("CreateNextId 失败:\tUniqueIdGenerator 未初始划，请先调用 Init()")
	}

	currentTimestamp := generator.getTimestamp()
	if currentTimestamp < generator.lastTimestamp {
		// 我们需要优化这里如果系统时钟出现问题的情况
		return 0, fmt.Errorf("CreateNextId failed:\t系统时钟出现回拨")
	}

	if currentTimestamp == generator.lastTimestamp {
		// 与最大值的结果为：若sequence到达了最大值，结果为0
		generator.currentSequence = (generator.currentSequence + 1) & generator.maxSequence
		// 序号已到达最大值，再次获取时间戳
		if generator.currentSequence == 0 {
			currentTimestamp = generator.getTimestamp()
		}
	} else {
		generator.currentSequence = 0
	}

	generator.lastTimestamp = currentTimestamp

	// 时间戳左移至最高位为1，即保证最终结果为20位数字，又保证或后的workID和序列号数据正确
	currentTimestamp = currentTimestamp << 3

	// 组装最后结果
	currentTimestamp = currentTimestamp | generator.currentSequence | generator.workIdAfterShift
	// generator.parseId(currentTimestamp)
	return currentTimestamp, nil
}

func (generator *UniqueIdGenerator) parseId(result uint64) {
	fmt.Printf("\n-------------- 解析 --------------\n")
	fmt.Printf("结果: %d\n", result)
	fmt.Printf("二进制结果: %s\n", biu.ToBinaryString(result))

	var timestamp uint64 = result >> (3 + (generator.workIdBitSize + generator.sequenceBitSize)) << (generator.workIdBitSize + generator.sequenceBitSize)
	fmt.Printf("时间戳: %d\n", timestamp)
	fmt.Printf("二进制时间戳: %s\n", biu.ToBinaryString(timestamp))

	var sequence uint64 = result << (generator.timestampBitSize + generator.workIdBitSize) >> (generator.timestampBitSize + generator.workIdBitSize)
	fmt.Printf("序列号: %d\n", sequence)
	fmt.Printf("序列号时间戳: %s\n", biu.ToBinaryString(sequence))

	var workId uint64 = result << generator.timestampBitSize >> (generator.timestampBitSize + generator.sequenceBitSize)
	fmt.Printf("机器编号: %d\n", workId)
	fmt.Printf("机器编号时间戳: %s\n", biu.ToBinaryString(workId))

	fmt.Printf("\n-------------- 解析完毕 --------------\n")
}

func (generator *UniqueIdGenerator) GetIdByCount(count int) ([]uint64, error) {
	var idSlice []uint64
	for index := 0; index < count; index++ {
		id0, err := generator.CreateNextId()
		if err != nil {
			return nil, err
		} else {
			idSlice = append(idSlice, id0)
		}
	}
	return idSlice, nil
}
