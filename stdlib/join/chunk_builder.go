package join

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute/table"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/values"
)

type chunkBuilder struct {
	cols     []flux.ColMeta
	builders []array.Builder
}

func newChunkBuilder(cols []flux.ColMeta, size int, mem memory.Allocator) *chunkBuilder {
	builders := make([]array.Builder, len(cols))
	for i, col := range cols {
		b := arrow.NewBuilder(col.Type, mem)
		b.Resize(size)
		builders[i] = b
	}
	return &chunkBuilder{cols: cols, builders: builders}
}

func (b *chunkBuilder) appendRecord(record values.Object) error {
	for i, col := range b.cols {
		v, ok := record.Get(col.Label)
		if !ok {
			return errors.Newf(codes.Internal, "could not find column %s in record", col.Label)
		}
		if err := arrow.AppendValue(b.builders[i], v); err != nil {
			return err
		}
	}
	return nil
}

func (b *chunkBuilder) build(key flux.GroupKey) table.Chunk {
	buf := arrow.TableBuffer{
		GroupKey: key,
		Columns:  b.cols,
	}
	vals := make([]array.Array, 0, len(b.builders))
	for _, builder := range b.builders {
		vals = append(vals, builder.NewArray())
	}
	buf.Values = vals
	return table.ChunkFromBuffer(buf)
}
