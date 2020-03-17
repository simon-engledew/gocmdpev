package pev // import "github.com/simon-engledew/gocmdpev/pev"

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	humanize "github.com/dustin/go-humanize"
	"github.com/fatih/color"
	wordwrap "github.com/mitchellh/go-wordwrap"
)

type EstimateDirection string

const (
	Over  EstimateDirection = "Over"
	Under                   = "Under"
)

type NodeType string

const (
	Limit           NodeType = "Limit"
	Append                   = "Append"
	Sort                     = "Sort"
	NestedLoop               = "Nested Loop"
	MergeJoin                = "Merge Join"
	Hash                     = "Hash"
	HashJoin                 = "Hash Join"
	Aggregate                = "Aggregate"
	Hashaggregate            = "Hashaggregate"
	SequenceScan             = "Seq Scan"
	IndexScan                = "Index Scan"
	IndexOnlyScan            = "Index Only Scan"
	BitmapHeapScan           = "Bitmap Heap Scan"
	BitmapIndexScan          = "Bitmap Index Scan"
	CTEScan                  = "CTE Scan"
)

var prefixFormat = color.New(color.FgHiBlack).SprintFunc()
var tagFormat = color.New(color.FgWhite, color.BgRed).SprintFunc()
var mutedFormat = color.New(color.FgHiBlack).SprintFunc()
var boldFormat = color.New(color.FgHiWhite).SprintFunc()
var goodFormat = color.New(color.FgGreen).SprintFunc()
var warningFormat = color.New(color.FgHiYellow).SprintFunc()
var criticalFormat = color.New(color.FgHiRed).SprintFunc()
var outputFormat = color.New(color.FgCyan).SprintFunc()

var Descriptions = map[NodeType]string{
	Append:          "Used in a UNION to merge multiple record sets by appending them together.",
	Limit:           "Returns a specified number of rows from a record set.",
	Sort:            "Sorts a record set based on the specified sort key.",
	NestedLoop:      "Merges two record sets by looping through every record in the first set and trying to find a match in the second set. All matching records are returned.",
	MergeJoin:       "Merges two record sets by first sorting them on a join key.",
	Hash:            "Generates a hash table from the records in the input recordset. Hash is used by Hash Join.",
	HashJoin:        "Joins to record sets by hashing one of them (using a Hash Scan).",
	Aggregate:       "Groups records together based on a GROUP BY or aggregate function (e.g. sum()).",
	Hashaggregate:   "Groups records together based on a GROUP BY or aggregate function (e.g. sum()). Hash Aggregate uses a hash to first organize the records by a key.",
	SequenceScan:    "Finds relevant records by sequentially scanning the input record set. When reading from a table, Seq Scans (unlike Index Scans) perform a single read operation (only the table is read).",
	IndexScan:       "Finds relevant records based on an Index. Index Scans perform 2 read operations: one to read the index and another to read the actual value from the table.",
	IndexOnlyScan:   "Finds relevant records based on an Index. Index Only Scans perform a single read operation from the index and do not read from the corresponding table.",
	BitmapHeapScan:  "Searches through the pages returned by the Bitmap Index Scan for relevant rows.",
	BitmapIndexScan: "Uses a Bitmap Index (index which uses 1 bit per page) to find all relevant pages. Results of this node are fed to the Bitmap Heap Scan.",
	CTEScan:         "Performs a sequential scan of Common Table Expression (CTE) query results. Note that results of a CTE are materialized (calculated and temporarily stored).",
}

type Explain struct {
	Plan          Plan          `json:"Plan"`
	PlanningTime  float64       `json:"Planning Time"`
	Triggers      []interface{} `json:"Triggers"`
	ExecutionTime float64       `json:"Execution Time"`
	TotalCost     float64
	MaxRows       uint64
	MaxCost       float64
	MaxDuration   float64
}

