// Package scheduler 内置轻量定时调度引擎。
//
// 设计目标：中小企业无需部署 DolphinScheduler，即可让质量任务 / 归档任务 /
// 物化视图刷新按 cron 定时自动执行；大规模调度场景仍可选 DolphinScheduler。
//
// cron 支持子集（标准 5 段：分 时 日 月 周，NextRun 为纯函数、可单测）：
//   - `*`   任意值（五个字段均支持）
//   - 数字  精确值（分 0-59 / 时 0-23 / 日 1-31 / 月 1-12 / 周 0-6，0 与 7 均为周日）
//   - */n  步进（五个字段均支持，如 */5 = 每 5 个单位一次）
//
// 明确不支持（够用原则，不引第三方 cron 库）：区间 a-b、列表 a,b,c、
// a-b/n、英文月/周名（JAN/SUN）、6 段含秒表达式、@hourly 等宏。
// 日(dow 之外)与周同时受限时，按标准 cron 语义取并集（任一命中即触发）。
package scheduler

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

// DefaultTick 默认巡检间隔：每 30s 检查一轮各任务是否到期
const DefaultTick = 30 * time.Second

// ErrBadCron cron 表达式非法（段数/取值/步进越界等）
var ErrBadCron = errors.New("cron 表达式非法")

// ErrNoMatch 4 年内找不到任何 cron 匹配时间（如 `0 0 31 2 *`：2 月永无 31 日）
var ErrNoMatch = errors.New("4 年内无 cron 匹配时间")

// cronRanges 五个字段的合法取值边界：分 时 日 月 周（周 0=周日，7 归一化为 0）
var cronRanges = [...][2]int{
	{0, 59}, // 分
	{0, 23}, // 时
	{1, 31}, // 日
	{1, 12}, // 月
	{0, 6},  // 周（0=周日）
}

// cronExpr 解析后的 cron：每个字段是"允许取值集合"，全量集合等价于未受限
type cronExpr struct {
	fields [5]map[int]bool
}

// ParseCron 解析 5 段 cron 表达式（仅支持子集：`*`、数字、`*/n`，见包注释）
func ParseCron(expr string) (*cronExpr, error) {
	segs := strings.Fields(expr)
	if len(segs) != 5 {
		return nil, fmt.Errorf("%w: %q 须为 5 段（分 时 日 月 周）", ErrBadCron, expr)
	}
	var e cronExpr
	for i, seg := range segs {
		set, err := parseField(seg, cronRanges[i])
		if err != nil {
			return nil, fmt.Errorf("%w: 第 %d 段 %q: %v", ErrBadCron, i+1, seg, err)
		}
		e.fields[i] = set
	}
	return &e, nil
}

// parseField 解析单段：`*` 全量、`*/n` 步进、数字精确值（周字段 7 归一化为周日）
func parseField(seg string, rg [2]int) (map[int]bool, error) {
	full := func() map[int]bool {
		set := make(map[int]bool, rg[1]-rg[0]+1)
		for v := rg[0]; v <= rg[1]; v++ {
			set[v] = true
		}
		return set
	}
	seg = strings.TrimSpace(seg)
	switch {
	case seg == "*":
		return full(), nil
	case strings.HasPrefix(seg, "*/"):
		// 步进：从字段最小值起每 n 个单位取一个（*/0 及非数字报错）
		n, err := strconv.Atoi(seg[2:])
		if err != nil || n <= 0 {
			return nil, fmt.Errorf("步进 %q 须为正整数", seg[2:])
		}
		set := make(map[int]bool)
		for v := rg[0]; v <= rg[1]; v += n {
			set[v] = true
		}
		return set, nil
	default:
		// 精确值：仅支持纯数字（区间/列表/英文名均不支持，直接报错）
		v, err := strconv.Atoi(seg)
		if err != nil {
			return nil, fmt.Errorf("仅支持 * / 数字 / */n，收到 %q", seg)
		}
		if rg[0] == 0 && rg[1] == 6 && v == 7 { // 周字段的 7 视为周日
			v = 0
		}
		if v < rg[0] || v > rg[1] {
			return nil, fmt.Errorf("取值 %d 超出范围 [%d,%d]", v, rg[0], rg[1])
		}
		return map[int]bool{v: true}, nil
	}
}

