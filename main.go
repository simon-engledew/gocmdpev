package main

import (
  "fmt"
  "strings"
  "log"
  "encoding/json"
  "io/ioutil"
  "os"
)

type Direction string
const (
  Over Direction = "Over"
  Under = "Under"
)

type NodeType string
const (
  Limit NodeType = "Limit"
  Sort = "Sort"
  NestedLoop = "Nested Loop"
  MergeJoin = "Merge Join"
  Hash = "Hash"
  HashJoin = "Hash Join"
  Aggregate = "Aggregate"
  Hashaggregate = "Hashaggregate"
  SequenceScan = "Seq Scan"
  IndexScan = "Index Scan"
  IndexOnlyScan = "Index Only Scan"
  BitmapHeapScan = "Bitmap Heap Scan"
  BitmapIndexScan = "Bitmap Index Scan"
  CTEScan = "CTE Scan"
)

type Explain struct {
  Plan Plan `json:"Plan"`
  PlanningTime float64 `json:"Planning Time"`
  Triggers []interface{} `json:"Triggers"`
  ExecutionTime float64 `json:"Execution Time"`
  TotalCost float64
  MaxRows uint64
  MaxCost float64
  MaxDuration float64
}

type Plan struct {
  NodeType NodeType `json:"Node Type"`
  StartupCost float64 `json:"Startup Cost"`
  TotalCost float64 `json:"Total Cost"`
  PlanRows uint64 `json:"Plan Rows"`
  PlanWidth uint64 `json:"Plan Width"`
  Alias string `json:"Alias"`
  Schema string `json:"Schema"`
  RelationName string `json:"Relation Name"`
  RowsRemovedByIndexRecheck uint64 `json:"Rows Removed by Index Recheck"`
  ParentRelationship string `json:"Parent Relationship"`
  ScanDirection string `json:"Scan Direction"`
  CTEName string `json:"CTE Name"`
  JoinType string `json:"Join Type"`
  IndexName string `json:"Index Name"`
  IndexCond string `json:"Index Cond"`
  Filter string `json:"Filter"`
  RowsRemovedByFilter uint64 `json:"Rows Removed by Filter"`
  HashCondition string `json:"Hash Cond"`
  HeapFetches uint64 `json:"Heap Fetches"`
  ActualStartupTime float64 `json:"Actual Startup Time"`
  ActualTotalTime float64 `json:"Actual Total Time"`
  ActualDuration float64
  ActualCost float64
  ActualRows uint64 `json:"Actual Rows"`
  ActualLoops uint64 `json:"Actual Loops"`
  Output []string `json:"Output"`
  SharedHitBlocks uint64 `json:"Shared Hit Blocks"`
  SharedReadBlocks uint64 `json:"Shared Read Blocks"`
  SharedDirtiedBlocks uint64 `json:"Shared Dirtied Blocks"`
  SharedWrittenBlocks uint64 `json:"Shared Written Blocks"`
  LocalHitBlocks uint64 `json:"Local Hit Blocks"`
  LocalReadBlocks uint64 `json:"Local Read Blocks"`
  LocalDirtiedBlocks uint64 `json:"Local Dirtied Blocks"`
  LocalWrittenBlocks uint64 `json:"Local Written Blocks"`
  TempReadBlocks uint64 `json:"Temp Read Blocks"`
  TempWrittenBlocks uint64 `json:"Temp Written Blocks"`
  IOReadTime float64 `json:"I/O Read Time"`
  IOWriteTime float64 `json:"I/O Write Time"`
  PlannerRowEstimateFactor float64
  PlannerRowEstimateDirection Direction
  Costliest bool
  Largest bool
  Slowest bool
  Plans []Plan `json:"Plans"`
}

func calculatePlannerEstimate(explain * Explain, plan * Plan) {
  plan.PlannerRowEstimateFactor = 0
  plan.PlannerRowEstimateDirection = Under;
  if plan.PlanRows != 0 {
    plan.PlannerRowEstimateFactor = float64(plan.ActualRows) / float64(plan.PlanRows);
  }

  if (plan.PlannerRowEstimateFactor < 1.0) {
    plan.PlannerRowEstimateFactor = 0
    plan.PlannerRowEstimateDirection = Over;
    if plan.ActualRows != 0 {
      plan.PlannerRowEstimateFactor = float64(plan.PlanRows) / float64(plan.ActualRows);
    }
  }
}

func calculateActuals(explain * Explain, plan * Plan) {
  plan.ActualDuration = plan.ActualTotalTime
  plan.ActualCost = plan.TotalCost

  for _, child := range plan.Plans {
    if child.NodeType != CTEScan {
      plan.ActualDuration = plan.ActualDuration - child.ActualTotalTime
      plan.ActualCost = plan.ActualCost - child.TotalCost
    }
  }

  if (plan.ActualCost < 0) {
    plan.ActualCost = 0
  }

  explain.TotalCost = explain.TotalCost + plan.ActualCost

  plan.ActualDuration = plan.ActualDuration * float64(plan.ActualLoops)
}