type Plan struct {
	ActualCost                  float64
	ActualDuration              float64
	ActualLoops                 uint64  `json:"Actual Loops"`
	ActualRows                  uint64  `json:"Actual Rows"`
	ActualStartupTime           float64 `json:"Actual Startup Time"`
	ActualTotalTime             float64 `json:"Actual Total Time"`
	Alias                       string  `json:"Alias"`
	Costliest                   bool
	CTEName                     string   `json:"CTE Name"`
	Filter                      string   `json:"Filter"`
	GroupKey                    []string `json:"Group Key"`
	HashCondition               string   `json:"Hash Cond"`
	HeapFetches                 uint64   `json:"Heap Fetches"`
	IndexCondition              string   `json:"Index Cond"`
	IndexName                   string   `json:"Index Name"`
	IOReadTime                  float64  `json:"I/O Read Time"`
	IOWriteTime                 float64  `json:"I/O Write Time"`
	JoinType                    string   `json:"Join Type"`
	Largest                     bool
	LocalDirtiedBlocks          uint64   `json:"Local Dirtied Blocks"`
	LocalHitBlocks              uint64   `json:"Local Hit Blocks"`
	LocalReadBlocks             uint64   `json:"Local Read Blocks"`
	LocalWrittenBlocks          uint64   `json:"Local Written Blocks"`
	NodeType                    NodeType `json:"Node Type"`
	Output                      []string `json:"Output"`
	ParentRelationship          string   `json:"Parent Relationship"`
	PlannerRowEstimateDirection EstimateDirection
	PlannerRowEstimateFactor    float64
	PlanRows                    uint64 `json:"Plan Rows"`
	PlanWidth                   uint64 `json:"Plan Width"`
	RelationName                string `json:"Relation Name"`
	RowsRemovedByFilter         uint64 `json:"Rows Removed by Filter"`
	RowsRemovedByIndexRecheck   uint64 `json:"Rows Removed by Index Recheck"`
	ScanDirection               string `json:"Scan Direction"`
	Schema                      string `json:"Schema"`
	SharedDirtiedBlocks         uint64 `json:"Shared Dirtied Blocks"`
	SharedHitBlocks             uint64 `json:"Shared Hit Blocks"`
	SharedReadBlocks            uint64 `json:"Shared Read Blocks"`
	SharedWrittenBlocks         uint64 `json:"Shared Written Blocks"`
	Slowest                     bool
	StartupCost                 float64 `json:"Startup Cost"`
	Strategy                    string  `json:"Strategy"`
	TempReadBlocks              uint64  `json:"Temp Read Blocks"`
	TempWrittenBlocks           uint64  `json:"Temp Written Blocks"`
	TotalCost                   float64 `json:"Total Cost"`
	Plans                       []Plan  `json:"Plans"`
}

func calculatePlannerEstimate(explain *Explain, plan *Plan) {
	plan.PlannerRowEstimateFactor = 0

	if plan.PlanRows == plan.ActualRows {
		return
	}

	plan.PlannerRowEstimateDirection = Under
	if plan.PlanRows != 0 {
		plan.PlannerRowEstimateFactor = float64(plan.ActualRows) / float64(plan.PlanRows)
	}

	if plan.PlannerRowEstimateFactor < 1.0 {
		plan.PlannerRowEstimateFactor = 0
		plan.PlannerRowEstimateDirection = Over
		if plan.ActualRows != 0 {
			plan.PlannerRowEstimateFactor = float64(plan.PlanRows) / float64(plan.ActualRows)
		}
	}
}

func calculateActuals(explain *Explain, plan *Plan) {
	plan.ActualDuration = plan.ActualTotalTime
	plan.ActualCost = plan.TotalCost

	for _, child := range plan.Plans {
		if child.NodeType != CTEScan {
			plan.ActualDuration = plan.ActualDuration - child.ActualTotalTime
			plan.ActualCost = plan.ActualCost - child.TotalCost
		}
	}

	if plan.ActualCost < 0 {
		plan.ActualCost = 0
	}

	explain.TotalCost = explain.TotalCost + plan.ActualCost

	plan.ActualDuration = plan.ActualDuration * float64(plan.ActualLoops)
}

func calculateOutlierNodes(explain *Explain, plan *Plan) {
	plan.Costliest = plan.ActualCost == explain.MaxCost
	plan.Largest = plan.ActualRows == explain.MaxRows
	plan.Slowest = plan.ActualDuration == explain.MaxDuration

	for index := range plan.Plans {
		calculateOutlierNodes(explain, &plan.Plans[index])
	}
}

