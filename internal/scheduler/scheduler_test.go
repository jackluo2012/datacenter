package scheduler

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

// base 固定基准时刻：2026-10-11 12:34:56 UTC（周日），避开整点便于断言
func base() time.Time {
	return time.Date(2026, 10, 11, 12, 34, 56, 0, time.UTC)
}

// mustNext 调 NextRun 并断言无错
func mustNext(t *testing.T, cron string, now, lastRun time.Time) time.Time {
	t.Helper()
	next, err := NextRun(cron, now, lastRun)
	if err != nil {
		t.Fatalf("NextRun(%q) 报错: %v", cron, err)
	}
	return next
}

// waitUntil 轮询等待条件成立（最多 2s），超时 fail
func waitUntil(t *testing.T, cond func() bool, msg string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("等待超时: %s", msg)
}

func TestNextRunEveryMinute(t *testing.T) {
	// * * * * *：下一分钟整
	next := mustNext(t, "* * * * *", base(), base())
	want := time.Date(2026, 10, 11, 12, 35, 0, 0, time.UTC)
	if !next.Equal(want) {
		t.Fatalf("* * * * * 期望 %s，得到 %s", want, next)
	}
}

func TestNextRunDaily(t *testing.T) {
	// 0 2 * * *：今天 02:00 已过（lastRun 12:00）→ 明天 02:00
	next := mustNext(t, "0 2 * * *", base(), base())
	want := time.Date(2026, 10, 12, 2, 0, 0, 0, time.UTC)
	if !next.Equal(want) {
		t.Fatalf("每日 02:00 期望 %s，得到 %s", want, next)
	}
	// lastRun 01:00 → 今天 02:00
	last := time.Date(2026, 10, 11, 1, 0, 0, 0, time.UTC)
	next = mustNext(t, "0 2 * * *", base(), last)
	want = time.Date(2026, 10, 11, 2, 0, 0, 0, time.UTC)
	if !next.Equal(want) {
		t.Fatalf("每日 02:00（lastRun 01:00）期望 %s，得到 %s", want, next)
	}
	// 秒/毫秒应被截断：lastRun 02:00:00.500 → 下一次是明天 02:00:00
	last = time.Date(2026, 10, 11, 2, 0, 0, 500_000_000, time.UTC)
	next = mustNext(t, "0 2 * * *", base(), last)
	want = time.Date(2026, 10, 12, 2, 0, 0, 0, time.UTC)
	if !next.Equal(want) {
		t.Fatalf("每日 02:00（lastRun 含毫秒）期望 %s，得到 %s", want, next)
	}
}

func TestNextRunStepEvery5Minutes(t *testing.T) {
	cases := []struct {
		last time.Time
		want time.Time
	}{
		// */5 * * * *：对齐到 0/5/10/.../55
		{time.Date(2026, 10, 11, 12, 2, 0, 0, time.UTC), time.Date(2026, 10, 11, 12, 5, 0, 0, time.UTC)},
		{time.Date(2026, 10, 11, 12, 5, 0, 0, time.UTC), time.Date(2026, 10, 11, 12, 10, 0, 0, time.UTC)},
		{time.Date(2026, 10, 11, 12, 59, 0, 0, time.UTC), time.Date(2026, 10, 11, 13, 0, 0, 0, time.UTC)},
		{time.Date(2026, 10, 11, 23, 56, 0, 0, time.UTC), time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)},
	}
	for i, c := range cases {
		next := mustNext(t, "*/5 * * * *", base(), c.last)
		if !next.Equal(c.want) {
			t.Fatalf("case %d: 期望 %s，得到 %s", i, c.want, next)
		}
	}
	// 小时/日/月步进抽查
	next := mustNext(t, "0 */6 * * *", base(), base()) // 12:34 → 18:00
	want := time.Date(2026, 10, 11, 18, 0, 0, 0, time.UTC)
	if !next.Equal(want) {
		t.Fatalf("0 */6 * * * 期望 %s，得到 %s", want, next)
	}
	next = mustNext(t, "0 0 */2 * *", base(), time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)) // 日步进自最小值 1 起：1,3,5,... → 10/2 → 10/3
	want = time.Date(2026, 10, 3, 0, 0, 0, 0, time.UTC)
	if !next.Equal(want) {
		t.Fatalf("0 0 */2 * * 期望 %s，得到 %s", want, next)
	}
	next = mustNext(t, "0 0 1 */3 *", base(), base()) // 月步进自最小值 1 起：1,4,7,10 → 10/11 已过 → 明年 1 月 1 日
	want = time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)
	if !next.Equal(want) {
		t.Fatalf("0 0 0 */3 * 期望 %s，得到 %s", want, next)
	}
}

