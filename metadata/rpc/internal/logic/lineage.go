package logic

import (
	"strconv"

	"datacenter/metadata/model"
)

// LineageEdgeData 血缘边的最小数据结构（与存储层解耦，便于纯函数单测）
type LineageEdgeData struct {
	FromId    int64
	ToId      int64
	PropsJson string
}

// defaultLineageDepth 血缘遍历默认深度（depth<=0 时按 1 跳处理）
const defaultLineageDepth = 1

// traverseLineage 从 start 出发按 direction 做 BFS，最多 depth 层。
// 参数 next(id) 返回与 id 直连的边：
//   - upstream：以 id 为终点(to_id)的边，沿 from_id 反向继续；
//   - downstream：以 id 为起点(from_id)的边，沿 to_id 正向继续。
//
// 返回：
//   - nodes：去重后的节点 id（含起点，按访问顺序）；
//   - edges：途经的边（同一条边只输出一次）；
//
// visited 集合天然完成环检测，环路不会导致死循环，也不会重复输出节点/边。
func traverseLineage(start int64, direction string, depth int,
	next func(int64) ([]LineageEdgeData, error)) ([]int64, []LineageEdgeData, error) {
	if depth <= 0 {
		depth = defaultLineageDepth
	}
	visited := map[int64]struct{}{start: {}}
	nodes := []int64{start}
	edges := make([]LineageEdgeData, 0)
	edgeSeen := make(map[string]struct{})
	frontier := []int64{start}

	for level := 0; level < depth && len(frontier) > 0; level++ {
		nextFrontier := make([]int64, 0, len(frontier))
		for _, cur := range frontier {
			neighbors, err := next(cur)
			if err != nil {
				return nil, nil, err
			}
			for _, e := range neighbors {
				// 同一条边只处理一次（防止邻接数据重复输出）
				key := edgeKey(e.FromId, e.ToId)
				if _, ok := edgeSeen[key]; ok {
					continue
				}
				edgeSeen[key] = struct{}{}
				edges = append(edges, e)

				// 计算下一跳节点：upstream 取边的起点，downstream 取边的终点
				nextId := e.ToId
				if direction == model.DirectionUpstream {
					nextId = e.FromId
				}
				// 环检测：已访问节点不再入队
				if _, ok := visited[nextId]; ok {
					continue
				}
				visited[nextId] = struct{}{}
				nodes = append(nodes, nextId)
				nextFrontier = append(nextFrontier, nextId)
			}
		}
		frontier = nextFrontier
	}
	return nodes, edges, nil
}

// edgeKey 边去重键：from->to
func edgeKey(from, to int64) string {
	return strconv.FormatInt(from, 10) + "->" + strconv.FormatInt(to, 10)
}
