package join

import (
	"context"
	"sync"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/table"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

// MergeJoinTransformation performs a sort-merge-join on two table streams.
// It makes the following assumptions about the input tables:
//   * They have the same group key
//   * They are sorted by the columns in the `on` parameter
//   * We are performing an equijoin
type MergeJoinTransformation struct {
	ctx                     context.Context
	on                      []ColumnPair
	as                      *JoinFn
	left, right             execute.DatasetID
	leftSchema, rightSchema []flux.ColMeta
	method                  string
	d                       *execute.TransportDataset
	mu                      sync.Mutex
	mem                     memory.Allocator

	leftFinished,
	rightFinished bool
}

func NewMergeJoinTransformation(
	id execute.DatasetID,
	s plan.ProcedureSpec,
	a execute.Administration,
) (*MergeJoinTransformation, error) {
	spec, ok := s.(*SortMergeJoinProcedureSpec)
	if !ok {
		return nil, errors.New(codes.Internal, "unsupported join spec - not a sortMergeJoin")
	}
	mem := a.Allocator()
	return &MergeJoinTransformation{
		ctx:    a.Context(),
		on:     spec.On,
		as:     NewJoinFn(spec.As),
		left:   a.Parents()[0],
		right:  a.Parents()[1],
		method: spec.Method,
		d:      execute.NewTransportDataset(id, mem),
		mem:    mem,
	}, nil
}

func (t *MergeJoinTransformation) ProcessMessage(m execute.Message) error {
	defer m.Ack()

	switch m := m.(type) {
	case execute.ProcessChunkMsg:
		chunk := m.TableChunk()
		id := m.SrcDatasetID()
		state, _ := t.d.Lookup(chunk.Key())
		s, ok, err := t.processChunk(chunk, state, id)
		if err != nil {
			return err
		}
		if ok {
			t.d.Set(chunk.Key(), s)
		}
	case execute.FlushKeyMsg:
		id := m.SrcDatasetID()
		key := m.Key()
		state, _ := t.d.Lookup(key)
		s, _ := state.(*joinState)
		if id == t.left {
			s.left.done = true
		} else if id == t.right {
			s.right.done = true
		}

		if s.finished() {
			if err := t.flush(s); err != nil {
				return err
			}
		}
	case execute.FinishMsg:
		err := m.Error()
		if err != nil {
			t.d.Finish(err)
			return nil
		}

		id := m.SrcDatasetID()
		if id == t.left {
			t.leftFinished = true
		} else if id == t.right {
			t.rightFinished = true
		}

		if t.isFinished() {
			err = t.d.Range(func(key flux.GroupKey, value interface{}) error {
				s, ok := value.(*joinState)
				if !ok {
					return errors.New(codes.Internal, "received bad joinState")
				}
				s.left.schema = t.leftSchema
				s.right.schema = t.rightSchema
				return t.flush(s)
			})
			t.d.Finish(err)
		}
	default:
		return errors.New(codes.Internal, "invalid message")
	}
	return nil
}

func (t *MergeJoinTransformation) initState(state interface{}) (*joinState, bool) {
	if state != nil {
		s, ok := state.(*joinState)
		return s, ok
	}
	s := joinState{}
	return &s, true
}

func (t *MergeJoinTransformation) processChunk(chunk table.Chunk, state interface{}, id execute.DatasetID) (*joinState, bool, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	s, ok := t.initState(state)
	if !ok {
		return nil, false, errors.New(codes.Internal, "invalid join state")
	}

	if chunk.Len() == 0 {
		return s, true, nil
	}
	chunk.Retain()

	var isLeft bool
	if id == t.left {
		isLeft = true
		s.left.schema = chunk.Cols()
		t.leftSchema = chunk.Cols()
	} else if id == t.right {
		isLeft = false
		s.right.schema = chunk.Cols()
		t.rightSchema = chunk.Cols()
	} else {
		return s, true, errors.New(codes.Internal, "invalid chunk passed to join - dataset id is neither left nor right")
	}

	t.mergeJoin(chunk, s, isLeft)

	return s, true, nil
}

func (t *MergeJoinTransformation) mergeJoin(chunk table.Chunk, s *joinState, isLeft bool) error {
	for {
		key, rows := s.merge(chunk, isLeft, t.on)
		if key == nil {
			break
		}
		i, canJoin := s.insert(key, rows, isLeft)
		if canJoin {
			joined, err := s.join(t.ctx, t.method, t.as, i, t.mem)
			if err != nil {
				return err
			}

			for _, chunk := range joined {
				err := t.d.Process(chunk)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// Produce whatever results we can from the remaining, unprocessed data.
func (t *MergeJoinTransformation) flush(s *joinState) error {
	lkey, lrows := s.left.flush(t.mem)
	if lkey != nil {
		_, _ = s.insert(lkey, lrows, true)
	}
	rkey, rrows := s.right.flush(t.mem)
	if rkey != nil {
		_, _ = s.insert(rkey, rrows, false)
	}

	joined, err := s.join(t.ctx, t.method, t.as, len(s.products)-1, t.mem)
	if err != nil {
		return err
	}
	for _, chunk := range joined {
		err = t.d.Process(chunk)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *MergeJoinTransformation) isFinished() bool {
	return t.leftFinished && t.rightFinished
}

type joinState struct {
	left, right sideState
	products    []joinProduct
}

// merge passes the chunk to the appropriate side of the transformation, sets
// the join key columns if they're not already set, and returns the output of scan()
func (s *joinState) merge(c table.Chunk, isLeft bool, on []ColumnPair) (*joinKey, joinRows) {
	if isLeft {
		if len(s.left.joinKeyCols) < 1 {
			s.left.setJoinKeyCols(getJoinKeyCols(on, isLeft), c)
		}
		return s.left.scan(c)
	} else {
		if len(s.right.joinKeyCols) < 1 {
			s.right.setJoinKeyCols(getJoinKeyCols(on, isLeft), c)
		}
		return s.right.scan(c)
	}
}

func getJoinKeyCols(on []ColumnPair, isLeft bool) []string {
	labels := make([]string, 0, len(on))
	for _, pair := range on {
		if isLeft {
			labels = append(labels, pair.Left)
		} else {
			labels = append(labels, pair.Right)
		}
	}
	return labels
}

// Inserts `rows` into s.products, while maintaining sort order. Returns a position and a bool.
// If the bool == true, that means it's safe to join all the items in s.products up to and including
// the returned position.
//
// Returns true under 2 circumstances:
//   (1) Inserting `rows` completes the left and right pair for a given product
//   (2) `rows` was inserted at an index greater than 0, and all of the products that
//        come before it only have entries on the opposite side.
//
// If condition 1 is true, we can join everything up to and including the index where
// `rows` was inserted.
//
// If condition 2 is true and condition 1 is false, we can join up to, but not including,
// the index where rows was inserted.
func (s *joinState) insert(key *joinKey, rows joinRows, isLeft bool) (int, bool) {
	if len(s.products) == 0 {
		p := newJoinProduct(key, rows, isLeft)
		s.products = []joinProduct{p}
		return 0, false
	}

	found := false
	position := 0
	for i, product := range s.products {
		if product.key.equal(*key) {
			if isLeft {
				product.left = rows
			} else {
				product.right = rows
			}
			s.products[i] = product
			found = true
			position = i
			break
		} else if key.less(product.key) {
			newProducts := make([]joinProduct, 0, len(s.products)+1)
			newProducts = append(newProducts, s.products[:i]...)
			newProducts = append(newProducts, newJoinProduct(key, rows, isLeft))
			newProducts = append(newProducts, s.products[i:]...)
			s.products = newProducts
			found = true
			position = i
			break
		}
		position = i
	}

	if !found {
		s.products = append(s.products, newJoinProduct(key, rows, isLeft))
		position = len(s.products) - 1
	}

	if s.products[position].isDone() {
		return position, true
	}

	// The only condition where we would not emit a signal to join and flush
	// the list of products is if we received a bunch of `joinRows` all
	// from the same side.
	//
	// So if the previous product is only populated on one side, then it is either the first
	// item in the list, or all of the items before it are also populated on the same side.
	canJoin := false
	if position > 0 {
		position--
		prev := s.products[position]
		if isLeft {
			canJoin = prev.left.len() == 0 && prev.right.len() > 0
		} else {
			canJoin = prev.left.len() > 0 && prev.right.len() == 0
		}
	}
	return position, canJoin
}

func (s *joinState) join(ctx context.Context, method string, fn *JoinFn, joinable int, mem memory.Allocator) ([]table.Chunk, error) {
	if fn.compiled == nil {
		err := fn.Prepare(s.left.schema, s.right.schema)
		if err != nil {
			return nil, err
		}
	}
	joined := make([]table.Chunk, 0, joinable+1)
	for i := 0; i <= joinable; i++ {
		prod := s.products[i]
		p, ok, err := prod.evaluate(ctx, method, *fn, mem)
		if err != nil {
			return joined, err
		}
		if !ok {
			continue
		}
		joined = append(joined, *p)
	}
	s.products = s.products[joinable+1:]
	return joined, nil
}

func (s *joinState) finished() bool {
	return s.left.done && s.right.done
}

type sideState struct {
	schema      []flux.ColMeta
	joinKeyCols []flux.ColMeta
	chunks      []table.Chunk
	keyStart    int
	keyEnd      int
	done        bool
}

func (s *sideState) setJoinKeyCols(labels []string, c table.Chunk) {
	cols := make([]flux.ColMeta, 0, len(labels))
	for _, label := range labels {
		cols = append(cols, c.Col(c.Index(label)))
	}
	s.joinKeyCols = cols
}

// scan calls and handles the outputs of advance(). If advance reports that it
// found a complete join key, scan returns the key, as well as a collection of the rows that match
// that join key. Otherwise, it stores the rows it just scanned in s.chunks
func (s *sideState) scan(c table.Chunk) (*joinKey, joinRows) {
	key, complete := s.advance(c)
	if !complete {
		s.addChunk(getChunkSlice(c, s.keyStart, c.Len()))
		return nil, nil
	}
	rows := s.consumeRows(c)
	return key, rows
}

func (s *sideState) addChunk(c table.Chunk) {
	s.chunks = append(s.chunks, c)
}

// advance iterates over each row of c until it either finds a new join key or
// reaches the end of the chunk. Returns true if it finds the end of a join key,
// along with the key itself.
func (s *sideState) advance(c table.Chunk) (*joinKey, bool) {
	startKey := joinKeyFromRow(s.joinKeyCols, c, s.keyEnd)
	s.keyStart = s.keyEnd
	s.keyEnd++

	complete := false
	for ; s.keyEnd < c.Len(); s.keyEnd++ {
		key := joinKeyFromRow(s.joinKeyCols, c, s.keyEnd)

		if !key.equal(startKey) {
			complete = true
			break
		}
	}

	// We hit the end of the chunk without finding a new join key.
	// Reset keyEnd to 0 since the next thing we process will be a new chunk.
	if !complete {
		s.keyEnd = 0
	}

	return &startKey, complete
}

// collects rows within the indices set by `advance()` and returns them in one
// data structure. It will discard any exhausted chunks or rows.
//
// Any chunks stored in s.chunks should have the same join key.
func (s *sideState) consumeRows(c table.Chunk) joinRows {
	rows := make([]table.Chunk, 0, len(s.chunks)+1)
	if len(s.chunks) > 0 {
		rows = append(rows, s.chunks...)
	}
	rows = append(rows, getChunkSlice(c, s.keyStart, s.keyEnd))
	s.chunks = []table.Chunk{}

	return rows
}

// flush returns any stored table chunks and their join key. It should not be possible for
// the returned chunks to have multiple join keys.
func (s *sideState) flush(mem memory.Allocator) (*joinKey, joinRows) {
	if len(s.chunks) < 1 {
		return nil, nil
	}
	key := joinKeyFromRow(s.joinKeyCols, s.chunks[0], 0)
	rows := joinRows(s.chunks)
	s.chunks = s.chunks[:0]
	return &key, rows
}

func getChunkSlice(chunk table.Chunk, start, stop int) table.Chunk {
	buf := arrow.TableBuffer{
		GroupKey: chunk.Key(),
		Columns:  chunk.Cols(),
		Values:   make([]array.Array, 0, chunk.NCols()),
	}
	for _, col := range chunk.Buffer().Values {
		arr := arrow.Slice(col, int64(start), int64(stop))
		buf.Values = append(buf.Values, arr)
	}
	return table.ChunkFromBuffer(buf)
}

// joinRows represents a collection of rows from the same input table that all
// have the same join key.
type joinRows []table.Chunk

func (r joinRows) len() int {
	return len(r)
}

func (r joinRows) nrows() int {
	nrows := 0
	for _, chunk := range r {
		nrows += chunk.Len()
	}
	return nrows
}

func (r joinRows) getRow(i int, typ semantic.MonoType) values.Object {
	var obj values.Object
	for _, chunk := range r {
		if i > chunk.Len()-1 {
			i -= chunk.Len() - 1
		} else {
			obj = rowFromChunk(chunk, i, typ)
			break
		}
	}
	return obj
}

// joinProduct represents a collection of rows from the left and right
// input tables that have the same join key, and can therefore be joined
// in the final output table.
type joinProduct struct {
	key         joinKey
	left, right joinRows
}

func newJoinProduct(key *joinKey, rows joinRows, isLeft bool) joinProduct {
	p := joinProduct{
		key: *key,
	}
	if isLeft {
		p.left = rows
	} else {
		p.right = rows
	}
	return p
}

func (p *joinProduct) isDone() bool {
	return p.left.len() > 0 && p.right.len() > 0
}

// Returns the joined output of the product, if there is any. If the returned bool is `true`,
// the returned table chunk contains joined data that should be passed along to the next node.
func (p *joinProduct) evaluate(ctx context.Context, method string, fn JoinFn, mem memory.Allocator) (*table.Chunk, bool, error) {
	return fn.Eval(ctx, p, method, mem)
}

func rowFromChunk(c table.Chunk, i int, mt semantic.MonoType) values.Object {
	obj := values.NewObject(mt)
	buf := c.Buffer()
	for j := 0; j < c.NCols(); j++ {
		col := c.Col(j).Label
		v := execute.ValueForRow(&buf, i, j)
		obj.Set(col, v)
	}
	return obj
}