// match 判断时刻 t 是否命中 cron（仅比较到分钟，秒忽略）
func (e *cronExpr) match(t time.Time) bool {
	if !e.fields[0][t.Minute()] || !e.fields[1][t.Hour()] || !e.fields[3][int(t.Month())] {
		return false
	}
	// 全量集合 = 未受限；日与周同时受限时按标准 cron 语义取并集
	domAny := len(e.fields[2]) == 31
	dowAny := len(e.fields[4]) == 7
	domOK := e.fields[2][t.Day()]
	dowOK := e.fields[4][int(t.Weekday())]
	switch {
	case domAny && dowAny:
		return true
	case domAny:
		return dowOK
	case dowAny:
		return domOK
	default:
		return domOK || dowOK
	}
}

// nextAfter 返回严格晚于 after 的最早 cron 匹配分钟。
// 逐分钟向前暴力搜索（上限 4 年）；一次调用最多约 210 万次轻量判断，开销可忽略。
func (e *cronExpr) nextAfter(after time.Time) (time.Time, error) {
	t := time.Date(after.Year(), after.Month(), after.Day(),
		after.Hour(), after.Minute(), 0, 0, after.Location()).Add(time.Minute)
	limit := after.AddDate(4, 0, 0)
	for ; t.Before(limit); t = t.Add(time.Minute) {
		if e.match(t) {
			return t, nil
		}
	}
	return time.Time{}, ErrNoMatch
}

// NextRun 计算任务下次应执行的时间（纯函数）：
//   - lastRun 为零值（从未执行）视为立即到期，直接返回 now；
//   - 否则返回严格晚于 lastRun 的最早 cron 匹配时间（可能早于/晚于 now）；
//
// cron 先行校验：表达式非法时无论 lastRun 是否为零值都返回错误，
// 保证"坏任务"在上游就能被识别并跳过。调用方用 Due（或自行比较
// NextRun 结果 <= now）判断是否到期。
func NextRun(cron string, now, lastRun time.Time) (time.Time, error) {
	e, err := ParseCron(cron)
	if err != nil {
		return time.Time{}, err
	}
	if lastRun.IsZero() {
		return now, nil
	}
	return e.nextAfter(lastRun)
}

// Due 判断任务此刻是否到期：NextRun 结果 <= now 即到期。
// cron 解析失败返回错误，调用方应跳过该任务并记日志（不拖垮其他任务）。
func Due(cron string, now time.Time, lastRun *time.Time) (bool, error) {
	var base time.Time
	if lastRun != nil {
		base = *lastRun
	}
	next, err := NextRun(cron, now, base)
	if err != nil {
		return false, err
	}
	return !next.After(now), nil
}

// Task 一个可调度的任务定义
type Task struct {
	Name    string                          // 全局唯一名（如 quality:job:3），重复注册忽略后到者
	Cron    string                          // 5 段 cron 表达式（子集见包注释）
	LastRun *time.Time                      // 上次执行时间；nil 视为从未执行，启动即到期一次
	Fn      func(ctx context.Context) error // 到期执行的业务动作（内部已 recover，勿自行吞 panic 外泄）
}

// taskState 任务运行态：注册时定稿 cron 解析结果；running 原子标记防同任务并发重入
type taskState struct {
	task    Task
	expr    *cronExpr
	running atomic.Bool
}

// Scheduler 轻量调度器：单实例运行（分布式锁后续接入），每 tick 巡检一轮
type Scheduler struct {
	tick  time.Duration
	mu    sync.Mutex
	tasks map[string]*taskState
	wg    sync.WaitGroup // 等待在跑任务收尾，保证优雅退出
}