func calculateMaximums(explain *Explain, plan *Plan) {
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

func durationToString(value float64) string {
	if value < 1 {
		return goodFormat("<1 ms")
	} else if value < 100 {
		return goodFormat(fmt.Sprintf("%.2f ms", value))
	} else if value < 1000 {
		return warningFormat(fmt.Sprintf("%.2f ms", value))
	} else if value < 60000 {
		return criticalFormat(fmt.Sprintf("%.2f s", value/2000.0))
	} else {
		return criticalFormat(fmt.Sprintf("%.2f m", value/60000.0))
	}
}

func processExplain(explain *Explain) {
	processPlan(explain, &explain.Plan)
	calculateOutlierNodes(explain, &explain.Plan)
}

func processPlan(explain *Explain, plan *Plan) {
	calculatePlannerEstimate(explain, plan)
	calculateActuals(explain, plan)
	calculateMaximums(explain, plan)

	for index := range plan.Plans {
		processPlan(explain, &plan.Plans[index])
	}
}

func writeExplain(writer io.Writer, explain *Explain, width uint) {
	fmt.Fprintf(writer, "○ Total Cost: %s\n", humanize.Commaf(explain.TotalCost))
	fmt.Fprintf(writer, "○ Planning Time: %s\n", durationToString(explain.PlanningTime))
	fmt.Fprintf(writer, "○ Execution Time: %s\n", durationToString(explain.ExecutionTime))
	fmt.Fprintf(writer, prefixFormat("┬\n"))

	writePlan(writer, explain, &explain.Plan, "", 0, width, len(explain.Plan.Plans) == 1)
}

func formatDetails(plan *Plan) string {
	var details []string

	if plan.ScanDirection != "" {
		details = append(details, plan.ScanDirection)
	}

	if plan.Strategy != "" {
		details = append(details, plan.Strategy)
	}

	if len(details) > 0 {
		return mutedFormat(fmt.Sprintf(" [%v]", strings.Join(details, ", ")))
	}

	return ""
}

func formatTag(tag string) string {
	return tagFormat(fmt.Sprintf(" %v ", tag))
}

func formatTags(plan *Plan) string {
	var tags []string

	if plan.Slowest {
		tags = append(tags, formatTag("slowest"))
	}
	if plan.Costliest {
		tags = append(tags, formatTag("costliest"))
	}
	if plan.Largest {
		tags = append(tags, formatTag("largest"))
	}
	if plan.PlannerRowEstimateFactor >= 100 {
		tags = append(tags, formatTag("bad estimate"))
	}

	return strings.Join(tags, " ")
}

func getTerminator(index int, plan *Plan) string {
	if index == 0 {
		if len(plan.Plans) == 0 {
			return "⌡► "
		} else {
			return "├►  "
		}
	} else {
		if len(plan.Plans) == 0 {
			return "   "
		} else {
			return "│  "
		}
	}
}

func wrapString(line string, width uint) string {
	if width == 0 {
		return line
	}
	return wordwrap.WrapString(line, width)
}

func writePlan(writer io.Writer, explain *Explain, plan *Plan, prefix string, depth int, width uint, lastChild bool) {
	currentPrefix := prefix

	var outputFn = func(format string, a ...interface{}) (int, error) {
		return fmt.Fprintf(writer, fmt.Sprintf("%s%s\n", prefixFormat(currentPrefix), format), a...)
	}

	outputFn(prefixFormat("│"))

	joint := "├"
	if len(plan.Plans) > 1 || lastChild {
		joint = "└"
	}

	outputFn("%v %v%v %v", prefixFormat(joint+"─⌠"), boldFormat(plan.NodeType), formatDetails(plan), formatTags(plan))

	if len(plan.Plans) > 1 || lastChild {
		prefix += "  "
	} else {
		prefix += "│ "
	}

	currentPrefix = prefix + "│ "

	cols := width - uint(len(currentPrefix))

	for _, line := range strings.Split(wrapString(Descriptions[plan.NodeType], cols), "\n") {
		outputFn("%v", mutedFormat(line))
	}

	outputFn("○ %v %v (%.0f%%)", "Duration:", durationToString(plan.ActualDuration), (plan.ActualDuration/explain.ExecutionTime)*100)

	outputFn("○ %v %v (%.0f%%)", "Cost:", humanize.Commaf(plan.ActualCost), (plan.ActualCost/explain.TotalCost)*100)

	outputFn("○ %v %v", "Rows:", humanize.Comma(int64(plan.ActualRows)))

	currentPrefix = currentPrefix + "  "

	if plan.JoinType != "" {
		outputFn("%v %v", plan.JoinType, mutedFormat("join"))
	}

	if plan.RelationName != "" {
		outputFn("%v %v.%v", mutedFormat("on"), plan.Schema, plan.RelationName)
	}

	if plan.IndexName != "" {
		outputFn("%v %v", mutedFormat("using"), plan.IndexName)
	}

	if plan.IndexCondition != "" {
		outputFn("%v %v", mutedFormat("condition"), plan.IndexCondition)
	}

	if plan.Filter != "" {
		outputFn("%v %v %v", mutedFormat("filter"), plan.Filter, mutedFormat(fmt.Sprintf("[-%v rows]", humanize.Comma(int64(plan.RowsRemovedByFilter)))))
	}

	if plan.HashCondition != "" {
		outputFn("%v %v", mutedFormat("on"), plan.HashCondition)
	}

	if plan.CTEName != "" {
		outputFn("CTE %v", plan.CTEName)
	}

	if plan.PlannerRowEstimateFactor != 0 {
		outputFn("%v %vestimated %v %.2fx", mutedFormat("rows"), plan.PlannerRowEstimateDirection, mutedFormat("by"), plan.PlannerRowEstimateFactor)
	}

	currentPrefix = prefix

	if len(plan.Output) > 0 {
		for index, line := range strings.Split(wrapString(strings.Join(plan.Output, " + "), cols), "\n") {
			outputFn(prefixFormat(getTerminator(index, plan)) + outputFormat(line))
		}
	}

	for index := range plan.Plans {
		writePlan(writer, explain, &plan.Plans[index], prefix, depth+1, width, index == len(plan.Plans)-1)
	}
}

func Visualize(writer io.Writer, reader io.Reader, width uint) error {
	var explain []Explain

	err := json.NewDecoder(reader).Decode(&explain)

	if err != nil {
		return err
	}

	for index := range explain {
		processExplain(&explain[index])
		writeExplain(writer, &explain[index], width)
	}

	return nil
}
