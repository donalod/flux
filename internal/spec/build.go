package spec

import (
	"context"
	"fmt"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/plan"
	"github.com/opentracing/opentracing-go"
)

type ider struct {
	id     *int
	lookup map[*flux.TableObject]flux.OperationID
}

func (i *ider) nextID() int {
	next := *i.id
	*i.id++
	return next
}

func (i *ider) get(t *flux.TableObject) (flux.OperationID, bool) {
	tableID, ok := i.lookup[t]
	return tableID, ok
}

func (i *ider) set(t *flux.TableObject, id int) flux.OperationID {
	opID := flux.OperationID(fmt.Sprintf("%s%d", t.Kind, id))
	i.lookup[t] = opID
	return opID
}

func (i *ider) ID(t *flux.TableObject) flux.OperationID {
	tableID, ok := i.get(t)
	if !ok {
		tableID = i.set(t, i.nextID())
	}
	return tableID
}

// FromEvaluation produces a spec from
// TODO mention when skipYields should be used
func FromEvaluation(
	ctx context.Context,
	ses []interpreter.SideEffect,
	now time.Time,
	// FIXME: skipYields is a misnomer. Better name needed.
	skipYields bool,
) (*flux.Spec, error) {
	var nextNodeID *int
	if value := ctx.Value(plan.NextPlanNodeIDKey); value != nil {
		nextNodeID = value.(*int)
	} else {
		nextNodeID = new(int)
	}
	ider := &ider{
		id:     nextNodeID,
		lookup: make(map[*flux.TableObject]flux.OperationID),
	}

	spec := &flux.Spec{Now: now}
	seen := make(map[*flux.TableObject]bool)
	objs := make([]*flux.TableObject, 0, len(ses))

	sideEffectCount := 0 // FIXME: I think these would all be results in the end?

	for _, se := range ses {

		if op, ok := se.Value.(*flux.TableObject); ok {
			sideEffectCount += 1
			fmt.Printf("se op=%q\n", op.Kind)
			s, cctx := opentracing.StartSpanFromContext(ctx, "toSpec")
			s.SetTag("opKind", op.Kind)
			if se.Node != nil {
				s.SetTag("loc", se.Node.Location().String())
			}

			if !isDuplicateTableObject(cctx, op, objs) {
				buildSpecWithTrace(cctx, op, ider, spec, seen, skipYields)
				objs = append(objs, op)
			}

			s.Finish()
		}
	}

	resultCount := sideEffectCount
	if sideEffectCount == 0 { // FIXME: this is wrong
		resultCount = 1
	}
	fmt.Printf(
		"\n\nresultCount=%d\nskipYields=%t\nspec=%s\n\n",
		resultCount,
		skipYields,
		flux.Formatted(spec),
	)

	// When skipYields is true, we're running a sub-program (a la tableFind).
	// In this case we ignore any yields but we also have an extra requirement:
	// there can only be 1 result. This is to say, if there is not exactly 1
	// operation, we error.
	if skipYields && resultCount != 1 {
		return nil,
			errors.Newf(
				codes.Invalid,
				// TODO: should we report the skipped yields in this count?
				"expected exactly 1 result from table stream, found %d", resultCount,
			)
	}
	if len(spec.Operations) == 0 {
		return nil,
			errors.New(codes.Invalid,
				"this Flux script returns no streaming data. "+
					"Consider adding a \"yield\" or invoking streaming functions directly, without performing an assignment")
	}
	return spec, nil
}

func isDuplicateTableObject(ctx context.Context, op *flux.TableObject, objs []*flux.TableObject) bool {
	s, _ := opentracing.StartSpanFromContext(ctx, "isDuplicate")
	defer s.Finish()

	for _, tableObject := range objs {
		if op == tableObject {
			return true
		}
	}
	return false
}

func buildSpecWithTrace(ctx context.Context, t *flux.TableObject, ider flux.IDer, spec *flux.Spec, visited map[*flux.TableObject]bool, skipYields bool) {
	s, _ := opentracing.StartSpanFromContext(ctx, "buildSpec")
	s.SetTag("opKind", t.Kind)
	buildSpec(t, ider, spec, visited, skipYields)
	s.Finish()
}

func buildSpec(t *flux.TableObject, ider flux.IDer, spec *flux.Spec, visited map[*flux.TableObject]bool, skipYields bool) {
	// Traverse graph upwards to first unvisited node.
	// Note: parents are sorted based on parameter name, so the visit order is consistent.
	for _, p := range t.Parents {

		if !visited[p] {
			// FIXME: if the parent is a yield, go directly to the parent's parent?
			if skipYields && p.Kind == "yield" {
				for _, pp := range p.Parents {
					if !visited[pp] {
						// recurse up parents
						buildSpec(pp, ider, spec, visited, skipYields)
					}
				}
			} else {
				// recurse up parents
				buildSpec(p, ider, spec, visited, skipYields)
			}
		}
	}

	// Assign ID to table object after visiting all ancestors.
	tableID := ider.ID(t)

	// Link table object to all parents after assigning ID.
	for _, p := range t.Parents {
		// FIXME: if the parent is a yield, go directly to the parent's parent?
		if skipYields && p.Kind == "yield" {
			for _, pp := range p.Parents {
				spec.Edges = append(spec.Edges, flux.Edge{
					Parent: ider.ID(pp),
					Child:  tableID,
				})
			}
		} else {
			spec.Edges = append(spec.Edges, flux.Edge{
				Parent: ider.ID(p),
				Child:  tableID,
			})
		}

	}

	visited[t] = true
	if !(skipYields && t.Kind == "yield") {
		spec.Operations = append(spec.Operations, t.Operation(ider))
	}
}

// FromTableObject returns a spec from a TableObject.
func FromTableObject(ctx context.Context, to *flux.TableObject, now time.Time) (*flux.Spec, error) {
	return FromEvaluation(ctx, []interpreter.SideEffect{{Value: to}}, now, true)
}

// FromScript returns a spec from a script expressed as a raw string.
// This is duplicate logic for what happens when a flux.Program runs.
// This function is used in tests that compare flux.Specs (e.g. in planner tests).
func FromScript(ctx context.Context, runtime flux.Runtime, now time.Time, script string) (*flux.Spec, error) {
	s, _ := opentracing.StartSpanFromContext(ctx, "parse")
	astPkg, err := runtime.Parse(script)
	if err != nil {
		return nil, err
	}
	s.Finish()

	deps := execute.NewExecutionDependencies(nil, &now, nil)
	ctx = deps.Inject(ctx)

	s, cctx := opentracing.StartSpanFromContext(ctx, "eval")
	sideEffects, scope, err := runtime.Eval(cctx, astPkg, nil, flux.SetNowOption(now))
	if err != nil {
		return nil, err
	}
	s.Finish()

	s, cctx = opentracing.StartSpanFromContext(ctx, "compile")
	defer s.Finish()
	nowOpt, ok := scope.Lookup(interpreter.NowOption)
	if !ok {
		return nil, fmt.Errorf("%q option not set", interpreter.NowOption)
	}
	nowTime, err := nowOpt.Function().Call(ctx, nil)
	if err != nil {
		return nil, err
	}

	return FromEvaluation(cctx, sideEffects, nowTime.Time().Time(), false)
}