func TestNextRunWeekdayAndMonth(t *testing.T) {
	// 30 8 * * 1：周一 08:30。2026-10-11 是周日 → 明天周一
	next := mustNext(t, "30 8 * * 1", base(), base())
	want := time.Date(2026, 10, 12, 8, 30, 0, 0, time.UTC)
	if !next.Equal(want) {
		t.Fatalf("每周一 08:30 期望 %s，得到 %s", want, next)
	}
	// 周字段 7 视作周日：2026-10-11（周日）12:34:56 → 12:35
	next = mustNext(t, "* * * * 7", base(), base())
	want = time.Date(2026, 10, 11, 12, 35, 0, 0, time.UTC)
	if !next.Equal(want) {
		t.Fatalf("周字段 7 期望 %s，得到 %s", want, next)
	}
	// 0 0 1 1 *：每年 1 月 1 日
	next = mustNext(t, "0 0 1 1 *", base(), base())
	want = time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)
	if !next.Equal(want) {
		t.Fatalf("每年元旦 期望 %s，得到 %s", want, next)
	}
	// 日+周同时受限（标准 cron 取并集）：0 0 13 * 1 = 每月 13 号或周一
	// 基准 10-11 周日 → 下一个命中是周一 10-12
	next = mustNext(t, "0 0 13 * 1", base(), base())
	want = time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)
	if !next.Equal(want) {
		t.Fatalf("日+周并集 期望 %s，得到 %s", want, next)
	}
	// 10-13（周二）00:01 之后 → 下一个命中是周一 10-19
	last := time.Date(2026, 10, 13, 0, 1, 0, 0, time.UTC)
	next = mustNext(t, "0 0 13 * 1", base(), last)
	want = time.Date(2026, 10, 19, 0, 0, 0, 0, time.UTC)
	if !next.Equal(want) {
		t.Fatalf("日+周并集（跨周） 期望 %s，得到 %s", want, next)
	}
}

func TestNextRunZeroLastRun(t *testing.T) {
	now := base()
	// lastRun 为零值（从未跑过）→ 立即到期，返回 now 本身
	next, err := NextRun("0 2 * * *", now, time.Time{})
	if err != nil {
		t.Fatalf("报错: %v", err)
	}
	if !next.Equal(now) {
		t.Fatalf("lastRun 为零值应返回 now %s，得到 %s", now, next)
	}
}

func TestNextRunFutureLastRun(t *testing.T) {
	// lastRun 在未来（如时钟回拨/人工预置）→ 下次时间更远，不应到期
	future := base().Add(2 * time.Hour)
	next := mustNext(t, "* * * * *", base(), future)
	if !next.After(base()) {
		t.Fatalf("lastRun 在未来时 next 应晚于 now，得到 %s", next)
	}
	due, err := Due("* * * * *", base(), &future)
	if err != nil || due {
		t.Fatalf("lastRun 在未来不应到期：due=%v err=%v", due, err)
	}
}

func TestDue(t *testing.T) {
	now := base()
	// lastRun 为 nil → 立即到期一次
	if due, err := Due("0 2 * * *", now, nil); err != nil || !due {
		t.Fatalf("lastRun 为 nil 应立即到期：due=%v err=%v", due, err)
	}
	// lastRun 昨天 02:00（每日任务）→ 今天已过期 → 到期
	yesterday := now.AddDate(0, 0, -1)
	yesterday = time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 2, 0, 0, 0, time.UTC)
	if due, err := Due("0 2 * * *", now, &yesterday); err != nil || !due {
		t.Fatalf("lastRun 在过去应到期：due=%v err=%v", due, err)
	}
	// lastRun 今天 02:00（每日任务）→ 下次明天 → 未到期
	today := time.Date(now.Year(), now.Month(), now.Day(), 2, 0, 0, 0, time.UTC)
	if due, err := Due("0 2 * * *", now, &today); err != nil || due {
		t.Fatalf("今日已跑不应到期：due=%v err=%v", due, err)
	}
	// cron 非法 → 返回错误
	if _, err := Due("bad cron", now, nil); !errors.Is(err, ErrBadCron) {
		t.Fatalf("cron 非法应返回 ErrBadCron，得到 %v", err)
	}
}

