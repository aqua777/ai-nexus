package lancedb

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/apache/arrow/go/v17/arrow"
	"github.com/apache/arrow/go/v17/arrow/array"
	"github.com/apache/arrow/go/v17/arrow/memory"
	"github.com/aqua777/ai-nexus/vectordb/v1/schema"
	lancedb "github.com/aqua777/go-lancedb"
)

// LanceDBStore is a vector store implementation using LanceDB.
type LanceDBStore struct {
	conn      *lancedb.Connection
	table     *lancedb.Table
	tableName string
}

// NewLanceDBStore creates a new LanceDBStore.
func NewLanceDBStore(uri string, tableName string) (*LanceDBStore, error) {
	conn, err := lancedb.Connect(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to lancedb: %w", err)
	}

	store := &LanceDBStore{
		conn:      conn,
		tableName: tableName,
	}

	// Check if table exists
	tblNames, err := conn.TableNames()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to list tables: %w", err)
	}

	for _, name := range tblNames {
		if name == tableName {
			table, err := conn.OpenTable(tableName)
			if err != nil {
				conn.Close()
				return nil, fmt.Errorf("failed to open table: %w", err)
			}
			store.table = table
			break
		}
	}

	return store, nil
}

// Close closes the connection.
func (s *LanceDBStore) Close() error {
	if s.table != nil {
		s.table.Close()
	}
	if s.conn != nil {
		s.conn.Close()
	}
	return nil
}

// Add adds nodes to the store.
func (s *LanceDBStore) Add(ctx context.Context, nodes []schema.Node) ([]string, error) {
	if len(nodes) == 0 {
		return nil, nil
	}

	// Define schema
	dim := len(nodes[0].Embedding)
	if dim == 0 {
		return nil, fmt.Errorf("first node has no embedding dimension")
	}

	arrowSchema := arrow.NewSchema(
		[]arrow.Field{
			{Name: "id", Type: arrow.BinaryTypes.String},
			{Name: "text", Type: arrow.BinaryTypes.String},
			{Name: "type", Type: arrow.BinaryTypes.String},
			{Name: "metadata", Type: arrow.BinaryTypes.String},
			{Name: "embedding", Type: arrow.FixedSizeListOf(int32(dim), arrow.PrimitiveTypes.Float32)},
		},
		nil,
	)

	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, arrowSchema)
	defer builder.Release()

	idBuilder := builder.Field(0).(*array.StringBuilder)
	textBuilder := builder.Field(1).(*array.StringBuilder)
	typeBuilder := builder.Field(2).(*array.StringBuilder)
	metadataBuilder := builder.Field(3).(*array.StringBuilder)
	embeddingBuilder := builder.Field(4).(*array.FixedSizeListBuilder)
	embeddingValueBuilder := embeddingBuilder.ValueBuilder().(*array.Float32Builder)

	ids := make([]string, len(nodes))

	for i, node := range nodes {
		ids[i] = node.ID
		idBuilder.Append(node.ID)
		textBuilder.Append(node.Text)
		typeBuilder.Append(string(node.Type))

		// Serialize metadata
		metaBytes, err := json.Marshal(node.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata for node %s: %w", node.ID, err)
		}
		metadataBuilder.Append(string(metaBytes))

		// Append embedding
		if len(node.Embedding) != dim {
			return nil, fmt.Errorf("node %s has embedding dimension %d, expected %d", node.ID, len(node.Embedding), dim)
		}
		embeddingBuilder.Append(true)
		for _, v := range node.Embedding {
			embeddingValueBuilder.Append(v)
		}
	}

	record := builder.NewRecord()
	defer record.Release()

	if s.table == nil {
		var err error
		s.table, err = s.conn.CreateTable(s.tableName)
		if err != nil {
			return nil, fmt.Errorf("failed to create table: %w", err)
		}
		if err := s.table.Add(record, lancedb.AddModeOverwrite); err != nil {
			return nil, fmt.Errorf("failed to add initial records: %w", err)
		}
	} else {
		if err := s.table.Add(record, lancedb.AddModeAppend); err != nil {
			return nil, fmt.Errorf("failed to add records: %w", err)
		}
	}

	return ids, nil
}

