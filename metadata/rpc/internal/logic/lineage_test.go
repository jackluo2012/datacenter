package logic

import (
	"errors"
	"reflect"
	"testing"

	"datacenter/metadata/model"
)

// memoryGraph 内存边集合（单测用，替代存储层）
type memoryGraph struct {
	edges []LineageEdgeData
}

// next 构造与 traverseLineage 匹配的邻居查询闭包（模拟 RelationModel.FindNeighbors）
func (g *memoryGraph) next(direction string) func(int64) ([]LineageEdgeData, error) {
	return func(id int64) ([]LineageEdgeData, error) {
		out := make([]LineageEdgeData, 0)
		for _, e := range g.edges {
			switch direction {
			case model.DirectionDownstream:
				if e.FromId == id {
					out = append(out, e)
				}
			case model.DirectionUpstream:
				if e.ToId == id {
					out = append(out, e)
				}
			}
		}
		return out, nil
	}
}

// mkEdge 构造一条血缘边
func mkEdge(from, to int64) LineageEdgeData {
	return LineageEdgeData{FromId: from, ToId: to, PropsJson: ""}
}

// collectEdgeKeys 把边列表转成 "from->to" 键集合，便于断言
func collectEdgeKeys(edges []LineageEdgeData) map[string]struct{} {
	keys := make(map[string]struct{}, len(edges))
	for _, e := range edges {
		keys[edgeKey(e.FromId, e.ToId)] = struct{}{}
	}
	return keys
}

// 测试下游多跳：1->2->3->4，depth=2 时只展开到第 2 跳
func TestTraverseLineageDownstreamMultiHop(t *testing.T) {
	g := &memoryGraph{edges: []LineageEdgeData{mkEdge(1, 2), mkEdge(2, 3), mkEdge(3, 4)}}

	nodes, edges, err := traverseLineage(1, model.DirectionDownstream, 2, g.next(model.DirectionDownstream))
	if err != nil {
		t.Fatalf("traverseLineage error: %v", err)
	}
	if !reflect.DeepEqual(nodes, []int64{1, 2, 3}) {
		t.Fatalf("nodes = %v, want [1 2 3]", nodes)
	}
	want := map[string]struct{}{"1->2": {}, "2->3": {}}
	if !reflect.DeepEqual(collectEdgeKeys(edges), want) {
		t.Fatalf("edges = %v, want %v", edges, want)
	}
}

// 测试深度足够时走完整条链
func TestTraverseLineageDownstreamFullChain(t *testing.T) {
	g := &memoryGraph{edges: []LineageEdgeData{mkEdge(1, 2), mkEdge(2, 3), mkEdge(3, 4)}}

	nodes, _, err := traverseLineage(1, model.DirectionDownstream, 10, g.next(model.DirectionDownstream))
	if err != nil {
		t.Fatalf("traverseLineage error: %v", err)
	}
	if !reflect.DeepEqual(nodes, []int64{1, 2, 3, 4}) {
		t.Fatalf("nodes = %v, want [1 2 3 4]", nodes)
	}
}

// 测试上游反向遍历：1->2->3->4，从 4 出发沿 from_id 反向找
func TestTraverseLineageUpstreamMultiHop(t *testing.T) {
	g := &memoryGraph{edges: []LineageEdgeData{mkEdge(1, 2), mkEdge(2, 3), mkEdge(3, 4)}}

	nodes, edges, err := traverseLineage(4, model.DirectionUpstream, 2, g.next(model.DirectionUpstream))
	if err != nil {
		t.Fatalf("traverseLineage error: %v", err)
	}
	if !reflect.DeepEqual(nodes, []int64{4, 3, 2}) {
		t.Fatalf("nodes = %v, want [4 3 2]", nodes)
	}
	want := map[string]struct{}{"3->4": {}, "2->3": {}}
	if !reflect.DeepEqual(collectEdgeKeys(edges), want) {
		t.Fatalf("edges = %v, want %v", edges, want)
	}
}

