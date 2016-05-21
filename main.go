package main

import (
  "encoding/json"
  "fmt"
  "io/ioutil"
  "log"
  "os"
  "io"
  "strings"
  "github.com/fatih/color"
  "github.com/mitchellh/go-wordwrap"
  "github.com/dustin/go-humanize"
)

type EstimateDirection string
const (
  Over EstimateDirection = "Over"
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

var PrefixFormat = color.New(color.FgHiBlack).SprintFunc()
var TagFormat = color.New(color.FgWhite, color.BgRed).SprintFunc()
var MutedFormat = color.New(color.FgHiBlack).SprintFunc()
var BoldFormat = color.New(color.FgHiWhite).SprintFunc()
var GoodFormat = color.New(color.FgGreen).SprintFunc()
var WarningFormat = color.New(color.FgHiYellow).SprintFunc()
var CriticalFormat = color.New(color.FgHiRed).SprintFunc()
var OutputFormat = color.New(color.FgYellow).SprintFunc()
  // LabelFormat := color.New(color.FgWhite, color.BgBlue).SprintfFunc()

var Descriptions = map[NodeType]string {
   Limit: "Returns a specified number of rows from a record set.",
   Sort: "Sorts a record set based on the specified sort key.",
   NestedLoop: "Merges two record sets by looping through every record in the first set and trying to find a match in the second set. All matching records are returned.",
   MergeJoin: "Merges two record sets by first sorting them on a join key.",
   Hash: "Generates a hash table from the records in the input recordset. Hash is used by Hash Join.",
   HashJoin: "Joins to record sets by hashing one of them (using a Hash Scan).",
   Aggregate: "Groups records together based on a GROUP BY or aggregate function (e.g. sum()).",
   Hashaggregate: "Groups records together based on a GROUP BY or aggregate function (e.g. sum()). Hash Aggregate uses a hash to first organize the records by a key.",
   SequenceScan: "Finds relevant records by sequentially scanning the input record set. When reading from a table, Seq Scans (unlike Index Scans) perform a single read operation (only the table is read).",
   IndexScan: "Finds relevant records based on an Index. Index Scans perform 2 read operations: one to read the index and another to read the actual value from the table.",
   IndexOnlyScan: "Finds relevant records based on an Index. Index Only Scans perform a single read operation from the index and do not read from the corresponding table.",
   BitmapHeapScan: "Searches through the pages returned by the Bitmap Index Scan for relevant rows.",
   BitmapIndexScan: "Uses a Bitmap Index (index which uses 1 bit per page) to find all relevant pages. Results of this node are fed to the Bitmap Heap Scan.",
   CTEScan: "Performs a sequential scan of Common Table Expression (CTE) query results. Note that results of a CTE are materialized (calculated and temporarily stored).",
}

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
  ActualCost float64
  ActualDuration float64
  ActualLoops uint64 `json:"Actual Loops"`
  ActualRows uint64 `json:"Actual Rows"`
  ActualStartupTime float64 `json:"Actual Startup Time"`
  ActualTotalTime float64 `json:"Actual Total Time"`
  Alias string `json:"Alias"`
  Costliest bool
  CTEName string `json:"CTE Name"`
  Filter string `json:"Filter"`
  GroupKey []string `json:"Group Key"`
  HashCondition string `json:"Hash Cond"`
  HeapFetches uint64 `json:"Heap Fetches"`
  IndexCondition string `json:"Index Cond"`
  IndexName string `json:"Index Name"`
  IOReadTime float64 `json:"I/O Read Time"`
  IOWriteTime float64 `json:"I/O Write Time"`
  JoinType string `json:"Join Type"`
  Largest bool
  LocalDirtiedBlocks uint64 `json:"Local Dirtied Blocks"`
  LocalHitBlocks uint64 `json:"Local Hit Blocks"`
  LocalReadBlocks uint64 `json:"Local Read Blocks"`
  LocalWrittenBlocks uint64 `json:"Local Written Blocks"`
  NodeType NodeType `json:"Node Type"`
  Output []string `json:"Output"`
  ParentRelationship string `json:"Parent Relationship"`
  PlannerRowEstimateDirection EstimateDirection
  PlannerRowEstimateFactor float64
  PlanRows uint64 `json:"Plan Rows"`
  PlanWidth uint64 `json:"Plan Width"`
  RelationName string `json:"Relation Name"`
  RowsRemovedByFilter uint64 `json:"Rows Removed by Filter"`
  RowsRemovedByIndexRecheck uint64 `json:"Rows Removed by Index Recheck"`
  ScanDirection string `json:"Scan Direction"`
  Schema string `json:"Schema"`
  SharedDirtiedBlocks uint64 `json:"Shared Dirtied Blocks"`
  SharedHitBlocks uint64 `json:"Shared Hit Blocks"`
  SharedReadBlocks uint64 `json:"Shared Read Blocks"`
  SharedWrittenBlocks uint64 `json:"Shared Written Blocks"`
  Slowest bool
  StartupCost float64 `json:"Startup Cost"`
  Strategy string `json:"Strategy"`
  TempReadBlocks uint64 `json:"Temp Read Blocks"`
  TempWrittenBlocks uint64 `json:"Temp Written Blocks"`
  TotalCost float64 `json:"Total Cost"`
  Plans []Plan `json:"Plans"`
}