// Query finds the top-k most similar nodes to the query embedding.
func (s *LanceDBStore) Query(ctx context.Context, query schema.VectorStoreQuery) ([]schema.NodeWithScore, error) {
	if s.table == nil {
		return nil, fmt.Errorf("table not initialized")
	}

	q := s.table.Query().
		NearestTo(query.Embedding).
		Limit(query.TopK)

	// Apply filters
	if query.Filters != nil && len(query.Filters.Filters) > 0 {
		whereClause, err := s.buildWhereClause(query.Filters)
		if err != nil {
			return nil, fmt.Errorf("failed to build where clause: %w", err)
		}
		if whereClause != "" {
			q = q.Where(whereClause)
		}
	}

	results, err := q.Execute()
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}

	// Parse results
	var nodes []schema.NodeWithScore
	for _, record := range results {
		// We need to extract columns.
		// Schema: id, text, type, metadata, embedding (and potentially _distance)
		// Note: record is an arrow.Record
		defer record.Release()

		idCol := record.Column(0).(*array.String)
		textCol := record.Column(1).(*array.String)
		typeCol := record.Column(2).(*array.String)
		metaCol := record.Column(3).(*array.String)
		// embeddingCol := record.Column(4).(*array.FixedSizeList)
		// Distance is usually not in the schema but returned if we ask for it?
		// go-lancedb might return it as a separate column or we need to check schema.
		// The documentation says "score" might be available.
		// Let's check if there is a "_distance" column.
		// In standard LanceDB, it's often "_distance".

		distIndex := -1
		for i, field := range record.Schema().Fields() {
			if field.Name == "_distance" {
				distIndex = i
				break
			}
		}

		for i := 0; i < int(record.NumRows()); i++ {
			id := idCol.Value(i)
			text := textCol.Value(i)
			nodeType := schema.NodeType(typeCol.Value(i))
			metaStr := metaCol.Value(i)

			var meta map[string]interface{}
			if err := json.Unmarshal([]byte(metaStr), &meta); err != nil {
				// log error or skip? strict for now
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}

			// Reconstruct embedding? The query result might not include it unless we select it.
			// But NearestTo typically returns all columns unless Select is used.
			// Let's skip reconstruction of embedding for now to save bandwidth/time if not strictly needed,
			// but schema.Node has it.
			// If we need it, we can grab it from column 4.

			score := 0.0
			if distIndex != -1 {
				// Distance is float32 usually
				distCol := record.Column(distIndex).(*array.Float32)
				score = float64(distCol.Value(i))
			}

			nodes = append(nodes, schema.NodeWithScore{
				Node: schema.Node{
					ID:       id,
					Text:     text,
					Type:     nodeType,
					Metadata: meta,
					// Embedding: ... // (Skipped for now)
				},
				Score: score,
			})
		}
	}

	return nodes, nil
}

func (s *LanceDBStore) buildWhereClause(filters *schema.MetadataFilters) (string, error) {
	var clauses []string
	for _, f := range filters.Filters {
		// Basic approximation for JSON string matching
		// For strict correctness we'd need JSON operators which LanceDB might support via DataFusion,
		// but for now let's handle ID specifically or fallback to basic string match on the JSON blob.

		// If key is "id", "text", "type", map to top-level columns
		switch f.Key {
		case "id", "text", "type":
			val := fmt.Sprintf("'%v'", f.Value)
			op := string(f.Operator)
			if op == "==" {
				op = "="
			}
			clauses = append(clauses, fmt.Sprintf("%s %s %s", f.Key, op, val))
		default:
			// For metadata fields, we are storing them in a `metadata` JSON string column.
			// We can try `metadata LIKE '%"key":"value"%'`
			// This is brittle but functional for simple equality.
			if f.Operator == schema.FilterOperatorEq {
				// Need to handle quoting carefully.
				// Assuming simple values.
				// Key is quoted in JSON: "key"
				// Value depends on type. If string, "value". If number, value.
				jsonSubStr := ""
				switch v := f.Value.(type) {
				case string:
					jsonSubStr = fmt.Sprintf("\"%s\":\"%s\"", f.Key, v)
				case int, int64, float64:
					jsonSubStr = fmt.Sprintf("\"%s\":%v", f.Key, v)
				default:
					// Fallback to string representation
					jsonSubStr = fmt.Sprintf("\"%s\":\"%v\"", f.Key, v)
				}
				// Escape single quotes in the json string if any (SQL injection prevention not full here)
				jsonSubStr = strings.ReplaceAll(jsonSubStr, "'", "''")
				clauses = append(clauses, fmt.Sprintf("metadata LIKE '%%%s%%'", jsonSubStr))
			} else {
				// Other operators are hard on a JSON string without JSON functions
				// Warn or skip?
				// Let's return error for now to be safe.
				return "", fmt.Errorf("unsupported operator %s for metadata field %s (only top-level fields support full ops)", f.Operator, f.Key)
			}
		}
	}
	return strings.Join(clauses, " AND "), nil
}