// 测试环检测：1->2->3->1 成环时不能死循环，且节点/边不重复输出
func TestTraverseLineageCycle(t *testing.T) {
	g := &memoryGraph{edges: []LineageEdgeData{mkEdge(1, 2), mkEdge(2, 3), mkEdge(3, 1)}}

	nodes, edges, err := traverseLineage(1, model.DirectionDownstream, 10, g.next(model.DirectionDownstream))
	if err != nil {
		t.Fatalf("traverseLineage error: %v", err)
	}
	if !reflect.DeepEqual(nodes, []int64{1, 2, 3}) {
		t.Fatalf("nodes = %v, want [1 2 3]", nodes)
	}
	if len(edges) != 3 {
		t.Fatalf("edges = %v, want 3 edges", edges)
	}
}

// 测试菱形去重：1->2、1->3、2->4、3->4，节点 4 只出现一次
func TestTraverseLineageDiamondDedup(t *testing.T) {
	g := &memoryGraph{edges: []LineageEdgeData{mkEdge(1, 2), mkEdge(1, 3), mkEdge(2, 4), mkEdge(3, 4)}}

	nodes, edges, err := traverseLineage(1, model.DirectionDownstream, 3, g.next(model.DirectionDownstream))
	if err != nil {
		t.Fatalf("traverseLineage error: %v", err)
	}
	if !reflect.DeepEqual(nodes, []int64{1, 2, 3, 4}) {
		t.Fatalf("nodes = %v, want [1 2 3 4]", nodes)
	}
	if len(edges) != 4 {
		t.Fatalf("edges = %v, want 4 edges", edges)
	}
}

// 测试重复边去重：邻接数据里出现两条相同边时只输出一次
func TestTraverseLineageDuplicateEdge(t *testing.T) {
	g := &memoryGraph{edges: []LineageEdgeData{mkEdge(1, 2), mkEdge(1, 2), mkEdge(2, 3)}}

	nodes, edges, err := traverseLineage(1, model.DirectionDownstream, 2, g.next(model.DirectionDownstream))
	if err != nil {
		t.Fatalf("traverseLineage error: %v", err)
	}
	if !reflect.DeepEqual(nodes, []int64{1, 2, 3}) {
		t.Fatalf("nodes = %v, want [1 2 3]", nodes)
	}
	if len(edges) != 2 {
		t.Fatalf("edges = %v, want 2 edges", edges)
	}
}

// 测试深度缺省：depth<=0 时按 1 跳处理
func TestTraverseLineageDepthDefault(t *testing.T) {
	g := &memoryGraph{edges: []LineageEdgeData{mkEdge(1, 2), mkEdge(2, 3)}}

	nodes, _, err := traverseLineage(1, model.DirectionDownstream, 0, g.next(model.DirectionDownstream))
	if err != nil {
		t.Fatalf("traverseLineage error: %v", err)
	}
	if !reflect.DeepEqual(nodes, []int64{1, 2}) {
		t.Fatalf("nodes = %v, want [1 2]", nodes)
	}
}

// 测试孤立节点：无边可达时只返回起点自身
func TestTraverseLineageIsolatedNode(t *testing.T) {
	g := &memoryGraph{edges: []LineageEdgeData{mkEdge(2, 3)}}

	nodes, edges, err := traverseLineage(1, model.DirectionDownstream, 3, g.next(model.DirectionDownstream))
	if err != nil {
		t.Fatalf("traverseLineage error: %v", err)
	}
	if !reflect.DeepEqual(nodes, []int64{1}) {
		t.Fatalf("nodes = %v, want [1]", nodes)
	}
	if len(edges) != 0 {
		t.Fatalf("edges = %v, want empty", edges)
	}
}

// 测试邻居查询失败时错误向上传递
func TestTraverseLineageNextError(t *testing.T) {
	wantErr := errors.New("db down")
	next := func(int64) ([]LineageEdgeData, error) { return nil, wantErr }

	_, _, err := traverseLineage(1, model.DirectionDownstream, 2, next)
	if !errors.Is(err, wantErr) {
		t.Fatalf("err = %v, want %v", err, wantErr)
	}
}