func CalculatePlannerEstimate(explain * Explain, plan * Plan) {
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

func CalculateActuals(explain * Explain, plan * Plan) {
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

func CalculateOutlierNodes(explain * Explain, plan * Plan) {
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
    CalculateOutlierNodes(explain, &plan.Plans[index])
  }
}

func CalculateMaximums(explain * Explain, plan * Plan) {
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
    return GoodFormat("<1 ms");
  } else if value > 1 && value < 1000 {
    return GoodFormat(fmt.Sprintf("%.2f ms", value))
  } else if value > 200 && value < 1000 {
    return WarningFormat(fmt.Sprintf("%.2f ms", value))
  } else if value >= 1000 && value < 60000 {
    return CriticalFormat(fmt.Sprintf("%.2f s", value / 2000.0))
  } else if value >= 60000 {
    return CriticalFormat(fmt.Sprintf("%.2f m", value / 60000.0))
  }
  return fmt.Sprintf("%f", value)
}

func ProcessExplain(explain * Explain) {
  ProcessPlan(explain, &explain.Plan)
  CalculateOutlierNodes(explain, &explain.Plan)
}

func ProcessPlan(explain * Explain, plan * Plan) {
  CalculatePlannerEstimate(explain, plan)
  CalculateActuals(explain, plan)
  CalculateMaximums(explain, plan)

  for index, _ := range plan.Plans {
    ProcessPlan(explain, &plan.Plans[index])
  }
}

func WriteExplain(writer io.Writer, explain * Explain) {
  fmt.Fprintf(writer, "○ Total Cost: %.2f\n", explain.TotalCost)
  fmt.Fprintf(writer, "○ Planning Time: %s\n", DurationToString(explain.PlanningTime))
  fmt.Fprintf(writer, "○ Execution Time: %s\n", DurationToString(explain.ExecutionTime))
  fmt.Fprintf(writer, PrefixFormat("┬\n"))

  WritePlan(writer, explain, &explain.Plan, "", 0, len(explain.Plan.Plans) == 1)
}

func WriteWithPrefix(writer io.Writer, prefix string) func(string, ...interface{}) (int, error) {
  return func(format string, a... interface{}) (int, error) {
    return fmt.Fprintf(writer, fmt.Sprintf("%s%s\n", PrefixFormat(prefix), format), a...)
  }
}

func WritePlan(writer io.Writer, explain * Explain, plan * Plan, prefix string, depth int, lastChild bool) {
  ParentOut := WriteWithPrefix(writer, prefix)

  ParentOut(PrefixFormat("│"))

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

  direction := ""
  if (plan.ScanDirection != "") {
    direction = MutedFormat(fmt.Sprintf("%v ", plan.ScanDirection))
  }

  strategy := ""

  if (plan.Strategy != "") {
    strategy = fmt.Sprintf(" (%v)", plan.Strategy)
  }

  ParentOut("%v %v%v%v %v", PrefixFormat(joint + "─⌠"), direction, BoldFormat(plan.NodeType), strategy, strings.Join(tags, " "));

  Out := WriteWithPrefix(writer, prefix + PrefixFormat("│ "))

  for _, line := range strings.Split(wordwrap.WrapString(Descriptions[plan.NodeType], 60), "\n") {
    Out("%v", MutedFormat(line))
  }

  Out("○ %v %v (%.0f%%)", "Duration:", DurationToString(plan.ActualDuration), (plan.ActualDuration / explain.ExecutionTime) * 100)

  Out("○ %v %.2f (%.0f%%)", "Cost:", plan.ActualCost, (plan.ActualCost / explain.TotalCost) * 100)

  Out("○ %v %v", "Rows:", humanize.Comma(int64(plan.ActualRows)))

  Out = WriteWithPrefix(writer, prefix + PrefixFormat("│   "))

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
    Out("%v %v [-%v rows]", MutedFormat("filter"), plan.Filter, humanize.Comma(int64(plan.RowsRemovedByFilter)));
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
    for index, line := range strings.Split(wordwrap.WrapString(strings.Join(plan.Output, " + "), 60), "\n") {
      if index == 0 {
        fmt.Fprintln(writer, PrefixFormat(prefix + joint + "► ") + OutputFormat(line))
      } else {
        if len(plan.Plans) == 0 {
          fmt.Fprintln(writer, PrefixFormat(prefix) + "   " + OutputFormat(line))
        } else {
          fmt.Fprintln(writer, PrefixFormat(prefix + "│  ") + OutputFormat(line))
        }
      }
    }
  }

  for index, _ := range plan.Plans {
    WritePlan(writer, explain, &plan.Plans[index], prefix, depth + 1, index == len(plan.Plans) - 1)
  }
}

func Visualize(writer io.Writer, buffer []byte) (error) {
  var explain []Explain

  err := json.Unmarshal(buffer, &explain)

  if err != nil {
    return err
  }

  for index, _ := range explain {
    ProcessExplain(&explain[index])
    WriteExplain(writer, &explain[index])
  }

  return nil
}

func main() {
  buffer, err := ioutil.ReadAll(os.Stdin)

  if err != nil {
    log.Fatalf("%v", err)
  }

  // fmt.Println(string(buffer))

  err = Visualize(os.Stdout, buffer)

  if err != nil {
    log.Fatalf("%v", err)
  }
}