func TestParseCronErrors(t *testing.T) {
	bad := []string{
		"",            // 空
		"* * * *",     // 4 段
		"* * * * * *", // 6 段（不支持含秒）
		"61 * * * *",  // 分越界
		"* 24 * * *",  // 时越界
		"0 0 0 * *",   // 日越界
		"0 0 * 13 *",  // 月越界
		"0 0 * * 8",   // 周越界（仅 0-7）
		"*/0 * * * *", // 步进为 0
		"*/x * * * *", // 步进非数字
		"1-5 * * * *", // 区间不支持
		"1,2 * * * *", // 列表不支持
		"5/2 * * * *", // 基值/步进不支持
		"JAN * * * *", // 英文名不支持
		"@hourly",     // 宏不支持
		"** * * * *",  // 乱写
	}
	for _, expr := range bad {
		if _, err := ParseCron(expr); !errors.Is(err, ErrBadCron) {
			t.Fatalf("ParseCron(%q) 应返回 ErrBadCron，得到 %v", expr, err)
		}
	}
}

func TestNextRunNoMatch(t *testing.T) {
	// 字段均合法（2 月 31 日语法上成立）→ 解析应成功
	if _, err := ParseCron("0 0 31 2 *"); err != nil {
		t.Fatalf("字段均合法，解析不应报错: %v", err)
	}
	// 2 月永无 31 日 → 4 年内无匹配，应返回 ErrNoMatch 而非死循环
	next, err := NextRun("0 0 31 2 *", base(), base())
	if !errors.Is(err, ErrNoMatch) {
		t.Fatalf("期望 ErrNoMatch，得到 next=%s err=%v", next, err)
	}
}