func findOutlierNodes(explain * Explain, plan * Plan) {
  if plan.ActualCost == explain.MaxCost {
    plan.Costliest = true
  }
  if plan.ActualRows == explain.MaxRows {
    plan.Largest = true
  }
  if plan.ActualDuration == explain.MaxDuration {
    plan.Slowest = true
  }

  for index, _ := range plan.Plans {
    findOutlierNodes(explain, &plan.Plans[index])
  }
}

func calculateMaximums(explain * Explain, plan * Plan) {
  if explain.MaxRows < plan.ActualRows {
    explain.MaxRows = plan.ActualRows
  }
  if explain.MaxCost < plan.ActualCost {
    explain.MaxCost = plan.ActualCost
  }
  if explain.MaxDuration < plan.ActualDuration {
    explain.MaxDuration = plan.ActualDuration
  }
}

func DurationToString(value float64) (string) {
  if value < 1 {
    return "<1 ms";
  } else if value > 1 && value < 1000 {
    return fmt.Sprintf("%.2f ms", value)
  } else if value > 200 && value < 1000 {
    return fmt.Sprintf("%.2f ms", value)
  } else if value >= 1000 && value < 60000 {
    return fmt.Sprintf("%.2f s", value / 2000.0)
  } else if value >= 60000 {
    return fmt.Sprintf("%.2f m", value / 60000.0)
  }
  return fmt.Sprintf("%f", value)
}

func ProcessExplain(explain * Explain) {
  ProcessPlan(explain, &explain.Plan)
  findOutlierNodes(explain, &explain.Plan)
}

func ProcessPlan(explain * Explain, plan * Plan) {
  calculatePlannerEstimate(explain, plan)
  calculateActuals(explain, plan)
  calculateMaximums(explain, plan)

  for index, _ := range plan.Plans {
    ProcessPlan(explain, &plan.Plans[index])
  }
}

func PrintExplain(explain * Explain) {
  fmt.Printf("Total Cost: %.2f\n", explain.TotalCost)
  fmt.Printf("Planning Time: %s\n", DurationToString(explain.PlanningTime))
  fmt.Printf("Execution Time: %s\n", DurationToString(explain.ExecutionTime))

  PrintPlan(explain, &explain.Plan, 0)
}

func WithIndent(indent int) func(string, ...interface{}) (int, error) {
  return func(format string, a... interface{}) (int, error) {
    return fmt.Printf(fmt.Sprintf("%s%s\n", strings.Repeat("\t", indent), format), a...)
  }
}

func PrintPlan(explain * Explain, plan * Plan, indent int) {
  fmt.Println()

  Out := WithIndent(indent)

  Out("%v", plan.NodeType);

  Out("%v", DurationToString(plan.ActualDuration))

  Out("%v (%.0f%%)", plan.ActualCost, (plan.ActualCost / explain.TotalCost) * 100)

  if plan.JoinType != "" {
    Out("%v join", plan.JoinType);
  }

  if plan.RelationName != "" {
    Out("on %v.%v", plan.Schema, plan.RelationName);
  }

  if plan.IndexName != "" {
    Out("using %v", plan.IndexName);
  }

  if (plan.HashCondition != "") {
    Out("on %v", plan.HashCondition);
  }

  if plan.CTEName != "" {
    Out("CTE %v", plan.CTEName);
  }

  if (plan.PlannerRowEstimateFactor != 0) {
    Out("%v estimated rows by %.2fx", plan.PlannerRowEstimateDirection, plan.PlannerRowEstimateFactor)
  }

  var tags []string

  if plan.Slowest {
    tags = append(tags, "slowest")
  }
  if plan.Costliest {
    tags = append(tags, "costliest")
  }
  if plan.Largest {
    tags = append(tags, "largest")
  }
  if plan.PlannerRowEstimateFactor >= 100 {
    tags = append(tags, "bad estimate")
  }

  if (len(tags) > 0) {
    Out(strings.Join(tags, " "))
  }

  for _, child := range plan.Plans {
    PrintPlan(explain, &child, indent + 1)
  }
}

func main() {
  bytes, err := ioutil.ReadAll(os.Stdin)

  if err != nil {
    log.Fatalf("%v", err)
  }

  var explain []Explain

  err = json.Unmarshal(bytes, &explain)

  if err != nil {
    log.Fatalf("%v", err)
  }

  ProcessExplain(&explain[0])
  PrintExplain(&explain[0])
}