// New 创建调度器；tick <= 0 时取默认 30s（测试可传毫秒级 tick 加速）
func New(tick time.Duration) *Scheduler {
	if tick <= 0 {
		tick = DefaultTick
	}
	return &Scheduler{tick: tick, tasks: make(map[string]*taskState)}
}

// Register 注册任务。cron 解析失败：记 Errorf 并返回错误、跳过该任务（不拖垮其他任务）；
// 同名任务已存在：忽略后到者（支持周期性重扫表的重复注册）。
func (s *Scheduler) Register(t Task) (bool, error) {
	expr, err := ParseCron(t.Cron)
	if err != nil {
		logx.Errorf("[scheduler] 任务 %s 注册失败，已跳过: %v", t.Name, err)
		return false, err
	}
	if t.Fn == nil {
		err := fmt.Errorf("任务 %s 缺少执行函数", t.Name)
		logx.Errorf("[scheduler] %v", err)
		return false, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.tasks[t.Name]; ok {
		return false, nil
	}
	s.tasks[t.Name] = &taskState{task: t, expr: expr}
	return true, nil
}

// Len 已注册任务数
func (s *Scheduler) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.tasks)
}

// Run 巡检循环：启动即先巡检一次（lastRun 为空的任务立即到期执行一次），
// 之后每 tick（默认 30s）一轮；ctx 取消后等待在跑任务收尾再退出。
func (s *Scheduler) Run(ctx context.Context) {
	logx.Infof("[scheduler] 启动：任务数=%d tick=%s", s.Len(), s.tick)
	ticker := time.NewTicker(s.tick)
	defer ticker.Stop()
	s.tickOnce(ctx)
	for {
		select {
		case <-ctx.Done():
			s.wg.Wait()
			logx.Infof("[scheduler] 已退出（在跑任务均已收尾）")
			return
		case <-ticker.C:
			s.tickOnce(ctx)
		}
	}
}

// tickOnce 单轮巡检：遍历任务 → 判到期 → 原子抢运行标记 → goroutine 执行
func (s *Scheduler) tickOnce(ctx context.Context) {
	now := time.Now()
	for _, st := range s.snapshot() {
		due, err := Due(st.task.Cron, now, st.task.LastRun)
		if err != nil {
			logx.Errorf("[scheduler] 任务 %s cron 非法，本轮跳过: %v", st.task.Name, err)
			continue
		}
		if !due {
			continue
		}
		if !st.running.CompareAndSwap(false, true) {
			logx.Infof("[scheduler] 任务 %s 上一轮仍在运行，本轮跳过", st.task.Name)
			continue
		}
		// 触发即推进 lastRun：即使 Fn 长时间运行，下一轮也不会重复触发
		start := now
		st.task.LastRun = &start
		s.wg.Add(1)
		go s.exec(ctx, st, start)
	}
}

// exec 单次任务执行：panic recover 兜底，成功/失败均记日志，结束释放运行标记
func (s *Scheduler) exec(ctx context.Context, st *taskState, start time.Time) {
	defer s.wg.Done()
	defer func() {
		if r := recover(); r != nil {
			logx.Errorf("[scheduler] 任务 %s panic（已恢复，不影响后续调度）: %v\n%s",
				st.task.Name, r, debug.Stack())
		}
		st.running.Store(false)
	}()
	logx.Infof("[scheduler] 触发任务 %s（cron=%s，计划时间=%s）",
		st.task.Name, st.task.Cron, start.Format(time.DateTime))
	if err := st.task.Fn(ctx); err != nil {
		logx.Errorf("[scheduler] 任务 %s 执行失败: %v", st.task.Name, err)
		return
	}
	logx.Infof("[scheduler] 任务 %s 执行完成", st.task.Name)
}

// snapshot 复制当前任务列表（与 Register 并发安全）
func (s *Scheduler) snapshot() []*taskState {
	s.mu.Lock()
	defer s.mu.Unlock()
	states := make([]*taskState, 0, len(s.tasks))
	for _, st := range s.tasks {
		states = append(states, st)
	}
	return states
}