// TestSchedulerRunsAndAdvancesLastRun 到期任务被触发执行，且触发后 lastRun 被推进
func TestSchedulerRunsAndAdvancesLastRun(t *testing.T) {
	var runs int32
	s := New(20 * time.Millisecond)
	if _, err := s.Register(Task{
		Name: "t1",
		Cron: "* * * * *", // 每分钟：首轮立即触发一次，之后靠手工拨 lastRun 再触发
		Fn:   func(context.Context) error { atomic.AddInt32(&runs, 1); return nil },
	}); err != nil {
		t.Fatalf("注册失败: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go s.Run(ctx)

	waitUntil(t, func() bool { return atomic.LoadInt32(&runs) == 1 }, "首轮应立即触发一次")
	st := s.tasks["t1"]
	if st.task.LastRun == nil {
		t.Fatal("触发后 lastRun 应被推进为非 nil")
	}
	time.Sleep(60 * time.Millisecond) // 再过几轮 tick
	if n := atomic.LoadInt32(&runs); n != 1 {
		t.Fatalf("lastRun 已推进，不应重复触发，实际跑了 %d 次", n)
	}
}

// TestSchedulerRecoversPanic 任务 panic 不拖垮调度器，恢复后可再次触发
// （直接驱动 tickOnce，避免测试并发写 lastRun 与 Run 巡检竞争）
func TestSchedulerRecoversPanic(t *testing.T) {
	var panics int32
	s := New(time.Second)
	if _, err := s.Register(Task{
		Name: "boom",
		Cron: "* * * * *",
		Fn: func(context.Context) error {
			atomic.AddInt32(&panics, 1)
			panic("任务内部炸了")
		},
	}); err != nil {
		t.Fatalf("注册失败: %v", err)
	}
	ctx := context.Background()
	s.tickOnce(ctx) // 首轮触发（panic 在 goroutine 内被 recover）

	waitUntil(t, func() bool { return atomic.LoadInt32(&panics) == 1 }, "panic 任务应已执行一次")
	st := s.tasks["boom"]
	// 等 panic goroutine 完全收尾（running 标记释放），再把 lastRun 拨回一小时前使其再次到期
	waitUntil(t, func() bool { return !st.running.Load() }, "running 标记应在收尾后释放")
	oneHourAgo := time.Now().Add(-time.Hour)
	st.task.LastRun = &oneHourAgo
	s.tickOnce(ctx)
	waitUntil(t, func() bool { return atomic.LoadInt32(&panics) == 2 }, "panic 恢复后应可再次触发")
	if st.running.Load() {
		t.Fatal("第二次触发后 running 标记应再次可释放")
	}
}

// TestSchedulerSkipsRunningTask 同一任务运行中，后续到期轮次应跳过（atomic 标记防重入）
func TestSchedulerSkipsRunningTask(t *testing.T) {
	release := make(chan struct{})
	var runs int32
	s := New(time.Second)
	if _, err := s.Register(Task{
		Name: "slow",
		Cron: "* * * * *",
		Fn: func(context.Context) error {
			atomic.AddInt32(&runs, 1)
			<-release // 模拟长任务
			return nil
		},
	}); err != nil {
		t.Fatalf("注册失败: %v", err)
	}
	ctx := context.Background()
	s.tickOnce(ctx) // 首轮触发，Fn 阻塞中
	waitUntil(t, func() bool { return atomic.LoadInt32(&runs) == 1 }, "任务应已开始执行")

	// 人为拨回 lastRun 使其"到期"，但上一轮仍在运行 → 本轮应跳过（running 标记占用）
	st := s.tasks["slow"]
	oneHourAgo := time.Now().Add(-time.Hour)
	st.task.LastRun = &oneHourAgo
	s.tickOnce(ctx) // 该轮巡检应命中"运行中跳过"分支，不产生第二次执行
	if n := atomic.LoadInt32(&runs); n != 1 {
		t.Fatalf("运行中的任务不应被重复触发，实际执行 %d 次", n)
	}
	if !st.running.Load() {
		t.Fatal("任务执行中 running 标记应为 true")
	}
	close(release) // 放行长任务收尾
	waitUntil(t, func() bool { return !st.running.Load() }, "结束后 running 标记应释放")
	if n := atomic.LoadInt32(&runs); n != 1 {
		t.Fatalf("全程只应执行 1 次，实际 %d 次", n)
	}
}

// TestSchedulerSkipInvalidCronTask cron 解析失败的任务被跳过，不拖垮其他任务
func TestSchedulerSkipInvalidCronTask(t *testing.T) {
	var runs int32
	s := New(time.Second)
	if _, err := s.Register(Task{Name: "bad", Cron: "not a cron", Fn: func(context.Context) error { return nil }}); err == nil {
		t.Fatal("cron 非法的注册应返回错误")
	}
	if s.Len() != 0 {
		t.Fatalf("cron 非法的任务不应入册，实际 %d 个", s.Len())
	}
	if _, err := s.Register(Task{
		Name: "good",
		Cron: "* * * * *",
		Fn:   func(context.Context) error { atomic.AddInt32(&runs, 1); return nil },
	}); err != nil {
		t.Fatalf("合法任务注册失败: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go s.Run(ctx)
	waitUntil(t, func() bool { return atomic.LoadInt32(&runs) == 1 }, "合法任务应正常触发")
}

// TestSchedulerGracefulCancel ctx 取消后等待在跑任务收尾再退出
func TestSchedulerGracefulCancel(t *testing.T) {
	finished := make(chan struct{})
	s := New(time.Second)
	if _, err := s.Register(Task{
		Name: "long",
		Cron: "* * * * *",
		Fn: func(ctx context.Context) error {
			time.Sleep(50 * time.Millisecond)
			close(finished)
			return nil
		},
	}); err != nil {
		t.Fatalf("注册失败: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	go s.Run(ctx)
	waitUntil(t, func() bool { return s.tasks["long"].running.Load() }, "任务应已开始执行")
	cancel()
	select {
	case <-finished:
		// Run 退出前任务正常收尾
	case <-time.After(2 * time.Second):
		t.Fatal("取消后未等待在跑任务收尾")
	}
}
