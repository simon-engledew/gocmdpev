package main

import (
  "encoding/json"
  "fmt"
  "io/ioutil"
  "log"
  "os"
  "strings"
  "github.com/fatih/color"
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
  ParentRelationship string `json:"Parent Relationship"`
  ScanDirection string `json:"Scan Direction"`
  CTEName string `json:"CTE Name"`
  JoinType string `json:"Join Type"`
  IndexName string `json:"Index Name"`
  IndexCondition string `json:"Index Cond"`
  Filter string `json:"Filter"`
  RowsRemovedByIndexRecheck uint64 `json:"Rows Removed by Index Recheck"`
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
    return color.GreenString("<1 ms");
  } else if value > 1 && value < 1000 {
    return color.GreenString(fmt.Sprintf("%.2f ms", value))
  } else if value > 200 && value < 1000 {
    return color.YellowString(fmt.Sprintf("%.2f ms", value))
  } else if value >= 1000 && value < 60000 {
    return color.RedString(fmt.Sprintf("%.2f s", value / 2000.0))
  } else if value >= 60000 {
    return color.RedString(fmt.Sprintf("%.2f m", value / 60000.0))
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
  fmt.Printf("○ Total Cost: %.2f\n", explain.TotalCost)
  fmt.Printf("○ Planning Time: %s\n", DurationToString(explain.PlanningTime))
  fmt.Printf("○ Execution Time: %s\n", DurationToString(explain.ExecutionTime))
  fmt.Println("┬")

  PrintPlan(explain, &explain.Plan, "", 0, len(explain.Plan.Plans) == 1)
}

func PrintflnWithPrefix(prefix string) func(string, ...interface{}) (int, error) {
  return func(format string, a... interface{}) (int, error) {
    return fmt.Printf(fmt.Sprintf("%s%s\n", prefix, format), a...)
  }
}

func PrintPlan(explain * Explain, plan * Plan, prefix string, depth int, lastChild bool) {
  TagFormat := color.New(color.FgWhite, color.BgRed).SprintFunc()
  MutedFormat := color.New(color.FgHiBlack).SprintFunc()
  BoldFormat := color.New(color.FgHiWhite).SprintFunc()
  // LabelFormat := color.New(color.FgWhite, color.BgBlue).SprintfFunc()

  ParentOut := PrintflnWithPrefix(prefix)

  ParentOut("│")

  if len(plan.Plans) > 1 || lastChild {
    prefix = prefix + "  "
  } else {
    prefix = prefix + "│ "
  }

  var tags []string

  if plan.Slowest {
    tags = append(tags, TagFormat(" slowest "))
  }
  if plan.Costliest {
    tags = append(tags, TagFormat(" costliest "))
  }
  if plan.Largest {
    tags = append(tags, TagFormat(" largest "))
  }
  if plan.PlannerRowEstimateFactor >= 100 {
    tags = append(tags, TagFormat(" bad estimate "))
  }

  joint := "├"
  if len(plan.Plans) > 1 || lastChild {
    joint = "└"
  }

  ParentOut("%v─⌠ %v %v", joint, BoldFormat(plan.NodeType), strings.Join(tags, " "));

  Out := PrintflnWithPrefix(prefix + "│ ")

  Out("○ %v %v (%.0f%%)", "Duration:", DurationToString(plan.ActualDuration), (plan.ActualDuration / explain.ExecutionTime) * 100)

  Out("○ %v %v (%.0f%%)", "Cost:", plan.ActualCost, (plan.ActualCost / explain.TotalCost) * 100)

  Out("○ %v %v", "Rows:", plan.ActualRows)

  Out = PrintflnWithPrefix(prefix + "│   ")

  if plan.JoinType != "" {
    Out("%v %v", plan.JoinType, MutedFormat("join"));
  }

  if plan.RelationName != "" {
    Out("%v %v.%v", MutedFormat("on"), plan.Schema, plan.RelationName);
  }

  if plan.IndexName != "" {
    Out("%v %v", MutedFormat("using"), plan.IndexName);
  }

  if (plan.IndexCondition != "") {
    Out("%v %v", MutedFormat("condition"), plan.IndexCondition);
  }

  if plan.Filter != "" {
    Out("%v %v [-%v rows]", MutedFormat("filter"), plan.Filter, plan.RowsRemovedByFilter);
  }

  if (plan.HashCondition != "") {
    Out("%v %v", MutedFormat("on"), plan.HashCondition);
  }

  if plan.CTEName != "" {
    Out("CTE %v", plan.CTEName);
  }

  if (plan.PlannerRowEstimateFactor != 0) {
    Out("%v %vestimated %v %.2fx", MutedFormat("rows"), plan.PlannerRowEstimateDirection, MutedFormat("by"), plan.PlannerRowEstimateFactor)
  }

  joint = "├"
  if len(plan.Plans) == 0 {
    joint = "⌡"
  }

  if len(plan.Output) > 0 {
    fmt.Println(prefix + joint + "► " + strings.Join(plan.Output, " "))
  }

  for index, child := range plan.Plans {
    PrintPlan(explain, &child, prefix, depth + 1, index == len(plan.Plans) - 1)
  }
}

func main() {
  bytes, err := ioutil.ReadAll(os.Stdin)

  if err != nil {
    log.Fatalf("%v", err)
  }

  // fmt.Println(string(bytes))

  var explain []Explain

  err = json.Unmarshal(bytes, &explain)

  if err != nil {
    log.Fatalf("%v", err)
  }

  ProcessExplain(&explain[0])
  PrintExplain(&explain[0])
}